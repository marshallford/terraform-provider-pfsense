package pfsense

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrReloadFirewallFilter = errors.New("failed to reload firewall filter")
)

func (pf *Client) ReloadFirewallFilter(ctx context.Context) error {
	u := url.URL{Path: "status_filter_reload.php"}
	v := url.Values{
		"reloadfilter": {"Reload Filter"},
	}

	resp, err := pf.call(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrApplyDNSResolverChange, err)
	}

	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return nil
}
