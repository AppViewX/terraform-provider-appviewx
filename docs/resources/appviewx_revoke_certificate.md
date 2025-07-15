# Certificate Revocation

The `appviewx_revoke_certificate` resource allows you to revoke an existing certificate in AppViewX and Delete the particular Certificate from Azure Key vault (AKV) by specifying its serial number and issuer common name.


## Prerequisites

- **`Necessary permissions to delete the Certificate and the associated Key in Azure Key Vault`**
- **`Azure Key Vault (AKV) need to be onboarded in AppViewX`**
- **`This Terraform version(tf) can be used only when there is a custom workflow enabled for pushing certs to AKV`**

## Process Overview

1. **Input Parameters**:
   - The resource requires the certificate's serial number, issuer common name, and a revocation reason. Optionally, you can provide comments for the revocation.

2. **Authentication**:
   - Authentication to AppViewX can be performed using either username/password or client ID/secret, provided via provider configuration or environment variables.

3. **Resource Lookup**:
   - The resource first looks up the certificate's resource ID using the provided serial number and issuer common name.

4. **Revocation Request**:
   - The certificate is revoked by sending a request to the AppViewX API with the resource ID and revocation reason.

5. **Delete Certificate Request**:
   - The Certificate is deleted by communicate with the AKV Device with the Certificate file named.

5. **Response Handling**:
   - The resource captures the HTTP status code, request ID, response message, and whether the revocation was successful. The workflow ID can be used to poll for status and displays the Result using the `appviewx_revoke_certificate_request_status` resource.

6. **State Management**:
   - The resource is create-only. Updates and deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`serial_number`** (string):  
  Serial number of the certificate to revoke. (e.g., `D1:CF:81:B0:43:8E:B3:D7:F6:CE:16:58:0B:82:E5:4F`)

- **`issuer_common_name`** (string):  
  Issuer common name of the certificate to revoke.

- **`reason`** (string):  
  Reason for certificate revocation. Allowed values:
  - Unspecified
  - Key compromise
  - CA compromise
  - Affiliation Changed
  - Superseded
  - Cessation of operation

### Optional Attributes

- **`comments`** (string):  
  Additional comments for revocation.

## Example Usage

```hcl
resource "appviewx_revoke_certificate" "revoke_cert" {
  serial_number      = "<Certificate Serial Number>"
  issuer_common_name = "<Issuer Common Name>"
  reason             = "Key compromise"
  comments           = "Revoked due to key compromise"
}
```

## RevokeCertificate.tf File

```hcl
provider "appviewx" {
  appviewx_environment_ip = "<AppViewX - FQDN or IP>"
  appviewx_environment_port = "<Port>"
  appviewx_environment_is_https = true
}

resource "appviewx_revoke_certificate" "cert_revoke" {
  serial_number = "74:00:00:0A:BD:5E:FD:73:A6:A3:8D:C7:A6:00:00:00:00:0A:BD"
  issuer_common_name = "AppViewX Intermediate CA"
  reason = "Superseded"
  comments = "Certificate replaced"
}

resource "appviewx_revoke_certificate_request_status" "revoke_cert_status" {
  request_id = appviewx_revoke_certificate.cert_revoke.request_id
  retry_count = 30
  retry_interval = 10
}
```

## Import

To import an existing revocation request into the Terraform state, use:

```bash
terraform import appviewx_revoke_certificate.revoke_cert <request_id>
```
Replace `<request_id>` with the actual request ID of the revocation.

---