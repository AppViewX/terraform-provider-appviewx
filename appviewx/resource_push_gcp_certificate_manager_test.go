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
	payload := buildGCPPushPayload("6a47c392e9692037365e62e1", "apigee-ingress-usc1-cert", "us-central1", "GCP connector", true, true, profiles)

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
