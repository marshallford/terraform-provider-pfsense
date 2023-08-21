resource "pfsense_dnsresolver_domainoverride" "example" {
  domain      = "servers.example.com"
  ip_address  = "10.10.10.1:53"
  description = "dedicated DHCP/DNS for servers"
}
