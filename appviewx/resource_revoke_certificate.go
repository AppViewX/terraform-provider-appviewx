package appviewx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
)

func ResourceRevokeCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRevokeCertificateCreate,
		ReadContext:   resourceRevokeCertificateRead,
		DeleteContext: resourceRevokeCertificateDelete,
		UpdateContext: resourceRevokeCertificateUpdate,

		Schema: map[string]*schema.Schema{
			"serial_number": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Serial number of the certificate to revoke",
			},
			"issuer_common_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Issuer common name of the certificate to revoke",
			},
			"reason": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Unspecified",
					"Key compromise",
					"CA compromise",
					"Affiliation Changed",
					"Superseded",
					"Cessation of operation",
				}, false),
				Description: "Reason for certificate revocation",
			},
			"comments": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional comments for revocation",
			},
			"status_code": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "HTTP status code of the revocation request",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Resource ID of the revoked certificate",
			},
			"request_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Request ID of the revocation request",
			},
			"response_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Response message from the server",
			},
			"revocation_success": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the revocation was successful",
			},
		},
	}
}

func resourceRevokeCertificateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n====================[CERTIFICATE REVOCATION]====================")
	log.Println("  🚀  Resource Revoke Certificate Create")
	log.Println("======================================================================\n")

	configAppViewXEnvironment := m.(*config.AppViewXEnvironment)

	// Extract configuration parameters
	appviewxUserName := configAppViewXEnvironment.AppViewXUserName
	appviewxPassword := configAppViewXEnvironment.AppViewXPassword
	appviewxClientId := configAppViewXEnvironment.AppViewXClientId
	appviewxClientSecret := configAppViewXEnvironment.AppViewXClientSecret
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS
	appviewxGwSource := "external"

	// Authenticate using either username/password or client ID/secret
	var appviewxSessionID, accessToken string
	var err error

	if appviewxUserName != "" && appviewxPassword != "" {
		appviewxSessionID, err = GetSession(appviewxUserName, appviewxPassword, appviewxEnvironmentIP, appviewxEnvironmentPort, "WEB", appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error in getting the session:")
			log.Println("   ", err)
			log.Println("----------------------------------------------------------------------\n")
			return diag.FromErr(err)
		}
	} else if appviewxClientId != "" && appviewxClientSecret != "" {
		accessToken, err = GetAccessToken(appviewxClientId, appviewxClientSecret, appviewxEnvironmentIP, appviewxEnvironmentPort, "WEB", appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error in getting the access token:")
			log.Println("   ", err)
			log.Println("----------------------------------------------------------------------\n")
			return diag.FromErr(err)
		}
	}

	// If both authentication methods failed, return error
	if appviewxSessionID == "" && accessToken == "" {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Authentication failed - provide either username/password or client ID/secret in Terraform File or in the Environment Variables:[APPVIEWX_TERRAFORM_CLIENT_ID, APPVIEWX_TERRAFORM_CLIENT_SECRET]\n")
		return diag.FromErr(errors.New("authentication failed - provide either username/password or client ID/secret in Terraform File or in the Environment Variables:[APPVIEWX_TERRAFORM_CLIENT_ID, APPVIEWX_TERRAFORM_CLIENT_SECRET]"))
	}

	// Get serial number and issuer common name from config
	serialNumber := d.Get("serial_number").(string)
	issuerCommonName := d.Get("issuer_common_name").(string)

	log.Printf("\n[CERTIFICATE REVOCATION][INFO] 🔍 Looking up certificate with serial: %s and issuer: %s\n", serialNumber, issuerCommonName)

	// Step 1: Call the execute-hook API to get resource ID
	resourceId, err := getResourceIdBySerialAndIssuer(appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxEnvironmentIsHTTPS, appviewxSessionID, accessToken, serialNumber, issuerCommonName)
	if err != nil {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error retrieving resource ID:")
		log.Println("   ", err)
		log.Println("----------------------------------------------------------------------\n")
		return diag.FromErr(err)
	}

	// Save the resource ID in the state
	d.Set("resource_id", resourceId)
	log.Printf("\n[CERTIFICATE REVOCATION][INFO] 🔄 Found certificate with resource ID: %s\n", resourceId)

	// Step 2: Revoke certificate using the resource ID
	// Prepare revocation request
	reason := d.Get("reason").(string)
	log.Printf("[CERTIFICATE REVOCATION][INFO] 📝 Revocation reason: %s\n", reason)

	// Build revocation payload
	payload := map[string]interface{}{
		"resourceId": resourceId,
		"reason":     reason,
	}

	// Add comments if provided
	if comments, ok := d.GetOk("comments"); ok {
		payload["comments"] = comments.(string)
		log.Printf("[CERTIFICATE REVOCATION][INFO] 💬 Revocation comments: %s\n", comments.(string))
	}

	// Set query parameters
	queryParams := map[string]string{
		"gwsource": appviewxGwSource,
	}

	// Get URL for the revoke endpoint
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, "certificate/revoke", queryParams, appviewxEnvironmentIsHTTPS)
	log.Printf("\n[CERTIFICATE REVOCATION][DEBUG] 🌐 Revoking certificate using URL: %s\n", url)

	// Prepare the request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error in marshalling the payload:")
		log.Println("   ", err)
		log.Printf("   Payload: %+v\n", payload)
		log.Println("----------------------------------------------------------------------\n")
		return diag.FromErr(err)
	}

	// Log the request for debugging
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("\n[CERTIFICATE REVOCATION][DEBUG] 📝 Revocation payload:\n%s\n", string(payloadBytes))

	// Create HTTP client
	client := &http.Client{Transport: HTTPTransport()}

	// Create request
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error in creating new request:")
		log.Println("   ", err)
		log.Println("----------------------------------------------------------------------\n")
		return diag.FromErr(err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication header
	if appviewxSessionID != "" {
		log.Printf("[CERTIFICATE REVOCATION][DEBUG] 🔑 Using session ID for authentication")
		req.Header.Set(constants.SESSION_ID, appviewxSessionID)
	} else if accessToken != "" {
		log.Printf("[CERTIFICATE REVOCATION][DEBUG] 🔑 Using access token for authentication\n")
		req.Header.Set(constants.TOKEN, accessToken)
	}

	// Log headers for debugging
	// headersBytes, _ := json.MarshalIndent(req.Header, "", "  ")
	// log.Printf("[CERTIFICATE REVOCATION][DEBUG] 🏷️ Request headers:\n%s\n", string(headersBytes))

	// Make the request
	log.Printf("[CERTIFICATE REVOCATION][INFO] 📤 Sending revocation request...\n")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Error in revoking certificate:")
		log.Println("   ", err)
		log.Println("----------------------------------------------------------------------\n")
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	log.Printf("[CERTIFICATE REVOCATION][INFO] 📊 Certificate revocation response status code: %s\n", resp.Status)

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Unable to read response body:")
		log.Println("   ", err)
		log.Println("----------------------------------------------------------------------\n")
		return diag.FromErr(err)
	}

	// Format and log JSON response for better readability
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err != nil {
		log.Printf("\n[CERTIFICATE REVOCATION][DEBUG] 📦 Revocation response body (raw):\n%s\n", string(responseBody))
	} else {
		log.Printf("\n[CERTIFICATE REVOCATION][DEBUG] 📦 Revocation response body (formatted JSON):\n%s\n", prettyJSON.String())
	}

	// Store response status
	d.Set("status_code", resp.StatusCode)

	// Parse response
	var responseObj map[string]interface{}
	var requestId string
	if err := json.Unmarshal(responseBody, &responseObj); err == nil {
		if response, ok := responseObj["response"].(map[string]interface{}); ok {
			if message, ok := response["message"].(string); ok {
				d.Set("response_message", message)
				log.Printf("[CERTIFICATE REVOCATION][INFO] 💬 Response message: %s\n", message)
			}
			if reqId, ok := response["requestId"].(string); ok && reqId != "" {
				requestId = reqId
				d.Set("request_id", requestId)
				log.Printf("[CERTIFICATE REVOCATION][INFO] 🔑 Found request ID: %s\n", requestId)
			}
		}
	}

	// Determine if revocation was successful based on status code and response
	revocationSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
	d.Set("revocation_success", revocationSuccess)

	// Check for error responses
	if !revocationSuccess {
		log.Println("\n[CERTIFICATE REVOCATION][ERROR] ❌ Revocation failed:")
		log.Printf("   Status: %s\n", resp.Status)
		log.Println("   Response:", string(responseBody))
		log.Println("----------------------------------------------------------------------\n")
		// We don't return an error here because we want to keep the resource info in state
		// even if revocation failed - this allows users to see what went wrong
	} else {
		log.Printf("\n[CERTIFICATE REVOCATION][INFO] ✅ Certificate with resource ID %s successfully revoked\n", resourceId)
	}

	// Set ID to resourceId to track this resource
	if requestId != "" {
		d.SetId(requestId)
		log.Printf("[CERTIFICATE REVOCATION][INFO] 📝 Setting resource ID to request ID: %s\n", requestId)
	} else {
		d.SetId(resourceId)
		log.Printf("[CERTIFICATE REVOCATION][INFO] 📝 Request ID not found, setting resource ID to original resource ID: %s\n", resourceId)
	}

	log.Println("\n[CERTIFICATE REVOCATION][INFO] ✅ Revocation process complete")
	log.Println("======================================================================\n")

	return nil
}

