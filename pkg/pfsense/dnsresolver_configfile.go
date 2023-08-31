package pfsense

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	dnsResolverConfDir = "/var/unbound/conf.d"
)

var (
	ErrDNSResolverConfigFileSkipInvalid = fmt.Errorf("skipping invalid config file")
	ErrDNSResolverConfigFileInvalidName = fmt.Errorf("config file name must only consist of lowercase alphanumeric characters (with dashes)")
)

type ConfigFile struct {
	Name    string
	Content string
}

func sanitizeHTMLMessage(text string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		return "", err
	}
	sanitize := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	return sanitize.ReplaceAllString(doc.Text(), ""), nil
}

func (cf ConfigFile) formatFileName() string {
	parts := []string{cf.Name, clientName, "conf"}
	return fmt.Sprintf("%s/%s", dnsResolverConfDir, strings.Join(parts, "."))
}

func (cf ConfigFile) formatContent() string {
	return base64.StdEncoding.EncodeToString([]byte(cf.Content))
}

func (cf *ConfigFile) setByHTMLTableRow(i int) error {
	if i < 3 {
		return ErrDNSResolverConfigFileSkipInvalid
	}
	return nil
}

func (cf *ConfigFile) setByHTMLTableCol(i int, text string) error {
	if i == 1 {
		parts := strings.Split(text, ".")
		if len(parts) != 3 || parts[1] != clientName {
			return ErrNotManagedByClient
		}
		return cf.SetName(parts[0])
	}

	return nil
}

func (cf *ConfigFile) SetName(name string) error {
	var isValidName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString
	if !isValidName(name) {
		return ErrDNSResolverConfigFileInvalidName
	}

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

func scrapeDNSResolverConfigFiles(doc *goquery.Document) (*ConfigFiles, error) {
	tableBody := doc.FindMatcher(goquery.Single("table tbody"))

	if tableBody.Length() == 0 {
		return nil, fmt.Errorf("%w, config files table not found", ErrUnableToScrapeHTML)
	}

	configFiles := ConfigFiles(scrapeHTMLTable[ConfigFile](tableBody))

	return &configFiles, nil
}

func (pf *Client) getDNSResolverConfigFile(ctx context.Context, cf ConfigFile) (*ConfigFile, error) {
	u := url.URL{Path: "diag_edit.php"}
	v := url.Values{
		"file":   {cf.formatFileName()},
		"action": {"load"},
	}

	resp, err := pf.do(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.Trim(string(b), "|"), "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("%w response", ErrUnableToParse)
	}

	returnCode, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w return code", ErrUnableToParse)
	}

	if returnCode != 0 {
		message, err := sanitizeHTMLMessage(parts[1])
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%w, '%s'", ErrResponse, message)
	}

	content, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	if err := cf.SetContent(string(content)); err != nil {
		return nil, err
	}

	return &cf, nil
}

func (pf *Client) GetDNSResolverConfigFiles(ctx context.Context) (*ConfigFiles, error) {
	pf.mutexes.DNSResolverConfigFile.Lock()
	defer pf.mutexes.DNSResolverConfigFile.Unlock()

	u := url.URL{Path: "vendor/filebrowser/browser.php"}
	q := u.Query()
	q.Set("path", dnsResolverConfDir)
	u.RawQuery = q.Encode()

	doc, err := pf.doHTML(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("%w config files, %w", ErrGetOperationFailed, err)
	}

	partialConfigFiles, err := scrapeDNSResolverConfigFiles(doc)
	if err != nil {
		return nil, fmt.Errorf("%w for config files, %w", ErrUnableToScrapeHTML, err)
	}

	var cfs []ConfigFile
	for _, pcf := range *partialConfigFiles {
		cf, err := pf.getDNSResolverConfigFile(ctx, pcf)
		if err != nil {
			return nil, fmt.Errorf("%w config files, %w", ErrGetOperationFailed, err)
		}
		cfs = append(cfs, *cf)
	}

	configFiles := ConfigFiles(cfs)

	return &configFiles, nil
}

func (pf *Client) GetDNSResolverConfigFile(ctx context.Context, name string) (*ConfigFile, error) {
	pf.mutexes.DNSResolverConfigFile.Lock()
	defer pf.mutexes.DNSResolverConfigFile.Unlock()

	var cf ConfigFile
	if err := cf.SetName(name); err != nil {
		return nil, fmt.Errorf("%w config file (name '%s'), %w", ErrGetOperationFailed, name, err)
	}

	configFile, err := pf.getDNSResolverConfigFile(ctx, cf)
	if err != nil {
		return nil, fmt.Errorf("%w config file (name '%s'), %w", ErrGetOperationFailed, name, err)
	}

	return configFile, nil
}

func (pf *Client) createOrUpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	u := url.URL{Path: "diag_edit.php"}
	v := url.Values{
		"file":   {configFileReq.formatFileName()},
		"action": {"save"},
		"data":   {configFileReq.formatContent()},
	}

	resp, err := pf.do(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	message, err := sanitizeHTMLMessage(strings.Trim(string(b), "|"))
	if err != nil {
		return nil, err
	}

	if !strings.Contains(message, "success") {
		return nil, fmt.Errorf("%w '%s'", ErrResponse, message)
	}

	return pf.getDNSResolverConfigFile(ctx, configFileReq)
}

func (pf *Client) CreateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	pf.mutexes.DNSResolverConfigFile.Lock()
	defer pf.mutexes.DNSResolverConfigFile.Unlock()

	cf, err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq)
	if err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrCreateOperationFailed, err)
	}
	return cf, nil
}

func (pf *Client) UpdateDNSResolverConfigFile(ctx context.Context, configFileReq ConfigFile) (*ConfigFile, error) {
	pf.mutexes.DNSResolverConfigFile.Lock()
	defer pf.mutexes.DNSResolverConfigFile.Unlock()

	cf, err := pf.createOrUpdateDNSResolverConfigFile(ctx, configFileReq)
	if err != nil {
		return nil, fmt.Errorf("%w config file, %w", ErrUpdateOperationFailed, err)
	}
	return cf, nil
}

func (pf *Client) DeleteDNSResolverConfigFile(ctx context.Context, name string) error {
	var cf ConfigFile
	if err := cf.SetName(name); err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	u := url.URL{Path: "diag_command.php"}
	v := url.Values{
		"txtCommand": {fmt.Sprintf("rm %s", cf.formatFileName())},
		"submit":     {"EXEC"},
	}

	resp, err := pf.do(ctx, http.MethodPost, u, &v)
	if err != nil {
		return fmt.Errorf("%w config file, %w", ErrDeleteOperationFailed, err)
	}

	defer resp.Body.Close()

	return nil
}
