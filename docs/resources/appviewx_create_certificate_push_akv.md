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

### Optional Attributes

- **`workflow_name`** (string):  
  The custom workflow name to execute the Create Certificate and Push to AKV Operation.

### Mandatory parameters

- **`assign_group`** (string): The name of the group to which the certificate belongs in AppViewX.

- **`azure_account`** (string): The name of the AKV Device which was onboarded in AppViewX.

- **`azure_key_vault`** (string): The name of the AKV Key Vault which was onboarded in AppViewX.

- **`logged_in_username`** (string): The name of the user used to login in the AppViewX.

- **`cert_type`** (string): Describes the Certificate category. Possible Values: [`Server`, `Client`, `CodeSigning`]

- **`ca`** (string): The name of the Certificate Authority (CA) to issue the certificate. Possible Values: [`AppViewX`, `Sectigo`, `OpenTrust`, `Microsoft Enterprise`, `DigiCert`]

- **`validity_unit`** (string): The unit of validity for the certificate. Possible values are [`Days`, `Months`, `Years`].

- **`validity_unit_value`** (string): The value for the validity unit

- **`cn_uploadcsr`** (string): The domain name or identifier for the certificate.

- **`hash_uploadcsr`** (string): Describes the Hashing algorithm. Possible Values are [`SHA160`, `SHA224`, `SHA256`, `SHA384`, `SHA512`, `SHA3-224`, `SHA3-256`]

- **`key_uploadcsr`** (string): The cryptographic algorithm for the key. Possible values are [`RSA`, `DSA`, `EC`]

- **`bit_uploadcsr`** (string): The size of the key in bits. Possible values are [`1024`, `2048`, `3072`, `4096`, `7680`, `8192`].

- **`entrust_cert_type`** (string): The Certificate Type that need to be enroll via Entrust.
  NOTE:
  - Mandatory Only for Entrust CA.
  - For other CA's - mention as NA.

- **`sectigo_cert_type`** (string): The Certificate Type that need to be enroll via Sectigo.
  NOTE:
  - Mandatory Only for Sectigo CA.
  - For other CA's - mention as NA.

- **`template_name`** (string): The Certificate Template name that need to be enroll via Microsoft Enterprise CA.
  NOTE:
  - Mandatory Only for Microsoft Enterprise CA.
  - For other CA's - mention as NA.

- **`digicert_division`** (string): The Digicert's Divison of the Certificate that need to be enroll via Sectigo.
  NOTE:
  - Mandatory Only for DigiCert CA.
  - For other CA's - mention as NA.

- **`digicert_cert_type`** (string): The Certificate Type that need to be enroll via Digicert.
  NOTE:
  - Mandatory Only for DigiCert CA.
  - For other CA's - mention as NA.

- **`digicert_server_type`** (string): The Server Type of the Certificate that need to be enroll via Digicert.
  NOTE:
  - Mandatory Only for DigiCert CA.
  - For other CA's - mention as NA.

