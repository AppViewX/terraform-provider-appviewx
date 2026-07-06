package appviewx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
	"terraform-provider-appviewx/appviewx/logger"
)

// constructGCPCertificateID builds the GCP Certificate Manager resource id in the
// same format as google_certificate_manager_certificate.id:
// projects/{project}/locations/{location}/certificates/{name}
func constructGCPCertificateID(project, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/certificates/%s", project, location, name)
}

// buildGCPPushPayload builds the /avxapi/certificate/pushToDevice request body for
// the cert-manager-only workflow (no LB binding).
func buildGCPPushPayload(certificateID, certificateName, location, connectorName string, isNewCertificate, pushAutomatically bool, selectedProfiles []string) map[string]interface{} {
	return map[string]interface{}{
		"generalInformation": map[string]interface{}{
			"category":               "cloud",
			"vendor":                 "GCP",
			"profileFilterSelection": ":Certificate Manager",
			"name":                   connectorName,
			"profileType":            "Push and Bind Profiles",
		},
		"certificateDetails": map[string]interface{}{
			"isNewCertificate":       isNewCertificate,
			"certificateName":        certificateName,
			"region":                 location,
			"profileFilterSelection": "Certificate Manager",
		},
		"pushDetails": map[string]interface{}{
			"scriptLocation":    "appviewx",
			"pushAutomatically": pushAutomatically,
		},
		"certificateId":    certificateID,
		"selectedProfiles": selectedProfiles,
	}
}

func ResourcePushGCPCertificateManager() *schema.Resource {
	return &schema.Resource{
		Create: resourcePushGCPCertificateManagerCreate,
		Read:   resourcePushGCPCertificateManagerRead,
		Update: resourcePushGCPCertificateManagerUpdate,
		Delete: resourcePushGCPCertificateManagerDelete,

		Schema: map[string]*schema.Schema{
			constants.CERTIFICATE_ID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "AppViewX certificate resource id (resourceId from appviewx_create_certificate)",
			},
			constants.CERTIFICATE_NAME: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the certificate in GCP Certificate Manager",
			},
			constants.GCP_PROJECT: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "GCP project id (used to construct the certificate id)",
			},
			constants.GCP_LOCATION: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "GCP location: a region (e.g. us-central1) or 'global'",
			},
			constants.GCP_CONNECTOR_NAME: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "AppViewX GCP connector name (generalInformation.name)",
			},
			constants.SELECTED_PROFILES: &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "AppViewX selected profiles, e.g. GCP.Single.All:<project>:Certificate Manager",
			},
			constants.IS_NEW_CERTIFICATE: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			constants.PUSH_AUTOMATICALLY: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			constants.WAIT_FOR_COMPLETION: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			constants.WAIT_TIMEOUT_SECONDS: &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      600,
				ValidateFunc: validation.IntAtLeast(1),
			},
			constants.POLL_INTERVAL_SECONDS: &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10,
				ValidateFunc: validation.IntAtLeast(1),
			},
			constants.CERT_MANAGER_CERTIFICATE_ID: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GCP Certificate Manager certificate id: projects/{project}/locations/{location}/certificates/{name}",
			},
			constants.REQUEST_ID: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			constants.CONNECTOR_ID: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			constants.STATUS_CODE: &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			constants.SUCCESS: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourcePushGCPCertificateManagerRead(d *schema.ResourceData, m interface{}) error {
	logger.Info("**************** READ (no-op, preserving state) - GCP CERT MANAGER PUSH ****************")
	return nil
}

func resourcePushGCPCertificateManagerUpdate(d *schema.ResourceData, m interface{}) error {
	logger.Info("**************** UPDATE (no-op) - GCP CERT MANAGER PUSH ****************")
	return nil
}

func resourcePushGCPCertificateManagerDelete(d *schema.ResourceData, m interface{}) error {
	logger.Info("**************** DELETE (state only; GCP cert not deleted) - GCP CERT MANAGER PUSH ****************")
	d.SetId("")
	return nil
}

