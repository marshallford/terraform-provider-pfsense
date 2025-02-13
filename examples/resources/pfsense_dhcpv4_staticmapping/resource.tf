# simple
resource "pfsense_dhcpv4_staticmapping" "example" {
  interface   = "lan"
  mac_address = "00:00:00:00:00:00"
  ip_address  = "192.168.1.10"
}

# advanced
resource "pfsense_dhcpv4_staticmapping" "advanced_example" {
  interface              = "lan"
  mac_address            = "00:00:00:00:00:00"
  client_identifier      = "server-a"
  ip_address             = "192.168.1.10"
  arp_table_static_entry = false
  hostname               = "server-a"
  description            = "server A"
  wins_servers           = ["192.168.1.2"]
  dns_servers            = ["1.1.1.1"]
  gateway                = "192.168.2.1"
  domain_name            = "example.com"
  domain_search_list     = ["example.internal"]
  default_lease_time     = "1h"
  maximum_lease_time     = "12h"
}
