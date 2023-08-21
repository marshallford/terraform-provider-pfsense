resource "pfsense_dnsresolver_hostoverride" "example" {
  for_each     = toset(["a", "b", "c"])
  host         = each.value
  domain       = "example.com"
  ip_addresses = ["10.10.10.10"]
  apply        = false
}

# apply once
resource "pfsense_dnsresolver_apply" "example" {
  lifecycle {
    replace_triggered_by = [
      pfsense_dnsresolver_hostoverride.example,
    ]
  }
}
