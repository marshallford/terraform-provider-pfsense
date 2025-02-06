resource "pfsense_dhcpdv4_staticmapping" "example" {
  for_each    = toset(["00:00:00:00:00:00", "00:00:00:00:00:01"])
  interface   = "lan"
  mac_address = each.value
  apply       = false
}

# apply once
resource "pfsense_dhcpdv4_apply" "example" {
  interface = "lan"
  lifecycle {
    replace_triggered_by = [
      pfsense_dhcpdv4_staticmapping.example,
    ]
  }
}
