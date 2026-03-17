resource "pfsense_dhcpv4_staticmapping" "example" {
  for_each    = toset(["00:00:00:00:00:00", "00:00:00:00:00:01"])
  interface   = "lan"
  mac_address = each.value
  apply       = false
}

resource "terraform_data" "dhcpv4_apply" {
  triggers_replace = pfsense_dhcpv4_staticmapping.example

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.pfsense_dhcpv4_apply.example]
    }
  }
}

# apply once
action "pfsense_dhcpv4_apply" "example" {
  config {
    interface = "lan"
  }
}
