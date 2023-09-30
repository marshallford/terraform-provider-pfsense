data "pfsense_dnsresolver_domainoverrides" "this" {}

output "domainoverrides" {
  value = data.pfsense_dnsresolver_domainoverrides.this.all
}
