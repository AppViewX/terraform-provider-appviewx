# Certificate Download

The `appviewx_download_certificate` resource allows you to download an existing certificate from the AppViewX platform. This resource provides flexibility to retrieve certificates based on their name and save them to a specified location in the desired format. In addition to the existing file-based download, the resource also supports storing the certificate and private key content directly in the Terraform state by enabling `store_certificate_in_state`.

## Process Overview

1. **Input Parameters**:
   - The resource accepts parameters such as `common_name`, `serial_number`, `resource_id`, `certificate_download_path`, `certificate_download_format`, `certificate_download_password`, and `certificate_chain_required`. These parameters are used to identify and download the certificate.

2. **Certificate Retrieval**:
   - When the resource is applied, it sends a request to the AppViewX API to retrieve the certificate based on the provided parameters like `common_name` and `serial_number` or `resource_id`.

3. **Certificate Download**:
   - The certificate is downloaded to the specified path in the desired format (e.g., `PEM`, `DER`, `P12`). Additional options like password-protecting the file are also supported.
   - If `store_certificate_in_state = true`, the certificate and private key are stored in the Terraform state instead of being written to the filesystem.

## Attributes

The `appviewx_download_certificate` resource supports the following attributes:

### Required Attributes

- **`resource_id`** (string): The unique identifier of the certificate to be downloaded. This is typically obtained from the `appviewx_create_certificate` resource. This `resource_id` would have been printed in the logs when the `appviewx_create_certificate` resource is applied.
- **`common_name`** (string): The common name (CN) of the certificate, typically representing the domain name or entity associated with the certificate.
- **`serial_number`** (string): The serial number (SN) of the certificate, a unique identifier assigned by the certificate authority.

> **Note**: Either `resource_id` or a combination of `common_name` and `serial_number` must be provided to identify the certificate.

### Optional Attributes

- **`certificate_download_path`** (string): The file path to download the certificate. Required only when `store_certificate_in_state` is `false` (default behavior).
- **`certificate_download_format`** (string): The format of the downloaded certificate. Possible values are `PEM`, `CER`, `CRT`, `DER`, `P12`, `PFX`.
- **`certificate_download_password`** (string): The password for the downloaded certificate file. If this password is defined in the provider configuration, it takes precedence over the resource-level password. Additionally, when specified in the provider, the password will not be stored in the Terraform state file for enhanced security.
- **`certificate_chain_required`** (boolean): Whether to include the certificate chain in the downloaded certificate.
- **`key_download_path`** (string): The file path to download the private key separately.
- **`key_download_password`** (string): The password for the downloaded private key. This is required to download the private key from AppViewX and by default the key is password protected from AppViewX. If specified in the provider block, the password will not be stored in the `.tfstate` file. When the password is defined in both the provider and the resource, the value from the provider takes precedence.
- **`download_password_protected_key`** (boolean): To specify whether the private key should be downloaded as password-protected or plain private key. If this is enabled then the password protected key is downloaded as such, but if this is disabled then the password protected key is decrypted using the provided password using OpenSSL and saved in the specified path automatically.
- **`store_certificate_in_state`** (boolean): When enabled, stores the certificate and private key content directly in the Terraform state instead of writing them to the filesystem. Default: `false`.

> **Note**: Private key download is optional and can be ignored if the certificate download format specified is `P12` or `PFX`.

### Computed Attributes

The following attributes are populated only when `store_certificate_in_state = true`:

- **`certificate_content`** (string): The complete PEM-formatted certificate content, including the `BEGIN CERTIFICATE` and `END CERTIFICATE` headers.
- **`key_content`** (string, sensitive): The complete PEM-formatted private key content. This attribute is marked as **sensitive** and is hidden from Terraform plan and apply output.

## Example Usage

### File-Based Download (Default)

```hcl
resource "appviewx_download_certificate" "example" {
   common_name                     = "sample.example.com"
   serial_number                   = "serial_number_of_certificate"
   certificate_download_path       = "/path/to/directory"
   certificate_download_format     = "PEM"
   certificate_download_password   = "password"
   certificate_chain_required      = true

   key_download_path               = "/path/to/directory"
   key_download_password           = "password"
   download_password_protected_key = false
}
```

### Store Certificate in Terraform State

```hcl
resource "appviewx_download_certificate" "example" {
   common_name                   = "sample.example.com"
   serial_number                 = "serial_number_of_certificate"
   certificate_download_format   = "PEM"
   certificate_chain_required    = true

   key_download_password         = "password"

   store_certificate_in_state    = true
}
```

The downloaded values can be referenced by other Terraform resources:

```hcl
resource "local_file" "certificate" {
  content  = appviewx_download_certificate.example.certificate_content
  filename = "${path.module}/certificate.pem"
}

resource "local_sensitive_file" "private_key" {
  content  = appviewx_download_certificate.example.key_content
  filename = "${path.module}/private_key.pem"
}
```

> **Note**: When `store_certificate_in_state = true`, no certificate or private key files are created on the local filesystem. The certificate content is stored in the Terraform state, and `key_content` is treated as a sensitive attribute.

> **Note**: This `appviewx_download_certificate` resource can be used to download the same or different certificates multiple times.

## Import

To import an existing certificate into the Terraform state, use the following command:

```bash
terraform import appviewx_download_certificate.downloadcert <resource_id>
```

Replace `<resource_id>` with the actual resource ID of the certificate you want to import.