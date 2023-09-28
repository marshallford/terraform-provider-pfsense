package pfsense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type firewallIPAliasResponse struct {
	Name        string `json:"name"`
	Description string `json:"descr"`
	Type        string `json:"type"`
	Addresses   string `json:"address"`
	Details     string `json:"detail"`
	ControlID   int    `json:"controlID"`
}

type FirewallIPAlias struct {
	Name        string
	Description string
	Type        string
	Entries     []FirewallIPAliasEntry
	controlID   int
}

type FirewallIPAliasEntry struct {
	Address     string // TODO can be a IP, CIDR, FQDN, or alias, how to validate?
	Description string
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

func (entry *FirewallIPAliasEntry) SetAddress(addr string) error {
	entry.Address = addr

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
	return nil, fmt.Errorf("firewall IP alias %w with name '%s'", ErrNotFound, name)
}

func (ipAliases FirewallIPAliases) GetControlIDByName(name string) (*int, error) {
	for _, ipAlias := range ipAliases {
		if ipAlias.Name == name {
			return &ipAlias.controlID, nil
		}
	}
	return nil, fmt.Errorf("firewall IP alias %w with name '%s'", ErrNotFound, name)
}

func (pf *Client) getFirewallIPAliases(ctx context.Context) (*FirewallIPAliases, error) {
	command := "$output = array();" +
		"array_walk($config['aliases']['alias'], function(&$v, $k) use (&$output) {" +
		"if (in_array($v['type'], array('host', 'network'))) {" +
		"$v['controlID'] = $k; array_push($output, $v);" +
		"}});" +
		"print_r(json_encode($output));"

	b, err := pf.runPHPCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	var ipAliasResp []firewallIPAliasResponse
	err = json.Unmarshal(b, &ipAliasResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	var ipAliases FirewallIPAliases
	for _, resp := range ipAliasResp {
		var ipAlias FirewallIPAlias
		var err error

		err = ipAlias.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w firewall IP alias response, %w", ErrUnableToParse, err)
		}

		err = ipAlias.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w firewall IP alias response, %w", ErrUnableToParse, err)
		}

		err = ipAlias.SetType(resp.Type)
		if err != nil {
			return nil, fmt.Errorf("%w firewall IP alias response, %w", ErrUnableToParse, err)
		}

		ipAlias.controlID = resp.ControlID

		if resp.Addresses == "" {
			ipAliases = append(ipAliases, ipAlias)
			continue
		}

		addresses := strings.Split(resp.Addresses, " ")
		details := strings.Split(resp.Details, "||")

		if len(addresses) != len(details) {
			return nil, fmt.Errorf("%w firewall IP alias response, addresses and descriptions do not match", ErrUnableToParse)
		}

		for i := range addresses {
			var entry FirewallIPAliasEntry
			var err error

			err = entry.SetAddress(addresses[i])
			if err != nil {
				return nil, fmt.Errorf("%w firewall IP alias response, %w", ErrUnableToParse, err)
			}

			err = entry.SetDescription(details[i])
			if err != nil {
				return nil, fmt.Errorf("%w firewall IP alias response, %w", ErrUnableToParse, err)
			}

			ipAlias.Entries = append(ipAlias.Entries, entry)
		}

		ipAliases = append(ipAliases, ipAlias)
	}

	return &ipAliases, nil
}

func (pf *Client) GetFirewallIPAliases(ctx context.Context) (*FirewallIPAliases, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP aliases, %w", ErrGetOperationFailed, err)
	}

	return ipAliases, nil
}

func (pf *Client) GetFirewallIPAlias(ctx context.Context, name string) (*FirewallIPAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP alias (name '%s'), %w", ErrGetOperationFailed, name, err)
	}

	return ipAliases.GetByName(name)
}

func (pf *Client) createOrUpdateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias, controlID *int) (*FirewallIPAlias, error) {
	u := url.URL{Path: "firewall_aliases_edit.php"}
	v := url.Values{
		"name":  {ipAliasReq.Name},
		"descr": {ipAliasReq.Description},
		"type":  {ipAliasReq.Type},
		"save":  {"Save"},
	}

	for i, entry := range ipAliasReq.Entries {
		v.Set(fmt.Sprintf("address%d", i), entry.Address)
		v.Set(fmt.Sprintf("detail%d", i), entry.Description)
	}

	if controlID != nil {
		q := u.Query()
		q.Set("id", strconv.Itoa(*controlID))
		u.RawQuery = q.Encode()
	}

	doc, err := pf.callHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	err = scrapeHTMLValidationErrors(doc)
	if err != nil {
		return nil, err
	}

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, err
	}

	ipAlias, err := ipAliases.GetByName(ipAliasReq.Name)
	if err != nil {
		return nil, err
	}

	return ipAlias, nil
}

func (pf *Client) CreateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias) (*FirewallIPAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	ipAlias, err := pf.createOrUpdateFirewallIPAlias(ctx, ipAliasReq, nil)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP alias, %w", ErrCreateOperationFailed, err)
	}

	return ipAlias, nil
}

func (pf *Client) UpdateFirewallIPAlias(ctx context.Context, ipAliasReq FirewallIPAlias) (*FirewallIPAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP alias, %w", ErrUpdateOperationFailed, err)
	}

	controlID, err := ipAliases.GetControlIDByName(ipAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP alias, %w", ErrUpdateOperationFailed, err)
	}

	ipAlias, err := pf.createOrUpdateFirewallIPAlias(ctx, ipAliasReq, controlID)
	if err != nil {
		return nil, fmt.Errorf("%w firewall IP alias, %w", ErrUpdateOperationFailed, err)
	}

	return ipAlias, nil
}

func (pf *Client) DeleteFirewallIPAlias(ctx context.Context, name string) error {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	ipAliases, err := pf.getFirewallIPAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w firewall IP alias, %w", ErrDeleteOperationFailed, err)
	}

	controlID, err := ipAliases.GetControlIDByName(name)
	if err != nil {
		return fmt.Errorf("%w firewall IP alias, %w", ErrDeleteOperationFailed, err)
	}

	u := url.URL{Path: "firewall_aliases.php"}
	v := url.Values{
		"act": {"del"},
		"id":  {strconv.Itoa(*controlID)},
	}

	_, err = pf.callHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w firewall IP alias, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
