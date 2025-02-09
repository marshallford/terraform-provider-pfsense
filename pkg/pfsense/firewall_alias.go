package pfsense

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

const (
	aliasEntryAddressSep     = " "
	aliasEntryDescriptionSep = "||"
)

func (pf *Client) deleteFirewallAlias(ctx context.Context, controlID int) error {
	relativeURL := url.URL{Path: "firewall_aliases.php"}
	values := url.Values{
		"act": {"del"},
		"id":  {strconv.Itoa(controlID)},
	}

	_, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)

	return err
}
