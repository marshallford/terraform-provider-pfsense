# port example
resource "pfsense_firewall_port_alias" "port_example" {
  name = "common"
  entries = [
    { port = "22", description = "ssh" },
    { port = "443", description = "https" },
  ]
}

# port range example
resource "pfsense_firewall_port_alias" "port_range_example" {
  name = "useful"
  entries = [
    { port = "20:21", description = "data and command ftp" },
    { port = "1024:65535", description = "unprivileged" },
  ]
}

# advanced example
resource "pfsense_firewall_port_alias" "advanced_example" {
  name = "extended"
  entries = [
    { port = pfsense_firewall_port_alias.port_example.name },
    { port = "53", description = "dns" },
  ]
}
