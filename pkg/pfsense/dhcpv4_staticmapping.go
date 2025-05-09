package pfsense

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type dhcpv4StaticMappingResponse struct {
	MACAddress          string   `json:"mac"`
	ClientIdentifier    string   `json:"cid"`
	IPAddress           string   `json:"ipaddr"`
	ARPTableStaticEntry *string  `json:"arp_table_static_entry"` //nolint:tagliatelle
	Hostname            string   `json:"hostname"`
	Description         string   `json:"descr"`
	WINSServers         []string `json:"winsserver"`
	DNSServers          []string `json:"dnsserver"`
	Gateway             string   `json:"gateway"`
	DomainName          string   `json:"domain"`
	DomainSearchList    string   `json:"domainsearchlist"`
	DefaultLeaseTime    string   `json:"defaultleasetime"`
	MaximumLeaseTime    string   `json:"maxleasetime"`
}

type DHCPv4StaticMapping struct {
	Interface           string
	MACAddress          net.HardwareAddr
	ClientIdentifier    string
	IPAddress           netip.Addr
	ARPTableStaticEntry bool
	Hostname            string
	Description         string
	WINSServers         []netip.Addr
	DNSServers          []netip.Addr
	Gateway             netip.Addr
	DomainName          string
	DomainSearchList    []string
	DefaultLeaseTime    time.Duration
	MaximumLeaseTime    time.Duration
}

func (sm DHCPv4StaticMapping) StringifyIPAddress() string {
	return safeAddrString(sm.IPAddress)
}

func (sm DHCPv4StaticMapping) StringifyWINSServers() []string {
	winsServers := make([]string, 0, len(sm.WINSServers))
	for _, winsServer := range sm.WINSServers {
		winsServers = append(winsServers, safeAddrString(winsServer))
	}

	return winsServers
}

func (sm DHCPv4StaticMapping) StringifyDNSServers() []string {
	dnsServers := make([]string, 0, len(sm.DNSServers))
	for _, dnsServer := range sm.DNSServers {
		dnsServers = append(dnsServers, safeAddrString(dnsServer))
	}

	return dnsServers
}

func (sm DHCPv4StaticMapping) StringifyGateway() string {
	return safeAddrString(sm.Gateway)
}

func (sm DHCPv4StaticMapping) formatDomainSearchList() string {
	return strings.Join(sm.DomainSearchList, staticMappingDomainSearchListSep)
}

func (sm DHCPv4StaticMapping) formatDefaultLeaseTime() string {
	if sm.DefaultLeaseTime == 0 {
		return ""
	}

	return strconv.FormatFloat(sm.DefaultLeaseTime.Seconds(), 'f', 0, 64)
}

func (sm DHCPv4StaticMapping) formatMaximumLeaseTime() string {
	if sm.MaximumLeaseTime == 0 {
		return ""
	}

	return strconv.FormatFloat(sm.MaximumLeaseTime.Seconds(), 'f', 0, 64)
}

func (sm *DHCPv4StaticMapping) SetInterface(iface string) error {
	sm.Interface = iface

	return nil
}

func (sm *DHCPv4StaticMapping) SetMACAddress(macAddress string) error {
	if macAddress == "" {
		return nil
	}

	mac, err := net.ParseMAC(macAddress)
	if err != nil {
		return err
	}

	sm.MACAddress = mac

	return nil
}

func (sm *DHCPv4StaticMapping) SetClientIdentifier(clientIdentifier string) error {
	sm.ClientIdentifier = clientIdentifier

	return nil
}

func (sm *DHCPv4StaticMapping) SetIPAddress(ipAddress string) error {
	if ipAddress == "" {
		return nil
	}

	addr, err := netip.ParseAddr(ipAddress)
	if err != nil {
		return err
	}

	sm.IPAddress = addr

	return nil
}

func (sm *DHCPv4StaticMapping) SetARPTableStaticEntry(arpTableStaticEntry bool) error {
	sm.ARPTableStaticEntry = arpTableStaticEntry

	return nil
}

func (sm *DHCPv4StaticMapping) SetHostname(hostname string) error {
	sm.Hostname = hostname

	return nil
}

func (sm *DHCPv4StaticMapping) SetDescription(description string) error {
	sm.Description = description

	return nil
}

func (sm *DHCPv4StaticMapping) SetWINSServers(winsServers []string) error {
	for _, winsServer := range winsServers {
		addr, err := netip.ParseAddr(winsServer)
		if err != nil {
			return err
		}
		sm.WINSServers = append(sm.WINSServers, addr)
	}

	return nil
}

func (sm *DHCPv4StaticMapping) SetDNSServers(dnsServers []string) error {
	for _, dnsServer := range dnsServers {
		addr, err := netip.ParseAddr(dnsServer)
		if err != nil {
			return err
		}
		sm.DNSServers = append(sm.DNSServers, addr)
	}

	return nil
}

