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
)

type hostOverrideResponse struct {
	Host        string                        `json:"host"`
	Domain      string                        `json:"domain"`
	IPAddresses string                        `json:"ip"`
	Description string                        `json:"descr"`
	Aliases     hostOverrideAliasItemResponse `json:"aliases"`
}

type hostOverrideAliasItemResponse struct {
	Item []hostOverrideAliasResponse `json:"item"`
}

type hostOverrideAliasResponse struct {
	Host        string `json:"host"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

type HostOverride struct {
	Host        string
	Domain      string
	IPAddresses []netip.Addr
	Description string
	Aliases     []HostOverrideAlias
}

type HostOverrideAlias struct {
	Host        string
	Domain      string
	Description string
}

func (p *hostOverrideAliasItemResponse) UnmarshalJSON(data []byte) error {
	if data[0] == '{' {
		type t hostOverrideAliasItemResponse
		var resp t
		if err := json.Unmarshal(data, &resp); err != nil {
			return err
		}
		*p = hostOverrideAliasItemResponse(resp)
	}

	return nil
}

// TODO replace with Terraform custom type for netip.Addr
func (ho HostOverride) StringifyIPAddresses() []string {
	addrs := make([]string, 0, len(ho.IPAddresses))
	for _, ipAddress := range ho.IPAddresses {
		addrs = append(addrs, ipAddress.String())
	}

	return addrs
}

func (ho HostOverride) formatIPAddresses() string {
	return strings.Join(ho.StringifyIPAddresses(), ",")
}

func (ho HostOverride) FQDN() string {
	return strings.Join(removeEmptyStrings([]string{ho.Host, ho.Domain}), ".")
}

func (hoa HostOverrideAlias) FQDN() string {
	return strings.Join([]string{hoa.Host, hoa.Domain}, ".")
}

func (ho *HostOverride) SetHost(host string) error {
	ho.Host = host

	return nil
}

func (ho *HostOverride) SetDomain(domain string) error {
	ho.Domain = domain

	return nil
}

func (ho *HostOverride) SetIPAddresses(ipAddresses []string) error {
	for _, ipAddress := range ipAddresses {
		addr, err := netip.ParseAddr(ipAddress)
		if err != nil {
			return err
		}
		ho.IPAddresses = append(ho.IPAddresses, addr)
	}

	return nil
}

func (ho *HostOverride) SetDescription(description string) error {
	ho.Description = description

	return nil
}

func (hoa *HostOverrideAlias) SetHost(host string) error {
	hoa.Host = host

	return nil
}

func (hoa *HostOverrideAlias) SetDomain(domain string) error {
	hoa.Domain = domain

	return nil
}

func (hoa *HostOverrideAlias) SetDescription(description string) error {
	hoa.Description = description

	return nil
}

type HostOverrides []HostOverride

func (hos HostOverrides) GetByFQDN(fqdn string) (*HostOverride, error) {
	for _, ho := range hos {
		if ho.FQDN() == fqdn {
			return &ho, nil
		}
	}

	return nil, fmt.Errorf("host override %w with FQDN '%s'", ErrNotFound, fqdn)
}

func (hos HostOverrides) GetControlIDByFQDN(fqdn string) (*int, error) {
	for index, ho := range hos {
		if ho.FQDN() == fqdn {
			return &index, nil
		}
	}

	return nil, fmt.Errorf("host override %w with FQDN '%s'", ErrNotFound, fqdn)
}

func (pf *Client) getDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	bytes, err := pf.getConfigJSON(ctx, "['unbound']['hosts']")
	if err != nil {
		return nil, err
	}

	var hoResp []hostOverrideResponse
	err = json.Unmarshal(bytes, &hoResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	hostOverrides := make(HostOverrides, 0, len(hoResp))
	for _, resp := range hoResp {
		var hostOverride HostOverride
		var err error

		err = hostOverride.SetHost(resp.Host)
		if err != nil {
			return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
		}

		err = hostOverride.SetDomain(resp.Domain)
		if err != nil {
			return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
		}

		err = hostOverride.SetIPAddresses(strings.Split(resp.IPAddresses, ","))
		if err != nil {
			return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
		}

		err = hostOverride.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
		}

		for _, aliasResp := range resp.Aliases.Item {
			var hostOverrideAlias HostOverrideAlias
			var err error

			err = hostOverrideAlias.SetHost(aliasResp.Host)
			if err != nil {
				return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
			}

			err = hostOverrideAlias.SetDomain(aliasResp.Domain)
			if err != nil {
				return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
			}

			err = hostOverrideAlias.SetDescription(aliasResp.Description)
			if err != nil {
				return nil, fmt.Errorf("%w host override response, %w", ErrUnableToParse, err)
			}

			hostOverride.Aliases = append(hostOverride.Aliases, hostOverrideAlias)
		}

		hostOverrides = append(hostOverrides, hostOverride)
	}

	return &hostOverrides, nil
}

func (pf *Client) GetDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides, %w", ErrGetOperationFailed, err)
	}

	return hostOverrides, nil
}

func (pf *Client) GetDNSResolverHostOverride(ctx context.Context, fqdn string) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host override (FQDN '%s'), %w", ErrGetOperationFailed, fqdn, err)
	}

	return hostOverrides.GetByFQDN(fqdn)
}

func (pf *Client) createOrUpdateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride, controlID *int) (*HostOverride, error) {
	relativeURL := url.URL{Path: "services_unbound_host_edit.php"}
	values := url.Values{
		"host":   {hostOverrideReq.Host},
		"domain": {hostOverrideReq.Domain},
		"ip":     {hostOverrideReq.formatIPAddresses()},
		"descr":  {hostOverrideReq.Description},
		"save":   {"Save"},
	}

	for index, alias := range hostOverrideReq.Aliases {
		values.Set(fmt.Sprintf("aliashost%d", index), alias.Host)
		values.Set(fmt.Sprintf("aliasdomain%d", index), alias.Domain)
		values.Set(fmt.Sprintf("aliasdescription%d", index), alias.Description)
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

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, err
	}

	hostOverride, err := hostOverrides.GetByFQDN(hostOverrideReq.FQDN())
	if err != nil {
		return nil, err
	}

	return hostOverride, nil
}

func (pf *Client) CreateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverride, err := pf.createOrUpdateDNSResolverHostOverride(ctx, hostOverrideReq, nil)
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrCreateOperationFailed, err)
	}

	return hostOverride, nil
}

func (pf *Client) UpdateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrUpdateOperationFailed, err)
	}

	controlID, err := hostOverrides.GetControlIDByFQDN(hostOverrideReq.FQDN())
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrUpdateOperationFailed, err)
	}

	hostOverride, err := pf.createOrUpdateDNSResolverHostOverride(ctx, hostOverrideReq, controlID)
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrUpdateOperationFailed, err)
	}

	return hostOverride, nil
}

func (pf *Client) DeleteDNSResolverHostOverride(ctx context.Context, fqdn string) error {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return fmt.Errorf("%w host override, %w", ErrDeleteOperationFailed, err)
	}

	controlID, err := hostOverrides.GetControlIDByFQDN(fqdn)
	if err != nil {
		return fmt.Errorf("%w host override, %w", ErrDeleteOperationFailed, err)
	}

	relativeURL := url.URL{Path: "services_unbound.php"}
	values := url.Values{
		"type": {"host"},
		"act":  {"del"},
		"id":   {strconv.Itoa(*controlID)},
	}

	_, err = pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w host override, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
