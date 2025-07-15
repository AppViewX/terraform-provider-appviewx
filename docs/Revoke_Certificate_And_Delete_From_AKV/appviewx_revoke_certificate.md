# Certificate Revocation

The `appviewx_revoke_certificate` resource allows you to revoke an existing certificate in AppViewX and Delete the particular Certificate from Azure Key vault (AKV) by specifying its serial number and issuer common name.

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
   - The resource captures the HTTP status code, request ID, response message, and whether the revocation was successful.

6. **State Management**:
   - The resource is create-only. Updates and deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`serial_number`** (string):  
  Serial number of the certificate to revoke.

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

### Computed Attributes

- **`status_code`** (int):  
  HTTP status code of the revocation request.

- **`resource_id`** (string):  
  Resource ID of the revoked certificate.

- **`request_id`** (string):  
  Request ID of the revocation request.

- **`response_message`** (string):  
  Response message from the server.

- **`revocation_success`** (bool):  
  Whether the revocation was successful.

## Example Usage

```hcl
resource "appviewx_revoke_certificate" "revoke_cert" {
  serial_number      = "<Certificate Serial Number>"
  issuer_common_name = "<Issuer Common Name>"
  reason             = "Key compromise"
  comments           = "Revoked due to key compromise"
}
```

## Import

To import an existing revocation request into the Terraform state, use:

```bash
terraform import appviewx_revoke_certificate.revoke_cert <request_id>
```
Replace `<request_id>` with the actual request ID of the revocation.

---