func (sm *DHCPv4StaticMapping) SetGateway(gateway string) error {
	if gateway == "" {
		return nil
	}

	addr, err := netip.ParseAddr(gateway)
	if err != nil {
		return err
	}

	sm.Gateway = addr

	return nil
}

func (sm *DHCPv4StaticMapping) SetDomainName(domainName string) error {
	sm.DomainName = domainName

	return nil
}

func (sm *DHCPv4StaticMapping) SetDomainSearchList(domainSearchList []string) error {
	sm.DomainSearchList = domainSearchList

	return nil
}

func (sm *DHCPv4StaticMapping) SetDefaultLeaseTime(defaultLeaseTime string) error {
	duration, err := time.ParseDuration(defaultLeaseTime)
	if err != nil {
		return err
	}

	sm.DefaultLeaseTime = duration

	return nil
}

func (sm *DHCPv4StaticMapping) SetMaximumLeaseTime(maximumLeaseTime string) error {
	duration, err := time.ParseDuration(maximumLeaseTime)
	if err != nil {
		return err
	}

	sm.MaximumLeaseTime = duration

	return nil
}

type DHCPv4StaticMappings []DHCPv4StaticMapping

func (sms DHCPv4StaticMappings) GetByMACAddress(macAddress net.HardwareAddr) (*DHCPv4StaticMapping, error) {
	for _, sm := range sms {
		if sm.MACAddress.String() == macAddress.String() {
			return &sm, nil
		}
	}

	return nil, fmt.Errorf("static mapping %w with mac address '%s'", ErrNotFound, macAddress)
}

func (sms DHCPv4StaticMappings) GetControlIDByMACAddress(macAddress net.HardwareAddr) (*int, error) {
	for index, do := range sms {
		if do.MACAddress.String() == macAddress.String() {
			return &index, nil
		}
	}

	return nil, fmt.Errorf("static mapping %w with mac address '%s'", ErrNotFound, macAddress)
}

//nolint:gocognit
func (pf *Client) getDHCPv4StaticMappings(ctx context.Context, iface string) (*DHCPv4StaticMappings, error) {
	unableToParseResErr := fmt.Errorf("%w static mapping response", ErrUnableToParse)
	command := fmt.Sprintf("print_r(json_encode($config['dhcpd']['%s']['staticmap']));", iface)
	var smResp []dhcpv4StaticMappingResponse
	if err := pf.executePHPCommand(ctx, command, &smResp); err != nil {
		return nil, err
	}

	staticMappings := make(DHCPv4StaticMappings, 0, len(smResp))
	for _, resp := range smResp {
		var staticMapping DHCPv4StaticMapping
		var err error

		if err = staticMapping.SetInterface(iface); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetMACAddress(resp.MACAddress); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetClientIdentifier(resp.ClientIdentifier); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetIPAddress(resp.IPAddress); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if resp.ARPTableStaticEntry != nil {
			if err = staticMapping.SetARPTableStaticEntry(true); err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}
		}

		if err = staticMapping.SetHostname(resp.Hostname); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetDescription(resp.Description); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetWINSServers(resp.WINSServers); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetDNSServers(resp.DNSServers); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetGateway(resp.Gateway); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetDomainName(resp.DomainName); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetDomainSearchList(safeSplit(resp.DomainSearchList, staticMappingDomainSearchListSep)); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetDefaultLeaseTime(durationSeconds(resp.DefaultLeaseTime)); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		if err = staticMapping.SetMaximumLeaseTime(durationSeconds(resp.MaximumLeaseTime)); err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		staticMappings = append(staticMappings, staticMapping)
	}

	return &staticMappings, nil
}

func (pf *Client) GetDHCPv4StaticMappings(ctx context.Context, iface string) (*DHCPv4StaticMappings, error) {
	defer pf.read(&pf.mutexes.DHCPv4StaticMapping)()

	staticMappings, err := pf.getDHCPv4StaticMappings(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mappings, %w", ErrGetOperationFailed, iface, err)
	}

	return staticMappings, nil
}

func (pf *Client) GetDHCPv4StaticMapping(ctx context.Context, iface string, macAddress net.HardwareAddr) (*DHCPv4StaticMapping, error) {
	defer pf.read(&pf.mutexes.DHCPv4StaticMapping)()

	staticMappings, err := pf.getDHCPv4StaticMappings(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mappings, %w", ErrGetOperationFailed, iface, err)
	}

	staticMapping, err := staticMappings.GetByMACAddress(macAddress)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping, %w", ErrGetOperationFailed, iface, err)
	}

	return staticMapping, nil
}

