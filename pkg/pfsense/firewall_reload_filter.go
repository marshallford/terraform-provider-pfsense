package pfsense

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (pf *Client) ReloadFirewallFilter(ctx context.Context) error {
	pf.mutexes.FirewallFilterReload.Lock()
	defer pf.mutexes.FirewallFilterReload.Unlock()

	relativeURL := url.URL{Path: "status_filter_reload.php"}
	values := url.Values{
		"reloadfilter": {"Reload Filter"},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w, failed to reload firewall filter, %w", ErrApplyOperationFailed, err)
	}

	defer resp.Body.Close() //nolint:errcheck
	_, _ = io.Copy(io.Discard, resp.Body)

	return nil
}