func resourcePushGCPCertificateManagerCreate(d *schema.ResourceData, m interface{}) error {
	logger.Info("**************** CREATE - GCP CERT MANAGER PUSH ****************")
	configAppViewXEnvironment := m.(*config.AppViewXEnvironment)

	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS
	appviewxGwSource := "WEB"

	// Authenticate (session id or access token).
	var appviewxSessionID, accessToken string
	var err error
	if configAppViewXEnvironment.AppViewXUserName != "" && configAppViewXEnvironment.AppViewXPassword != "" {
		appviewxSessionID, err = GetSession(configAppViewXEnvironment.AppViewXUserName, configAppViewXEnvironment.AppViewXPassword, appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxGwSource, appviewxEnvironmentIsHTTPS)
		if err != nil {
			logger.Error("Error getting session: ", err)
		}
	}
	if appviewxSessionID == "" && configAppViewXEnvironment.AppViewXClientId != "" && configAppViewXEnvironment.AppViewXClientSecret != "" {
		accessToken, err = GetAccessToken(configAppViewXEnvironment.AppViewXClientId, configAppViewXEnvironment.AppViewXClientSecret, appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxGwSource, appviewxEnvironmentIsHTTPS)
		if err != nil {
			logger.Error("Error getting access token: ", err)
			return err
		}
	}
	if appviewxSessionID == "" && accessToken == "" {
		return errors.New("authentication failed - provide either username/password or client ID/secret")
	}

	// Gather inputs.
	certificateID := d.Get(constants.CERTIFICATE_ID).(string)
	certificateName := d.Get(constants.CERTIFICATE_NAME).(string)
	project := d.Get(constants.GCP_PROJECT).(string)
	location := d.Get(constants.GCP_LOCATION).(string)
	connectorName := d.Get(constants.GCP_CONNECTOR_NAME).(string)
	isNewCertificate := d.Get(constants.IS_NEW_CERTIFICATE).(bool)
	pushAutomatically := d.Get(constants.PUSH_AUTOMATICALLY).(bool)

	rawProfiles := d.Get(constants.SELECTED_PROFILES).([]interface{})
	selectedProfiles := make([]string, len(rawProfiles))
	for i, v := range rawProfiles {
		selectedProfiles[i] = v.(string)
	}

	// Build and send the pushToDevice request.
	payload := buildGCPPushPayload(certificateID, certificateName, location, connectorName, isNewCertificate, pushAutomatically, selectedProfiles)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	logger.Debug("\nGCP push payload:\n%s\n", string(payloadBytes))

	queryParams := map[string]string{constants.GW_SOURCE: appviewxGwSource}
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, "certificate/pushToDevice", queryParams, appviewxEnvironmentIsHTTPS)

	client := &http.Client{Transport: HTTPTransport()}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if appviewxSessionID != "" {
		req.Header.Set(constants.SESSION_ID, appviewxSessionID)
	} else {
		req.Header.Set(constants.TOKEN, accessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	d.Set(constants.STATUS_CODE, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logger.Info("\nGCP push response:\n%s\n", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gcp certificate push failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response[0].requestId + connectorId.
	var responseObj map[string]interface{}
	if err := json.Unmarshal(body, &responseObj); err != nil {
		return fmt.Errorf("unable to parse push response: %v", err)
	}
	var requestID, connectorID string
	if arr, ok := responseObj["response"].([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			if v, ok := first["requestId"].(string); ok {
				requestID = v
			}
			if v, ok := first["connectorId"].(string); ok {
				connectorID = v
			}
		}
	}
	if requestID == "" {
		return fmt.Errorf("gcp certificate push did not return a requestId: %s", string(body))
	}
	d.Set(constants.REQUEST_ID, requestID)
	d.Set(constants.CONNECTOR_ID, connectorID)

	// Optionally wait for completion by polling the request status.
	if d.Get(constants.WAIT_FOR_COMPLETION).(bool) {
		waitTimeout := d.Get(constants.WAIT_TIMEOUT_SECONDS).(int)
		pollInterval := d.Get(constants.POLL_INTERVAL_SECONDS).(int)
		if err := waitForPushCompletion(configAppViewXEnvironment, appviewxSessionID, accessToken, appviewxGwSource, requestID, waitTimeout, pollInterval); err != nil {
			return err
		}
	}

	// Construct and expose the GCP certificate id.
	gcpCertID := constructGCPCertificateID(project, location, certificateName)
	d.Set(constants.CERT_MANAGER_CERTIFICATE_ID, gcpCertID)
	d.Set(constants.SUCCESS, true)
	d.SetId(gcpCertID)

	return nil
}

func waitForPushCompletion(cfg *config.AppViewXEnvironment, sessionID, accessToken, gwSource, requestID string, waitTimeoutSeconds, pollIntervalSeconds int) error {
	logger.Info("Waiting for GCP push completion (requestId=%s, timeout=%ds)", requestID, waitTimeoutSeconds)
	deadline := time.Now().Add(time.Duration(waitTimeoutSeconds) * time.Second)

	for {
		statusCode, body, err := pollWorkflowStatus(cfg.AppViewXEnvironmentIP, cfg.AppViewXEnvironmentPort, cfg.AppViewXIsHTTPS, sessionID, accessToken, requestID, gwSource)
		if err != nil {
			return fmt.Errorf("error polling push status: %v", err)
		}
		if statusCode < 200 || statusCode >= 300 {
			return fmt.Errorf("error polling push status, http %d: %s", statusCode, string(body))
		}

		var responseObj map[string]interface{}
		if err := json.Unmarshal(body, &responseObj); err != nil {
			return fmt.Errorf("error parsing push status response: %v", err)
		}
		code, completed := getWorkflowStatusCode(responseObj)
		if completed {
			if code == STATUS_SUCCESS {
				logger.Info("GCP push completed successfully (requestId=%s)", requestID)
				return nil
			}
			return fmt.Errorf("gcp push failed with status code %d (requestId=%s): %s", code, requestID, string(body))
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %ds waiting for gcp push to complete (requestId=%s)", waitTimeoutSeconds, requestID)
		}
		time.Sleep(time.Duration(pollIntervalSeconds) * time.Second)
	}
}
