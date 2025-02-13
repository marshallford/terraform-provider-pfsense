package pfsense

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (pf *Client) ApplyDHCPv4Changes(ctx context.Context, iface string) error {
	pf.mutexes.DHCPv4Apply.Lock()
	defer pf.mutexes.DHCPv4Apply.Unlock()

	relativeURL := url.URL{Path: "services_dhcp.php"}
	query := relativeURL.Query()
	query.Set("if", iface)
	relativeURL.RawQuery = query.Encode()
	values := url.Values{
		"apply": {"Apply Changes"},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w '%s' dhcpv4 changes, %w", ErrApplyOperationFailed, iface, err)
	}

	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return nil
}
