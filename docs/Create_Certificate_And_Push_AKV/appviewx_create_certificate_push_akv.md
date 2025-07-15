# Certificate Creation and Push to Azure Key Vault

The `appviewx_certificate_push_akv` resource automates the creation of a certificate and its push to Azure Key Vault (AKV) using AppViewX workflows.

## Process Overview

1. **Input Parameters**:
   - The resource accepts a single required parameter, `field_info`, which is a JSON string containing all certificate and key vault configuration details. This includes certificate subject details, key parameters, CA settings, and Azure Key Vault information.

2. **Workflow Execution**:
   - The resource triggers a pre-configured AppViewX workflow (default: `Desjardins Create Certificate Push to AKV`) to create and push the certificate to AKV.

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
  Example:
  ```hcl
  provider "appviewx" {
  appviewx_environment_ip = "gs-dev-apvx-n36.lab.appviewx.net"
  appviewx_environment_port = "31443"
  appviewx_environment_is_https = true
}

resource "appviewx_automation" "certificate_creation_and_push_to_akv"{
 payload= <<EOF
 {
  "payload" : {
    "data" : {
      "input" : {
        "requestData" : [ {
          "sequenceNo" : 1,
          "scenario" : "scenario",
          "fieldInfo" : {
            "assign_group" : "<mandatory field>",
            "device_name" : "<mandatory field>",
            "key_vault" : "<mandatory field>",
            "logged_in_username" : "<%loggedInUsername%>",
            "cert_type" : "Server,Client,CodeSigning",
            "ca" : "AppViewX,Microsoft Enterprise,OpenTrust,Sectigo,DigiCert",
            "entrust_cert_type" : "<mandatory field>",
            "sectigo_cert_type" : "<mandatory field>",
            "template_name" : "<mandatory field>",
            "digicert_division" : "<mandatory field>",
            "digicert_cert_type" : "<mandatory field>",
            "digicert_server_type" : "<mandatory field>",
            "digicert_payment_method" : "<mandatory field>",
            "validity_unit" : "Days,Months,Years",
            "validity_unit_value" : "<mandatory field>",
            "cn_uploadcsr" : "<mandatory field>",
            "dns_uploadcsr" : "",
            "org_uploadcsr" : "",
            "org_address" : "",
            "locality" : "",
            "org_unit" : "",
            "state" : "",
            "country" : "",
            "email_address" : "",
            "challenge_pwd" : "",
            "confirm_pwd" : "",
            "challenge_pwd_uploadcsr" : "",
            "confirm_pwd_uploadcsr" : "",
            "hash_uploadcsr" : "<mandatory field>",
            "key_uploadcsr" : "<mandatory field>",
            "bit_uploadcsr" : "<mandatory field>",
            "end_entity_username" : "",
            "prevalidation" : "",
            "isapiuser" : "yes",
            "D_Resp_exploitation-adresse" : "",
            "D_Demandeur-adresse" : "",
            "D_Demandeur" : "",
            "D_Serveur-nom" : "",
            "D_Serveur-IP" : "",
            "D_No_Projet" : "",
            "D_Commentaires" : "",
            "D_Casewise-Bizzdesign" : "",
            "D_VPTI_proprietaire" : "",
            "D_Contact_tech-adresse" : "",
            "D_Environnement" : "",
            "D_CT_Logs" : "",
            "D_infonuagique" : "",
            "D_En_Utilisation" : "",
            "D_Nom_Proprietaire_TI" : "",
            "D_Site_Externe" : "",
            "D_Notes_client" : "",
            "D_Numero_derogation" : "",
            "D_Localite" : "Toronto,Montreal"
          }
        } ]
      },
      "task_action" : 1
    },
    "header" : {
      "workflowName" : "Desjardins Create Certificate Push to AKV"
    }
  }
}
EOF
action_id= "visualworkflow-submit-request"

  }
    ...
  })
  ```
  > See your workflow requirements for the full list of supported fields.

### Optional Attributes

- **`workflow_name`** (string):  
  The workflow name to execute. Defaults to `Desjardins Create Certificate Push to AKV`.

### Computed Attributes

- **`status_code`** (int):  
  HTTP status code from the workflow submission response.

- **`workflow_id`** (string):  
  Workflow request ID if successful. Use this to poll for status.

- **`success`** (bool):  
  Whether the workflow request was successfully submitted.

- **`certificate_common_name`** (string):  
  The common name of the certificate being pushed (extracted from `field_info`).

## Example Usage

```hcl
provider "appviewx" {
  appviewx_environment_ip = "gs-dev-apvx-n36.lab.appviewx.net"
  appviewx_environment_port = "31443"
  appviewx_environment_is_https = true
  log_level= "DEBUG"
}

resource "appviewx_certificate_push_akv" "create_and_push_certificate" {
  field_info = jsonencode({
    "assign_group": "Default",
    "device_name": "AKV1",
    "key_vault": "KeyVault-AVX",
    "logged_in_username": "savx@appviewx.com",
    "cert_type": "Server",
    "ca": "Microsoft Enterprise",
    "template_name": "WebServer",
    "validity_unit": "Years",
    "validity_unit_value": "1",
    "cn_uploadcsr": "Microsoft.certplus.in",
    "hash_uploadcsr": "SHA256",
    "key_uploadcsr": "RSA",
    "bit_uploadcsr": "2048",
    ...
  })
}
```

## Import

To import an existing workflow request into the Terraform state, use:

```bash
terraform import appviewx_certificate_push_akv.create_and_push_certificate <workflow_id>
```
Replace `<workflow_id>` with the actual workflow request ID.

---
