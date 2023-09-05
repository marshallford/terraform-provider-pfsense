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

// TODO pfSense allows for more than one domain override entry with the same domain

type domainOverrideResponse struct {
	Domain      string  `json:"domain"`
	IPAddress   string  `json:"ip"`
	TLSQueries  *string `json:"forward_tls_upstream"`
	TLSHostname string  `json:"tls_hostname"`
	Description string  `json:"descr"`
}

type DomainOverride struct {
	Domain      string
	IPAddress   netip.AddrPort
	TLSQueries  bool
	TLSHostname string
	Description string
}

func (do DomainOverride) formatIPAddress() string {
	index := strings.LastIndex(do.IPAddress.String(), ":")
	return strings.Join([]string{do.IPAddress.String()[:index], do.IPAddress.String()[index+1:]}, "@")
}

func (do *DomainOverride) SetDomain(domain string) error {
	do.Domain = domain

	return nil
}

// TODO support address without port specified (default to 53/853)
func (do *DomainOverride) SetIPAddress(ipAddress string) error {
	addr, err := netip.ParseAddrPort(ipAddress)
	if err != nil {
		return err
	}

	do.IPAddress = addr

	return nil
}

func (do *DomainOverride) SetTLSQueries(value bool) error {
	do.TLSQueries = value

	return nil
}

func (do *DomainOverride) SetDescription(description string) error {
	do.Description = description

	return nil
}

func (do *DomainOverride) SetTLSHostname(hostname string) error {
	do.TLSHostname = hostname

	return nil
}

type DomainOverrides []DomainOverride

func (dos DomainOverrides) GetByDomain(domain string) (*DomainOverride, error) {
	for _, do := range dos {
		if do.Domain == domain {
			return &do, nil
		}
	}
	return nil, fmt.Errorf("domain override %w with domain '%s'", ErrNotFound, domain)
}

func (dos DomainOverrides) GetControlIDByDomain(domain string) (*int, error) {
	for i, do := range dos {
		if do.Domain == domain {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("domain override %w with domain '%s'", ErrNotFound, domain)
}

func (pf *Client) getDNSResolverDomainOverrides(ctx context.Context) (*DomainOverrides, error) {
	b, err := pf.getConfigJSON(ctx, "['unbound']['domainoverrides']")
	if err != nil {
		return nil, err
	}

	var doResp []domainOverrideResponse
	err = json.Unmarshal(b, &doResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	var domainOverrides DomainOverrides
	for _, resp := range doResp {
		var domainOverride DomainOverride
		var err error

		err = domainOverride.SetDomain(resp.Domain)
		if err != nil {
			return nil, fmt.Errorf("%w domain override response, %w", ErrUnableToParse, err)
		}

		index := strings.LastIndex(resp.IPAddress, "@")
		err = domainOverride.SetIPAddress(strings.Join([]string{resp.IPAddress[:index], resp.IPAddress[index+1:]}, ":"))
		if err != nil {
			return nil, fmt.Errorf("%w domain override response, %w", ErrUnableToParse, err)
		}

		if resp.TLSQueries != nil {
			err = domainOverride.SetTLSQueries(true)
			if err != nil {
				return nil, err
			}
		}

		err = domainOverride.SetTLSHostname(resp.TLSHostname)
		if err != nil {
			return nil, fmt.Errorf("%w domain override response, %w", ErrUnableToParse, err)
		}

		err = domainOverride.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w domain override response, %w", ErrUnableToParse, err)
		}

		domainOverrides = append(domainOverrides, domainOverride)
	}

	return &domainOverrides, nil
}

func (pf *Client) GetDNSResolverDomainOverrides(ctx context.Context) (*DomainOverrides, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w domain overrides, %w", ErrGetOperationFailed, err)
	}

	return domainOverrides, nil
}

func (pf *Client) GetDNSResolverDomainOverride(ctx context.Context, domain string) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w domain override (domain '%s'), %w", ErrGetOperationFailed, domain, err)
	}

	return domainOverrides.GetByDomain(domain)
}

func (pf *Client) createOrUpdateDNSResolverDomainOverride(ctx context.Context, domainOverrideReq DomainOverride, controlID *int) (*DomainOverride, error) {
	u := url.URL{Path: "services_unbound_domainoverride_edit.php"}
	v := url.Values{
		"domain":       {domainOverrideReq.Domain},
		"ip":           {domainOverrideReq.formatIPAddress()},
		"tls_hostname": {domainOverrideReq.TLSHostname},
		"descr":        {domainOverrideReq.Description},
		"save":         {"Save"},
	}

	if domainOverrideReq.TLSQueries {
		v.Set("forward_tls_upstream", "yes")
	}

	if controlID != nil {
		q := u.Query()
		q.Set("id", strconv.Itoa(*controlID))
		u.RawQuery = q.Encode()
	}

	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	err = scrapeHTMLValidationErrors(doc)
	if err != nil {
		return nil, err
	}

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, err
	}

	domainOverride, err := domainOverrides.GetByDomain(domainOverrideReq.Domain)
	if err != nil {
		return nil, err
	}

	return domainOverride, nil
}

func (pf *Client) CreateDNSResolverDomainOverride(ctx context.Context, domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverride, err := pf.createOrUpdateDNSResolverDomainOverride(ctx, domainOverrideReq, nil)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrCreateOperationFailed, err)
	}

	return domainOverride, nil
}

func (pf *Client) UpdateDNSResolverDomainOverride(ctx context.Context, domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrUpdateOperationFailed, err)
	}

	controlID, err := domainOverrides.GetControlIDByDomain(domainOverrideReq.Domain)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrUpdateOperationFailed, err)
	}

	domainOverride, err := pf.createOrUpdateDNSResolverDomainOverride(ctx, domainOverrideReq, controlID)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrUpdateOperationFailed, err)
	}

	return domainOverride, nil
}

func (pf *Client) DeleteDNSResolverDomainOverride(ctx context.Context, domain string) error {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return fmt.Errorf("%w domain override, %w", ErrDeleteOperationFailed, err)
	}

	controlID, err := domainOverrides.GetControlIDByDomain(domain)
	if err != nil {
		return fmt.Errorf("%w domain override, %w", ErrDeleteOperationFailed, err)
	}

	u := url.URL{Path: "services_unbound.php"}
	v := url.Values{
		"type": {"doverride"},
		"act":  {"del"},
		"id":   {strconv.Itoa(*controlID)},
	}

	_, err = pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w domain override, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
