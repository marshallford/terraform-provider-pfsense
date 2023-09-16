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
	u := url.URL{Path: "pkg_mgr_install.php"}
	v := url.Values{
		"ajax":       {"ajax"},
		"getversion": {"yes"},
	}

	resp, err := pf.call(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	if !json.Valid(b) {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	var r SystemVersion
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, fmt.Errorf("%w system version response as JSON, %w", ErrUnableToParse, err)
	}

	return &r, nil
}
