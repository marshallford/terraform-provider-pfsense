package pfsense

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"

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
	re2 := regexp.MustCompile(`:([^:]*)$`)
	return re2.ReplaceAllString(do.IPAddress.String(), `@$1`)
}

func (do *DomainOverride) setByHTMLTableRow(i int) error {
	return do.setControlID(i)
}

func (do *DomainOverride) setByHTMLTableCol(i int, text string) error {
	switch i {
	case 0:
		return do.SetDomain(text)
	case 1:
		re := regexp.MustCompile(`@([^@]*)$`)
		str := re.ReplaceAllString(text, `:$1`)
		err := do.SetIPAddress(str)
		return err
	case 2:
		id, description, err := parseDescription(text)
		if err != nil {
			return err
		}
		err = do.SetID(id)
		if err != nil {
			return err
		}
		err = do.SetDescription(description)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("domain override does not have matching field for column %d", i)
	}
	return nil
}

func (do *DomainOverride) SetID(id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("domain override ID not valid UUID")
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
	return nil, fmt.Errorf("domain override with domain %s not found ", domain)
}

func (dos DomainOverrides) GetByID(id string) (*DomainOverride, error) {
	for _, do := range dos {
		if do.ID.String() == id {
			return &do, nil
		}
	}
	return nil, fmt.Errorf("domain override with ID %s not found ", id)
}

func scrapeDNSResolverDomainOverrides(doc *goquery.Document) (*DomainOverrides, error) {
	tableBody := doc.FindMatcher(goquery.Single("div.panel:has(h2:contains('Domain Overrides')) table tbody"))

	if tableBody.Length() == 0 {
		return nil, errors.New("domain overrides table not found")
	}

	domainOverrides := DomainOverrides(scrapeHTMLTable[DomainOverride](tableBody))

	return &domainOverrides, nil
}

func (pf *Client) getDNSResolverDomainOverrides() (*DomainOverrides, error) {
	u := pf.Options.URL.ResolveReference(&url.URL{Path: "services_unbound.php"})

	resp, err := pf.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get DNS resolver page, %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return scrapeDNSResolverDomainOverrides(doc)
}

func (pf *Client) GetDNSResolverDomainOverrides() (*DomainOverrides, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	return pf.getDNSResolverDomainOverrides()
}

func (pf *Client) GetDNSResolverDomainOverride(id string) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrides, err := pf.getDNSResolverDomainOverrides()
	if err != nil {
		return nil, err
	}

	return domainOverrides.GetByID(id)
}

func (pf *Client) CreateDNSResolverDomainOverride(domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	domainOverrideReq.ID = uuid.New()

	u := pf.Options.URL.ResolveReference(&url.URL{Path: "services_unbound_domainoverride_edit.php"})

	resp, err := pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey: {pf.token},
		"domain":    {domainOverrideReq.Domain},
		"ip":        {domainOverrideReq.formatIPAddress()},
		"descr":     {domainOverrideReq.formatDescription()},
		"save":      {"Save"},
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create domain override, %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain override, %w", err)
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

func (pf *Client) UpdateDNSResolverDomainOverride(domainOverrideReq DomainOverride) (*DomainOverride, error) {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	if domainOverrideReq.ID == uuid.Nil {
		return nil, errors.New("domain override missing ID")
	}

	// get current control ID of domain override
	domainOverrides, err := pf.getDNSResolverDomainOverrides()
	if err != nil {
		return nil, err
	}

	currentDomainOverride, err := domainOverrides.GetByID(domainOverrideReq.ID.String())
	if err != nil {
		return nil, err
	}

	controlID := currentDomainOverride.controlID

	// update domain override
	u := pf.Options.URL.ResolveReference(&url.URL{Path: "services_unbound_domainoverride_edit.php"})
	q := u.Query()
	q.Set("id", controlID)
	u.RawQuery = q.Encode()

	resp, err := pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey: {pf.token},
		"domain":    {domainOverrideReq.Domain},
		"ip":        {domainOverrideReq.formatIPAddress()},
		"descr":     {domainOverrideReq.formatDescription()},
		"id":        {controlID},
		"save":      {"Save"},
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update domain override, %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	err = scrapeValidationErrors(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update domain override, %w", err)
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

func (pf *Client) DeleteDNSResolverDomainOverride(id string) error {
	pf.mutexes.DNSResolverDomainOverride.Lock()
	defer pf.mutexes.DNSResolverDomainOverride.Unlock()

	// get current control ID of domain override
	domainOverrides, err := pf.getDNSResolverDomainOverrides()
	if err != nil {
		return err
	}

	freshDomainOverride, err := domainOverrides.GetByID(id)
	if err != nil {
		return err
	}

	controlID := freshDomainOverride.controlID

	// delete domain override
	u := pf.Options.URL.ResolveReference(&url.URL{Path: "services_unbound.php"})

	resp, err := pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey: {pf.token},
		"type":      {"doverride"},
		"act":       {"del"},
		"id":        {controlID},
	})
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete domain override, %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
