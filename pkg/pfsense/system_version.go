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
	Current string `json:"installed_version"` //nolint:tagliatelle
	Latest  string `json:"version"`           //nolint:tagliatelle
}

func (pf *Client) GetSystemVersion(ctx context.Context) (*SystemVersion, error) {
	relativeURL := url.URL{Path: "pkg_mgr_install.php"}
	values := url.Values{
		"ajax":       {"ajax"},
		"getversion": {"yes"},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	if !json.Valid(bytes) {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	var result SystemVersion
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, fmt.Errorf("%w system version response as JSON, %w", ErrUnableToParse, err)
	}

	return &result, nil
}
