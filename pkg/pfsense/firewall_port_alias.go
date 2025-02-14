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

	return nil, fmt.Errorf("port alias %w with name '%s'", ErrNotFound, name)
}

func (portAliases FirewallPortAliases) GetControlIDByName(name string) (*int, error) {
	for _, portAlias := range portAliases {
		if portAlias.Name == name {
			return &portAlias.controlID, nil
		}
	}

	return nil, fmt.Errorf("port alias %w with name '%s'", ErrNotFound, name)
}

func (pf *Client) getFirewallPortAliases(ctx context.Context) (*FirewallPortAliases, error) {
	unableToParseResErr := fmt.Errorf("%w port alias response", ErrUnableToParse)
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
		return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
	}

	portAliases := make(FirewallPortAliases, 0, len(portAliasResp))
	for _, resp := range portAliasResp {
		var portAlias FirewallPortAlias
		var err error

		err = portAlias.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = portAlias.SetDescription(resp.Description)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		portAlias.controlID = resp.ControlID

		if resp.Addresses == "" {
			portAliases = append(portAliases, portAlias)

			continue
		}

		ports := safeSplit(resp.Addresses, aliasEntryAddressSep)
		descriptions := safeSplit(resp.Details, aliasEntryDescriptionSep)

		if len(ports) != len(descriptions) {
			return nil, fmt.Errorf("%w, ports and descriptions do not match", unableToParseResErr)
		}

		for index := range ports {
			var entry FirewallPortAliasEntry
			var err error

			err = entry.SetPort(ports[index])
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			err = entry.SetDescription(descriptions[index])
			if err != nil {
				return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
			}

			portAlias.Entries = append(portAlias.Entries, entry)
		}

		portAliases = append(portAliases, portAlias)
	}

	return &portAliases, nil
}

func (pf *Client) GetFirewallPortAliases(ctx context.Context) (*FirewallPortAliases, error) {
	defer pf.read(&pf.mutexes.FirewallAlias)()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w port aliases, %w", ErrGetOperationFailed, err)
	}

	return portAliases, nil
}

func (pf *Client) GetFirewallPortAlias(ctx context.Context, name string) (*FirewallPortAlias, error) {
	defer pf.read(&pf.mutexes.FirewallAlias)()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w port aliases, %w", ErrGetOperationFailed, err)
	}

	portAlias, err := portAliases.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("%w port alias, %w", ErrGetOperationFailed, err)
	}

	return portAlias, nil
}

func (pf *Client) createOrUpdateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias, controlID *int) error {
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
		return err
	}

	return scrapeHTMLValidationErrors(doc)
}

func (pf *Client) CreateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias) (*FirewallPortAlias, error) {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	if err := pf.createOrUpdateFirewallPortAlias(ctx, portAliasReq, nil); err != nil {
		return nil, fmt.Errorf("%w port alias, %w", ErrCreateOperationFailed, err)
	}

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w port aliases after creating, %w", ErrGetOperationFailed, err)
	}

	portAlias, err := portAliases.GetByName(portAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w port alias after creating, %w", ErrGetOperationFailed, err)
	}

	return portAlias, nil
}

func (pf *Client) UpdateFirewallPortAlias(ctx context.Context, portAliasReq FirewallPortAlias) (*FirewallPortAlias, error) {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w port aliases, %w", ErrGetOperationFailed, err)
	}

	controlID, err := portAliases.GetControlIDByName(portAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w port alias, %w", ErrGetOperationFailed, err)
	}

	if err := pf.createOrUpdateFirewallPortAlias(ctx, portAliasReq, controlID); err != nil {
		return nil, fmt.Errorf("%w port alias, %w", ErrUpdateOperationFailed, err)
	}

	portAliases, err = pf.getFirewallPortAliases(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w port aliases after updating, %w", ErrGetOperationFailed, err)
	}

	portAlias, err := portAliases.GetByName(portAliasReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w port alias after updating, %w", ErrGetOperationFailed, err)
	}

	return portAlias, nil
}

func (pf *Client) DeleteFirewallPortAlias(ctx context.Context, name string) error {
	defer pf.write(&pf.mutexes.FirewallAlias)()

	portAliases, err := pf.getFirewallPortAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w port aliases, %w", ErrGetOperationFailed, err)
	}

	controlID, err := portAliases.GetControlIDByName(name)
	if err != nil {
		return fmt.Errorf("%w port alias, %w", ErrGetOperationFailed, err)
	}

	if err := pf.deleteFirewallAlias(ctx, *controlID); err != nil {
		return fmt.Errorf("%w port alias, %w", ErrDeleteOperationFailed, err)
	}

	portAliases, err = pf.getFirewallPortAliases(ctx)
	if err != nil {
		return fmt.Errorf("%w port aliases after deleting, %w", ErrGetOperationFailed, err)
	}

	if _, err := portAliases.GetByName(name); err == nil {
		return fmt.Errorf("%w port alias, still exists", ErrDeleteOperationFailed)
	}

	return nil
}
