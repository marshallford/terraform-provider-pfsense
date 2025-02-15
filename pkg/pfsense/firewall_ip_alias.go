package pfsense

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type firewallIPAliasResponse struct {
	Name        string `json:"name"`
	Description string `json:"descr"`
	Type        string `json:"type"`
	Addresses   string `json:"address"`
	Details     string `json:"detail"`
	ControlID   int    `json:"controlID"` //nolint:tagliatelle
}

type FirewallIPAlias struct {
	Name        string
	Description string
	Type        string
	Entries     []FirewallIPAliasEntry
	controlID   int
}

type FirewallIPAliasEntry struct {
	IP          string
	Description string
}

func (FirewallIPAlias) Types() []string {
	return []string{"host", "network"}
}

func (ipAlias *FirewallIPAlias) SetName(name string) error {
	ipAlias.Name = name

	return nil
}

func (ipAlias *FirewallIPAlias) SetDescription(description string) error {
	ipAlias.Description = description

	return nil
}

func (ipAlias *FirewallIPAlias) SetType(t string) error {
	ipAlias.Type = t

	return nil
}

func (entry *FirewallIPAliasEntry) SetIP(ip string) error {
	entry.IP = ip

	return nil
}

func (entry *FirewallIPAliasEntry) SetDescription(description string) error {
	entry.Description = description

	return nil
}

type FirewallIPAliases []FirewallIPAlias

func (ipAliases FirewallIPAliases) GetByName(name string) (*FirewallIPAlias, error) {
	for _, ipAlias := range ipAliases {
		if ipAlias.Name == name {
			return &ipAlias, nil
		}
	}

	return nil, fmt.Errorf("ip alias %w with name '%s'", ErrNotFound, name)
}

func (ipAliases FirewallIPAliases) GetControlIDByName(name string) (*int, error) {
	for _, ipAlias := range ipAliases {
		if ipAlias.Name == name {
			return &ipAlias.controlID, nil
		}
	}

	return nil, fmt.Errorf("ip alias %w with name '%s'", ErrNotFound, name)
}

func (pf *Client) getFirewallIPAliases(ctx context.Context) (*FirewallIPAliases, error) {
	unableToParseResErr := fmt.Errorf("%w ip alias response", ErrUnableToParse)
	command := "$output = array();" +
		"array_walk($config['aliases']['alias'], function(&$v, $k) use (&$output) {" +
		"if (in_array($v['type'], array('host', 'network'))) {" +
		"$v['controlID'] = $k; array_push($output, $v);" +
		"}});" +
		"print_r(json_encode($output));"
	var ipAliasResp []firewallIPAliasResponse
	if err := pf.ExecutePHPCommand(ctx, command, &ipAliasResp); err != nil {
		return nil, err
	}

	ipAliases := make(FirewallIPAliases, 0, len(ipAliasResp))
	for _, resp := range ipAliasResp {
		var ipAlias FirewallIPAlias
		var err error

		err = ipAlias.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = ipAlias.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = ipAlias.SetType(resp.Type)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		ipAlias.controlID = resp.ControlID

		if resp.Addresses == "" {
			ipAliases = append(ipAliases, ipAlias)

			continue
		}

		ips := safeSplit(resp.Addresses, aliasEntryAddressSep)
		descriptions := safeSplit(resp.Details, aliasEntryDescriptionSep)

		if len(ips) != len(descriptions) {
			return nil, fmt.Errorf("%w, ips and descriptions do not match", unableToParseResErr)
		}

		for index := range ips {
			var entry FirewallIPAliasEntry
			var err error

			err = entry.SetIP(ips[index])
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			err = entry.SetDescription(descriptions[index])
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			ipAlias.Entries = append(ipAlias.Entries, entry)
		}

		ipAliases = append(ipAliases, ipAlias)
	}

	return &ipAliases, nil
}

func (pf *Client) GetFirewallIPAliases(ctx context.Context) (*FirewallIPAliases, error) {
	defer pf.read(&pf.mutexes.FirewallAlias)()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w ip aliases, %w", ErrGetOperationFailed, err)
	}

	return ipAliases, nil
}

func (pf *Client) GetFirewallIPAlias(ctx context.Context, name string) (*FirewallIPAlias, error) {
	defer pf.read(&pf.mutexes.FirewallAlias)()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w ip aliases, %w", ErrGetOperationFailed, err)
	}

	ipAlias, err := ipAliases.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("%w ip alias, %w", ErrGetOperationFailed, err)
	}

	return ipAlias, nil
}

func (pf *Client) createOrUpdateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias, controlID *int) error {
	relativeURL := url.URL{Path: "firewall_aliases_edit.php"}
	values := url.Values{
		"name":  {ipAliasReq.Name},
		"descr": {ipAliasReq.Description},
		"type":  {ipAliasReq.Type},
		"save":  {"Save"},
	}

	for index, entry := range ipAliasReq.Entries {
		values.Set(fmt.Sprintf("address%d", index), entry.IP)
		values.Set(fmt.Sprintf("detail%d", index), entry.Description)
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

func (pf *Client) CreateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias) (*FirewallIPAlias, error) {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	if err := pf.createOrUpdateFirewallIPAlias(ctx, ipAliasReq, nil); err != nil {
		return nil, fmt.Errorf("%w ip alias, %w", ErrCreateOperationFailed, err)
	}

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w ip aliases after creating, %w", ErrGetOperationFailed, err)
	}

	ipAlias, err := ipAliases.GetByName(ipAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w ip alias after creating, %w", ErrGetOperationFailed, err)
	}

	return ipAlias, nil
}

func (pf *Client) UpdateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias) (*FirewallIPAlias, error) {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w ip aliases, %w", ErrGetOperationFailed, err)
	}

	controlID, err := ipAliases.GetControlIDByName(ipAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w ip alias, %w", ErrGetOperationFailed, err)
	}

	if err := pf.createOrUpdateFirewallIPAlias(ctx, ipAliasReq, controlID); err != nil {
		return nil, fmt.Errorf("%w ip alias, %w", ErrUpdateOperationFailed, err)
	}

	ipAliases, err = pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w ip aliases after updating, %w", ErrGetOperationFailed, err)
	}

	ipAlias, err := ipAliases.GetByName(ipAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w ip alias after updating, %w", ErrGetOperationFailed, err)
	}

	return ipAlias, nil
}

func (pf *Client) DeleteFirewallIPAlias(ctx context.Context, name string) error {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w ip aliases, %w", ErrGetOperationFailed, err)
	}

	controlID, err := ipAliases.GetControlIDByName(name)
	if err != nil {
		return fmt.Errorf("%w ip alias, %w", ErrGetOperationFailed, err)
	}

	if err := pf.deleteFirewallAlias(ctx, *controlID); err != nil {
		return fmt.Errorf("%w ip alias, %w", ErrDeleteOperationFailed, err)
	}

	ipAliases, err = pf.getFirewallIPAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w ip aliases after deleting, %w", ErrGetOperationFailed, err)
	}

	if _, err := ipAliases.GetByName(name); err == nil {
		return fmt.Errorf("%w ip alias, still exists", ErrDeleteOperationFailed)
	}

	return nil
}
