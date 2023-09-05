resource "pfsense_dnsresolver_hostoverride" "example" {
  host         = "foobar"
  domain       = "example.com"
  ip_addresses = ["1.1.1.1"]
  description  = "an example"
}

resource "pfsense_dnsresolver_hostoverride" "multi_example" {
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
