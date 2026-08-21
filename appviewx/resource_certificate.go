package appviewx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
	"terraform-provider-appviewx/appviewx/fileops"
)

func ResourceCertificateServer() *schema.Resource {
	//fmt.Println("****************** Logging for test purpose")
	return &schema.Resource{
		Create: resourceCertificateServerCreate,
		Read:   resourceCertificateServerRead,
		Update: resourceCertificateServerUpdate,
		Delete: resourceCertificateServerDelete,

		Schema: map[string]*schema.Schema{
			constants.COMMON_NAME: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.HASH_FUNCTION: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.KEY_TYPE: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.BIT_LENGTH: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.DNS_NAMES: &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			constants.CUSTOM_FIELDS: &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			constants.VENDOR_SPECIFIC_FIELDS: &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			constants.CERTIFICATE_AUTHORITY: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CERTIFICATE_GROUP_NAME: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CA_SETTING_NAME: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CERTIFICATE_TYPE: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.VALIDITY: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.VALIDITY_UNIT: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.VALIDITY_UNIT_VALUE: &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			constants.IS_SYNC: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			constants.CERTIFICATE_DOWNLOAD_PATH: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CERTIFICATE_DOWNLOAD_FORMAT: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CERTIFICATE_DOWNLOAD_PASSWORD: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.CERTIFICATE_CHAIN_REQUIRED: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			constants.RESOURCE_ID: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			constants.KEY_DOWNLOAD_PATH: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.KEY_DOWNLOAD_PASSWORD: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.DOWNLOAD_PASSWORD_PROTECTED_KEY: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceCertificateImport,
		},
	}
}

func resourceCertificateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	id := d.Id()

	parameters := strings.Split(id, ",")

	log.Println("parameters = ", parameters)

	return []*schema.ResourceData{d}, nil
}

func resourceCertificateServerRead(d *schema.ResourceData, m interface{}) error {
	log.Println("[INFO]  **************** GET OPERATION NOT SUPPORTED FOR THIS RESOURCE **************** ")
	// Since the resource is for stateless operation, only nil returned
	return nil
}

func resourceCertificateServerUpdate(resourceData *schema.ResourceData, m interface{}) error {
	log.Println("[INFO]  **************** UPDATE OPERATION NOT SUPPORTED FOR THIS RESOURCE **************** ")
	//Update implementation is empty since this resource is for the stateless generic api invocation
	return errors.New("Update not supported")
}

func resourceCertificateServerDelete(d *schema.ResourceData, m interface{}) error {
	log.Println("[INFO]  **************** DELETE OPERATION NOT SUPPORTED FOR THIS RESOURCE **************** ")
	// Delete implementation is empty since this resoruce is for the stateless generic api invocation
	d.SetId("")
	return nil
}

