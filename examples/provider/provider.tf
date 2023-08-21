# Configuration-based authentication
provider "pfsense" {
  username = "admin"
  password = var.pfsense_password
  host     = "http://192.168.1.1"
}
