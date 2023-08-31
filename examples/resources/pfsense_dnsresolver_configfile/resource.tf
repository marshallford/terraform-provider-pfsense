# https://docs.netgate.com/pfsense/en/latest/services/dns/wildcards.html#dns-resolver-unbound
resource "pfsense_dnsresolver_configfile" "example" {
  name    = "wildcard-record-example"
  content = <<-EOT
  server:
  local-zone: "subdomain.example.com" redirect
  local-data: "subdomain.example.com 3600 IN A 10.10.10.10"
  EOT
}
