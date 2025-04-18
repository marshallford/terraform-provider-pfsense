package pfsense

import (
	"context"
	"encoding/base64"
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

func (cf ConfigFile) formatName() string {
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
	unableToParseResErr := fmt.Errorf("%w config file response", ErrUnableToParse)
	command := "print_r(json_encode(array_map(function ($filename) {" +
		fmt.Sprintf("$configs['name'] = basename($filename, '.%s');", dnsResolverConfigFileExt) +
		"$configs['content'] = file_get_contents($filename);" +
		"return $configs;" +
		fmt.Sprintf("}, glob('%s/*.%s'))));", dnsResolverConfigFileDir, dnsResolverConfigFileExt)
	var cfResp []configFileResponse
	if err := pf.executePHPCommand(ctx, command, &cfResp); err != nil {
		return nil, err
	}

	configFiles := make(ConfigFiles, 0, len(cfResp))
	for _, resp := range cfResp {
		var configFile ConfigFile
		var err error

		err = configFile.SetName(resp.Name)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		err = configFile.SetContent(resp.Content)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", unableToParseResErr, err)
		}

		configFiles = append(configFiles, configFile)
	}

	return &configFiles, nil
}

func (pf *Client) GetDNSResolverConfigFiles(ctx context.Context) (*ConfigFiles, error) {
	defer pf.read(&pf.mutexes.DNSResolverConfigFile)()

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config files, %w", ErrGetOperationFailed, err)
	}

	return configFiles, nil
}

func (pf *Client) GetDNSResolverConfigFile(ctx context.Context, name string) (*ConfigFile, error) {
	defer pf.read(&pf.mutexes.DNSResolverConfigFile)()

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config files, %w", ErrGetOperationFailed, err)
	}

	configFile, err := configFiles.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrGetOperationFailed, err)
	}

	return configFile, nil
}

func (pf *Client) createOrUpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) error {
	relativeURL := url.URL{Path: "diag_edit.php"}
	values := url.Values{
		"file":   {configFileReq.formatName()},
		"action": {"save"},
		"data":   {configFileReq.formatContent()},
	}

	resp, err := pf.call(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	bytes, err := io.ReadAll(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err
	}

	message, err := sanitizeHTMLMessage(strings.Trim(string(bytes), "|"))
	if err != nil {
		return err
	}

	if !strings.Contains(message, "success") {
		return fmt.Errorf("%w '%s'", ErrServerValidation, message)
	}

	return nil
}

func (pf *Client) CreateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	defer pf.write(&pf.mutexes.DNSResolverConfigFile)()

	if err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq); err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrCreateOperationFailed, err)
	}

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config files after creating, %w", ErrGetOperationFailed, err)
	}

	configFile, err := configFiles.GetByName(configFileReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w config file after creating, %w", ErrGetOperationFailed, err)
	}

	return configFile, nil
}

func (pf *Client) UpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	defer pf.write(&pf.mutexes.DNSResolverConfigFile)()

	if err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq); err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrUpdateOperationFailed, err)
	}

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w config files after updating, %w", ErrGetOperationFailed, err)
	}

	configFile, err := configFiles.GetByName(configFileReq.Name)
	if err != nil {
		return nil, fmt.Errorf("%w config file after updating, %w", ErrGetOperationFailed, err)
	}

	// TODO equality check.
	return configFile, nil
}

func (pf *Client) deleteDNSResolverConfigFile(ctx context.Context, formattedName string) error {
	relativeURL := url.URL{Path: "diag_command.php"}
	values := url.Values{
		"txtCommand": {fmt.Sprintf("rm %s", formattedName)},
		"submit":     {"EXEC"},
	}

	_, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)

	return err
}

func (pf *Client) DeleteDNSResolverConfigFile(ctx context.Context, name string) error {
	defer pf.write(&pf.mutexes.DNSResolverConfigFile)()

	var config ConfigFile
	if err := config.SetName(name); err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	if err := pf.deleteDNSResolverConfigFile(ctx, config.formatName()); err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	configFiles, err := pf.getDNSResolverConfigFiles(ctx)
	if err != nil {
		return fmt.Errorf("%w config files after deleting, %w", ErrGetOperationFailed, err)
	}

	if _, err := configFiles.GetByName(name); err == nil {
		return fmt.Errorf("%w config file, still exists", ErrDeleteOperationFailed)
	}

	return nil
}