// getResourceIdBySerialAndIssuer calls the execute-hook API to get the resource ID
func getResourceIdBySerialAndIssuer(appviewxEnvironmentIP, appviewxEnvironmentPort string, appviewxEnvironmentIsHTTPS bool, appviewxSessionID, accessToken, serialNumber, issuerCommonName string) (string, error) {
	// Create payload for execute-hook API
	payload := map[string]interface{}{
		"payload": map[string]interface{}{
			"hook": map[string]interface{}{
				"name": "Desjardins - Get Resource Id",
			},
			"input": map[string]interface{}{
				"serial_number":      serialNumber,
				"issuer_common_name": issuerCommonName,
			},
		},
	}

	// Set query parameters
	queryParams := map[string]string{
		"gwsource": "external",
	}

	// Get URL for execute-hook API
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, "execute-hook", queryParams, appviewxEnvironmentIsHTTPS)
	log.Printf("\n[CERTIFICATE LOOKUP][DEBUG] 🌐 Looking up resource ID using URL: %s\n", url)

	// Prepare the request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %v", err)
	}

	// Log the request for debugging
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("\n[CERTIFICATE LOOKUP][DEBUG] 📝 Resource ID lookup payload:\n%s\n", string(payloadBytes))

	// Create HTTP client
	client := &http.Client{Transport: HTTPTransport()}

	// Create request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication header
	if appviewxSessionID != "" {
		log.Printf("[CERTIFICATE LOOKUP][DEBUG] 🔑 Using session ID for authentication")
		req.Header.Set(constants.SESSION_ID, appviewxSessionID)
	} else if accessToken != "" {
		log.Printf("[CERTIFICATE LOOKUP][DEBUG] 🔑 Using access token for authentication\n")
		req.Header.Set(constants.TOKEN, accessToken)
	}

	// Make the request
	log.Printf("[CERTIFICATE LOOKUP][INFO] 📤 Sending resource ID lookup request...\n")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Format and log JSON response for debugging
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err == nil {
		log.Printf("\n[CERTIFICATE LOOKUP][DEBUG] 📦 Resource ID lookup response:\n%s\n", prettyJSON.String())
	} else {
		log.Printf("\n[CERTIFICATE LOOKUP][DEBUG] 📦 Resource ID lookup response (raw):\n%s\n", string(responseBody))
	}

	// Parse response to extract resource ID
	var responseObj map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseObj); err != nil {
		return "", fmt.Errorf("error parsing response JSON: %v", err)
	}

	// FIXED: Updated path to extract resource ID based on the actual response structure
	// The resource ID is in response.output[0]._id rather than response.data.resourceId
	if resp, ok := responseObj["response"].(map[string]interface{}); ok {
		if output, ok := resp["output"].([]interface{}); ok && len(output) > 0 {
			if firstOutput, ok := output[0].(map[string]interface{}); ok {
				if resourceId, ok := firstOutput["_id"].(string); ok && resourceId != "" {
					log.Printf("[CERTIFICATE LOOKUP][INFO] ✅ Found resource ID: %s\n", resourceId)
					return resourceId, nil
				}
			}
		}

		// Additional check for success without resource ID
		if status, ok := resp["status"].(string); ok && status == "Success" {
			if output, ok := resp["output"].([]interface{}); ok && len(output) == 0 {
				return "", fmt.Errorf("certificate not found: successful response but no certificate matched the criteria")
			}
		}
	}

	// Check for error in response
	if resp, ok := responseObj["response"].(map[string]interface{}); ok {
		if errMsg, ok := resp["message"].(string); ok && errMsg != "" {
			return "", fmt.Errorf("API returned error: %s", errMsg)
		}

		// Additional status check
		if status, ok := resp["status"].(string); ok && status != "Success" {
			return "", fmt.Errorf("API returned non-success status: %s", status)
		}
	}

	// Dump the full response for debugging
	fullResponseBytes, _ := json.MarshalIndent(responseObj, "", "  ")
	log.Printf("\n[CERTIFICATE LOOKUP][ERROR] ❌ Could not find resource ID in response structure:\n%s\n", string(fullResponseBytes))

	return "", fmt.Errorf("resource ID not found in response or certificate not found")
}

func resourceRevokeCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE REVOCATION][INFO] ℹ️  GET OPERATION FOR REVOKED CERTIFICATE")
	log.Println("   Since revocation is a one-time operation, returning existing state")
	log.Println("======================================================================\n")

	// Preserve all state values
	for _, key := range []string{"serial_number", "issuer_common_name", "reason", "comments",
		"status_code", "resource_id", "response_message", "revocation_success"} {
		if val, ok := d.GetOk(key); ok {
			d.Set(key, val)
		}
	}

	return nil
}

func resourceRevokeCertificateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE REVOCATION][INFO] 🗑️  DELETE OPERATION FOR CERTIFICATE REVOCATION")
	log.Println("   Revocation is a one-way operation, removing resource from state only")
	log.Println("======================================================================\n")
	// Revocation is a one-way operation, so deletion from terraform doesn't actually delete anything on AppViewX
	// We just remove the resource from state
	d.SetId("")
	return nil
}

func resourceRevokeCertificateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE REVOCATION][INFO] 🗑️  Update OPERATION FOR CERTIFICATE REVOCATION")
	return nil
}
