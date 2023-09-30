data "pfsense_dnsresolver_hostoverrides" "this" {}

output "hostoverrides" {
  value = data.pfsense_dnsresolver_hostoverrides.this.all
}
