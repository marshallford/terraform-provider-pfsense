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
    { address = each.value },
  ]
}

# reload once
resource "pfsense_firewall_filter_reload" "example" {
  lifecycle {
    replace_triggered_by = [
      pfsense_firewall_ip_alias.example,
    ]
  }
}
