package pfsense

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	dnsResolverConfigFileDir = "/var/unbound/conf.d"
	dnsResolverConfigFileExt = "conf"
)

type configFileResponse struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ConfigFile struct {
	Name    string
	Content string
}

func (cf ConfigFile) formatFileName() string {
	return fmt.Sprintf("%s/%s.%s", dnsResolverConfigFileDir, cf.Name, dnsResolverConfigFileExt)
}

func (cf ConfigFile) formatContent() string {
	return base64.StdEncoding.EncodeToString([]byte(cf.Content))
}

func (cf *ConfigFile) SetName(name string) error {
	cf.Name = name

	return nil
}

func (cf *ConfigFile) SetContent(content string) error {
	cf.Content = content

	return nil
}

type ConfigFiles []ConfigFile

func (cfs ConfigFiles) GetByName(name string) (*ConfigFile, error) {
	for _, cf := range cfs {
		if cf.Name == name {
			return &cf, nil
		}
	}

	return nil, fmt.Errorf("config file %w with name '%s'", ErrNotFound, name)
}

func (pf *Client) getDNSResolverConfigFiles(ctx context.Context) (*ConfigFiles, error) {
	command := "print_r(json_encode(array_map(function ($filename) {" +
		fmt.Sprintf("$configs['name'] = basename($filename, '.%s');", dnsResolverConfigFileExt) +
		"$configs['content'] = file_get_contents($filename);" +
		"return $configs;" +
		fmt.Sprintf("}, glob('%s/*.%s'))));", dnsResolverConfigFileDir, dnsResolverConfigFileExt)

	bytes, err := pf.runPHPCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	var cfResp []configFileResponse
	err = json.Unmarshal(bytes, &cfResp)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrUnableToParse, err)
	}

	configFiles := make(ConfigFiles, 0, len(cfResp))
	for _, resp := range cfResp {
		var configFile ConfigFile
		var err error

		err = configFile.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w config file response, %w", ErrUnableToParse, err)
		}

		err = configFile.SetContent(resp.Content)
		if err != nil {
			return nil, fmt.Errorf("%w config file response, %w", ErrUnableToParse, err)
		}

		configFiles = append(configFiles, configFile)
	}

	return &configFiles, nil
}

func (pf *Client) GetDNSResolverConfigFiles(ctx context.Context) (*ConfigFiles, error) {
	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config files, %w", ErrGetOperationFailed, err)
	}

	return configFiles, nil
}

func (pf *Client) GetDNSResolverConfigFile(ctx context.Context, name string) (*ConfigFile, error) {
	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config file (name '%s'), %w", ErrGetOperationFailed, name, err)
	}

	return configFiles.GetByName(name)
}

func (pf *Client) createOrUpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	relativeURL := url.URL{Path: "diag_edit.php"}
	values := url.Values{
		"file":   {configFileReq.formatFileName()},
		"action": {"save"},
		"data":   {configFileReq.formatContent()},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, err
	}

	message, err := sanitizeHTMLMessage(strings.Trim(string(bytes), "|"))
	if err != nil {
		return nil, err
	}

	if !strings.Contains(message, "success") {
		return nil, fmt.Errorf("%w '%s'", ErrServerValidation, message)
	}

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, err
	}

	configFile, err := configFiles.GetByName(configFileReq.Name)
	if err != nil {
		return nil, err
	}

	return configFile, nil
}

func (pf *Client) CreateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	cf, err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq)
	if err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrCreateOperationFailed, err)
	}

	return cf, nil
}

func (pf *Client) UpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	cf, err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq)
	if err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrUpdateOperationFailed, err)
	}

	return cf, nil
}

func (pf *Client) DeleteDNSResolverConfigFile(ctx context.Context, name string) error {
	var config ConfigFile
	if err := config.SetName(name); err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	relativeURL := url.URL{Path: "diag_command.php"}
	values := url.Values{
		"txtCommand": {fmt.Sprintf("rm %s", config.formatFileName())},
		"submit":     {"EXEC"},
	}

	_, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	return nil
}