// TODO: cleanup to be done
func resourceCertificateServerCreate(resourceData *schema.ResourceData, m interface{}) error {

	log.Println("****************** Resource Certificate Server Create ******************")
	configAppViewXEnvironment := m.(*config.AppViewXEnvironment)

	appviewxUserName := configAppViewXEnvironment.AppViewXUserName
	appviewxPassword := configAppViewXEnvironment.AppViewXPassword
	appviewxClientId := configAppViewXEnvironment.AppViewXClientId
	appviewxClientSecret := configAppViewXEnvironment.AppViewXClientSecret
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS
	appviewxGwSource := "WEB"
	var appviewxSessionID, accessToken string
	var err error
	if appviewxUserName != "" && appviewxPassword != "" {
		appviewxSessionID, err = GetSession(appviewxUserName, appviewxPassword, appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxGwSource, appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Println("[ERROR] Error in getting the session due to : ", err)
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	} else if appviewxClientId != "" && appviewxClientSecret != "" {
		accessToken, err = GetAccessTokenWithRotation(configAppViewXEnvironment, appviewxGwSource)
		if err != nil {
			log.Println("[ERROR] Error in getting the access token due to : ", err)
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	result, err := createCertificate(resourceData, configAppViewXEnvironment, appviewxSessionID, accessToken)
	if err != nil {
		log.Println("[ERROR] Error in creating the certificate due to : ", err)
		return err
	}
	if result.Response["resourceId"] == "" {
		log.Println("[ERROR] Resource ID is not obtained from the certificate creation response to proceed with certificate download")
		return errors.New("[ERROR] Resource ID is not obtained to proceed with certificate download")
	}
	resourceID := result.Response["resourceId"]
	resourceData.Set(constants.RESOURCE_ID, resourceID)
	resourceData.SetId(resourceID)
	log.Println("[INFO] resource_id data is set in payload")

	if resourceData.Get(constants.IS_SYNC) == nil || !resourceData.Get(constants.IS_SYNC).(bool) {
		log.Println("[INFO] Certificate is created in ASYNC mode so download can be done once the certificate is issued.")
		log.Println("[INFO] ***** Use this resource ID to download the certificate", resourceID)
		resourceData.SetId(strconv.Itoa(rand.Int()))
		return nil
	} else {
		log.Println("[INFO] Certificate is created in SYNC mode so proceeding with download.")
		if err := downloadCertificate(resourceData, resourceID, appviewxSessionID, accessToken, configAppViewXEnvironment); err != nil {
			return err
		}
		if resourceData.Get(constants.KEY_DOWNLOAD_PATH).(string) != "" {
			log.Println("[INFO] Key download path is provided in the payload hence proceeding with key download")
			if err := downloadKey(resourceData, resourceID, appviewxSessionID, accessToken, configAppViewXEnvironment); err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadCertificate(resourceData *schema.ResourceData, resourceID string, appviewxSessionID string, accessToken string, configAppViewXEnvironment *config.AppViewXEnvironment) error {
	var isChainRequired, ok bool
	var downloadPassword string
	commonName := resourceData.Get(constants.COMMON_NAME).(string)

	downloadFormat := GetDownloadFormat(resourceData)
	downloadPath, err := GetDownloadFilePath(resourceData, commonName, downloadFormat)
	if err != nil {
		log.Println("[ERROR] Failed to validate certificate download path: " + err.Error())
		return err
	}
	if downloadPassword, ok = GetDownloadPassword(resourceData, downloadFormat, configAppViewXEnvironment); !ok {
		return errors.New("[ERROR] Error in getting the download password")
	}
	isChainRequired = resourceData.Get(constants.CERTIFICATE_CHAIN_REQUIRED).(bool)

	if downloadSuccess := downloadCertificateFromAppviewx(resourceID, "", "", downloadFormat, downloadPassword, downloadPath, isChainRequired, appviewxSessionID, accessToken, configAppViewXEnvironment); downloadSuccess {
		log.Println("[INFO] Certificate downloaded successfully in the specified path")
		resourceData.SetId(strconv.Itoa(rand.Int()))
	} else {
		log.Println("[ERROR] Certificate was not downloaded in the specified path")
		return errors.New("[ERROR] Certificate was not downloaded in the specified path")
	}
	return nil
}

func GetAccessToken(appviewxClientId, appviewxClientSecret, appviewxEnvironmentIP,
	appviewxEnvironmentPort,
	appviewxGwSource string,
	appviewxEnvironmentIsHTTPS bool) (string, error) {
	log.Println("[INFO] Request received for fetching access token")

	headers := make(map[string]interface{})
	headers[constants.CONTENT_TYPE] = constants.APPLICATION_URL_ENCODED
	headers[constants.ACCEPT] = constants.APPLICATION_JSON

	actionID := constants.APPVIEWX_GET_ACCESS_TOKEN_ACTION_ID

	queryParams := make(map[string]string)
	queryParams[constants.GW_SOURCE] = appviewxGwSource

	payload := url.Values{}
	payload.Set(constants.GRANT_TYPE, constants.CLIENT_CREDENTIALS)

	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, actionID, queryParams, appviewxEnvironmentIsHTTPS)

	client := &http.Client{Transport: HTTPTransport()}
	req, err := http.NewRequest(constants.POST, url, strings.NewReader(payload.Encode()))
	req.SetBasicAuth(appviewxClientId, appviewxClientSecret)
	if err != nil {
		log.Println("[ERROR] Error in creating the new reqeust", err)
		return "", err
	}

	for key, value := range headers {
		value1 := fmt.Sprintf("%v", value)
		key1 := fmt.Sprintf("%v", key)
		req.Header.Add(key1, value1)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] Error in executing the request", err)
		return "", err
	}
	defer resp.Body.Close()
	log.Println("[INFO] Response status code : ", resp.Status)

	// Read body once so it can be inspected before parsing.
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[ERROR] error in reading the response body", err)
		return "", err
	}

	// Detect expired client secret regardless of HTTP status code.
	if strings.Contains(string(responseBody), "Client secret expired") {
		log.Println("[ERROR] Client secret has expired")
		return "", errors.New(
			"Client secret has expired. " +
				"The provider will attempt to regenerate it automatically. " +
				"If regeneration fails, you must manually update your client secret:\n" +
				"  - If using .tfvars file: Update appviewx_client_secret in terraform.tfvars\n" +
				"  - If using environment variable: Update APPVIEWX_TERRAFORM_CLIENT_SECRET\n" +
				"  - If using hardcoded secret in provider block: Update appviewx_client_secret in your .tf file",
		)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		log.Println("[ERROR] Client credentials rejected (HTTP "+strconv.Itoa(resp.StatusCode)+"): ", string(responseBody))
		return "", errors.New(
			"Authentication failed (HTTP " + strconv.Itoa(resp.StatusCode) + "): " +
				"Client credentials are invalid or expired. " +
				"The provider will attempt automatic secret regeneration. " +
				"Your configuration source will be updated as follows:\n" +
				"  - .tfvars file: Will be automatically updated with the new secret\n" +
				"  - Environment variable: A .appviewx_secret.env file will be created with the new secret\n" +
				"  - Hardcoded provider block: You must manually update appviewx_client_secret",
		)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("[ERROR] Response obtained : ", string(responseBody))
		return "", errors.New("error in getting the access token (HTTP " + strconv.Itoa(resp.StatusCode) + "): " + string(responseBody))
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Println("[ERROR] Error in Unmarshalling the responseContents", err)
		return "", err
	}

	if response[constants.RESPONSE] != nil {
		log.Println("[INFO] Access token retrieval success, access token will be used for AppViewX API calls")
		return response[constants.RESPONSE].(string), nil
	}
	log.Println("[ERROR] Access token retrieval failed")
	return "", errors.New("access token retrieval failed")
}

// RegenerateClientSecret calls the acctmgmt-regenerate-client-secret endpoint using the same
// headers and query parameters as GetAccessToken and returns the newly issued client secret.
func RegenerateClientSecret(clientId, clientSecret, ip, port, gwSource string, isHTTPS bool) (string, error) {
	log.Println("[INFO] Requesting client secret regeneration from AppViewX")

	queryParams := map[string]string{
		constants.GW_SOURCE: gwSource,
	}

	// API requires JSON body — url-encoded body returns HTTP 415.
	bodyMap := map[string]string{
		"client_id":     clientId,
		"client_secret": clientSecret,
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return "", fmt.Errorf("could not marshal regenerate-client-secret payload: %w", err)
	}

	endpoint := GetURL(ip, port, constants.APPVIEWX_REGENERATE_CLIENT_SECRET_ACTION_ID, queryParams, isHTTPS)

	client := &http.Client{Transport: HTTPTransport()}
	req, err := http.NewRequest(constants.POST, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("could not create regenerate-client-secret request: %w", err)
	}
	req.Header.Set(constants.CONTENT_TYPE, constants.APPLICATION_JSON)
	req.Header.Set(constants.ACCEPT, constants.APPLICATION_JSON)
	req.SetBasicAuth(clientId, clientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("regenerate-client-secret request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read regenerate-client-secret response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("regenerate-client-secret returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("could not parse regenerate-client-secret response: %w", err)
	}

	// Expected response shape: {"response": {"clientSecret": "..."}}
	if responseVal, ok := result[constants.RESPONSE]; ok {
		if responseMap, ok := responseVal.(map[string]interface{}); ok {
			if secret, ok := responseMap[constants.CLIENT_SECRET_RESPONSE_KEY].(string); ok && secret != "" {
				log.Println("[INFO] Client secret regeneration successful")
				return secret, nil
			}
		}
	}

	return "", errors.New("regenerate-client-secret: 'clientSecret' not found in response")
}

// GetAccessTokenWithRotation fetches an access token.  On a 401 or 403 response it
// automatically calls RegenerateClientSecret, persists the new secret to the .tfvars file
// defined by configEnv.TfVarsFilePath (auto-discovered when empty), updates the in-memory
// config, and retries GetAccessToken once with the new secret.
func GetAccessTokenWithRotation(configEnv *config.AppViewXEnvironment, gwSource string) (string, error) {
	token, err := GetAccessToken(
		configEnv.AppViewXClientId, configEnv.AppViewXClientSecret,
		configEnv.AppViewXEnvironmentIP, configEnv.AppViewXEnvironmentPort,
		gwSource, configEnv.AppViewXIsHTTPS,
	)
	if err == nil {
		return token, nil
	}

	// Only attempt rotation on authentication failures or expired secret
	// Use broad "expired" check to catch all variations: "secret expired", "secret has expired", etc.
	isAuthError := strings.Contains(err.Error(), "HTTP 401") ||
		strings.Contains(err.Error(), "HTTP 403")
	isExpiredError := strings.Contains(strings.ToLower(err.Error()), "expired") ||
		strings.Contains(err.Error(), "unauthorized")
	
	// If it's NOT an auth/expiry error, return immediately without attempting regeneration
	if !isAuthError && !isExpiredError {
		return "", err
	}

	log.Println("[INFO] Access token request failed due to expired/invalid client secret — attempting automatic regeneration")
	newSecret, regenErr := RegenerateClientSecret(
		configEnv.AppViewXClientId, configEnv.AppViewXClientSecret,
		configEnv.AppViewXEnvironmentIP, configEnv.AppViewXEnvironmentPort,
		gwSource, configEnv.AppViewXIsHTTPS,
	)
	if regenErr != nil {
		// If regeneration failed with 401, it means the secret has already been invalidated
		// by a previous regeneration attempt. Provide clear guidance for different config sources.
		if strings.Contains(regenErr.Error(), "HTTP 401") {
			log.Println("[ERROR] ============================================================")
			log.Println("[ERROR] CRITICAL: Client secret regeneration failed with HTTP 401")
			log.Println("[ERROR] This typically means your secret was already regenerated")
			log.Println("[ERROR] in a previous terraform run and is now permanently invalid.")
			log.Println("[ERROR] ============================================================")
			log.Println("[ERROR] HOW TO FIX:")
			log.Println("[ERROR] 1. Log in to the AppViewX UI")
			log.Println("[ERROR] 2. Navigate to Service Account tab in Platform")
			log.Println("[ERROR] 3. Copy the newly generated client secret")
			log.Println("[ERROR] 4. UPDATE your terraform configuration:")
			log.Println("[ERROR]    - If using hardcoded secret in provider block: Update appviewx_client_secret directly in your .tf file")
			log.Println("[ERROR]    - If using .tfvars file: Update appviewx_client_secret in terraform.tfvars")
			log.Println("[ERROR]    - If using environment variable: Update APPVIEWX_TERRAFORM_CLIENT_SECRET and source it again")
			log.Println("[ERROR] 5. Re-run: terraform apply")
			log.Println("[ERROR] ============================================================")
			return "", fmt.Errorf(
				"Client secret is permanently invalid and cannot be auto-regenerated.\n" +
					"Reason: Previous regeneration attempt already invalidated this secret at the AppViewX server.\n" +
					"Solution: Manually generate a new secret in the AppViewX UI, then update your terraform configuration with the new secret.",
			)
		}
		return "", fmt.Errorf("access token failed (%v); secret regeneration also failed: %w", err, regenErr)
	}

	// Persist new secret to .tfvars before updating in-memory state so it survives restarts
	if updateErr := fileops.UpdateTfVarsClientSecret(configEnv.TfVarsFilePath, configEnv.AppViewXClientId, newSecret); updateErr != nil {
		log.Printf("[WARN] New client secret obtained but could not be persisted: %v", updateErr)
	}

	configEnv.AppViewXClientSecret = newSecret
	log.Println("[INFO] Retrying access token request with regenerated client secret")
	return GetAccessToken(
		configEnv.AppViewXClientId, newSecret,
		configEnv.AppViewXEnvironmentIP, configEnv.AppViewXEnvironmentPort,
		gwSource, configEnv.AppViewXIsHTTPS,
	)
}

func createCertificate(resourceData *schema.ResourceData, configAppViewXEnvironment *config.AppViewXEnvironment, appviewxSessionID, accessToken string) (config.AppviewxCreateCertResponse, error) {
	var result config.AppviewxCreateCertResponse
	httpMethod := config.HTTPMethodPost
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS
	queryParams := frameQueryParams()
	if resourceData.Get(constants.IS_SYNC) != nil {
		isSync := resourceData.Get(constants.IS_SYNC).(bool)
		queryParams["isSync"] = strconv.FormatBool(isSync)
	}
	headers := frameHeaders()
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, config.CreateCertificateActionId, queryParams, appviewxEnvironmentIsHTTPS)
	payload := frameCertificatePayload(resourceData)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Println("[ERROR] error in Marshalling the payload ", payload, err)
		return result, err
	}
	client := &http.Client{Transport: HTTPTransport()}

	printRequest(httpMethod, url, headers, requestBody)

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("[ERROR] error in creating new Request", err)
		return result, err
	}

	for key, value := range headers {
		value1 := fmt.Sprintf("%v", value)
		key1 := fmt.Sprintf("%v", key)
		req.Header.Add(key1, value1)
	}
	if appviewxSessionID != "" {
		req.Header.Add(constants.SESSION_ID, appviewxSessionID)
	} else {
		req.Header.Add(constants.TOKEN, accessToken)
	}

	httpResponse, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] Error in making certificate create request due to ", err)
		return result, err
	} else {
		log.Println("[INFO] Certificate creation request submitted successfully")
	}
	log.Println("[INFO] Response status code : ", httpResponse.Status)
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		responseBody, err := io.ReadAll(httpResponse.Body)
		if err == nil {
			log.Println("[ERROR] Response obtained : ", string(responseBody))
			return result, errors.New("error in creating the certificate due to " + string(responseBody))
		}
	}
	responseByte, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		log.Println(err)
		return result, err
	} else {
		err = json.Unmarshal(responseByte, &result)
		if err != nil {
			log.Println("[ERROR] Unable to unmarshall the response due to ", err)
			return result, err
		} else {
			log.Println("[INFO] Response obtained successfully for certificate create")
		}
	}
	return result, nil

}

