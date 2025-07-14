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

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
)

func ResourceSearchCertificateByKeyword() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSearchCertificateByKeywordCreate,
		ReadContext:   resourceSearchCertificateByKeywordRead,
		UpdateContext: resourceSearchCertificateByKeywordUpdate,
		DeleteContext: resourceSearchCertificateByKeywordDelete,

		Schema: map[string]*schema.Schema{
			"category": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Category of certificate (e.g., Server)",
			},
			"cert_serial_no": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate serial number",
			},
			"cert_issuer": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate issuer",
			},
			"cert_cn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate common name",
			},
			"cert_san": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate SAN",
			},
			"max_results": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Maximum number of results to return",
			},
			"start_index": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Start index for pagination",
			},
			"sort_column": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "commonName",
				Description: "Column to sort results by",
			},
			"sort_order": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "desc",
				Description: "Sort order (asc or desc)",
			},
			"total_records": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of records found",
			},
		},
	}
}

func resourceSearchCertificateByKeywordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n====================[CERTIFICATE SEARCH]====================")
	log.Println("  🚀  Resource Search Certificate By Keyword Create")
	log.Println("==================================================================\n")

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
			log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error in getting the session:")
			log.Println("   ", err)
			log.Println("------------------------------------------------------------------\n")
			return diag.FromErr(err)
		}
	} else if appviewxClientId != "" && appviewxClientSecret != "" {
		accessToken, err = GetAccessToken(appviewxClientId, appviewxClientSecret, appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxGwSource, appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error in getting the access token:")
			log.Println("   ", err)
			log.Println("------------------------------------------------------------------\n")
			return diag.FromErr(err)
		}
	}

	result, err := searchCertificatesByKeyword(d, configAppViewXEnvironment, appviewxSessionID, accessToken)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set resource ID to something unique to identify this search
	searchId := fmt.Sprintf("cert_search_%s", d.Get("category").(string))
	d.SetId(searchId)

	// Only store non-sensitive metadata in state
	d.Set("total_records", result.TotalRecords)

	// DO NOT store certificates or raw response in state

	log.Printf("\n[CERTIFICATE SEARCH][INFO] ✅ Search complete with %d total records\n", result.TotalRecords)

	return nil
}

// Structure for search results
type CertificateSearchResult struct {
	RawResponse  string
	TotalRecords int
	Certificates []Certificate
}

// Structure for certificate details
type Certificate struct {
	ID                 string
	UUID               string
	CommonName         string
	SerialNumber       string
	Issuer             string
	Status             string
	ValidFrom          string
	ValidTo            string
	ValidFor           string
	KeyAlgorithm       string
	SignatureAlgorithm string
	ThumbPrint         string
	ResourceID         string
}

