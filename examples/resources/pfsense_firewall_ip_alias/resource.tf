# host example
resource "pfsense_firewall_ip_alias" "host_example" {
  name = "access_points"
  type = "host"
  entries = [
    { ip = "192.168.1.5", description = "hallway" },
    { ip = "192.168.1.6", description = "bedroom" },
    { ip = "192.168.1.7", description = "kitchen" },
  ]
}

# network example
resource "pfsense_firewall_ip_alias" "network_example" {
  name        = "all_servers"
  description = "virtual machines"
  type        = "network"
  entries = [
    { ip = "10.10.0.0/22" },
    { ip = "10.10.4.0/22" },
    { ip = "10.10.8.0/21" },
  ]
}

# advanced example
resource "pfsense_firewall_ip_alias" "advanced_example" {
  name = "poe"
  type = "host"
  entries = [
    { ip = pfsense_firewall_ip_alias.host_example.name },
    { ip = "192.168.1.10" },
    { ip = "ipcam01.lan" },
  ]
}
