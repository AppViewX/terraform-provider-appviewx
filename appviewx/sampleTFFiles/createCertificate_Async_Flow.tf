# Credentials are loaded from terraform.tfvars (copied from credentials.auto.tfvars.example).
# To rotate an expired client_secret, update the value in terraform.tfvars and re-run terraform apply.
provider "appviewx" {
  appviewx_client_id            = var.appviewx_client_id
  appviewx_client_secret        = var.appviewx_client_secret
  appviewx_environment_is_https = var.appviewx_environment_is_https
  appviewx_environment_ip       = var.appviewx_environment_ip
  appviewx_environment_port     = var.appviewx_environment_port
}
resource "appviewx_create_certificate" "createcert"{
	      common_name="sample.domain.com"
	    hash_function="SHA256"
      	key_type="RSA"
     	bit_length="2048"
     	certificate_authority="AppViewX CA"
     	ca_setting_name="appviewx ca"
     	certificate_type="SSL Certificate"
     	dns_names=["domain.com","domain123.com"]
     	custom_fields={"field_name":"value"}
     	vendor_specific_fields={"field_name":"value"}
     	validity_unit="years"
     	validity_unit_value=1
     	certificate_group_name="Certificate-Group"
     	is_sync=false
}

resource "time_sleep" "wait" {
    depends_on = [appviewx_create_certificate.createcert]
    create_duration = "10s"
}

resource "appviewx_download_certificate" "downloadcert"{
	depends_on = [time_sleep.wait]
	resource_id=appviewx_create_certificate.createcert.resource_id
    certificate_download_path="<Directory/filename where certificate to be downloaded>"
    certificate_download_format="<P12/PEM/CRT/DER/JKS/PFX>"
    certificate_download_password="<Mandatory for P12,JKS and PFX>"

    #If trust tore certificates to be included in the certificate which will be downloaded
    certificate_chain_required=true

    key_download_path="<Directory/filename where private key to be downloaded>" #Key download related fields are optional
    key_download_password="<Mandatory to download private key>"

    #If download_password_protected_key is true then key will be downloaded as a password #protected key which can be used with the password specified in field key_download_password
    download_password_protected_key=false
}
