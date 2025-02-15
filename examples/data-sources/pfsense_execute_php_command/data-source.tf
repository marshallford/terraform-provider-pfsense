# get pfsense php version
data "pfsense_execute_php_command" "php_version" {
  command = "print(json_encode(phpversion()));"
}

output "php_version" {
  value = data.pfsense_execute_php_command.php_version.result
}

# get system dns servers
data "pfsense_execute_php_command" "dns_servers" {
  command = "print(json_encode($config['system']['dnsserver']));"
}

output "dns_servers" {
  value = data.pfsense_execute_php_command.dns_servers.result
}
