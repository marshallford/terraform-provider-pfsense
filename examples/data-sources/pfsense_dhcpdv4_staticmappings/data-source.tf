data "pfsense_dhcpdv4_staticmappings" "this" {
  interface = "lan"
}

output "staticmappings" {
  value = data.pfsense_dhcpdv4_staticmappings.this.all
}
