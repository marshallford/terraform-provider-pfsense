package pfsense

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	DefaultTimeout        = 30 * time.Second
	DefaultConnectTimeout = 5 * time.Second
	DefaultMaxIdleConns   = 10
	DefaultURL            = "https://192.168.1.1"
	DefaultUsername       = "admin"
	DefaultTLSSkipVerify  = false
)

type Options struct {
	URL           *url.URL
	Username      string
	Password      string
	TLSSkipVerify *bool
}

type mutexes struct {
	DNSResolverApply          sync.Mutex
	DNSResolverHostOverride   sync.Mutex
	DNSResolverDomainOverride sync.Mutex
}

type Client struct {
	Options    *Options
	token      string
	tokenKey   string
	httpClient *http.Client
	mutexes    *mutexes
}

func (pf *Client) updateToken(doc *goquery.Document) error {
	head := doc.FindMatcher(goquery.Single("head")).Text()
	if head == "" {
		return ErrUnableToScrapeHTML
	}
	tokenKey := regexp.MustCompile(`var csrfMagicName = "([^"]+)";`)
	token := regexp.MustCompile(`var csrfMagicToken = "([^"]+)";`)
	tokenKeyMatches := tokenKey.FindStringSubmatch(head)
	tokenMatches := token.FindStringSubmatch(head)

	if len(tokenKeyMatches) < 1 {
		return fmt.Errorf("%w, token key not found", ErrLoginFailed)
	}

	pf.tokenKey = tokenKeyMatches[1]

	if len(tokenMatches) < 1 {
		return fmt.Errorf("%w, token not found", ErrLoginFailed)
	}

	pf.token = tokenMatches[1]

	return nil
}

func NewClient(ctx context.Context, opts *Options) (*Client, error) {
	var err error

	if opts.URL.String() == "" {
		url, err := url.Parse(DefaultURL)

		if err != nil {
			return nil, err
		}

		opts.URL = url
	}

	if opts.Username == "" {
		opts.Username = DefaultUsername
	}

	if opts.Password == "" {
		return nil, fmt.Errorf("%w, password required", ErrClientValidation)
	}

	if opts.TLSSkipVerify == nil {
		b := DefaultTLSSkipVerify
		opts.TLSSkipVerify = &b
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// #nosec G402
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   DefaultTimeout,
				KeepAlive: DefaultTimeout,
			}).DialContext,
			MaxIdleConns:    DefaultMaxIdleConns,
			IdleConnTimeout: DefaultTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *opts.TLSSkipVerify,
			},
			TLSHandshakeTimeout: DefaultConnectTimeout,
		},
	}

	pf := &Client{
		Options:    opts,
		httpClient: client,
		mutexes:    &mutexes{},
	}

	u := url.URL{Path: "/"}

	// get initial token
	doc, err := pf.doHTML(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	err = pf.updateToken(doc)
	if err != nil {
		return nil, err
	}

	// login
	v := url.Values{
		"usernamefld": {pf.Options.Username},
		"passwordfld": {pf.Options.Password},
		"login":       {"Sign In"},
	}

	doc, err = pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrLoginFailed, err)
	}

	body := doc.FindMatcher(goquery.Single("body"))

	if body.Length() != 1 {
		return nil, fmt.Errorf("%w, %w", ErrLoginFailed, ErrUnableToScrapeHTML)
	}

	if strings.Contains(body.Text(), "Username or Password incorrect") {
		return nil, fmt.Errorf("%w, username or password incorrect", ErrLoginFailed)
	}

	err = pf.updateToken(doc)
	if err != nil {
		return nil, err
	}

	return pf, nil
}

func (pf *Client) do(ctx context.Context, method string, relativeURL url.URL, values *url.Values) (*http.Response, error) {
	var reqBody io.Reader
	if values != nil {
		if pf.tokenKey != "" && pf.token != "" {
			values.Set(pf.tokenKey, pf.token)
		}
		reqBody = strings.NewReader(values.Encode())
	}

	url := pf.Options.URL.ResolveReference(&relativeURL).String()
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create request, %s %s %w", method, relativeURL.String(), err)
	}

	req.Header.Set("User-Agent", "go-pfsense")

	if values != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := pf.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request, %s %s %w", method, relativeURL.String(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w, %s %s %s", ErrResponse, resp.Status, method, relativeURL.String())
	}

	return resp, nil
}

func (pf *Client) doHTML(ctx context.Context, method string, relativeURL url.URL, values *url.Values) (*goquery.Document, error) {
	resp, err := pf.do(ctx, method, relativeURL, values)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (pf *Client) doPHPCommand(ctx context.Context, command string) ([]byte, error) {
	u := url.URL{Path: "diag_command.php"}
	v := url.Values{
		"txtPHPCommand": {command},
		"submit":        {"EXECPHP"},
	}
	doc, err := pf.doHTML(ctx, http.MethodPost, u, &v)
	if err != nil {
		return nil, err
	}

	resp := doc.FindMatcher(goquery.Single("pre"))

	if resp.Length() != 1 {
		return nil, fmt.Errorf("%w, php command response not found", ErrUnableToScrapeHTML)
	}

	return []byte(resp.Text()), nil
}

func (pf *Client) getConfigJSON(ctx context.Context, value string) (json.RawMessage, error) {
	resp, err := pf.doPHPCommand(ctx, fmt.Sprintf("print_r(json_encode($config%s));", value))
	if err != nil {
		return nil, err
	}

	if !json.Valid(resp) {
		return nil, fmt.Errorf("%w php command response as JSON, %w", ErrUnableToParse, err)
	}

	return resp, nil
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
