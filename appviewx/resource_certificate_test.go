package appviewx

import (
	"testing"

	"terraform-provider-appviewx/appviewx/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFrameCertificatePayload_AutoRenewalEnabled(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceCertificateServer().Schema, map[string]interface{}{
		constants.COMMON_NAME:             "dev.api.example.com",
		constants.IS_AUTO_RENEWAL:         true,
		constants.RENEW_BEFORE:            "30",
		constants.AUTO_REGENERATE_ENABLED: false,
	})

	payload := frameCertificatePayload(d)

	if payload.CaConnectorInfo.IsAutoRenewal != "true" {
		t.Fatalf("expected isAutoRenewal \"true\", got %q", payload.CaConnectorInfo.IsAutoRenewal)
	}
	if payload.CaConnectorInfo.RenewBefore != "30" {
		t.Fatalf("expected renewBefore \"30\", got %q", payload.CaConnectorInfo.RenewBefore)
	}
	if payload.CaConnectorInfo.AutoRegenerateEnabled == nil || *payload.CaConnectorInfo.AutoRegenerateEnabled != false {
		t.Fatalf("expected autoRegenerateEnabled pointer to false, got %v", payload.CaConnectorInfo.AutoRegenerateEnabled)
	}
}

func TestFrameCertificatePayload_AutoRenewalDisabledByDefault(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceCertificateServer().Schema, map[string]interface{}{
		constants.COMMON_NAME: "dev.api.example.com",
	})

	payload := frameCertificatePayload(d)

	if payload.CaConnectorInfo.IsAutoRenewal != "" {
		t.Fatalf("expected empty isAutoRenewal when not enabled, got %q", payload.CaConnectorInfo.IsAutoRenewal)
	}
	if payload.CaConnectorInfo.AutoRegenerateEnabled != nil {
		t.Fatalf("expected nil autoRegenerateEnabled when not enabled, got %v", *payload.CaConnectorInfo.AutoRegenerateEnabled)
	}
}

func TestBuildRevokePayload_WithComments(t *testing.T) {
	got := buildRevokePayload("6a47c392e9692037365e62e1", "Cessation of operation", "decommissioned")
	if got["resourceId"] != "6a47c392e9692037365e62e1" {
		t.Fatalf("unexpected resourceId: %v", got["resourceId"])
	}
	if got["reason"] != "Cessation of operation" {
		t.Fatalf("unexpected reason: %v", got["reason"])
	}
	if got["comments"] != "decommissioned" {
		t.Fatalf("expected comments to be set, got: %v", got["comments"])
	}
}

func TestBuildRevokePayload_WithoutComments(t *testing.T) {
	got := buildRevokePayload("abc", "Superseded", "")
	if _, ok := got["comments"]; ok {
		t.Fatalf("expected comments to be omitted when empty, got: %v", got["comments"])
	}
}
