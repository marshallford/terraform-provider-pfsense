package pfsense

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SystemVersion struct {
	Current string `json:"installed_version"`
	Latest  string `json:"version"`
}

func (pf *Client) GetSystemVersion(ctx context.Context) (*SystemVersion, error) {
	pf.mutexes.Version.Lock()
	defer pf.mutexes.Version.Unlock()

	u := url.URL{Path: "pkg_mgr_install.php"}
	v := url.Values{
		"ajax":       {"ajax"},
		"getversion": {"yes"},
	}

	resp, err := pf.do(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s %s", resp.Status, http.MethodPost, u.String())
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !json.Valid(b) {
		return nil, err
	}

	var r SystemVersion
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	if r.Current == "" {
		return nil, errors.New("current version not returned")
	}

	if r.Latest == "" {
		return nil, errors.New("latest version not returned")
	}

	return &r, nil
}
