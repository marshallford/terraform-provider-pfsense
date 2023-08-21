package pfsense

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (pf *Client) ApplyDNSResolverChanges() (*uuid.UUID, error) {
	pf.mutexes.DNSResolverHostOverride.Lock()
	defer pf.mutexes.DNSResolverHostOverride.Unlock()

	u := pf.Options.URL.ResolveReference(&url.URL{Path: "services_unbound.php"})

	resp, err := pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey: {pf.token},
		"apply":     {"Apply Changes"},
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to apply DNS resolver changes, %d %s", resp.StatusCode, resp.Status)
	}

	uuid := uuid.New()

	return &uuid, nil
}
