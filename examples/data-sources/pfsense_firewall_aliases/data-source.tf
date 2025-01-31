data "pfsense_firewall_aliases" "this" {}

output "firewall_ip_aliases" {
  value = data.pfsense_firewall_aliases.this.ip
}

output "firewall_port_aliases" {
  value = data.pfsense_firewall_aliases.this.port
}
