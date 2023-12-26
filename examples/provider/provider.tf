# self-signed Web GUI certificate
provider "pfsense" {
  url             = "https://192.168.1.1"
  password        = var.pfsense_password
  tls_skip_verify = true
}

# trusted Web GUI certificate
provider "pfsense" {
  url      = "https://pfsense.lan"
  password = var.pfsense_password
}

# custom user
provider "pfsense" {
  url      = "https://10.0.0.1"
  username = "some-user"
  password = var.pfsense_password
}
