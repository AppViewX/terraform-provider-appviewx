package appviewx

import (
	"reflect"
	"testing"
)

func TestConstructGCPCertificateID(t *testing.T) {
	got := constructGCPCertificateID("my-project", "us-central1", "apigee-ingress-usc1-cert")
	want := "projects/my-project/locations/us-central1/certificates/apigee-ingress-usc1-cert"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildGCPPushPayload(t *testing.T) {
	profiles := []string{"GCP.Single.All:my-project:Certificate Manager"}
	payload := buildGCPPushPayload("6a47c392e9692037365e62e1", "apigee-ingress-usc1-cert", "us-central1", "GCP connector", "Push and Bind Profiles", true, true, profiles)

	if payload["certificateId"] != "6a47c392e9692037365e62e1" {
		t.Fatalf("unexpected certificateId: %v", payload["certificateId"])
	}
	if !reflect.DeepEqual(payload["selectedProfiles"], profiles) {
		t.Fatalf("unexpected selectedProfiles: %v", payload["selectedProfiles"])
	}

	gi := payload["generalInformation"].(map[string]interface{})
	if gi["vendor"] != "GCP" || gi["category"] != "cloud" {
		t.Fatalf("unexpected generalInformation: %v", gi)
	}
	if gi["name"] != "GCP connector" {
		t.Fatalf("unexpected connector name: %v", gi["name"])
	}
	if gi["profileType"] != "Push and Bind Profiles" {
		t.Fatalf("unexpected profileType: %v", gi["profileType"])
	}
	if gi["profileFilterSelection"] != ":Certificate Manager" {
		t.Fatalf("unexpected generalInformation.profileFilterSelection: %v", gi["profileFilterSelection"])
	}

	cd := payload["certificateDetails"].(map[string]interface{})
	if cd["certificateName"] != "apigee-ingress-usc1-cert" || cd["region"] != "us-central1" {
		t.Fatalf("unexpected certificateDetails: %v", cd)
	}
	if cd["isNewCertificate"] != true {
		t.Fatalf("expected isNewCertificate true, got: %v", cd["isNewCertificate"])
	}
	if cd["profileFilterSelection"] != "Certificate Manager" {
		t.Fatalf("unexpected certificateDetails.profileFilterSelection: %v", cd["profileFilterSelection"])
	}

	pd := payload["pushDetails"].(map[string]interface{})
	if pd["scriptLocation"] != "appviewx" || pd["pushAutomatically"] != true {
		t.Fatalf("unexpected pushDetails: %v", pd)
	}
}

func TestExtractPushIDs_ObjectShape(t *testing.T) {
	response := []interface{}{
		map[string]interface{}{"requestId": "2710282", "connectorId": "1783753921336"},
	}
	requestID, connectorID := extractPushIDs(response)
	if requestID != "2710282" {
		t.Fatalf("unexpected requestID: %q", requestID)
	}
	if connectorID != "1783753921336" {
		t.Fatalf("unexpected connectorID: %q", connectorID)
	}
}

func TestExtractPushIDs_StringShape(t *testing.T) {
	response := []interface{}{"1783800974255"}
	requestID, connectorID := extractPushIDs(response)
	if requestID != "" {
		t.Fatalf("expected empty requestID, got %q", requestID)
	}
	if connectorID != "1783800974255" {
		t.Fatalf("unexpected connectorID: %q", connectorID)
	}
}

func TestExtractPushIDs_Empty(t *testing.T) {
	requestID, connectorID := extractPushIDs(nil)
	if requestID != "" || connectorID != "" {
		t.Fatalf("expected empty ids, got %q / %q", requestID, connectorID)
	}
}
