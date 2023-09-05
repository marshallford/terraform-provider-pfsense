# simple
resource "pfsense_dnsresolver_domainoverride" "example" {
  domain      = "servers.example.com"
  ip_address  = "10.10.10.1:53"
  description = "dedicated DHCP/DNS for servers"
}

# SSL/TLS
resource "pfsense_dnsresolver_domainoverride" "tls_example" {
  domain       = "servers.example.com"
  ip_address   = "192.168.2.1:853"
  tls_queries  = true
  tls_hostname = "some.host.name.com"
}
