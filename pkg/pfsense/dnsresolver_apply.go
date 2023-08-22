package pfsense

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (pf *Client) ApplyDNSResolverChanges(ctx context.Context) (*uuid.UUID, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	u := url.URL{Path: "services_unbound.php"}
	v := url.Values{
		"apply": {"Apply Changes"},
	}

	_, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("failed to apply DNS resolver changes, %w", err)
	}

	uuid := uuid.New()

	return &uuid, nil
}
