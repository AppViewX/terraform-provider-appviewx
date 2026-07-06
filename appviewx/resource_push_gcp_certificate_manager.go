package appviewx

import (
	"fmt"
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
