# AppViewX Terraform Provider

[AppViewX](https://appviewx.com) protects many of the world’s brands with the industry’s most advanced cloud-native Certificate Lifecycle Management (CLM) and Public Key Infrastructure (PKI) platform. Our solutions safeguard customers and enable digital transformation in the largest and most security-conscious enterprise organizations globally.

**AVX ONE** is the industry’s most advanced and fastest growing cloud-native Certificate Lifecycle Management (CLM) platform. It provides a suite of market leading capabilities including Smart Discovery, Crypto Resilience Scorecard, Closed-looped Automation and Infrastructure Context Awareness.

Powered by the market’s only out-of-the-box workflow engine, AVX ONE allows customers to realize immediate value from complete certificate lifecycle management, enterprise-wide Kubernetes TLS automation, scalable PKI-as-a-Service, secure code signing, easy Microsoft CA migration, IoT security, SSH and key management, and PQC-forward controls in even the most complex multi-cloud, hybrid, and edge environments.

Seamlessly enforce enterprise policies and strict access controls, ensure cryptographic agility, and prevent attacks that exploit expired, rogue, and non-compliant digital certificate identities.

AppViewX Terraform Provider allows you to manage certificates using the AppViewX platform. This provider enables certificate creation and download through Terraform configurations.

## Requirements

- Terraform 1.0 or later
- AppViewX Service Account Credentials
- Configurations in AppViewX like Certificate Authority, Certificate Group, and Policy.

## Installation

1. Download the `terraform-provider-appviewx` binary from the [AppViewX Terraform GitHub](https://github.com/AppViewX/terraform-provider-appviewx).
2. Place the binary in your Terraform plugins directory.
3. Configure the provider in your Terraform configuration file.

## Provider Configuration

The AppViewX Terraform Provider supports configuring authentication credentials directly in the provider configuration or through Terraform variables.

```hcl
provider "appviewx" {
    appviewx_username="username"
    appviewx_password="password"
    appviewx_client_id="clientid"
    appviewx_client_secret="clientsecret"
    appviewx_environment_is_https=true
    appviewx_environment_ip="appviewx_environment_ip or appviewx_environment_fqdn"
    appviewx_environment_port="appviewx_port"
    certificate_download_password="certificate_password"
    key_download_password="key_password"
    log_level="INFO"
}
```

### Using Terraform Variables for Provider Credentials

Use Terraform variables to keep credentials out of your configuration files:

**variables.tf:**
```hcl
variable "appviewx_client_id" {
  type      = string
}

variable "appviewx_client_secret" {
  type      = string
  sensitive = true
}
```

**provider configuration:**
```hcl
provider "appviewx" {
  appviewx_client_id     = var.appviewx_client_id
  appviewx_client_secret = var.appviewx_client_secret
  appviewx_environment_is_https = true
  appviewx_environment_ip = "appviewx_environment_ip_or_fqdn"
  appviewx_environment_port = "appviewx_port"
}
```

**terraform.tfvars** (add to `.gitignore`):
```hcl
appviewx_client_id     = "your_actual_client_id"
appviewx_client_secret = "your_actual_client_secret"
```

### Service Account Secret Rotation

The provider automatically handles service account secret expiration:

1. Provider detects secret expiration
2. Provider generates new secret via AppViewX API
3. If `appviewx_tfvars_file_path` is configured, the specified file is updated with the new secret
4. Provider retries the operation with the new secret
5. Next Terraform operation uses the updated secret

#### Configuration for Secret Rotation

**Option 1: With `.tfvars` file (Recommended)**
```hcl
provider "appviewx" {
  appviewx_client_id     = var.appviewx_client_id
  appviewx_client_secret = var.appviewx_client_secret
  appviewx_tfvars_file_path = "${path.module}/prod-secrets.tfvars"
  # ... other configuration
}
```
The provider will automatically update `prod-secrets.tfvars` when the secret expires.

**Option 2: Without file path (Manual Update Required)**
```hcl
provider "appviewx" {
  appviewx_client_secret = "hardcoded_secret"
  # ... no appviewx_tfvars_file_path specified
}
```
If the secret expires, the provider will display the new secret(regenerated secret) in terraform logs. You must manually update your configuration and re-run terraform.

#### Error Handling During Rotation

If `appviewx_tfvars_file_path` is not configured and the secret expires, you will see an error message like:
```
Error: no tfvars file configured

=== NEW CLIENT SECRET REGENERATED ===
Client ID: your-client-id
Client Secret: newly-generated-secret

Please update your configuration and re-run: terraform apply
```

> **Note**: The `.tfvars` file (if configured) must be writable by the user running Terraform.
> **Note**: When secrets are regenerated, ensure you re-run `terraform apply` to complete the operation.
> **Security**: Keep credentials out of version control and state files. Use `.tfvars` files with `.gitignore` and protect your `terraform.tfstate` with access controls.

## Attributes

- `appviewx_username`:
    - The username used to authenticate with the AppViewX API.
    - This is provided by your AppViewX administrator.
    - **Environment Variable:** If not specified in the provider block, the value will be read from `APPVIEWX_TERRAFORM_USERNAME`.

- `appviewx_password`:
    - The password associated with the AppViewX username.
    - Used for secure authentication.
    - **Environment Variable:** If not specified in the provider block, the value will be read from `APPVIEWX_TERRAFORM_PASSWORD`.

- `appviewx_client_id`:
    - The client ID used to authenticate with the AppViewX API.
    - This is provided by your AppViewX administrator.
    - The value can be provided directly in the provider configuration or through a Terraform variable.
    - **Environment Variable:** If not specified in the provider block, the value will be read from `APPVIEWX_TERRAFORM_CLIENT_ID`.

- `appviewx_client_secret`:
    - The client secret associated with the client ID. This is used for secure authentication.
    - The value can be provided directly in the provider configuration or through a Terraform variable.
    - For service account secret rotation, using a Terraform variable backed by a `.tfvars` file allows the provider to update the secret automatically when the service account secret is rotated.
    - **Environment Variable:** If not specified in the provider block, the value will be read from `APPVIEWX_TERRAFORM_CLIENT_SECRET`.

- `appviewx_tfvars_file_path`:
    - The file path to the `.tfvars` file that contains the `appviewx_client_secret` value.
    - When provided, the provider will automatically update this file with a newly regenerated client secret when the current secret expires.
    - Only the specified file is updated; other `.tfvars` files remain untouched.
    - If not specified, the provider will not persist regenerated secrets to any file. In this case, the new secret must be manually updated in the provider configuration.
    - **Recommended Use:** Specify this when using a custom `.tfvars` file path or multiple `.tfvars` files in your Terraform configuration.
    - **Example:**
      ```hcl
      provider "appviewx" {
        appviewx_tfvars_file_path = "${path.module}/prod-secrets.tfvars"
      }
      ```

- `appviewx_environment_is_https`:
    - A boolean value indicating whether the AppViewX environment uses HTTPS.
    - Set this to `true` if your environment is secured with HTTPS.

- `appviewx_environment_ip`:
    - The IP address or fully qualified domain name (FQDN) of the AppViewX environment.
    - Only the IP or FQDN should be provided, without any port or other values.
    - For on-premise AppViewX, use the IP or FQDN of the gateway.
    - For SaaS, provide the FQDN of the AppViewX Tenant.

- `appviewx_environment_port`:
    - The port number used to connect to the AppViewX environment.
    - Ensure this matches the port configured for API access.
    - For on-premise AppViewX, use `31443`.
    - For SaaS, use `443`.

- `certificate_download_password`:
    - The password used to download the created certificate or provided certificate details in formats such as P12 or PFX.
    - If specified in the provider block, the password will not be stored in the `.tfstate` file.
    - When the password is defined in both the provider and the resource, the value from the provider takes precedence.

- `key_download_password`:
    - The password used to download the private key associated with the certificate.
    - Similar to `certificate_download_password`, if specified in the provider block, it will not be stored in the `.tfstate` file.
    - If defined in both the provider and the resource, the value from the provider takes precedence.

- `log_level`:
    - Describes the log level.
    - Default is set to `INFO`.
    - Possible values are [`INFO`, `DEBUG`].

**Example environment variable usage:**

```sh
export APPVIEWX_TERRAFORM_CLIENT_ID="your_client_id"
export APPVIEWX_TERRAFORM_CLIENT_SECRET="your_client_secret"
export APPVIEWX_TERRAFORM_USERNAME="your_username"
export APPVIEWX_TERRAFORM_PASSWORD="your_password"
```

## Support

For support, please contact [AppViewX Support](https://helpcenter.appviewx.com/login)

## Certificate Management

The AppViewX Terraform Provider simplifies certificate management by enabling seamless integration with the AppViewX platform. Using this provider, you can automate the creation and retrieval of certificates, ensuring secure and efficient workflows for your infrastructure.

Below are the available certificate management operations:

- [Create Certificate](resources/appviewx_create_certificate.md)
- [Download Certificate](resources/appviewx_download_certificate.md)