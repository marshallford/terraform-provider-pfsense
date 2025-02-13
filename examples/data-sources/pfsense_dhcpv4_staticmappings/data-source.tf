data "pfsense_dhcpv4_staticmappings" "this" {
  interface = "lan"
}

output "staticmappings" {
  value = data.pfsense_dhcpv4_staticmappings.this.all
}
