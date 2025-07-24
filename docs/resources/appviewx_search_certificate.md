# Certificate Search

The `appviewx_search_certificate` resource allows you to search for certificates in AppViewX using Serial Number and Issuer Common Name and retrieve metadata about matching certificates.

## Process Overview

1. **Input Parameters**:
   - The resource accepts search filters such as category, serial number, issuer and common name. You can also control pagination and sorting.

2. **Authentication**:
   - Authentication to AppViewX can be performed using either username/password or client ID/secret, provided via provider configuration or environment variables.

3. **Search Execution**:
   - The resource sends a search request to the AppViewX API with the provided filters and retrieves a list of matching certificates.

4. **Response Handling**:
   - The resource captures the total number of records found. Certificate details are not stored in the Terraform state for security and privacy.

5. **State Management**:
   - The resource is create-only. Updates trigger a new search. Deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`category`** (string):  
  Category of certificate. Allowed values:
  - Server
  - Client
  - CodeSigning

- **`cert_serial_no`** (string):  
  Certificate serial number to search for. (e.g., `A1:CF:81:B0:23:8E:B3:D7:H2:CE:16:58:0B:82:G5:4F`)

- **`cert_issuer`** (string):  
  Certificate Issuer Common Name to search for.


## Example Usage

```hcl
resource "appviewx_search_certificate_by_keyword" "search_cert" {
  category        = "Server"
  cert_serial_no  = "<Certificate Serial Number>"
  cert_issuer     = "<Issuer Common Name>"
}
```

## SearchCertificate.tf File

```hcl
provider "appviewx" {
  appviewx_environment_ip = "<AppViewX - FQDN or IP>"
  appviewx_environment_port = "<Port>"
  appviewx_environment_is_https = true
}

resource "appviewx_search_certificate" "cert_search" {
  category = "<Category>"
  cert_serial_no = "<Certificate Serial Number>"
  cert_issuer = "<Issuer Common Name>"
}
```
## Response
```bash
  You will get the Certificate Meta data and parameters as a JSON.
```

## Destroy

To destroy the Certificate details in the Terraform State file, use:

```bash
terraform destroy
```

- This is mainly to ensure that certificates (or any cryptographic material) are not stored in the Terraform state file.
- This feature is crucial for maintaining the security and confidentiality of sensitive cryptographic materials.

---