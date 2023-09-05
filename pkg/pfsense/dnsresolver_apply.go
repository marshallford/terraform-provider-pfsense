package pfsense

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var (
	ErrApplyDNSResolverChange = errors.New("failed to apply DNS resolver changes")
)

func (pf *Client) ApplyDNSResolverChanges(ctx context.Context) error {
	pf.mutexes.DNSResolverApply.Lock()
	defer pf.mutexes.DNSResolverApply.Unlock()

	u := url.URL{Path: "services_unbound.php"}
	v := url.Values{
		"apply": {"Apply Changes"},
	}

	resp, err := pf.do(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w, %w", ErrApplyDNSResolverChange, err)
	}

	defer resp.Body.Close()

	return nil
}
