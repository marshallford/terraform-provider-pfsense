package pfsense

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

// TODO TLS queries
type DomainOverride struct {
	ID          uuid.UUID
	controlID   string
	Domain      string
	IPAddress   netip.AddrPort
	Description string
}

func (do DomainOverride) formatDescription() string {
	return formatDescription(do.ID.String(), do.Description)
}

func (do DomainOverride) formatIPAddress() string {
	index := strings.LastIndex(do.IPAddress.String(), ":")
	return strings.Join([]string{do.IPAddress.String()[:index], do.IPAddress.String()[index+1:]}, "@")
}

func (do *DomainOverride) setByHTMLTableRow(i int) error {
	return do.setControlID(i)
}

func (do *DomainOverride) setByHTMLTableCol(i int, text string) error {
	switch i {
	case 0:
		return do.SetDomain(text)
	case 1:
		index := strings.LastIndex(text, "@")
		return do.SetIPAddress(strings.Join([]string{text[:index], text[index+1:]}, ":"))
	case 2:
		id, description, err := parseDescription(text)
		if err != nil {
			return err
		}

		err = do.SetID(id)
		if err != nil {
			return err
		}

		return do.SetDescription(description)
	}
	return nil
}

func (do *DomainOverride) SetID(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w ID", ErrUnableToParse)
	}

	do.ID = uuid

	return nil
}

func (do *DomainOverride) setControlID(id int) error {
	do.controlID = strconv.Itoa(id)

	return nil
}

func (do *DomainOverride) SetDomain(domain string) error {
	do.Domain = domain

	return nil
}

func (do *DomainOverride) SetIPAddress(ipAddress string) error {
	addr, err := netip.ParseAddrPort(ipAddress)
	if err != nil {
		return err
	}

	do.IPAddress = addr

	return nil
}

func (do *DomainOverride) SetDescription(description string) error {
	do.Description = description

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

func (dos DomainOverrides) GetByID(id string) (*DomainOverride, error) {
	for _, do := range dos {
		if do.ID.String() == id {
			return &do, nil
		}
	}
	return nil, fmt.Errorf("domain override %w with ID '%s'", ErrNotFound, id)
}

func scrapeDNSResolverDomainOverrides(doc *goquery.Document) (*DomainOverrides, error) {
	tableBody := doc.FindMatcher(goquery.Single("div.panel:has(h2:contains('Domain Overrides')) table tbody"))

	if tableBody.Length() == 0 {
		return nil, fmt.Errorf("%w, domain overrides table not found", ErrUnableToScrapeHTML)
	}

	domainOverrides := DomainOverrides(scrapeHTMLTable[DomainOverride](tableBody))

	return &domainOverrides, nil
}

func (pf *Client) getDNSResolverDomainOverrides(ctx context.Context) (*DomainOverrides, error) {
	u := url.URL{Path: "services_unbound.php"}

	doc, err := pf.doHTML(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	return scrapeDNSResolverDomainOverrides(doc)
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

func (pf *Client) GetDNSResolverDomainOverride(ctx context.Context, id string) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w domain override (id '%s'), %w", ErrGetOperationFailed, id, err)
	}

	return domainOverrides.GetByID(id)
}

func (pf *Client) CreateDNSResolverDomainOverride(ctx context.Context, domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrideReq.ID = uuid.New()

	u := url.URL{Path: "services_unbound_domainoverride_edit.php"}
	v := url.Values{
		"domain": {domainOverrideReq.Domain},
		"ip":     {domainOverrideReq.formatIPAddress()},
		"descr":  {domainOverrideReq.formatDescription()},
		"save":   {"Save"},
	}

	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrCreateOperationFailed, err)
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrCreateOperationFailed, err)
	}

	domainOverrides, err := scrapeDNSResolverDomainOverrides(doc)
	if err != nil {
		return nil, err
	}

	domainOverride, err := domainOverrides.GetByID(domainOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	return domainOverride, nil
}

func (pf *Client) UpdateDNSResolverDomainOverride(ctx context.Context, domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	if domainOverrideReq.ID == uuid.Nil {
		return nil, fmt.Errorf("domain override %w 'ID'", ErrMissingField)
	}

	// get current control ID of domain override
	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return nil, err
	}

	currentDomainOverride, err := domainOverrides.GetByID(domainOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	controlID := currentDomainOverride.controlID

	// update domain override
	u := url.URL{Path: "services_unbound_domainoverride_edit.php"}
	q := u.Query()
	q.Set("id", controlID)
	u.RawQuery = q.Encode()
	v := url.Values{
		"domain": {domainOverrideReq.Domain},
		"ip":     {domainOverrideReq.formatIPAddress()},
		"descr":  {domainOverrideReq.formatDescription()},
		"id":     {controlID},
		"save":   {"Save"},
	}

	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrUpdateOperationFailed, err)
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("%w domain override, %w", ErrUpdateOperationFailed, err)
	}

	domainOverrides, err = scrapeDNSResolverDomainOverrides(doc)
	if err != nil {
		return nil, err
	}

	domainOverride, err := domainOverrides.GetByID(domainOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	return domainOverride, nil
}

func (pf *Client) DeleteDNSResolverDomainOverride(ctx context.Context, id string) error {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	// get current control ID of domain override
	domainOverrides, err := pf.getDNSResolverDomainOverrides(ctx)
	if err != nil {
		return err
	}

	freshDomainOverride, err := domainOverrides.GetByID(id)
	if err != nil {
		return err
	}

	controlID := freshDomainOverride.controlID

	// delete domain override
	u := url.URL{Path: "services_unbound.php"}
	v := url.Values{
		"type": {"doverride"},
		"act":  {"del"},
		"id":   {controlID},
	}

	_, err = pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w domain override, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
