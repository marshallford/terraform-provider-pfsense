package pfsense

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (pf *Client) ApplyDNSResolverChanges(ctx context.Context) error {
	pf.mutexes.DNSResolverApply.Lock()
	defer pf.mutexes.DNSResolverApply.Unlock()

	relativeURL := url.URL{Path: "services_unbound.php"}
	values := url.Values{
		"apply": {"Apply Changes"},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w dns resolver changes, %w", ErrApplyOperationFailed, err)
	}

	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return nil
}