func searchCertificatesByKeyword(d *schema.ResourceData, configAppViewXEnvironment *config.AppViewXEnvironment, appviewxSessionID, accessToken string) (CertificateSearchResult, error) {
	var result CertificateSearchResult
	httpMethod := config.HTTPMethodPost
	appviewxEnvironmentIP := configAppViewXEnvironment.AppViewXEnvironmentIP
	appviewxEnvironmentPort := configAppViewXEnvironment.AppViewXEnvironmentPort
	appviewxEnvironmentIsHTTPS := configAppViewXEnvironment.AppViewXIsHTTPS

	// Set query parameters exactly as in curl
	queryParams := map[string]string{
		"gwkey":    "f000ca01",
		"gwsource": "external",
	}

	// Get URL
	url := GetURL(appviewxEnvironmentIP, appviewxEnvironmentPort, "certificate/search", queryParams, appviewxEnvironmentIsHTTPS)

	log.Printf("\n[CERTIFICATE SEARCH][DEBUG] 🔍 Searching certificates using URL: %s\n", url)

	// Build search payload
	payload := buildSearchPayload(d)

	// Pretty print payload
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("\n[CERTIFICATE SEARCH][DEBUG] 📝 Search payload:\n%s\n", string(payloadBytes))

	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error in Marshalling the payload:")
		log.Println("   ", err)
		log.Printf("   Payload: %+v\n", payload)
		log.Println("------------------------------------------------------------------\n")
		return result, err
	}

	client := &http.Client{Transport: HTTPTransport()}

	log.Printf("\n[CERTIFICATE SEARCH][INFO] 🌐 Making request to %s\n", url)

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error in creating new Request:")
		log.Println("   ", err)
		log.Println("------------------------------------------------------------------\n")
		return result, err
	}

	// Set headers directly instead of using frameHeaders()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add session ID header
	if appviewxSessionID != "" {
		log.Printf("[CERTIFICATE SEARCH][DEBUG] 🔑 Using session ID for authentication")
		req.Header.Set(constants.SESSION_ID, appviewxSessionID)
	} else if accessToken != "" {
		log.Printf("[CERTIFICATE SEARCH][DEBUG] 🔑 Using access token for authentication\n")
		req.Header.Set(constants.TOKEN, accessToken)
	}

	// Debug headers with pretty print
	// headersBytes, _ := json.MarshalIndent(req.Header, "", "  ")
	// log.Printf("[CERTIFICATE SEARCH][DEBUG] 🏷️ Request headers:\n%s\n", string(headersBytes))

	httpResponse, err := client.Do(req)
	if err != nil {
		log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error in searching certificates:")
		log.Println("   ", err)
		log.Println("------------------------------------------------------------------\n")
		return result, err
	}
	defer httpResponse.Body.Close()

	log.Printf("[CERTIFICATE SEARCH][INFO] 📊 Search certificates response status code: %s\n", httpResponse.Status)

	// Read response body
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		log.Println("[ERROR] Unable to read response body: ", err)
		return result, err
	}

	// Log full response for debugging
	// Format and log JSON response for better readability
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err != nil {
		log.Printf("\n[CERTIFICATE SEARCH][DEBUG] 📦 Search response body (raw):\n%s\n", string(responseBody))
	} else {
		log.Printf("\n[CERTIFICATE SEARCH][DEBUG] 📦 Search response body (formatted JSON):\n%s\n", prettyJSON.String())
	}

	// Store raw response
	result.RawResponse = string(responseBody)

	// Check for error responses
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		log.Println("\n[CERTIFICATE SEARCH][ERROR] ❌ Error response received:")
		log.Println("   Status code:", httpResponse.Status)
		log.Printf("   Response: %s\n", string(responseBody))
		log.Println("------------------------------------------------------------------\n")
		return result, errors.New("error in searching certificates: " + string(responseBody))
	}

	// Parse response
	var responseObj map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseObj); err != nil {
		log.Println("[ERROR] Unable to unmarshal the response: ", err)
		return result, err
	}

	// Extract total records and certificates from the proper path in the response structure
	if topResponse, ok := responseObj["response"].(map[string]interface{}); ok {
		if innerResponse, ok := topResponse["response"].(map[string]interface{}); ok {
			// Extract totalRecords
			if totalRecords, ok := innerResponse["totalRecords"].(float64); ok {
				result.TotalRecords = int(totalRecords)
			}

			// Extract certificate records from "objects" array
			if objects, ok := innerResponse["objects"].([]interface{}); ok {
				for _, record := range objects {
					if certMap, ok := record.(map[string]interface{}); ok {
						cert := Certificate{}

						if val, ok := certMap["resourceId"].(string); ok {
							cert.ResourceID = val
						}
						if val, ok := certMap["uuid"].(string); ok {
							cert.UUID = val
							cert.ID = val // Use UUID as ID if no specific ID field
						}
						if val, ok := certMap["commonName"].(string); ok {
							cert.CommonName = val
						}
						if val, ok := certMap["serialNumber"].(string); ok {
							cert.SerialNumber = val
						}
						if val, ok := certMap["issuerCommonName"].(string); ok {
							cert.Issuer = val
						} else if val, ok := certMap["issuer"].(string); ok {
							cert.Issuer = val
						}
						if val, ok := certMap["status"].(string); ok {
							cert.Status = val
						}

						// Extract additional fields
						if val, ok := certMap["validFrom"].(float64); ok {
							cert.ValidFrom = fmt.Sprintf("%f", val)
						}
						if val, ok := certMap["validTo"].(float64); ok {
							cert.ValidTo = fmt.Sprintf("%f", val)
						}
						if val, ok := certMap["validFor"].(string); ok {
							cert.ValidFor = val
						}
						if val, ok := certMap["keyAlgorithmAndSize"].(string); ok {
							cert.KeyAlgorithm = val
						}
						if val, ok := certMap["signatureAlgorithm"].(string); ok {
							cert.SignatureAlgorithm = val
						}
						if val, ok := certMap["thumbPrint"].(string); ok {
							cert.ThumbPrint = val
						}

						result.Certificates = append(result.Certificates, cert)
					}
				}
			}
		}
	}

	log.Printf("\n[CERTIFICATE SEARCH][DEBUG] ✅ Extracted %d certificates from response\n", len(result.Certificates))
	log.Println("==================================================================\n")
	return result, nil
}

func buildSearchPayload(d *schema.ResourceData) map[string]interface{} {
	category := d.Get("category").(string)
	maxResults := d.Get("max_results").(int)
	startIndex := d.Get("start_index").(int)
	sortColumn := d.Get("sort_column").(string)
	sortOrder := d.Get("sort_order").(string)

	// Build keyword search payload
	keywordSearch := make(map[string]interface{})

	// Add search criteria if provided
	if serialNo, ok := d.GetOk("cert_serial_no"); ok {
		keywordSearch["certserialno"] = serialNo.(string)
	}
	if issuer, ok := d.GetOk("cert_issuer"); ok {
		keywordSearch["certissuer"] = issuer.(string)
	}
	if cn, ok := d.GetOk("cert_cn"); ok {
		keywordSearch["certcn"] = cn.(string)
	}
	if san, ok := d.GetOk("cert_san"); ok {
		keywordSearch["certsan"] = san.(string)
	}

	// Build filter payload
	filter := map[string]interface{}{
		"max":        fmt.Sprintf("%d", maxResults),
		"start":      fmt.Sprintf("%d", startIndex),
		"sortColumn": sortColumn,
		"sortOrder":  sortOrder,
	}

	// Build the complete payload
	payload := map[string]interface{}{
		"input": map[string]interface{}{
			"category":      category,
			"keywordSearch": keywordSearch,
		},
		"filter": filter,
	}

	// Log the final search criteria
	log.Printf("[CERTIFICATE SEARCH][DEBUG] 🔍 Search criteria: Category=%s, Results=%d-%d, Sort=%s %s\n",
		category, startIndex, startIndex+maxResults-1, sortColumn, sortOrder)

	return payload
}

func resourceSearchCertificateByKeywordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE SEARCH][INFO] ℹ️ GET OPERATION RETURNS EXISTING DATA\n")
	return nil
}

func resourceSearchCertificateByKeywordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE SEARCH][INFO] 🔄 UPDATE OPERATION TRIGGERS NEW SEARCH\n")
	return resourceSearchCertificateByKeywordCreate(ctx, d, m)
}

func resourceSearchCertificateByKeywordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("\n[CERTIFICATE SEARCH][INFO] 🗑️ Removing certificate search resource from state\n")
	d.SetId("")
	return nil
}
