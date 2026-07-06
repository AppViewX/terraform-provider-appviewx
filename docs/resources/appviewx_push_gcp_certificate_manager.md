# appviewx_push_gcp_certificate_manager

Pushes an AppViewX-issued certificate into **GCP Certificate Manager** (cert-manager-only; no load balancer binding). Exposes the GCP certificate id so a load balancer can reference it.

## Example

```hcl
resource "appviewx_push_gcp_certificate_manager" "this" {
  certificate_id      = appviewx_create_certificate.this.resource_id
  certificate_name    = "apigee-ingress-usc1-cert"
  project             = var.project_id
  location            = var.region                # region (e.g. us-central1) or "global"
  gcp_connector_name  = "GCP connector"
  selected_profiles   = ["GCP.Single.All:${var.project_id}:Certificate Manager"]
  wait_for_completion = true
}

resource "google_compute_region_target_https_proxy" "internal" {
  # ...
  certificate_manager_certificates = appviewx_push_gcp_certificate_manager.this.certificate_manager_certificate_id
}
```

## Argument Reference

| Name | Required | Default | Description |
|---|---|---|---|
| `certificate_id` | yes | | AppViewX certificate resource id (`resource_id` from `appviewx_create_certificate`). Changing forces recreation. |
| `certificate_name` | yes | | Certificate name in GCP Certificate Manager. Changing forces recreation. |
| `project` | yes | | GCP project id (used to construct the exposed id). Changing forces recreation. |
| `location` | yes | | GCP region (e.g. `us-central1`) or `global`. Changing forces recreation. |
| `gcp_connector_name` | yes | | AppViewX GCP connector name. Changing forces recreation. |
| `profile_type` | no | `Push and Bind Profiles` | AppViewX `profileType`. Use the push-only variant your AppViewX accepts to push to Certificate Manager without binding to a load balancer. The exact string is environment-specific. |
| `selected_profiles` | yes | | List of AppViewX profiles (the valid values are configured in your AppViewX; format `<profile>:<project>:Certificate Manager`). Must be non-empty. Changing forces recreation. |
| `is_new_certificate` | no | `true` | Whether this is a new certificate in Certificate Manager. Changing forces recreation. |
| `push_automatically` | no | `true` | Re-push to the target automatically after the certificate is renewed/reissued. |
| `wait_for_completion` | no | `true` | Block until the push request completes (correct dependency ordering). |
| `wait_timeout_seconds` | no | `600` | Max seconds to wait when `wait_for_completion = true`. Must be >= 1. |
| `poll_interval_seconds` | no | `10` | Seconds between status polls. Must be >= 1. |

## Attribute Reference

| Name | Description |
|---|---|
| `certificate_manager_certificate_id` | `projects/{project}/locations/{location}/certificates/{certificate_name}` — reference this from `certificate_manager_certificates`. |
| `request_id` | AppViewX push request id. |
| `connector_id` | AppViewX connector id from the push response. |
| `status_code` | HTTP status code of the push request. |
| `success` | Whether the push (and wait, if enabled) succeeded. |

## Notes

- **Delete** removes the resource from Terraform state only; it does **not** delete the certificate from GCP Certificate Manager or AppViewX. The certificate lifecycle (including revocation) belongs to `appviewx_create_certificate`.
