package appviewx

import (
	"errors"
	"log"
	"math/rand"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-appviewx/appviewx/config"
	"terraform-provider-appviewx/appviewx/constants"
)

func ResourceDownloadCertificateServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceDownloadCertificate,
		Read:   resourceCertificateServerRead,
		Update: resourceCertificateServerUpdate,
		Delete: resourceCertificateServerDelete,

		Schema: map[string]*schema.Schema{
			constants.COMMON_NAME: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.SERIAL_NUMBER: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.RESOURCE_ID: &schema.Schema{
				Type:     schema.TypeString,
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
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for certificate download (resource level) - provider level password takes priority",
			},
			constants.CERTIFICATE_CHAIN_REQUIRED: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			constants.KEY_DOWNLOAD_PATH: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			constants.KEY_DOWNLOAD_PASSWORD: &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for private key download (resource level) - provider level password takes priority",
			},
			constants.DOWNLOAD_PASSWORD_PROTECTED_KEY: &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

// TODO: cleanup to be done
func resourceDownloadCertificate(resourceData *schema.ResourceData, m interface{}) error {

	log.Println("****************** Resource Download Certificate ******************")
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
			return nil
		}
	} else if appviewxClientId != "" && appviewxClientSecret != "" {
		accessToken, err = GetAccessToken(appviewxClientId, appviewxClientSecret, appviewxEnvironmentIP, appviewxEnvironmentPort, appviewxGwSource, appviewxEnvironmentIsHTTPS)
		if err != nil {
			log.Println("[ERROR] Error in getting the access token due to : ", err)
			return nil
		}
	}

	commonName := resourceData.Get(constants.COMMON_NAME).(string)
	serialNumber := resourceData.Get(constants.SERIAL_NUMBER).(string)
	var downloadPath, downloadFormat, downloadPassword string
	log.Println("[INFO] CommonName =================================================================== ", commonName)
	var isChainRequired, ok bool
	downloadFormat = GetDownloadFormat(resourceData)
	downloadPath = GetDownloadFilePath(resourceData, commonName, downloadFormat)

	// Get certificate download password with provider priority
	providerCertPassword := configAppViewXEnvironment.ProviderCertDownloadPassword
	resourceCertPassword := resourceData.Get(constants.CERTIFICATE_DOWNLOAD_PASSWORD).(string)
	downloadPassword = getPasswordWithPriority(providerCertPassword, resourceCertPassword)

	// Validate password for formats that require it
	if downloadPassword == "" && (downloadFormat == "PFX" || downloadFormat == "JKS" || downloadFormat == "P12") {
		log.Println("[ERROR] Password not found for the specified download format - " + downloadFormat)
		return errors.New("[ERROR] Password not found for the specified download format - " + downloadFormat)
	} else if downloadPassword != "" && (downloadFormat == "PFX" || downloadFormat == "JKS" || downloadFormat == "P12") {
		log.Println("[INFO] Password found for download format - " + downloadFormat)
		ok = true
	} else {
		ok = true
	}

	if !ok {
		return errors.New("[ERROR] Error in getting the download password")
	}
	isChainRequired = resourceData.Get(constants.CERTIFICATE_CHAIN_REQUIRED).(bool)

	var resourceId = resourceData.Get(constants.RESOURCE_ID).(string)
	if commonName != "" && serialNumber != "" {
		log.Println("[INFO] Common Name and Serial Number are provided in payload hence proceeding with certificate download")
	} else if resourceId != "" {
		log.Println("[INFO] Resource id = ", resourceId, " is available in payload hence proceeding with certificate download")
	} else {
		log.Println("[ERROR] CommonName, SerialNumber or Resource ID details are not available to proceed with certificate download")
		return errors.New("[ERROR] CommonName, SerialNumber or Resource ID details are not available to proceed with certificate download")
	}
	if downloadSuccess := downloadCertificateFromAppviewx(resourceId, commonName, serialNumber, downloadFormat, downloadPassword, downloadPath, isChainRequired, appviewxSessionID, accessToken, configAppViewXEnvironment); downloadSuccess {
		log.Println("[INFO] Certificate downloaded successfully in the specified path")
		resourceData.SetId(strconv.Itoa(rand.Int()))
	} else {
		log.Println("[ERROR] Certificate was not downloaded in the specified path")
		return errors.New("[ERROR] Certificate was not downloaded in the specified path")
	}
	if resourceData.Get(constants.KEY_DOWNLOAD_PATH) != "" {
		log.Println("[INFO] Key download path is provided in the payload hence proceeding with key download")
		if err := downloadKeyWithPriority(resourceData, resourceId, appviewxSessionID, accessToken, configAppViewXEnvironment); err != nil {
			return err
		}
	}
	return nil
}

// getPasswordWithPriority returns password with provider priority over resource
func getPasswordWithPriority(providerPassword, resourcePassword string) string {
	if providerPassword != "" {
		log.Println("[INFO] -------------------Provider password is considered------------------")
		log.Println("[INFO] Using provider-level password")
		return providerPassword
	}
	if resourcePassword != "" {
		log.Println("[INFO] -------------------Resource password is considered------------------")
		log.Println("[INFO] Using resource-level password")
		return resourcePassword
	}
	log.Println("[INFO] No password provided at provider or resource level")
	return ""
}

// downloadKeyWithPriority downloads key using provider-level password priority
func downloadKeyWithPriority(resourceData *schema.ResourceData, resourceID, appviewxSessionID, accessToken string, configAppViewXEnvironment *config.AppViewXEnvironment) error {
	commonName := resourceData.Get(constants.COMMON_NAME).(string)
	downloadPath := GetDownloadFilePathForKey(resourceData, commonName+"_key", "PEM")

	// Get key download password with provider priority
	providerKeyPassword := configAppViewXEnvironment.ProviderKeyDownloadPassword
	resourceKeyPassword := resourceData.Get(constants.KEY_DOWNLOAD_PASSWORD).(string)
	downloadPassword := getPasswordWithPriority(providerKeyPassword, resourceKeyPassword)

	downloadPasswordProtectedKey := resourceData.Get(constants.DOWNLOAD_PASSWORD_PROTECTED_KEY).(bool)

	if downloadPassword == "" {
		log.Println("[ERROR] Password not found for private key download at provider or resource level")
		return errors.New("[ERROR] Password not found for private key download at provider or resource level")
	}

	// Use the existing downloadKeyFromAppviewx function with the priority-based password
	searchResponse := searchCertificate(resourceID, appviewxSessionID, accessToken, configAppViewXEnvironment)
	if searchResponse.AppviewxResponse.ResponseObject.Objects != nil && searchResponse.AppviewxResponse.ResponseObject.Objects[0].UUID == "" {
		log.Println("[ERROR] Cannot find the UUID for the resource id " + resourceID + " to proceed with key download")
		return errors.New("[ERROR] Certificate details was not found to download the private key")
	}
	uuid := searchResponse.AppviewxResponse.ResponseObject.Objects[0].UUID
	log.Println("[INFO] UUID for the resource id " + resourceID + " was obtained successfully")
	if downloadSuccess := downloadKeyFromAppviewx(uuid, downloadPassword, downloadPath, downloadPasswordProtectedKey, appviewxSessionID, accessToken, configAppViewXEnvironment); downloadSuccess {
		log.Println("[INFO] Private key downloaded successfully in the specified path")
		resourceData.SetId(strconv.Itoa(rand.Int()))
	} else {
		log.Println("[ERROR] Private key was not downloaded in the specified path")
		return errors.New("[ERROR] Private key was not downloaded in the specified path")
	}
	return nil
}
