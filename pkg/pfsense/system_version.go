package pfsense

import (
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

func (pf *Client) GetSystemVersion() (*SystemVersion, error) {
	pf.mutexes.Version.Lock()
	defer pf.mutexes.Version.Unlock()

	u := pf.Options.URL.ResolveReference(&url.URL{Path: "pkg_mgr_install.php"})

	resp, err := pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey:  {pf.token},
		"ajax":       {"ajax"},
		"getversion": {"yes"},
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit version form, %d %s", resp.StatusCode, resp.Status)
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
