resource "pfsense_firewall_ip_alias" "example" {
  for_each = {
    a = "192.168.1.1"
    b = "192.168.1.2"
    c = "192.168.1.3"
  }
  name  = each.key
  type  = "host"
  apply = false
  entries = [
    { ip = each.value },
  ]
}

resource "terraform_data" "firewall_filter_reload" {
  triggers_replace = pfsense_firewall_ip_alias.example

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.pfsense_firewall_filter_reload.example]
    }
  }
}

# reload once
action "pfsense_firewall_filter_reload" "example" {}