func frameCertificatePayload(resourceData *schema.ResourceData) config.CreateCertificatePayload {
	var payload config.CreateCertificatePayload
	var csrParams config.CSRParameters
	csrParams.CommonName = resourceData.Get(constants.COMMON_NAME).(string)
	csrParams.HashFunction = resourceData.Get(constants.HASH_FUNCTION).(string)
	csrParams.KeyType = resourceData.Get(constants.KEY_TYPE).(string)
	csrParams.BitLength = resourceData.Get(constants.BIT_LENGTH).(string)
	dnsNames, ok := resourceData.GetOk(constants.DNS_NAMES)
	var enhancedSAN config.EnhancedSANTypes
	if ok {
		dns := dnsNames.([]interface{})
		var dnsValues = make([]string, len(dns))
		for key, value := range dns {
			dnsValues[key] = value.(string)
		}
		enhancedSAN.DNSNames = dnsValues
		csrParams.EnhancedSANTypes = enhancedSAN
	}
	csrParams.CertificateCategories = []string{"Server", "Client"}
	payload.CaConnectorInfo.CSRParameters = csrParams
	payload.CaConnectorInfo.CASettingName = resourceData.Get(constants.CA_SETTING_NAME).(string)
	payload.CaConnectorInfo.CertificateAuthority = resourceData.Get(constants.CERTIFICATE_AUTHORITY).(string)
	payload.CaConnectorInfo.CAConnectorName = payload.CaConnectorInfo.CertificateAuthority + " Connector  Terraform"
	payload.CaConnectorInfo.ValidityInDays = resourceData.Get(constants.VALIDITY).(int)
	payload.CaConnectorInfo.ValidityUnit = resourceData.Get(constants.VALIDITY_UNIT).(string)
	payload.CaConnectorInfo.ValidityUnitValue = resourceData.Get(constants.VALIDITY_UNIT_VALUE).(int)
	payload.CaConnectorInfo.CertificateType = resourceData.Get(constants.CERTIFICATE_TYPE).(string)
	payload.CertificateGroup.CertificateGroupName = resourceData.Get(constants.CERTIFICATE_GROUP_NAME).(string)
	customFields, ok := resourceData.GetOk(constants.CUSTOM_FIELDS)
	if ok {
		var customFieldValues = make(map[string]string)
		customFields := customFields.(map[string]interface{})
		for key, values := range customFields {
			customFieldValues[key] = values.(string)
		}
		payload.CaConnectorInfo.CustomAttributes = customFieldValues
	}
	vendorSpecFields, ok := resourceData.GetOk(constants.VENDOR_SPECIFIC_FIELDS)
	if ok {
		var vendorFields = make(map[string]string)
		vendorSpecFieldList := vendorSpecFields.(map[string]interface{})
		for key, values := range vendorSpecFieldList {
			vendorFields[key] = values.(string)
		}
		payload.CaConnectorInfo.VendorSpecificfields = vendorFields
	}
	return payload
}

func frameHeaders() map[string]interface{} {
	var headers = make(map[string]interface{})
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	return headers
}

func frameQueryParams() map[string]string {
	var queryParams = make(map[string]string)
	queryParams[constants.GW_SOURCE] = "WEB"
	return queryParams
}
