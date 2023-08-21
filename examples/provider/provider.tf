# Configuration-based authentication
provider "pfsense" {
  username        = "admin"
  password        = var.pfsense_password
  host            = "https://192.168.1.1"
  tls_skip_verify = true
}
