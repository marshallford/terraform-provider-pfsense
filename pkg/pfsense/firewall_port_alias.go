package pfsense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type firewallPortAliasResponse struct {
	Name        string `json:"name"`
	Description string `json:"descr"`
	Type        string `json:"type"`
	Addresses   string `json:"address"`
	Details     string `json:"detail"`
	ControlID   int    `json:"controlID"` //nolint:tagliatelle
}

type FirewallPortAlias struct {
	Name        string
	Description string
	Entries     []FirewallPortAliasEntry
	controlID   int
}

type FirewallPortAliasEntry struct {
	Port        string
	Description string
}

func (portAlias *FirewallPortAlias) SetName(name string) error {
	portAlias.Name = name

	return nil
}

func (portAlias *FirewallPortAlias) SetDescription(description string) error {
	portAlias.Description = description

	return nil
}

func (entry *FirewallPortAliasEntry) SetPort(port string) error {
	entry.Port = port

	return nil
}

func (entry *FirewallPortAliasEntry) SetDescription(description string) error {
	entry.Description = description

	return nil
}

type FirewallPortAliases []FirewallPortAlias

func (portAliases FirewallPortAliases) GetByName(name string) (*FirewallPortAlias, error) {
	for _, portAlias := range portAliases {
		if portAlias.Name == name {
			return &portAlias, nil
		}
	}

	return nil, fmt.Errorf("firewall port alias %w with name '%s'", ErrNotFound, name)
}

func (portAliases FirewallPortAliases) GetControlIDByName(name string) (*int, error) {
	for _, portAlias := range portAliases {
		if portAlias.Name == name {
			return &portAlias.controlID, nil
		}
	}

	return nil, fmt.Errorf("firewall port alias %w with name '%s'", ErrNotFound, name)
}

func (pf *Client) getFirewallPortAliases(ctx context.Context) (*FirewallPortAliases, error) {
	command := "$output = array();" +
		"array_walk($config['aliases']['alias'], function(&$v, $k) use (&$output) {" +
		"if (in_array($v['type'], array('port'))) {" +
		"$v['controlID'] = $k; array_push($output, $v);" +
		"}});" +
		"print_r(json_encode($output));"

	bytes, err := pf.runPHPCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	var portAliasResp []firewallPortAliasResponse
	err = json.Unmarshal(bytes, &portAliasResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	portAliases := make(FirewallPortAliases, 0, len(portAliasResp))
	for _, resp := range portAliasResp {
		var portAlias FirewallPortAlias
		var err error

		err = portAlias.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w firewall port alias response, %w", ErrUnableToParse, err)
		}

		err = portAlias.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w firewall port alias response, %w", ErrUnableToParse, err)
		}

		portAlias.controlID = resp.ControlID

		if resp.Addresses == "" {
			portAliases = append(portAliases, portAlias)

			continue
		}

		addresses := safeSplit(resp.Addresses, " ")
		details := safeSplit(resp.Details, "||")

		if len(addresses) != len(details) {
			return nil, fmt.Errorf("%w firewall port alias response, addresses and descriptions do not match", ErrUnableToParse)
		}

		for index := range addresses {
			var entry FirewallPortAliasEntry
			var err error

			err = entry.SetPort(addresses[index])
			if err != nil {
				return nil, fmt.Errorf("%w firewall port alias response, %w", ErrUnableToParse, err)
			}

			err = entry.SetDescription(details[index])
			if err != nil {
				return nil, fmt.Errorf("%w firewall port alias response, %w", ErrUnableToParse, err)
			}

			portAlias.Entries = append(portAlias.Entries, entry)
		}

		portAliases = append(portAliases, portAlias)
	}

	return &portAliases, nil
}

func (pf *Client) GetFirewallPortAliases(ctx context.Context) (*FirewallPortAliases, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port aliases, %w", ErrGetOperationFailed, err)
	}

	return portAliases, nil
}

func (pf *Client) GetFirewallPortAlias(ctx context.Context, name string) (*FirewallPortAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port alias (name '%s'), %w", ErrGetOperationFailed, name, err)
	}

	return portAliases.GetByName(name)
}

func (pf *Client) createOrUpdateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias, controlID *int) (*FirewallPortAlias, error) {
	relativeURL := url.URL{Path: "firewall_aliases_edit.php"}
	values := url.Values{
		"name":  {portAliasReq.Name},
		"descr": {portAliasReq.Description},
		"type":  {"port"},
		"save":  {"Save"},
	}

	for index, entry := range portAliasReq.Entries {
		values.Set(fmt.Sprintf("address%d", index), entry.Port)
		values.Set(fmt.Sprintf("detail%d", index), entry.Description)
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

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, err
	}

	portAlias, err := portAliases.GetByName(portAliasReq.Name)
	if err != nil {
		return nil, err
	}

	return portAlias, nil
}

func (pf *Client) CreateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias) (*FirewallPortAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	portAlias, err := pf.createOrUpdateFirewallPortAlias(ctx, portAliasReq, nil)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port alias, %w", ErrCreateOperationFailed, err)
	}

	return portAlias, nil
}

func (pf *Client) UpdateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias) (*FirewallPortAlias, error) {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port alias, %w", ErrUpdateOperationFailed, err)
	}

	controlID, err := portAliases.GetControlIDByName(portAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port alias, %w", ErrUpdateOperationFailed, err)
	}

	portAlias, err := pf.createOrUpdateFirewallPortAlias(ctx, portAliasReq, controlID)
	if err != nil {
		return nil, fmt.Errorf("%w firewall port alias, %w", ErrUpdateOperationFailed, err)
	}

	return portAlias, nil
}

func (pf *Client) DeleteFirewallPortAlias(ctx context.Context, name string) error {
	pf.mutexes.FirewallAlias.Lock()
	defer pf.mutexes.FirewallAlias.Unlock()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w firewall port alias, %w", ErrDeleteOperationFailed, err)
	}

	controlID, err := portAliases.GetControlIDByName(name)
	if err != nil {
		return fmt.Errorf("%w firewall port alias, %w", ErrDeleteOperationFailed, err)
	}

	relativeURL := url.URL{Path: "firewall_aliases.php"}
	values := url.Values{
		"act": {"del"},
		"id":  {strconv.Itoa(*controlID)},
	}

	_, err = pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w firewall port alias, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
