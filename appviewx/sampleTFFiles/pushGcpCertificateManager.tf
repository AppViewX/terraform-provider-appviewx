terraform {
  required_providers {
    appviewx = {
      source = "appviewx.com/provider/appviewx"
    }
  }
}

provider "appviewx" {
  appviewx_environment_ip       = var.appviewx_ip
  appviewx_environment_port     = var.appviewx_port
  appviewx_environment_is_https = true
  # Auth via APPVIEWX_TERRAFORM_CLIENT_ID / APPVIEWX_TERRAFORM_CLIENT_SECRET env vars
}

resource "appviewx_create_certificate" "this" {
  common_name            = "dev.api.example.com"
  dns_names              = ["dev.api.example.com", "usc1.dev.api.example.com"]
  hash_function          = "SHA384"
  key_type               = "RSA"
  bit_length             = "4096"
  certificate_authority  = "Sectigo"
  ca_setting_name        = "Example CA Setting"
  certificate_type       = "External Certificates - customer facing or internet facing (public root)"
  certificate_group_name = "Application-Load-Balancers-GCP"
  validity_days          = 7
  validity_unit          = "days"
  validity_unit_value    = 7

  is_auto_renewal = true
  renew_before    = "30"

  # Block until the cert is issued so the push below runs against a ready cert
  # (create is async; without this the push can race ahead of issuance).
  wait_for_issuance = true

  revoke_on_destroy = true
  revoke_reason     = "Cessation of operation"
}

resource "appviewx_push_gcp_certificate_manager" "this" {
  certificate_id      = appviewx_create_certificate.this.resource_id
  certificate_name    = "apigee-ingress-usc1-cert"
  project             = var.project_id
  location            = var.region
  gcp_connector_name  = "GCP connector"
  selected_profiles   = ["GCP.Single.All:${var.project_id}:Certificate Manager"]
  wait_for_completion = true

  # profile_type defaults to "Push only Profiles" (push to Certificate Manager
  # without binding to a load balancer). Set to "Push and Bind Profiles" to bind.
  # profile_type = "Push only Profiles"
}

output "certificate_manager_certificate_id" {
  value = appviewx_push_gcp_certificate_manager.this.certificate_manager_certificate_id
}
