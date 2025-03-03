# TODO list

- [X] Use HardwareAddr type for Static Mapping MAC Address
- [ ] Add interfaces data source
- [ ] Refactor Domain Override to consider domain and IP together
- [X] Rename IP Alias Entry "address" to "ip"
- [X] Rename "DHCPD" to "DHCP"
- [ ] Add provider docs for user permissions/group requirements
- [X] Add provider docs to explain concurrency errors
- [X] Add provider option to make all requests serial to minimize errors
- [X] Command data source
- [X] Command resource
- [ ] Validate DNS resolver config file with `unbound-checkconf` before saving
- [ ] Add struct equality checks to confirm update calls are successful
- [ ] Support apply/reload on destroy
- [ ] Firewall Alias entry resources (non-authoritative)
- [ ] Add timeouts
- [ ] Smoke test nil vs empty slice/string/etc
- [ ] Smoke test case where apply fails after change to host/domain override
