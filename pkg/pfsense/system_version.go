package pfsense

import (
	"context"
	"encoding/json"
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
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	if !json.Valid(b) {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	var r SystemVersion
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrUnableToParse, err)
	}

	if r.Current == "" {
		return nil, fmt.Errorf("system version %w 'current'", ErrMissingField)
	}

	if r.Latest == "" {
		return nil, fmt.Errorf("system version %w 'latest'", ErrMissingField)
	}

	return &r, nil
}