func (pf *Client) createOrUpdateDHCPv4StaticMapping(ctx context.Context, staticMappingReq DHCPv4StaticMapping, controlID *int) error {
	relativeURL := url.URL{Path: "services_dhcp_edit.php"}
	query := relativeURL.Query()
	query.Set("if", staticMappingReq.Interface)
	relativeURL.RawQuery = query.Encode()
	values := url.Values{
		"mac":              {staticMappingReq.MACAddress.String()},
		"cid":              {staticMappingReq.ClientIdentifier},
		"ipaddr":           {staticMappingReq.StringifyIPAddress()},
		"hostname":         {staticMappingReq.Hostname},
		"descr":            {staticMappingReq.Description},
		"gateway":          {staticMappingReq.StringifyGateway()},
		"domain":           {staticMappingReq.DomainName},
		"domainsearchlist": {staticMappingReq.formatDomainSearchList()},
		"deftime":          {staticMappingReq.formatDefaultLeaseTime()},
		"maxtime":          {staticMappingReq.formatMaximumLeaseTime()},
		"save":             {"Save"},
	}

	if staticMappingReq.ARPTableStaticEntry {
		values.Set("arp_table_static_entry", "yes")
	}

	for index, winsServer := range staticMappingReq.WINSServers {
		values.Add(fmt.Sprintf("wins%d", index+1), safeAddrString(winsServer))
	}

	for index, dnsServer := range staticMappingReq.DNSServers {
		values.Add(fmt.Sprintf("dns%d", index+1), safeAddrString(dnsServer))
	}

	if controlID != nil {
		q := relativeURL.Query()
		q.Set("id", strconv.Itoa(*controlID))
		relativeURL.RawQuery = q.Encode()
	}

	doc, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return err
	}

	return scrapeHTMLValidationErrors(doc)
}

func (pf *Client) CreateDHCPv4StaticMapping(ctx context.Context, staticMappingReq DHCPv4StaticMapping) (*DHCPv4StaticMapping, error) {
	defer pf.write(&pf.mutexes.DHCPv4StaticMapping)()

	if err := pf.createOrUpdateDHCPv4StaticMapping(ctx, staticMappingReq, nil); err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping, %w", ErrCreateOperationFailed, staticMappingReq.Interface, err)
	}

	staticMappings, err := pf.getDHCPv4StaticMappings(ctx, staticMappingReq.Interface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mappings after creating, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	staticMapping, err := staticMappings.GetByMACAddress(staticMappingReq.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping after creating, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	return staticMapping, nil
}

func (pf *Client) UpdateDHCPv4StaticMapping(ctx context.Context, staticMappingReq DHCPv4StaticMapping) (*DHCPv4StaticMapping, error) {
	defer pf.write(&pf.mutexes.DHCPv4StaticMapping)()

	staticMappings, err := pf.getDHCPv4StaticMappings(ctx, staticMappingReq.Interface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mappings, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	controlID, err := staticMappings.GetControlIDByMACAddress(staticMappingReq.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	if err := pf.createOrUpdateDHCPv4StaticMapping(ctx, staticMappingReq, controlID); err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping, %w", ErrUpdateOperationFailed, staticMappingReq.Interface, err)
	}

	staticMappings, err = pf.getDHCPv4StaticMappings(ctx, staticMappingReq.Interface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mappings after creating, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	staticMapping, err := staticMappings.GetByMACAddress(staticMappingReq.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' static mapping after creating, %w", ErrGetOperationFailed, staticMappingReq.Interface, err)
	}

	// TODO equality check.
	return staticMapping, nil
}

func (pf *Client) deleteDHCPv4StaticMapping(ctx context.Context, iface string, controlID int) error {
	relativeURL := url.URL{Path: "services_dhcp.php"}
	values := url.Values{
		"if":  {iface},
		"act": {"del"},
		"id":  {strconv.Itoa(controlID)},
	}

	_, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)

	return err
}

func (pf *Client) DeleteDHCPv4StaticMapping(ctx context.Context, iface string, macAddress net.HardwareAddr) error {
	defer pf.write(&pf.mutexes.DHCPv4StaticMapping)()

	staticMappings, err := pf.getDHCPv4StaticMappings(ctx, iface)
	if err != nil {
		return fmt.Errorf("%w '%s' static mappings, %w", ErrGetOperationFailed, iface, err)
	}

	controlID, err := staticMappings.GetControlIDByMACAddress(macAddress)
	if err != nil {
		return fmt.Errorf("%w '%s' static mapping, %w", ErrGetOperationFailed, iface, err)
	}

	if err := pf.deleteDHCPv4StaticMapping(ctx, iface, *controlID); err != nil {
		return fmt.Errorf("%w '%s' static mapping, %w", ErrDeleteOperationFailed, iface, err)
	}

	staticMappings, err = pf.getDHCPv4StaticMappings(ctx, iface)
	if err != nil {
		return fmt.Errorf("%w '%s' static mappings after deleting, %w", ErrGetOperationFailed, iface, err)
	}

	if _, err := staticMappings.GetByMACAddress(macAddress); err == nil {
		return fmt.Errorf("%w '%s' static mapping, still exists", ErrDeleteOperationFailed, iface)
	}

	return nil
}
