# simple
resource "pfsense_dnsresolver_hostoverride" "example" {
  host         = "foobar"
  domain       = "example.com"
  ip_addresses = ["1.1.1.1"]
  description  = "an example"
}

# additional aliases example
resource "pfsense_dnsresolver_hostoverride" "aliases_example" {
  host         = "multi"
  domain       = "example.com"
  ip_addresses = ["2.2.2.2"]
  aliases = [
    {
      host   = "second"
      domain = "example.com"
    },
    {
      host   = "third"
      domain = "example.com"
    },
  ]
}
