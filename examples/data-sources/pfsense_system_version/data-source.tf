data "pfsense_system_version" "this" {}

output "versions" {
  value = data.pfsense_system_version.this
}
