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

const (
	hostOverrideIPAddressesSep = ","
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

// TODO replace with Terraform custom type for netip.Addr.
func (ho HostOverride) StringifyIPAddresses() []string {
	addrs := make([]string, 0, len(ho.IPAddresses))
	for _, ipAddress := range ho.IPAddresses {
		addrs = append(addrs, safeAddrString(ipAddress))
	}

	return addrs
}

func (ho HostOverride) formatIPAddresses() string {
	return strings.Join(ho.StringifyIPAddresses(), hostOverrideIPAddressesSep)
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

	return nil, fmt.Errorf("host override %w with fqdn '%s'", ErrNotFound, fqdn)
}

func (hos HostOverrides) GetControlIDByFQDN(fqdn string) (*int, error) {
	for index, ho := range hos {
		if ho.FQDN() == fqdn {
			return &index, nil
		}
	}

	return nil, fmt.Errorf("host override %w with fqdn '%s'", ErrNotFound, fqdn)
}

func (pf *Client) getDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	unableToParseResErr := fmt.Errorf("%w host override response", ErrUnableToParse)
	command := "print_r(json_encode($config['unbound']['hosts']));"
	var hoResp []hostOverrideResponse
	if err := pf.executePHPCommand(ctx, command, &hoResp); err != nil {
		return nil, err
	}

	hostOverrides := make(HostOverrides, 0, len(hoResp))
	for _, resp := range hoResp {
		var hostOverride HostOverride
		var err error

		err = hostOverride.SetHost(resp.Host)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = hostOverride.SetDomain(resp.Domain)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = hostOverride.SetIPAddresses(safeSplit(resp.IPAddresses, hostOverrideIPAddressesSep))
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = hostOverride.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		for _, aliasResp := range resp.Aliases.Item {
			var hostOverrideAlias HostOverrideAlias
			var err error

			err = hostOverrideAlias.SetHost(aliasResp.Host)
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			err = hostOverrideAlias.SetDomain(aliasResp.Domain)
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			err = hostOverrideAlias.SetDescription(aliasResp.Description)
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			hostOverride.Aliases = append(hostOverride.Aliases, hostOverrideAlias)
		}

		hostOverrides = append(hostOverrides, hostOverride)
	}

	return &hostOverrides, nil
}

func (pf *Client) GetDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	defer pf.read(&pf.mutexes.DNSResolverHostOverride)()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides, %w", ErrGetOperationFailed, err)
	}

	return hostOverrides, nil
}

func (pf *Client) GetDNSResolverHostOverride(ctx context.Context, fqdn string) (*HostOverride, error) {
	defer pf.read(&pf.mutexes.DNSResolverHostOverride)()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides, %w", ErrGetOperationFailed, err)
	}

	hostOverride, err := hostOverrides.GetByFQDN(fqdn)
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrGetOperationFailed, err)
	}

	return hostOverride, nil
}

func (pf *Client) createOrUpdateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride, controlID *int) error {
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
		return err
	}

	return scrapeHTMLValidationErrors(doc)
}

func (pf *Client) CreateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	defer pf.write(&pf.mutexes.DNSResolverHostOverride)()

	if err := pf.createOrUpdateDNSResolverHostOverride(ctx, hostOverrideReq, nil); err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrCreateOperationFailed, err)
	}

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides after creating, %w", ErrGetOperationFailed, err)
	}

	hostOverride, err := hostOverrides.GetByFQDN(hostOverrideReq.FQDN())
	if err != nil {
		return nil, fmt.Errorf("%w host override after creating, %w", ErrGetOperationFailed, err)
	}

	return hostOverride, nil
}

func (pf *Client) UpdateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	defer pf.write(&pf.mutexes.DNSResolverHostOverride)()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides, %w", ErrGetOperationFailed, err)
	}

	controlID, err := hostOverrides.GetControlIDByFQDN(hostOverrideReq.FQDN())
	if err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrGetOperationFailed, err)
	}

	if err := pf.createOrUpdateDNSResolverHostOverride(ctx, hostOverrideReq, controlID); err != nil {
		return nil, fmt.Errorf("%w host override, %w", ErrUpdateOperationFailed, err)
	}

	hostOverrides, err = pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w host overrides after updating, %w", ErrGetOperationFailed, err)
	}

	hostOverride, err := hostOverrides.GetByFQDN(hostOverrideReq.FQDN())
	if err != nil {
		return nil, fmt.Errorf("%w host override after updating, %w", ErrGetOperationFailed, err)
	}

	// TODO equality check.
	return hostOverride, nil
}

func (pf *Client) deleteDNSResolverHostOverride(ctx context.Context, controlID int) error {
	relativeURL := url.URL{Path: "services_unbound.php"}
	values := url.Values{
		"type": {"host"},
		"act":  {"del"},
		"id":   {strconv.Itoa(controlID)},
	}

	_, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)

	return err
}

func (pf *Client) DeleteDNSResolverHostOverride(ctx context.Context, fqdn string) error {
	defer pf.write(&pf.mutexes.DNSResolverHostOverride)()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return fmt.Errorf("%w host overrides, %w", ErrGetOperationFailed, err)
	}

	controlID, err := hostOverrides.GetControlIDByFQDN(fqdn)
	if err != nil {
		return fmt.Errorf("%w host override, %w", ErrGetOperationFailed, err)
	}

	if err := pf.deleteDNSResolverHostOverride(ctx, *controlID); err != nil {
		return fmt.Errorf("%w host override, %w", ErrDeleteOperationFailed, err)
	}

	hostOverrides, err = pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return fmt.Errorf("%w host overrides after deleting, %w", ErrGetOperationFailed, err)
	}

	if _, err := hostOverrides.GetByFQDN(fqdn); err == nil {
		return fmt.Errorf("%w host override, still exists", ErrDeleteOperationFailed)
	}

	return nil
}
