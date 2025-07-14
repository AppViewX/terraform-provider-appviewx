package appviewx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
)

// Status code constants
const (
	STATUS_IN_PROGRESS = 0
	STATUS_SUCCESS     = 1
)

// Failed status codes
var failedStatusCodes = []int{2, 3, 8, 9, 10, 11}

func CreatePushCertificateRequestStatus() *schema.Resource {
	return &schema.Resource{
		Create: createPushCertificateRequestStatusCreate,
		Read:   createPushCertificateRequestStatusRead,
		Delete: createPushCertificateRequestStatusDelete,
		Update: createPushCertificateRequestStatusUpdate,

		Schema: map[string]*schema.Schema{
			"request_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Request ID from a workflow execution",
			},
			"retry_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      10,
				Description:  "Number of times to retry checking workflow status (default: 10)",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"retry_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				Description:  "Seconds to wait between retry attempts (default: 20)",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"status_code": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "HTTP status code from the response",
			},
			"workflow_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the workflow",
			},
			"workflow_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the workflow (In Progress, Success, Failed)",
			},
			"workflow_status_code": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Status code of the workflow (0=InProgress, 1=Success, others=Failed)",
			},
			"log_data": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON string containing all tasks and logs",
			},
			"task_summary": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Summary of all task statuses",
			},
			"failed_task_logs": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Detailed logs of any failed tasks",
			},
			"failure_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Extracted failure reason from failed task logs",
			},
			"response_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Summary response message from the workflow",
			},
			"success": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the workflow completed successfully",
			},
			"completed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the workflow has completed (success or failure)",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User who created the workflow request",
			},
			"created_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the workflow request was created",
			},
			"completion_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the workflow completed or polling ended",
			},
			"last_polled_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last time the status was polled",
			},
			"certificate_resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource ID of the created certificate (extracted from 'Trigger Certificate Creation' task)",
			},
			"is_download_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to download the certificate after workflow completion",
			},
			"certificate_download_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Password for the downloaded certificate (if applicable)",
				Sensitive: true,
			},
			"certificate_download_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to download the certificate to",
			},
			"certificate_download_format": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "CRT",
				// ForceNew:    true,
				Description: "Format for the downloaded certificate (e.g., CRT, PFX)",
			},
			"certificate_chain_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				// ForceNew:    true,
				Description: "Whether to include the certificate chain in the download",
			},
			"downloaded_certificate_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Full path to the downloaded certificate file",
			},
			"certificate_common_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Common name of the certificate",
			},
			"certificate_serial_number": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Serial number of the certificate",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceCertificateImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},
	}
}

func createPushCertificateRequestStatusRead(d *schema.ResourceData, m interface{}) error {
	// log.Println("[INFO] **************** READ OPERATION - WORKFLOW LOGS ****************")

	// // Preserve all fields to avoid drift warnings
	// schemaKeys := []string{
	// 	"request_id", "retry_count", "retry_interval", "success", "workflow_status",
	// 	"workflow_status_code", "task_summary", "failed_task_logs", "certificate_resource_id",
	// 	"is_download_required", "certificate_download_path", "certificate_download_format",
	// 	"certificate_chain_required", "downloaded_certificate_path",
	// 	"certificate_common_name", "certificate_serial_number",
	// }

	// for _, key := range schemaKeys {
	// 	if v, ok := d.GetOk(key); ok {
	// 		d.Set(key, v)
	// 	}
	// }

	return nil
}

func createPushCertificateRequestStatusDelete(d *schema.ResourceData, m interface{}) error {
	log.Println("[INFO] **************** DELETE OPERATION FOR WORKFLOW LOGS **************** ")
	// Since this is a read-only resource, deletion just removes it from state
	return nil
}

func createPushCertificateRequestStatusUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[INFO] **************** UPDATE OPERATION FOR WORKFLOW LOGS **************** ")
	// Since this is a read-only resource, update just removes it from state
	return nil
}

