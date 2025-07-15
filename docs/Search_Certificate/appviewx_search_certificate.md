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
  Category of certificate (e.g., `Server, Client, CodeSigning`).

- **`cert_serial_no`** (string):  
  Certificate serial number to search for.

- **`cert_issuer`** (string):  
  Certificate issuer to search for.

### Optional Attributes

- **`max_results`** (int):  
  Maximum number of results to return (default: 10).

- **`start_index`** (int):  
  Start index for pagination (default: 1).

- **`sort_column`** (string):  
  Column to sort results by (default: `commonName`).

- **`sort_order`** (string):  
  Sort order, either `asc` or `desc` (default: `desc`).

### Computed Attributes

- **`total_records`** (int):  
  Total number of records found for the search criteria.

## Example Usage

```hcl
resource "appviewx_search_certificate_by_keyword" "search_cert" {
  category        = "Server"
  cert_serial_no  = "<Serial Number>"
  cert_issuer     = "<Issuer Common Name>"
}
```

## Import

To import an existing search into the Terraform state, use:

```bash
terraform import appviewx_search_certificate_by_keyword.search_cert <search_id>
```
Replace `<search_id>` with the unique identifier for your search (typically based on the category).

---