package pfsense

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

// TODO additional names for host
type HostOverride struct {
	ID          uuid.UUID
	controlID   string
	Host        string
	Domain      string
	IPAddresses []netip.Addr
	Description string
}

func (ho HostOverride) formatDescription() string {
	return formatDescription(ho.ID.String(), ho.Description)
}

func (ho HostOverride) formatIPAddresses() string {
	var addrs []string
	for _, ipAddress := range ho.IPAddresses {
		addrs = append(addrs, ipAddress.String())
	}
	return strings.Join(addrs, ",")
}

func (ho HostOverride) FQDN() string {
	return strings.Join([]string{ho.Host, ho.Domain}, ".")
}

func (ho *HostOverride) setByHTMLTableRow(i int) error {
	return ho.setControlID(i)
}

func (ho *HostOverride) setByHTMLTableCol(i int, text string) error {
	switch i {
	case 0:
		return ho.SetHost(text)
	case 1:
		return ho.SetDomain(text)
	case 2:
		return ho.SetIPAddress(strings.Split(text, ","))
	case 3:
		id, description, err := parseDescription(text)
		if err != nil {
			return err
		}

		err = ho.SetID(id)
		if err != nil {
			return err
		}

		err = ho.SetDescription(description)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("host override does not have matching field for column %d", i)
	}
	return nil
}

func (ho *HostOverride) SetID(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("host override ID not valid UUID")
	}

	ho.ID = uuid

	return nil
}

func (ho *HostOverride) setControlID(id int) error {
	ho.controlID = strconv.Itoa(id)

	return nil
}

func (ho *HostOverride) SetHost(host string) error {
	ho.Host = host

	return nil
}

func (ho *HostOverride) SetDomain(domain string) error {
	ho.Domain = domain

	return nil
}

func (ho *HostOverride) SetIPAddress(ipAddresses []string) error {
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

type HostOverrides []HostOverride

func (hos HostOverrides) GetByFQDN(fqdn string) (*HostOverride, error) {
	for _, ho := range hos {
		if ho.FQDN() == fqdn {
			return &ho, nil
		}
	}
	return nil, fmt.Errorf("host override with FQDN %s not found ", fqdn)
}

func (hos HostOverrides) GetByID(id string) (*HostOverride, error) {
	for _, ho := range hos {
		if ho.ID.String() == id {
			return &ho, nil
		}
	}
	return nil, fmt.Errorf("host override with ID %s not found ", id)
}

func scrapeDNSResolverHostOverrides(doc *goquery.Document) (*HostOverrides, error) {
	tableBody := doc.FindMatcher(goquery.Single("div.panel:has(h2:contains('Host Overrides')) table tbody"))

	if tableBody.Length() == 0 {
		return nil, errors.New("host overrides table not found")
	}

	hostOverrides := HostOverrides(scrapeHTMLTable[HostOverride](tableBody))

	return &hostOverrides, nil
}

func (pf *Client) getDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	u := url.URL{Path: "services_unbound.php"}

	doc, err := pf.doHTML(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get host override, %w", err)
	}

	return scrapeDNSResolverHostOverrides(doc)
}

func (pf *Client) GetDNSResolverHostOverrides(ctx context.Context) (*HostOverrides, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	return pf.getDNSResolverHostOverrides(ctx)
}

func (pf *Client) GetDNSResolverHostOverride(ctx context.Context, id string) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, err
	}

	return hostOverrides.GetByID(id)
}

func (pf *Client) CreateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	hostOverrideReq.ID = uuid.New()

	u := url.URL{Path: "services_unbound_host_edit.php"}
	v := url.Values{
		"host":   {hostOverrideReq.Host},
		"domain": {hostOverrideReq.Domain},
		"ip":     {hostOverrideReq.formatIPAddresses()},
		"descr":  {hostOverrideReq.formatDescription()},
		"save":   {"Save"},
	}

	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to create host override, %w", err)
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create host override, %w", err)
	}

	hostOverrides, err := scrapeDNSResolverHostOverrides(doc)
	if err != nil {
		return nil, err
	}

	hostOverride, err := hostOverrides.GetByID(hostOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	return hostOverride, nil
}

func (pf *Client) UpdateDNSResolverHostOverride(ctx context.Context, hostOverrideReq HostOverride) (*HostOverride, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	if hostOverrideReq.ID == uuid.Nil {
		return nil, errors.New("host override missing ID")
	}

	// get current control ID of host override
	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return nil, err
	}

	currentHostOverride, err := hostOverrides.GetByID(hostOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	controlID := currentHostOverride.controlID

	// update host override
	u := url.URL{Path: "services_unbound_host_edit.php"}
	q := u.Query()
	q.Set("id", controlID)
	u.RawQuery = q.Encode()
	v := url.Values{
		"host":   {hostOverrideReq.Host},
		"domain": {hostOverrideReq.Domain},
		"ip":     {hostOverrideReq.formatIPAddresses()},
		"descr":  {hostOverrideReq.formatDescription()},
		"id":     {controlID},
		"save":   {"Save"},
	}

	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to update host override, %w", err)
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update host override, %w", err)
	}

	hostOverrides, err = scrapeDNSResolverHostOverrides(doc)
	if err != nil {
		return nil, err
	}

	hostOverride, err := hostOverrides.GetByID(hostOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	return hostOverride, nil
}

func (pf *Client) DeleteDNSResolverHostOverride(ctx context.Context, id string) error {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	// get current control ID of host override
	hostOverrides, err := pf.getDNSResolverHostOverrides(ctx)
	if err != nil {
		return err
	}

	freshHostOverride, err := hostOverrides.GetByID(id)
	if err != nil {
		return err
	}

	controlID := freshHostOverride.controlID

	// delete host override
	u := url.URL{Path: "services_unbound.php"}
	v := url.Values{
		"type": {"host"},
		"act":  {"del"},
		"id":   {controlID},
	}

	_, err = pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("failed to delete host override, %w", err)
	}

	return nil
}