- **`digicert_payment_method`** (string): The Payment method of the Certificate that need to be enroll via Digicert.
  NOTE:
  - Mandatory Only for DigiCert CA.
  - For other CA's - mention as NA.


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
    "assign_group": "Default",
    "azure_account": "AKV",
    "azure_key_vault": "KeyVault-AVX",
    "logged_in_username": "avx@gmail.com",
    "cert_type": "Server",
    "ca": "AppViewX",
    "entrust_cert_type": "NA",
    "sectigo_cert_type": "NA",
    "template_name":"NA",
    "digicert_division": "NA",
    "digicert_cert_type": "NA",
    "digicert_server_type": "NA",
    "digicert_payment_method": "NA",
    "validity_unit": "Days",
    "validity_unit_value": "4",
    "cn_uploadcsr": "appviewxCertificate.certplus.in",
    "dns_uploadcsr": "",
    "org_uploadcsr": "",
    "org_address": "",
    "locality": "",
    "org_unit": "",
    "state": "",
    "country": "",
    "email_address": "",
    "challenge_pwd": "",
    "confirm_pwd": "",
    "challenge_pwd_uploadcsr": "",
    "confirm_pwd_uploadcsr": "",
    "hash_uploadcsr": "SHA256",
    "key_uploadcsr": "RSA",
    "bit_uploadcsr": "2048",
    "end_entity_username": "",
    "prevalidation": "",
    "isapiuser": "yes",
    "D_Resp_exploitation-adresse": "",
    "D_Demandeur-adresse": "",
    "D_Demandeur": "",
    "D_Serveur-nom": "",
    "D_Serveur-IP": "",
    "D_No_Projet": "",
    "D_Commentaires": "",
    "D_Casewise-Bizzdesign": "",
    "D_VPTI_proprietaire": "",
    "D_Contact_tech-adresse": "",
    "D_Environnement": "",
    "D_CT_Logs": "",
    "D_infonuagique": "",
    "D_En_Utilisation": "",
    "D_Nom_Proprietaire_TI": "",
    "D_Site_Externe": "",
    "D_Notes_client": "",
    "D_Numero_derogation": "",
    "D_Localite": "Toronto"
  })

  resource "appviewx_create_push_certificate_request_status" "create_and_push_certificate_status" {
  request_id = appviewx_certificate_push_akv.create_and_push_certificate.workflow_id
  retry_count = 20
  retry_interval = 10
  certificate_common_name = appviewx_certificate_push_akv.create_and_push_certificate.certificate_common_name
  certificate_download_path = "</path/to/directory or /path/to/directory/filename>"
  certificate_download_format = "CRT"
  certificate_chain_required = true
  is_download_required = true
  depends_on = [appviewx_certificate_push_akv.create_and_push_certificate]
}
}
```


### Certificate Creation with DigiCert CA

```hcl
resource "appviewx_certificate_push_akv" "create_and_push_certificate" {
  field_info = jsonencode({
    "assign_group": "Default",
    "azure_account": "AKV",
    "azure_key_vault": "KeyVault-AVX",
    "logged_in_username": "avx@gmail.com",
    "cert_type": "Server",
    "ca": "DigiCert",
    "entrust_cert_type": "NA",
    "sectigo_cert_type": "NA",
    "template_name": "NA",
    "digicert_division": "AppViewX",
    "digicert_cert_type": "SSL",
    "digicert_server_type": "Server1",
    "digicert_payment_method": "balance",
    "validity_unit": "Days",
    "validity_unit_value": "365",
    "cn_uploadcsr": "digicertCertificate.certplus.in",
    "dns_uploadcsr": "",
    "org_uploadcsr": "",
    "org_address": "",
    "locality": "",
    "org_unit": "",
    "state": "",
    "country": "",
    "email_address": "",
    "challenge_pwd": "",
    "confirm_pwd": "",
    "challenge_pwd_uploadcsr": "",
    "confirm_pwd_uploadcsr": "",
    "hash_uploadcsr": "SHA256",
    "key_uploadcsr": "RSA",
    "bit_uploadcsr": "2048",
    "end_entity_username": "",
    "prevalidation": "",
    "isapiuser": "yes",
    "D_Resp_exploitation-adresse": "",
    "D_Demandeur-adresse": "",
    "D_Demandeur": "",
    "D_Serveur-nom": "",
    "D_Serveur-IP": "",
    "D_No_Projet": "",
    "D_Commentaires": "",
    "D_Casewise-Bizzdesign": "",
    "D_VPTI_proprietaire": "",
    "D_Contact_tech-adresse": "",
    "D_Environnement": "",
    "D_CT_Logs": "",
    "D_infonuagique": "",
    "D_En_Utilisation": "",
    "D_Nom_Proprietaire_TI": "",
    "D_Site_Externe": "",
    "D_Notes_client": "",
    "D_Numero_derogation": "",
    "D_Localite": "Montreal"
  })
}
```

### Certificate Creation with Microsoft Enterprise CA

```hcl
resource "appviewx_certificate_push_akv" "create_and_push_certificate" {
  field_info = jsonencode({
    "assign_group": "Default",
    "azure_account": "AKV",
    "azure_key_vault": "KeyVault-AVX",
    "logged_in_username": "avx@gmail.com",
    "cert_type": "Server",
    "ca": "Microsoft Enterprise",
    "entrust_cert_type": "NA",
    "sectigo_cert_type": "NA",
    "template_name": "Server",
    "digicert_division": "NA",
    "digicert_cert_type": "NA",
    "digicert_server_type": "NA",
    "digicert_payment_method": "NA",
    "validity_unit": "Years",
    "validity_unit_value": "1",
    "cn_uploadcsr": "MicrosoftCertificate.certplus.in",
    "dns_uploadcsr": "",
    "org_uploadcsr": "",
    "org_address": "",
    "locality": "",
    "org_unit": "",
    "state": "",
    "country": "",
    "email_address": "",
    "challenge_pwd": "",
    "confirm_pwd": "",
    "challenge_pwd_uploadcsr": "",
    "confirm_pwd_uploadcsr": "",
    "hash_uploadcsr": "SHA256",
    "key_uploadcsr": "RSA",
    "bit_uploadcsr": "2048",
    "end_entity_username": "",
    "prevalidation": "",
    "isapiuser": "yes",
    "D_Resp_exploitation-adresse": "",
    "D_Demandeur-adresse": "",
    "D_Demandeur": "",
    "D_Serveur-nom": "",
    "D_Serveur-IP": "",
    "D_No_Projet": "",
    "D_Commentaires": "",
    "D_Casewise-Bizzdesign": "",
    "D_VPTI_proprietaire": "",
    "D_Contact_tech-adresse": "",
    "D_Environnement": "",
    "D_CT_Logs": "",
    "D_infonuagique": "",
    "D_En_Utilisation": "",
    "D_Nom_Proprietaire_TI": "",
    "D_Site_Externe": "",
    "D_Notes_client": "",
    "D_Numero_derogation": "",
    "D_Localite": "Toronto"
}
)
}
```

### Certificate Creation with Sectigo CA

```hcl
resource "appviewx_certificate_push_akv" "create_and_push_certificate" {
  field_info = jsonencode({
    "assign_group": "Default",
    "azure_account": "AKV",
    "azure_key_vault": "KeyVault-AVX",
    "logged_in_username": "avx@gmail.com",
    "cert_type": "Server",
    "ca": "Sectigo",
    "entrust_cert_type": "NA",
    "sectigo_cert_type": "SSL Certificate",
    "template_name": "NA",
    "digicert_division": "NA",
    "digicert_cert_type": "NA",
    "digicert_server_type": "NA",
    "digicert_payment_method": "NA",
    "validity_unit": "Years",
    "validity_unit_value": "1",
    "cn_uploadcsr": "SectigoCertificate.certplus.in",
    "dns_uploadcsr": "",
    "org_uploadcsr": "",
    "org_address": "",
    "locality": "",
    "org_unit": "",
    "state": "",
    "country": "",
    "email_address": "",
    "challenge_pwd": "",
    "confirm_pwd": "",
    "challenge_pwd_uploadcsr": "",
    "confirm_pwd_uploadcsr": "",
    "hash_uploadcsr": "SHA256",
    "key_uploadcsr": "RSA",
    "bit_uploadcsr": "2048",
    "end_entity_username": "",
    "prevalidation": "",
    "isapiuser": "yes",
    "D_Resp_exploitation-adresse": "",
    "D_Demandeur-adresse": "",
    "D_Demandeur": "",
    "D_Serveur-nom": "",
    "D_Serveur-IP": "",
    "D_No_Projet": "",
    "D_Commentaires": "",
    "D_Casewise-Bizzdesign": "",
    "D_VPTI_proprietaire": "",
    "D_Contact_tech-adresse": "",
    "D_Environnement": "",
    "D_CT_Logs": "",
    "D_infonuagique": "",
    "D_En_Utilisation": "",
    "D_Nom_Proprietaire_TI": "",
    "D_Site_Externe": "",
    "D_Notes_client": "",
    "D_Numero_derogation": "",
    "D_Localite": "Toronto"
}
)
}
```

## Import

To import an existing workflow request into the Terraform state, use:

```bash
terraform import appviewx_certificate_push_akv.create_and_push_certificate <workflow_id>
```
Replace `<workflow_id>` with the actual workflow request ID.

---

## Destroy

To destroy the Certificate details in the Terraform State file, use:

```bash
terraform destroy
```

- This is mainly to ensure that certificates (or any cryptographic material) are not stored in the Terraform state file.
- This feature is crucial for maintaining the security and confidentiality of sensitive cryptographic materials.

---