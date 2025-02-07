package pfsense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	staticMappingDomainSearchListSep = ";"
	StaticMappingMaxWINSServers      = 2
	StaticMappingMaxDNSServers       = 4
)

type dhcpdV4StaticMappingResponse struct {
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

type DHCPDV4StaticMapping struct {
	Interface           string
	MACAddress          string
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

func (sm DHCPDV4StaticMapping) StringifyIPAddress() string {
	return safeAddrString(sm.IPAddress)
}

func (sm DHCPDV4StaticMapping) StringifyWINSServers() []string {
	winsServers := make([]string, 0, len(sm.WINSServers))
	for _, winsServer := range sm.WINSServers {
		winsServers = append(winsServers, safeAddrString(winsServer))
	}

	return winsServers
}

func (sm DHCPDV4StaticMapping) StringifyDNSServers() []string {
	dnsServers := make([]string, 0, len(sm.DNSServers))
	for _, dnsServer := range sm.DNSServers {
		dnsServers = append(dnsServers, safeAddrString(dnsServer))
	}

	return dnsServers
}

func (sm DHCPDV4StaticMapping) StringifyGateway() string {
	return safeAddrString(sm.Gateway)
}

func (sm DHCPDV4StaticMapping) formatDomainSearchList() string {
	return strings.Join(sm.DomainSearchList, staticMappingDomainSearchListSep)
}

func (sm DHCPDV4StaticMapping) formatDefaultLeaseTime() string {
	if sm.DefaultLeaseTime == 0 {
		return ""
	}

	return strconv.FormatFloat(sm.DefaultLeaseTime.Seconds(), 'f', 0, 64)
}

func (sm DHCPDV4StaticMapping) formatMaximumLeaseTime() string {
	if sm.MaximumLeaseTime == 0 {
		return ""
	}

	return strconv.FormatFloat(sm.MaximumLeaseTime.Seconds(), 'f', 0, 64)
}

func (sm *DHCPDV4StaticMapping) SetInterface(iface string) error {
	sm.Interface = iface

	return nil
}

func (sm *DHCPDV4StaticMapping) SetMACAddress(macAddress string) error {
	sm.MACAddress = macAddress

	return nil
}

func (sm *DHCPDV4StaticMapping) SetClientIdentifier(clientIdentifier string) error {
	sm.ClientIdentifier = clientIdentifier

	return nil
}

func (sm *DHCPDV4StaticMapping) SetIPAddress(ipAddress string) error {
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

func (sm *DHCPDV4StaticMapping) SetARPTableStaticEntry(arpTableStaticEntry bool) error {
	sm.ARPTableStaticEntry = arpTableStaticEntry

	return nil
}

func (sm *DHCPDV4StaticMapping) SetHostname(hostname string) error {
	sm.Hostname = hostname

	return nil
}

func (sm *DHCPDV4StaticMapping) SetDescription(description string) error {
	sm.Description = description

	return nil
}

func (sm *DHCPDV4StaticMapping) SetWINSServers(winsServers []string) error {
	for _, winsServer := range winsServers {
		addr, err := netip.ParseAddr(winsServer)
		if err != nil {
			return err
		}
		sm.WINSServers = append(sm.WINSServers, addr)
	}

	return nil
}

func (sm *DHCPDV4StaticMapping) SetDNSServers(dnsServers []string) error {
	for _, dnsServer := range dnsServers {
		addr, err := netip.ParseAddr(dnsServer)
		if err != nil {
			return err
		}
		sm.DNSServers = append(sm.DNSServers, addr)
	}

	return nil
}

func (sm *DHCPDV4StaticMapping) SetGateway(gateway string) error {
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

func (sm *DHCPDV4StaticMapping) SetDomainName(domainName string) error {
	sm.DomainName = domainName

	return nil
}

func (sm *DHCPDV4StaticMapping) SetDomainSearchList(domainSearchList []string) error {
	sm.DomainSearchList = domainSearchList

	return nil
}

func (sm *DHCPDV4StaticMapping) SetDefaultLeaseTime(defaultLeaseTime string) error {
	duration, err := time.ParseDuration(defaultLeaseTime)
	if err != nil {
		return err
	}

	sm.DefaultLeaseTime = duration

	return nil
}

func (sm *DHCPDV4StaticMapping) SetMaximumLeaseTime(maximumLeaseTime string) error {
	duration, err := time.ParseDuration(maximumLeaseTime)
	if err != nil {
		return err
	}

	sm.MaximumLeaseTime = duration

	return nil
}

type DHCPDV4StaticMappings []DHCPDV4StaticMapping

func (sms DHCPDV4StaticMappings) GetByMACAddress(macAddress string) (*DHCPDV4StaticMapping, error) {
	for _, sm := range sms {
		if sm.MACAddress == macAddress {
			return &sm, nil
		}
	}

	return nil, fmt.Errorf("dhcpd v4 static mapping %w with MAC address '%s'", ErrNotFound, macAddress)
}

func (sms DHCPDV4StaticMappings) GetControlIDByMACAddress(macAddress string) (*int, error) {
	for index, do := range sms {
		if do.MACAddress == macAddress {
			return &index, nil
		}
	}

	return nil, fmt.Errorf("dhcpd v4 static mapping %w with MAC address '%s'", ErrNotFound, macAddress)
}

//nolint:gocognit
func (pf *Client) getDHCPDV4StaticMappings(ctx context.Context, iface string) (*DHCPDV4StaticMappings, error) {
	unableToParseResErr := fmt.Errorf("%w static mapping response", ErrUnableToParse)
	bytes, err := pf.getConfigJSON(ctx, fmt.Sprintf("['dhcpd']['%s']['staticmap']", iface))
	if err != nil {
		return nil, err
	}

	var smResp []dhcpdV4StaticMappingResponse
	err = json.Unmarshal(bytes, &smResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	staticMappings := make(DHCPDV4StaticMappings, 0, len(smResp))
	for _, resp := range smResp {
		var staticMapping DHCPDV4StaticMapping
		var err error

		if err = staticMapping.SetInterface(iface); err != nil {
			return nil, fmt.Errorf("%w %w", ErrClientValidation, err)
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

func (pf *Client) GetDHCPDV4StaticMappings(ctx context.Context, iface string) (*DHCPDV4StaticMappings, error) {
	pf.mutexes.DHCPDV4StaticMapping.Lock()
	defer pf.mutexes.DHCPDV4StaticMapping.Unlock()

	staticMappings, err := pf.getDHCPDV4StaticMappings(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mappings, %w", ErrGetOperationFailed, iface, err)
	}

	return staticMappings, nil
}

func (pf *Client) GetDHCPDV4StaticMapping(ctx context.Context, iface string, macAddress string) (*DHCPDV4StaticMapping, error) {
	pf.mutexes.DHCPDV4StaticMapping.Lock()
	defer pf.mutexes.DHCPDV4StaticMapping.Unlock()

	staticMappings, err := pf.getDHCPDV4StaticMappings(ctx, iface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrGetOperationFailed, iface, err)
	}

	return staticMappings.GetByMACAddress(macAddress)
}

func (pf *Client) createOrUpdateDHCPDV4StaticMapping(ctx context.Context, staticMappingReq DHCPDV4StaticMapping, controlID *int) (*DHCPDV4StaticMapping, error) {
	relativeURL := url.URL{Path: "services_dhcp_edit.php"}
	query := relativeURL.Query()
	query.Set("if", staticMappingReq.Interface)
	relativeURL.RawQuery = query.Encode()
	values := url.Values{
		"mac":              {staticMappingReq.MACAddress},
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
		return nil, err
	}

	err = scrapeHTMLValidationErrors(doc)
	if err != nil {
		return nil, err
	}

	staticMappings, err := pf.getDHCPDV4StaticMappings(ctx, staticMappingReq.Interface)
	if err != nil {
		return nil, err
	}

	staticMapping, err := staticMappings.GetByMACAddress(staticMappingReq.MACAddress)
	if err != nil {
		return nil, err
	}

	return staticMapping, nil
}

func (pf *Client) CreateDHCPDV4StaticMapping(ctx context.Context, staticMappingReq DHCPDV4StaticMapping) (*DHCPDV4StaticMapping, error) {
	pf.mutexes.DHCPDV4StaticMapping.Lock()
	defer pf.mutexes.DHCPDV4StaticMapping.Unlock()

	staticMapping, err := pf.createOrUpdateDHCPDV4StaticMapping(ctx, staticMappingReq, nil)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrCreateOperationFailed, staticMappingReq.Interface, err)
	}

	return staticMapping, nil
}

func (pf *Client) UpdateDHCPDV4StaticMapping(ctx context.Context, staticMappingReq DHCPDV4StaticMapping) (*DHCPDV4StaticMapping, error) {
	pf.mutexes.DHCPDV4StaticMapping.Lock()
	defer pf.mutexes.DHCPDV4StaticMapping.Unlock()

	staticMappings, err := pf.getDHCPDV4StaticMappings(ctx, staticMappingReq.Interface)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrUpdateOperationFailed, staticMappingReq.Interface, err)
	}

	controlID, err := staticMappings.GetControlIDByMACAddress(staticMappingReq.MACAddress)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrUpdateOperationFailed, staticMappingReq.Interface, err)
	}

	staticMapping, err := pf.createOrUpdateDHCPDV4StaticMapping(ctx, staticMappingReq, controlID)
	if err != nil {
		return nil, fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrUpdateOperationFailed, staticMappingReq.Interface, err)
	}

	return staticMapping, nil
}

func (pf *Client) DeleteDHCPDV4StaticMapping(ctx context.Context, iface string, macAddress string) error {
	pf.mutexes.DHCPDV4StaticMapping.Lock()
	defer pf.mutexes.DHCPDV4StaticMapping.Unlock()

	staticMappings, err := pf.getDHCPDV4StaticMappings(ctx, iface)
	if err != nil {
		return fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrDeleteOperationFailed, iface, err)
	}

	controlID, err := staticMappings.GetControlIDByMACAddress(macAddress)
	if err != nil {
		return fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrDeleteOperationFailed, iface, err)
	}

	relativeURL := url.URL{Path: "services_dhcp.php"}
	values := url.Values{
		"if":  {iface},
		"act": {"del"},
		"id":  {strconv.Itoa(*controlID)},
	}

	_, err = pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w '%s' dhcpd v4 static mapping, %w", ErrDeleteOperationFailed, iface, err)
	}

	return nil
}
