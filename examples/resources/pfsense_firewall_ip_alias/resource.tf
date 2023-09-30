# simple IP addresses example
resource "pfsense_firewall_ip_alias" "example" {
  name = "access_points"
  type = "host"
  entries = [
    { address = "192.168.1.5", description = "hallway" },
    { address = "192.168.1.6", description = "bedroom" },
    { address = "192.168.1.7", description = "kitchen" },
  ]
}

# network example
resource "pfsense_firewall_ip_alias" "network_example" {
  name        = "all_servers"
  description = "virtual machines"
  type        = "network"
  entries = [
    { address = "10.10.0.0/22" },
    { address = "10.10.4.0/22" },
    { address = "10.10.8.0/21" },
  ]
}

# mixed example
resource "pfsense_firewall_ip_alias" "mixed_example" {
  name = "poe"
  type = "host"
  entries = [
    { address = pfsense_firewall_ip_alias.example.name },
    { address = "192.168.1.10" },
    { address = "ipcam01.lan" },
  ]
}
