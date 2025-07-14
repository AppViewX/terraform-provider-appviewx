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
				Type:     schema.TypeString,
				Optional: true,
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
				Type:     schema.TypeString,
				Optional: true,
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
	if downloadPassword, ok = GetDownloadPassword(resourceData, downloadFormat); !ok {
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
		if err := downloadKey(resourceData, resourceId, appviewxSessionID, accessToken, configAppViewXEnvironment); err != nil {
			return err
		}
	}
	return nil
}
