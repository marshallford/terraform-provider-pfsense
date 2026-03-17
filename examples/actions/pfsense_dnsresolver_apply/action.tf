resource "pfsense_dnsresolver_hostoverride" "example" {
  for_each     = toset(["a", "b", "c"])
  host         = each.value
  domain       = "example.com"
  ip_addresses = ["10.10.10.10"]
  apply        = false
}

resource "terraform_data" "dnsresolver_apply" {
  triggers_replace = pfsense_dnsresolver_hostoverride.example

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.pfsense_dnsresolver_apply.example]
    }
  }
}

# apply once
action "pfsense_dnsresolver_apply" "example" {}
