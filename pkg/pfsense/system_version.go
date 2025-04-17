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

func (pf *Client) getSystemVersion(ctx context.Context) (*SystemVersion, error) {
	unableToParseResErr := fmt.Errorf("%w system version response", ErrUnableToParse)
	relativeURL := url.URL{Path: "pkg_mgr_install.php"}
	values := url.Values{
		"ajax":       {"ajax"},
		"getversion": {"yes"},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	bytes, err := io.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, err
	}

	if !json.Valid(bytes) {
		return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
	}

	var result SystemVersion
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
	}

	return &result, nil
}

func (pf *Client) GetSystemVersion(ctx context.Context) (*SystemVersion, error) {
	systemVersion, err := pf.getSystemVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w system version, %w", ErrGetOperationFailed, err)
	}

	return systemVersion, nil
}
