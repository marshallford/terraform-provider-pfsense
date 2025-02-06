data "pfsense_dhcpdv4_staticmappings" "this" {}

output "staticmappings" {
  value = data.pfsense_dhcpdv4_staticmappings.this.all
}
