resource "pfsense_dnsresolver_hostoverride" "example" {
  host         = "foobar"
  domain       = "example.com"
  ip_addresses = ["1.1.1.1"]
  description  = "an example"
}