func createPushCertificateRequestStatusCreate(d *schema.ResourceData, m interface{}) error {
	log.Println("[INFO] **************** CREATE OPERATION FOR WORKFLOW LOGS **************** ")
	configAppViewXEnvironment := m.(*config.AppViewXEnvironment)

	// d.Partial(true)

	// Get request ID and retry parameters
	requestID := d.Get("request_id").(string)

	if requestID == "" {
		log.Println("[INFO] No request ID provided, skipping workflow status polling")

		// Set a placeholder ID
		d.SetId(fmt.Sprintf("revoke-workflow-log-skipped-%s", strconv.Itoa(rand.Int())))

		// Set default values for computed fields
		d.Set("workflow_status", "Skipped")
		d.Set("workflow_status_code", -1) // Special code for skipped
		d.Set("completed", true)
		d.Set("success", false)
		d.Set("response_message", "Workflow polling was skipped because no request ID was provided")
		d.Set("last_polled_time", time.Now().Format(time.RFC3339))
		d.Set("completion_time", time.Now().Format(time.RFC3339))

		return nil
	}

	retryCount := d.Get("retry_count").(int)
	retryInterval := d.Get("retry_interval").(int)

	log.Printf("[INFO] Starting polling for workflow request ID: %s (max %d retries, %d second intervals)",
		requestID, retryCount, retryInterval)

	// Set resource ID early to ensure it's set even if polling fails
	d.SetId(fmt.Sprintf("workflow-log-%s-%s", requestID, strconv.Itoa(rand.Int())))

	// Authentication credentials
	appviewxUserName := configAppViewXEnvironment.AppViewXUserName
	appviewxPassword := configAppViewXEnvironment.AppViewXPassword
	appviewxClientId := configAppViewXEnvironment.AppViewXClientId
	appviewxClientSecret := configAppViewXEnvironment.AppViewXClientSecret
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS
	appviewxGwSource := "external"

	// Keep track of completion
	var completed bool = false
	var finalStatusCode int = STATUS_IN_PROGRESS
	var lastResponse map[string]interface{}

	// Start polling
	for attempt := 1; attempt <= retryCount; attempt++ {
		log.Printf("[INFO] Polling attempt %d/%d for workflow request ID: %s", attempt, retryCount, requestID)

		// Get authentication token for this request
		appviewxSessionID, accessToken, err := authenticate(
			appviewxUserName, appviewxPassword,
			appviewxClientId, appviewxClientSecret,
			appviewxEnvironmentIP, appviewxEnvironmentPort,
			appviewxEnvironmentIsHTTPS)

		if err != nil {
			log.Printf("[ERROR] Authentication failed on polling attempt %d: %v", attempt, err)
			// If we're on the last attempt, return the error
			if attempt == retryCount {
				return err
			}
			// Otherwise, try again after delay
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		// Poll the workflow status
		statusCode, respBody, err := pollWorkflowStatus(
			appviewxEnvironmentIP, appviewxEnvironmentPort,
			appviewxEnvironmentIsHTTPS, appviewxSessionID,
			accessToken, requestID, appviewxGwSource)

		if err != nil {
			log.Printf("[ERROR] Failed to poll workflow status on attempt %d: %v", attempt, err)
			// If we're on the last attempt, return the error
			if attempt == retryCount {
				return err
			}
			// Otherwise, try again after delay
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		// Parse the response
		var responseObj map[string]interface{}
		if err := json.Unmarshal(respBody, &responseObj); err != nil {
			log.Printf("[ERROR] Failed to parse response JSON on attempt %d: %v", attempt, err)
			if attempt == retryCount {
				return err
			}
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		// Store the last response
		lastResponse = responseObj

		// Record last polled time
		d.Set("last_polled_time", time.Now().Format(time.RFC3339))

		// Check if the workflow has completed
		statusCode, completed = getWorkflowStatusCode(responseObj)
		finalStatusCode = statusCode

		// If workflow has completed (success or failure), break out of the loop
		if completed {
			log.Printf("[INFO] Workflow completed with status code %d after %d polling attempts",
				statusCode, attempt)
			break
		}

		// If we're not done yet and not on the last attempt, wait before trying again
		if attempt < retryCount {
			log.Printf("[INFO] Workflow Request ID: %s is in progress (status code: %d). Waiting %d seconds before next poll...",
				requestID, statusCode, retryInterval)
			time.Sleep(time.Duration(retryInterval) * time.Second)
		}
	}

	// Record completion time
	d.Set("completion_time", time.Now().Format(time.RFC3339))

	// If we've exhausted retries and workflow is still not complete
	if !completed {
		log.Printf("[WARN] Maximum retry count (%d) reached, but workflow is still in progress", retryCount)
	}

	// Process and store the final response data
	if lastResponse != nil {
		processWorkflowResponse(d, m, lastResponse, finalStatusCode, completed)
	} else {
		return fmt.Errorf("no valid response received after %d attempts", retryCount)
	}

	return createPushCertificateRequestStatusRead(d, m)
}

func authenticate(username, password, clientId, clientSecret, envIP, envPort string, isHTTPS bool) (string, string, error) {
	var sessionID, accessToken string
	var err error

	// Try username/password authentication
	if username != "" && password != "" {
		sessionID, err = GetSession(username, password, envIP, envPort, "WEB", isHTTPS)
		if err != nil {
			log.Println("[DEBUG] Session authentication failed, trying client credentials")
		} else {
			return sessionID, "", nil
		}
	}

	// If username/password failed or wasn't provided, try client ID/secret
	if sessionID == "" && clientId != "" && clientSecret != "" {
		accessToken, err = GetAccessToken(clientId, clientSecret, envIP, envPort, "WEB", isHTTPS)
		if err != nil {
			log.Println("[ERROR] Client credentials authentication failed")
			return "", "", err
		}
		return "", accessToken, nil
	}

	// If both authentication methods failed
	if sessionID == "" && accessToken == "" {
		return "", "", errors.New("authentication failed - provide either username/password or client ID/secret in Terraform File or in the Environment Variables:[APPVIEWX_TERRAFORM_CLIENT_ID, APPVIEWX_TERRAFORM_CLIENT_SECRET]")
	}

	return sessionID, accessToken, nil
}

func pollWorkflowStatus(envIP, envPort string, isHTTPS bool, sessionID, accessToken, requestID, gwSource string) (int, []byte, error) {
	// Set query parameters
	queryParams := map[string]string{
		"gwsource": gwSource,
		"ids":      requestID,
	}

	// Get URL for visualworkflow-request-logs
	url := GetURL(envIP, envPort, "visualworkflow-request-logs", queryParams, isHTTPS)
	log.Printf("[DEBUG] 🌐 Fetching workflow request details using URL: %s", url)

	// Create HTTP client
	client := &http.Client{Transport: HTTPTransport()}

	// Create request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication header
	if sessionID != "" {
		req.Header.Set(constants.SESSION_ID, sessionID)
	} else if accessToken != "" {
		req.Header.Set(constants.TOKEN, accessToken)
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("error reading response body: %v", err)
	}

	return resp.StatusCode, body, nil
}

func getWorkflowStatusCode(responseObj map[string]interface{}) (int, bool) {
	// Extract workflow status code from response
	if resp, ok := responseObj["response"].(map[string]interface{}); ok {
		if requestList, ok := resp["requestList"].([]interface{}); ok && len(requestList) > 0 {
			if firstRequest, ok := requestList[0].(map[string]interface{}); ok {
				if statusCode, ok := firstRequest["statusCode"].(float64); ok {
					intStatusCode := int(statusCode)

					// Check if completed based on status code
					if intStatusCode == STATUS_SUCCESS {
						return intStatusCode, true // Success, completed
					}

					// Check if failed (any of the failure codes)
					for _, failCode := range failedStatusCodes {
						if intStatusCode == failCode {
							return intStatusCode, true // Failed, but completed
						}
					}

					// If we get here, it's still in progress
					return intStatusCode, false
				}
			}
		}
	}

	// Default to in-progress if we can't determine status
	return STATUS_IN_PROGRESS, false
}

func processWorkflowResponse(d *schema.ResourceData, m interface{}, responseObj map[string]interface{}, statusCode int, completed bool) {
	// Store the full response JSON
	// prettyJSON, _ := json.MarshalIndent(responseObj, "", "  ")
	// d.Set("log_data", string(prettyJSON))
	d.Set("status_code", statusCode)
	d.Set("workflow_status_code", statusCode)
	d.Set("completed", completed)

	// Define variables for response messages
	var responseMessage, failureReason string

	// Process response data
	if resp, ok := responseObj["response"].(map[string]interface{}); ok {
		if requestList, ok := resp["requestList"].([]interface{}); ok && len(requestList) > 0 {
			if firstRequest, ok := requestList[0].(map[string]interface{}); ok {
				// Extract workflow details
				if workflowName, ok := firstRequest["workflowName"].(string); ok {
					d.Set("workflow_name", workflowName)
				}

				if status, ok := firstRequest["status"].(string); ok {
					d.Set("workflow_status", status)
					log.Printf("[INFO] Workflow status: %s (code: %d)", status, statusCode)
				}

				if createdBy, ok := firstRequest["created_by"].(string); ok {
					d.Set("created_by", createdBy)
				}

				if createdTime, ok := firstRequest["created_time"].(float64); ok {
					// Convert Unix timestamp to readable format
					t := time.Unix(int64(createdTime)/1000, 0)
					d.Set("created_time", t.Format(time.RFC3339))
				}

				// Set success flag based on status code
				isSuccess := statusCode == STATUS_SUCCESS
				d.Set("success", isSuccess)
				// log.Printf("[INFO] Workflow success: %t (completed: %t)", isSuccess, completed)

				// Pretty logging for success or failure
				requestId := d.Get("request_id").(string)
				commonName := d.Get("certificate_common_name").(string)

				if isSuccess {
					// Create a success summary in JSON format
					successData := map[string]interface{}{
						"operation":    "Certificate Creation and Push",
						"status":       "Successful",
						"workflow_id":  requestId,
						"status_code":  statusCode,
						"completed_at": time.Now().Format(time.RFC3339),
					}

					if commonName != "" {
						successData["certificate_common_name"] = commonName
					}

					// Process tasks to extract certificate resource ID if needed
					if tasks, ok := firstRequest["tasks"].([]interface{}); ok {
						// Log how many tasks we found
						log.Printf("[DEBUG] Found %d tasks in workflow response", len(tasks))

						// Extract certificate resource ID if workflow succeeded
						resourceId := extractCertificateResourceId(tasks)
						if resourceId != "" {
							d.Set("certificate_resource_id", resourceId)
							log.Printf("[INFO] Saved certificate resource ID to state: %s", resourceId)
						}
					}

					// Check if certificate download is required and handle it
					if d.Get("is_download_required").(bool) {
						log.Printf("[INFO] Certificate download is required, initiating download...")
						// Call the download function with the necessary parameters
						downloadCertificateIfRequired(d, m, true)

						resourceId := d.Get("certificate_resource_id").(string)
						successData["resource_id"] = resourceId

						certificateDownloadPath := d.Get("certificate_download_path").(string)
						successData["certificate_download_path"] = certificateDownloadPath

					} else {
						log.Printf("[INFO] Certificate download not requested (is_download_required=false)")
					}
					successJSON, _ := json.MarshalIndent(successData, "", "  ")
					successMessage := fmt.Sprintf("\n[CERTIFICATE CREATION][SUCCESS] ✅ Operation Result:\n%s\n", string(successJSON))
					log.Println(successMessage)
				} else if completed {
					// Create a failure summary for completed but failed workflows
					failureData := map[string]interface{}{
						"operation":    "Certificate Creation and Push",
						"status":       "Failed",
						"workflow_id":  requestId,
						"status_code":  statusCode,
						"completed_at": time.Now().Format(time.RFC3339),
					}

					// Add certificate common name if available
					if commonName != "" {
						failureData["certificate_common_name"] = commonName
					}

					// Process tasks and extract failure information
					var taskSummary, failedTasksLog string
					failureReason = ""

					if tasks, ok := firstRequest["tasks"].([]interface{}); ok {
						taskSummary, failedTasksLog, failureReason = processTasks(tasks, isSuccess)
						d.Set("task_summary", taskSummary)
						d.Set("failed_task_logs", failedTasksLog)

						if failureReason != "" && failureReason != "No specific failure reason found in logs" {
							failureData["failure_reason"] = failureReason
						} else {
							// Try to find failure info directly in the workflow response
							if message, ok := firstRequest["message"].(string); ok && message != "" {
								if containsAny(message, []string{"Failed", "Error", "failed", "error"}) {
									failureReason = message
									failureData["failure_reason"] = failureReason
								}
							}

							// If still no reason, check if there's a tooltip
							if tooltip, ok := firstRequest["toolTip"].(string); ok && tooltip != "" {
								failureReason = tooltip
								failureData["failure_reason"] = failureReason
							}
						}
					}

					failureJSON, _ := json.MarshalIndent(failureData, "", "  ")
					failureMessage := fmt.Sprintf("\n[CERTIFICATE CREATION AND PUSH TO AKV][FAILURE] ❌ Operation Result:\n%s\n", string(failureJSON))
					log.Println(failureMessage)
				} else {
					// For incomplete operations (timed out)
					timeoutData := map[string]interface{}{
						"operation":      "Certificate Creation and Push",
						"status":         "Timeout",
						"workflow_id":    requestId,
						"status_code":    statusCode,
						"completed":      false,
						"message":        "Polling timed out before workflow completion",
						"last_polled_at": time.Now().Format(time.RFC3339),
					}

					// Add certificate common name if available
					if commonName != "" {
						timeoutData["certificate_common_name"] = commonName
					}

					timeoutJSON, _ := json.MarshalIndent(timeoutData, "", "  ")
					timeoutMessage := fmt.Sprintf("\n[CERTIFICATE CREATION][TIMEOUT] ⏱️ Operation Result:\n%s\n", string(timeoutJSON))
					log.Println(timeoutMessage)
				}
			}
		}
	}

	// Add the failure reason to the response message if it exists
	if failureReason != "" && failureReason != "No specific failure reason found in logs" {
		log.Printf("[INFO] Failure reason: %s", failureReason)
	}

	// Set the response message and failure reason
	d.Set("response_message", responseMessage)
	d.Set("failure_reason", failureReason)
}

func buildResponseMessage(requestData map[string]interface{}, statusCode int, failureReason string) string {
	var message bytes.Buffer

	// Extract basic workflow info
	workflowName, _ := requestData["workflowName"].(string)
	requestId, _ := requestData["requestId"].(string)
	status, _ := requestData["status"].(string)

	// Format the message as JSON
	responseData := map[string]interface{}{
		"workflow_name": workflowName,
		"request_id":    requestId,
		"status":        status,
		"status_code":   statusCode,
		"completed":     statusCode != STATUS_IN_PROGRESS,
		"successful":    statusCode == STATUS_SUCCESS,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	// If there's a failure, add error information
	if statusCode != STATUS_IN_PROGRESS && statusCode != STATUS_SUCCESS {
		responseData["error"] = "Workflow execution failed"

		// Add the failure reason if we have one
		if failureReason != "" && failureReason != "No specific failure reason found in logs" {
			responseData["failure_reason"] = failureReason
		}
	}

	// Create pretty JSON
	prettyJSON, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		message.WriteString(fmt.Sprintf("Error creating response message: %v", err))
	} else {
		message.WriteString(string(prettyJSON))
	}

	return message.String()
}

// Update the extractFailureReason function with a simpler approach
func extractFailureReason(logs []interface{}) string {
	// If there are no logs, we can't extract a failure reason
	if len(logs) == 0 {
		return "No logs found to determine failure reason"
	}

	// Get the second-to-last log entry if available (often contains the failure message)
	// If not available, try the last entry
	var relevantLog map[string]interface{}

	if len(logs) >= 2 {
		if logEntry, ok := logs[len(logs)-2].(map[string]interface{}); ok {
			relevantLog = logEntry
		}
	}

	// If we couldn't get the second-to-last, try the last one
	if relevantLog == nil && len(logs) > 0 {
		if logEntry, ok := logs[len(logs)-1].(map[string]interface{}); ok {
			relevantLog = logEntry
		}
	}

	// If we found a relevant log entry, extract the message
	if relevantLog != nil {
		if message, ok := relevantLog["message"].(string); ok && message != "" {
			// Format the failure message as pretty JSON
			failureData := map[string]interface{}{
				"type":    "Workflow Task Failure",
				"message": message,
			}

			// Add user if available
			if user, ok := relevantLog["user"].(string); ok && user != "" {
				failureData["reported_by"] = user
			}

			// Add timestamp if available
			if timestamp, ok := relevantLog["time"].(float64); ok && timestamp > 0 {
				t := time.Unix(int64(timestamp)/1000, 0)
				failureData["timestamp"] = t.Format(time.RFC3339)
			}

			return message
		}
	}

	return "No specific failure reason found in logs"
}

// Update the processTasks function to focus on the actual failure message
func processTasks(tasks []interface{}, isSuccess bool) (string, string, string) {
	var taskSummary bytes.Buffer
	var failedTaskLogs bytes.Buffer
	var failureReason string

	taskSummary.WriteString("Task Status Summary:\n")
	taskSummary.WriteString("-------------------\n")

	// First find any failed tasks
	var failedTasks []map[string]interface{}

	for _, t := range tasks {
		task, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		taskState := getIntValue(task, "state")
		taskName := getStringValue(task, "task_name")
		taskStatus := getStringValue(task, "task_status")

		// Add to summary regardless of status
		taskSummary.WriteString(fmt.Sprintf("- %s: %s (State: %d)\n", taskName, taskStatus, taskState))

		if isFailedState(taskState) {
			failedTasks = append(failedTasks, task)
		}
	}

	// If we found failed tasks, focus on them
	if len(failedTasks) > 0 {
		log.Printf("[INFO] Found %d failed tasks in workflow", len(failedTasks))

		// Get the first failed task for the primary error message
		failedTask := failedTasks[0]
		taskName := getStringValue(failedTask, "task_name")
		taskStatus := getStringValue(failedTask, "task_status")

		failedTaskLogs.WriteString(fmt.Sprintf("\n== FAILED TASK: %s ==\n", taskName))
		failedTaskLogs.WriteString(fmt.Sprintf("Status: %s\n\n", taskStatus))

		// Extract logs for the failed task
		if logs, ok := failedTask["logs"].([]interface{}); ok {
			failedTaskLogs.WriteString("Logs:\n")

			// Get the failure reason from the logs
			failureReason = extractFailureReason(logs)

			// Log all messages for this failed task
			for _, l := range logs {
				logEntry, ok := l.(map[string]interface{})
				if !ok {
					continue
				}

				user := getStringValue(logEntry, "user")
				message := getStringValue(logEntry, "message")
				timestamp := getFloatValue(logEntry, "time")

				// Format the log entry
				timeStr := ""
				if timestamp > 0 {
					t := time.Unix(int64(timestamp)/1000, 0)
					timeStr = t.Format(time.RFC3339)
				}

				failedTaskLogs.WriteString(fmt.Sprintf("[%s] %s: %s\n", timeStr, user, message))
			}
		}
	}

	return taskSummary.String(), failedTaskLogs.String(), failureReason
}

// Helper function to check if a string contains any of the given substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// Clean up error messages that might contain Python structures, JSON, etc.
func cleanErrorMessage(message string) string {
	// If message contains "Failure Reason", extract that part
	if strings.Contains(message, "Failure Reason:") {
		return extractReasonFromMessage(message)
	}

	// Remove Python list/dict syntax if present
	re := regexp.MustCompile(`\[.*?\]|\{.*?\}`)
	message = re.ReplaceAllString(message, "")

	// Remove extra whitespace
	message = strings.TrimSpace(message)
	reSpace := regexp.MustCompile(`\s+`)
	message = reSpace.ReplaceAllString(message, " ")

	// If we have a message after all the cleaning, return it
	if message != "" && message != ":" && len(message) > 3 {
		return message
	}

	return ""
}

// Extract the specific reason from a "Failure Reason:" message
func extractReasonFromMessage(message string) string {
	// Look for the pattern "Failure Reason: [something]" or similar
	if strings.Contains(message, "Failure Reason:") {
		parts := strings.SplitN(message, "Failure Reason:", 2)
		if len(parts) > 1 {
			failureMsg := strings.TrimSpace(parts[1])

			// If the failure message contains error info inside Python-like structures
			// Try to extract the actual message text
			if strings.Contains(failureMsg, "'message':") {
				// Extract text between 'message': and the next comma or closing bracket
				re := regexp.MustCompile(`'message':\s*'([^']+)'`)
				matches := re.FindStringSubmatch(failureMsg)
				if len(matches) > 1 {
					return matches[1]
				}
			}

			// If the failure message is a Python-like list
			if strings.HasPrefix(failureMsg, "[") && strings.HasSuffix(failureMsg, "]") {
				// Try to extract the most useful part - often the last message
				if strings.Contains(failureMsg, "'message':") {
					re := regexp.MustCompile(`'message':\s*'([^']+)'`)
					matches := re.FindAllStringSubmatch(failureMsg, -1)
					if len(matches) > 0 {
						// Return the last message as it's often the most specific
						return matches[len(matches)-1][1]
					}
				}
			}

			// Return the whole message if we couldn't extract a specific part
			return failureMsg
		}
	}

	// Direct error messages without "Failure Reason:" prefix
	errorIndicators := []string{"Error:", "Failed:", "Unable to"}
	for _, indicator := range errorIndicators {
		if strings.Contains(message, indicator) {
			parts := strings.SplitN(message, indicator, 2)
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// Helper functions for getting values from maps
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

func isFailedState(state int) bool {
	// Check if the state indicates failure
	for _, failedState := range failedStatusCodes {
		if state == failedState {
			return true
		}
	}
	return false
}

func extractCertificateResourceId(tasks []interface{}) string {
	// Look for the specific task
	for _, t := range tasks {
		task, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		taskName := getStringValue(task, "task_name")
		taskStatus := getStringValue(task, "task_status")

		// Check if this is the "Trigger Certificate Creation" task and it succeeded
		if taskName == "Trigger Certificate Creation" && taskStatus == "Success" {
			// Extract logs to find the resource ID
			if logs, ok := task["logs"].([]interface{}); ok {
				// Try different extraction strategies

				// Strategy 1: Look for JSON string containing resourceId
				for _, l := range logs {
					logEntry, ok := l.(map[string]interface{})
					if !ok {
						continue
					}

					message := getStringValue(logEntry, "message")

					// Skip empty messages
					if message == "" {
						continue
					}

					// Look for JSON response containing resourceId
					if strings.Contains(message, "resourceId") {
						log.Printf("[DEBUG] Found message containing resourceId: %s", message)

						// Strategy 1.1: Extract using regex for Python dict format
						reDict := regexp.MustCompile(`'resourceId':\s*'([^']+)'`)
						matches := reDict.FindStringSubmatch(message)
						if len(matches) > 1 {
							log.Printf("[INFO] Extracted certificate resource ID (Python dict): %s", matches[1])
							return matches[1]
						}

						// Strategy 1.2: Extract using regex for JSON format
						reJson := regexp.MustCompile(`"resourceId":\s*"([^"]+)"`)
						matches = reJson.FindStringSubmatch(message)
						if len(matches) > 1 {
							log.Printf("[INFO] Extracted certificate resource ID (JSON): %s", matches[1])
							return matches[1]
						}

						// Strategy 1.3: Try to parse the JSON string
						if strings.Contains(message, "{") && strings.Contains(message, "}") {
							jsonStart := strings.Index(message, "{")
							jsonEnd := strings.LastIndex(message, "}") + 1

							if jsonStart >= 0 && jsonEnd > jsonStart {
								jsonStr := message[jsonStart:jsonEnd]

								var jsonData map[string]interface{}
								if err := json.Unmarshal([]byte(jsonStr), &jsonData); err == nil {
									if resp, ok := jsonData["response"].(map[string]interface{}); ok {
										if resourceId, ok := resp["resourceId"].(string); ok && resourceId != "" {
											log.Printf("[INFO] Extracted certificate resource ID (JSON parse): %s", resourceId)
											return resourceId
										}
									}
								} else {
									log.Printf("[DEBUG] Failed to parse JSON: %v", err)
								}
							}
						}
					}
				}

				// Strategy 2: Look for specific log messages about resource creation
				for _, l := range logs {
					logEntry, ok := l.(map[string]interface{})
					if !ok {
						continue
					}

					message := getStringValue(logEntry, "message")

					// Look for resource creation messages
					if strings.Contains(message, "Certificate created with resource ID") {
						re := regexp.MustCompile(`Certificate created with resource ID[:\s]+([a-zA-Z0-9]+)`)
						matches := re.FindStringSubmatch(message)
						if len(matches) > 1 {
							log.Printf("[INFO] Extracted certificate resource ID from creation message: %s", matches[1])
							return matches[1]
						}
					}
				}
			}
		}
	}

	// If we didn't find the specific task, look in all tasks as a fallback
	for _, t := range tasks {
		task, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		if logs, ok := task["logs"].([]interface{}); ok {
			for _, l := range logs {
				logEntry, ok := l.(map[string]interface{})
				if !ok {
					continue
				}

				message := getStringValue(logEntry, "message")

				// Skip empty messages
				if message == "" {
					continue
				}

				// Look for resourceId in any log message
				if strings.Contains(message, "resourceId") {
					log.Printf("[DEBUG] Found message containing resourceId in task %s: %s",
						getStringValue(task, "task_name"), message)

					// Try regex extraction
					re := regexp.MustCompile(`['"](resourceId)['"]:\s*['"]([^'"]+)['"]`)
					matches := re.FindStringSubmatch(message)
					if len(matches) > 2 {
						log.Printf("[INFO] Extracted certificate resource ID from general logs: %s", matches[2])
						return matches[2]
					}
				}
			}
		}
	}

	log.Printf("[INFO] No certificate resource ID found in workflow logs")
	return ""
}

// Add this function to your resource_workflow_logs.go file

// fetchCertificateDetails retrieves certificate details using the resource ID
func fetchCertificateDetails(resourceId, certType, appviewxSessionID, accessToken string, configAppViewXEnvironment *config.AppViewXEnvironment) (string, string, error) {
	log.Printf("[INFO] Fetching certificate details for resource ID: %s", resourceId)

	// Extract configuration parameters
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS

	// Set query parameters
	queryParams := map[string]string{
		"gwsource": "external",
	}

	// Get URL for the certificate search endpoint
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, "certificate/search", queryParams, appviewxEnvironmentIsHTTPS)
	log.Printf("Certificate Type :::::::::::::::::::::::::::::::: %s", certType)
	// Build search payload
	payload := map[string]interface{}{
		"input": map[string]interface{}{
			"resourceId": resourceId,
			"category":   certType,
		},
		"filter": map[string]interface{}{
			"start": 1,
			"max":   1,
		},
	}

	// Prepare the request
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Error marshalling certificate search payload: %v", err)
		return "", "", err
	}

	// Create HTTP client
	client := &http.Client{Transport: HTTPTransport()}

	// Create request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("[ERROR] Error creating certificate search request: %v", err)
		return "", "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication header
	if appviewxSessionID != "" {
		req.Header.Set(constants.SESSION_ID, appviewxSessionID)
	} else if accessToken != "" {
		req.Header.Set(constants.TOKEN, accessToken)
	}

	log.Printf("[DEBUG] Sending certificate search request to: %s", url)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Error making certificate search request: %v", err)
		return "", "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Error reading certificate search response: %v", err)
		return "", "", err
	}

	// Format and log JSON response for debugging
	// var prettyJSON bytes.Buffer
	// if json.Indent(&prettyJSON, body, "", "  ") == nil {
	// 	log.Printf("[DEBUG] Certificate search response body (formatted JSON):\n%s", prettyJSON.String())
	// } else {
	// 	log.Printf("[DEBUG] Certificate search response body (raw):\n%s", string(body))
	// }

	// Parse response to extract certificate details
	var responseObj map[string]interface{}
	if err := json.Unmarshal(body, &responseObj); err != nil {
		log.Printf("[ERROR] Error parsing certificate search response: %v", err)
		return "", "", err
	}

	// Extract certificate common name and serial number from response
	var commonName, serialNumber string

	if resp, ok := responseObj["response"].(map[string]interface{}); ok {
		if innerResp, ok := resp["response"].(map[string]interface{}); ok {
			if objects, ok := innerResp["objects"].([]interface{}); ok && len(objects) > 0 {
				if cert, ok := objects[0].(map[string]interface{}); ok {
					if cn, ok := cert["commonName"].(string); ok {
						commonName = cn
						log.Printf("[INFO] Found certificate common name: %s", commonName)
					}

					if sn, ok := cert["serialNumber"].(string); ok {
						serialNumber = sn
						log.Printf("[INFO] Found certificate serial number: %s", serialNumber)
					}
				}
			}
		}
	}

	if commonName == "" || serialNumber == "" {
		log.Printf("[WARN] Could not extract certificate details from response")
		return "", "", fmt.Errorf("certificate details not found in response")
	}

	return commonName, serialNumber, nil
}

// downloadCertificateIfRequired handles certificate downloading if requested in configuration
func downloadCertificateIfRequired(d *schema.ResourceData, m interface{}, isSuccess bool) {
	// Only proceed if workflow succeeded and download is requested
	if !isSuccess || !d.Get("is_download_required").(bool) {
		return
	}

	resourceId := d.Get("certificate_resource_id").(string)
	certCommonName := d.Get("certificate_common_name").(string)
	if certCommonName== "" {
		log.Printf("[INFO] Certificate Common Name not found in the Input, Proceeding with the Default Certificate Name")
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	    certCommonName = "certificate-" + resourceId + "-" + timestamp
	}
	log.Printf("[DEBUG] Certificate common name: %s", certCommonName)
	if resourceId == "" {
		log.Printf("[WARN] Cannot download certificate: No certificate resource ID found in workflow response")
		return
	}

	log.Printf("[INFO] Initiating certificate download for resource ID: %s", resourceId)

	// Get authentication tokens
	configAppViewXEnvironment := m.(*config.AppViewXEnvironment)
	appviewxUserName := configAppViewXEnvironment.AppViewXUserName
	appviewxPassword := configAppViewXEnvironment.AppViewXPassword
	appviewxClientId := configAppViewXEnvironment.AppViewXClientId
	appviewxClientSecret := configAppViewXEnvironment.AppViewXClientSecret
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS

	var appviewxSessionID, accessToken string
	var err error

	// Use same authentication as the rest of the function
	if appviewxUserName != "" && appviewxPassword != "" {
		appviewxSessionID, err = GetSession(appviewxUserName, appviewxPassword, appviewxEnvironmentIP, appviewxEnvironmentPort, "WEB", appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Printf("[ERROR] Error getting session for certificate download: %v", err)
			return
		}
	} else if appviewxClientId != "" && appviewxClientSecret != "" {
		accessToken, err = GetAccessToken(appviewxClientId, appviewxClientSecret, appviewxEnvironmentIP, appviewxEnvironmentPort, "WEB", appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Printf("[ERROR] Error getting access token for certificate download: %v", err)
			return
		}
	}

	// Get download parameters
	downloadPath := d.Get("certificate_download_path").(string)
	downloadFormat := d.Get("certificate_download_format").(string)
	certificateChainRequired := d.Get("certificate_chain_required").(bool)

	if downloadPath == "" {
		log.Printf("[WARN] Cannot download certificate: No download path specified")
		return
	}

	// Prepare download path with certificate common name
	fullDownloadPath := downloadPath
	if !strings.HasSuffix(fullDownloadPath, "/") {
		fullDownloadPath += "/"
	}

	// Sanitize common name for filename
	safeCommonName := strings.ReplaceAll(certCommonName, "*", "wildcard")
	safeCommonName = strings.ReplaceAll(safeCommonName, ".", "_")
	safeCommonName = strings.ReplaceAll(safeCommonName, " ", "_")

	fullDownloadPath += safeCommonName + "." + strings.ToLower(downloadFormat)

	log.Printf("[INFO] Downloading certificate to: %s", fullDownloadPath)
	// Get certificate download password if required
	certDownloadPassword := d.Get("certificate_download_password").(string)
	// Download certificate
	downloadSuccess := downloadCertificateFromAppviewx(
		resourceId,
		certCommonName,
		"",
		downloadFormat,
		certDownloadPassword,
		fullDownloadPath,
		certificateChainRequired,
		appviewxSessionID,
		accessToken,
		configAppViewXEnvironment,
	)

	if downloadSuccess {
		log.Printf("[INFO] Certificate downloaded successfully to: %s", fullDownloadPath)
		d.Set("downloaded_certificate_path", fullDownloadPath)
	} else {
		log.Printf("[ERROR] Failed to download certificate")
	}
}
