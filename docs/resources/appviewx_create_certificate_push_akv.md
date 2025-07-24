# Certificate Creation and Push to Azure Key Vault

The `appviewx_certificate_push_akv` resource automates the creation of a certificate and its push to Azure Key Vault (AKV) using AppViewX workflows.

## Prerequisites

- **`Necessary permissions to delete the Certificate and the associated Key in Azure Key Vault`**
- **`Azure Key Vault (AKV) need to be onboarded in AppViewX`**
- **`This Terraform version(tf) can be used only when there is a custom workflow enabled for pushing certs to AKV`**

## Process Overview

1. **Input Parameters**:
   - The resource accepts a single required parameter, `field_info`, which is a JSON string containing all certificate and key vault configuration details. This includes certificate subject details, key parameters, CA settings, and Azure Key Vault information.

2. **Workflow Execution**:
   - The resource triggers a pre-configured AppViewX Custom workflow to create and push the certificate to AKV.

3. **Authentication**:
   - Authentication to AppViewX can be performed using either username/password or client ID/secret, provided via provider configuration or environment variables.

4. **Response Handling**:
   - The resource captures the workflow request ID, HTTP status code, and whether the request was successful. The workflow ID can be used to poll for status and download the certificate using the `appviewx_create_push_certificate_request_status` resource.

5. **State Management**:
   - The resource is create-only. Updates and deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`field_info`** (string, sensitive):  
  JSON string containing all certificate and key vault configuration.  

- **`workflow_name`** (string):  
  The custom workflow name to execute the Create Certificate and Push to AKV Operation.

### NOTE:
- These mandatory and optional attributes might differ based on the custom workflow used in AppViewX.

### Mandatory parameters

- **`certificate_group_name`** (string): The name of the group to which the certificate belongs in AppViewX.

- **`azure_account_name`** (string): The name of the AKV Device which was onboarded in AppViewX.

- **`azure_key_vault_name`** (string): The name of the AKV Key Vault which was onboarded in AppViewX.

- **`certificate_type`** (string): Describes the Certificate category. Possible Values: [`Server`, `Client`, `CodeSigning`]

- **`certificate_authority`** (string): The name of the Certificate Authority (CA) to issue the certificate. Possible Values: [`AppViewX`, `Sectigo`, `OpenTrust`, `Microsoft Enterprise`, `DigiCert`]

- **`validity_unit`** (string): The unit of validity for the certificate. Possible values are [`Days`, `Months`, `Years`].

- **`validity_unit_value`** (string): The value for the validity unit

- **`common_name`** (string): The domain name or identifier for the certificate.

- **`hash_algorithm`** (string): Describes the Hashing algorithm. Possible Values are [`SHA160`, `SHA224`, `SHA256`, `SHA384`, `SHA512`, `SHA3-224`, `SHA3-256`]

- **`key_type`** (string): The cryptographic algorithm for the key. Possible values are [`RSA`, `DSA`, `EC`]

- **`key_bit_length`** (string): The size of the key in bits. Possible values are 
  - RSA : [`1024`, `2048`, `3072`, `4096`, `7680`, `8192`].
  - DSA : [`1024`, `2048`].
  - EC : [`160`, `163`, `191`, `192`, `193`, `224`, `233`, `239`, `256`, `283`, `320`, `359`, `384`, `409`, `431`, `512`, `521`, `571`]
  - ECDSA Curve : [`ECDSA Curve that appviewx is supporting`]

## Example Usage

### Certificate Creation with AppViewX CA

```hcl
provider "appviewx" {
  appviewx_environment_ip = "<AppViewX - FQDN or IP>"
  appviewx_environment_port = "<Port>"
  appviewx_environment_is_https = true
}

resource "appviewx_certificate_push_akv" "create_and_push_certificate" {
  field_info = jsonencode({
    "certificate_group_name": "Group1",
    "azure_account_name": "AKV",
    "azure_key_vault_name": "KeyVault",
    "certificate_type": "Server",
    "certificate_authority": "AppViewX Certificate Authority",
    "validity_unit": "Days",
    "validity_unit_value": "4",
    "common_name": "appviewxCertificate.xxxxx.yy",
    "hash_algorithm": "SHA256",
    "key_type": "RSA",
    "key_bit_length": "2048"
  })
  workflow_name = "Create Cert Workflow"

  resource "appviewx_create_push_certificate_request_status" "create_and_push_certificate_status" {
  request_id = appviewx_certificate_push_akv.create_and_push_certificate.workflow_id
  retry_count = 20
  retry_interval = 20
  depends_on = [appviewx_certificate_push_akv.create_and_push_certificate]
}
}
```

## Response of the Resource

Response of the appviewx_certificate_push_akv resource

```bash
{
  "response": {
    "workorderId": "0",
    "requestType": "sample",
    "requestId": "2642",
    "workflowVersion": "version1",
    "message": "Workflow Request is created with Id 2642 . Request submitted to workflow engine for processing workorder.",
    "status": "In Progress",
    "statusCode": 0
  },
  "message": "Success",
  "appStatusCode": null,
  "tags": null,
  "headers": {
    "X-WorkFlowName": "Create Certificate Push to AKV"
  }

```

Final Response of this Request after pooling the Status of the Certificate Creation and pushing to AKV process

```bash
[CERTIFICATE CREATION][SUCCESS] ✅ Operation Result:
{
  "completed_at": "<Timestamp>",
  "operation": "Certificate Creation and Push",
  "status": "Successful",
  "status_code": 1,
  "workflow_id": "2645"
}
```
## Destroy

To destroy the Certificate details in the Terraform State file, use:

```bash
terraform destroy
```

- This is mainly to ensure that certificates (or any cryptographic material) are not stored in the Terraform state file.
- This feature is crucial for maintaining the security and confidentiality of sensitive cryptographic materials.

---