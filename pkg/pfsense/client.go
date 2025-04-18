package pfsense

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-cleanhttp"
)

const (
	DefaultURL                       = "https://192.168.1.1"
	DefaultUsername                  = "admin"
	DefaultTLSSkipVerify             = false
	DefaultRetryMinWait              = time.Second
	DefaultRetryMaxWait              = 5 * time.Second
	DefaultMaxAttempts               = 3
	DefaultConcurrentWrites          = false
	staticMappingDomainSearchListSep = ";"
	StaticMappingMaxWINSServers      = 2
	StaticMappingMaxDNSServers       = 4
)

type Options struct {
	URL              *url.URL
	Username         string
	Password         string
	TLSSkipVerify    *bool
	RetryMinWait     *time.Duration
	RetryMaxWait     *time.Duration
	MaxAttempts      *int
	ConcurrentWrites *bool // TODO atomic.Bool.
}

type mutexes struct {
	GlobalWrite               sync.Mutex
	DHCPv4Apply               sync.Mutex   // TODO one per iface.
	DHCPv4StaticMapping       sync.RWMutex // TODO one per iface.
	DNSResolverApply          sync.Mutex
	DNSResolverConfigFile     sync.RWMutex
	DNSResolverHostOverride   sync.RWMutex
	DNSResolverDomainOverride sync.RWMutex
	ExecutePHPCommand         sync.RWMutex
	FirewallAlias             sync.RWMutex
	FirewallFilterReload      sync.Mutex
}

type Client struct {
	Options    *Options
	token      string
	tokenKey   string
	httpClient *http.Client
	mutexes    *mutexes
}

func (opts Options) newHTTPClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: *opts.TLSSkipVerify} // #nosec G402

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	return client
}

func (pf *Client) updateToken(doc *goquery.Document) error {
	head := doc.FindMatcher(goquery.Single("head")).Text()
	if head == "" {
		return fmt.Errorf("%w, html head not found", ErrUnableToScrapeHTML)
	}
	tokenKey := regexp.MustCompile(`var csrfMagicName = "([^"]+)";`)
	token := regexp.MustCompile(`var csrfMagicToken = "([^"]+)";`)
	tokenKeyMatches := tokenKey.FindStringSubmatch(head)
	tokenMatches := token.FindStringSubmatch(head)

	if len(tokenKeyMatches) < 1 {
		return fmt.Errorf("%w, token key not found", ErrUnableToScrapeHTML)
	}

	pf.tokenKey = tokenKeyMatches[1]

	if len(tokenMatches) < 1 {
		return fmt.Errorf("%w, token not found", ErrUnableToScrapeHTML)
	}

	pf.token = tokenMatches[1]

	return nil
}

func (pf *Client) read(mutex *sync.RWMutex) func() {
	mutex.RLock()

	return func() {
		mutex.RUnlock()
	}
}

func (pf *Client) write(mutex *sync.RWMutex) func() {
	globalWriteLocking := !*pf.Options.ConcurrentWrites

	if globalWriteLocking {
		pf.mutexes.GlobalWrite.Lock()
	}
	mutex.Lock()

	return func() {
		mutex.Unlock()
		if globalWriteLocking {
			pf.mutexes.GlobalWrite.Unlock()
		}
	}
}

func NewClient(ctx context.Context, opts *Options) (*Client, error) {
	var err error

	if opts.URL.String() == "" {
		url, err := url.Parse(DefaultURL)
		if err != nil {
			return nil, fmt.Errorf("%w, %w", ErrClientValidation, err)
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

	if opts.RetryMinWait == nil {
		td := DefaultRetryMinWait
		opts.RetryMinWait = &td
	}

	if opts.RetryMaxWait == nil {
		td := DefaultRetryMaxWait
		opts.RetryMaxWait = &td
	}

	if opts.MaxAttempts == nil {
		i := DefaultMaxAttempts
		opts.MaxAttempts = &i
	}

	if opts.ConcurrentWrites == nil {
		b := DefaultConcurrentWrites
		opts.ConcurrentWrites = &b
	}

	pfsense := &Client{
		Options:    opts,
		httpClient: opts.newHTTPClient(),
		mutexes:    &mutexes{},
	}

	rootURL := url.URL{Path: "/"}

	// get initial token
	doc, err := pfsense.callHTML(ctx, http.MethodGet, rootURL, nil)
	if err != nil {
		return nil, err
	}

	err = pfsense.updateToken(doc)
	if err != nil {
		return nil, err
	}

	// login
	values := url.Values{
		"usernamefld": {pfsense.Options.Username},
		"passwordfld": {pfsense.Options.Password},
		"login":       {"Sign In"},
	}

	doc, err = pfsense.callHTML(ctx, http.MethodPost, rootURL, &values)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrLoginFailed, err)
	}

	body := doc.FindMatcher(goquery.Single("body"))

	if body.Length() != 1 {
		return nil, fmt.Errorf("%w, html body not found", ErrUnableToScrapeHTML)
	}

	if strings.Contains(body.Text(), "Username or Password incorrect") {
		return nil, fmt.Errorf("%w, username or password incorrect", ErrLoginFailed)
	}

	err = pfsense.updateToken(doc)
	if err != nil {
		return nil, err
	}

	return pfsense, nil
}

func (pf *Client) callHTML(ctx context.Context, method string, relativeURL url.URL, formValues *url.Values) (*goquery.Document, error) {
	resp, err := pf.call(ctx, method, relativeURL, formValues)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	_, _ = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (pf *Client) executePHPCommand(ctx context.Context, command string, value any) error {
	relativeURL := url.URL{Path: "diag_command.php"}
	values := url.Values{
		"txtPHPCommand": {command},
		"submit":        {"EXECPHP"},
	}
	doc, err := pf.callHTML(ctx, http.MethodPost, relativeURL, &values)
	if err != nil {
		return err
	}

	resp := doc.FindMatcher(goquery.Single("pre"))
	if resp.Length() != 1 {
		return fmt.Errorf("%w, php command response not found", ErrUnableToScrapeHTML)
	}

	commandErr := doc.FindMatcher(goquery.Single("#errdiv"))
	if commandErr.Length() == 1 {
		return fmt.Errorf("%w, php command failed, %s", ErrServerValidation, resp.Text())
	}

	err = json.Unmarshal([]byte(resp.Text()), &value)
	if err != nil {
		return fmt.Errorf("%w php command response as JSON, %w", ErrUnableToParse, err)
	}

	return nil
}

func removeEmptyStrings(s []string) []string {
	var results []string
	for _, str := range s {
		if str != "" {
			results = append(results, str)
		}
	}

	return results
}

func durationSeconds(s string) string {
	if s == "" {
		return "0s"
	}

	return fmt.Sprintf("%ss", s)
}

type ValidAddrStringer interface {
	String() string
	IsValid() bool
}

func safeAddrString(v ValidAddrStringer) string {
	if v.IsValid() {
		return v.String()
	}

	return ""
}

func safeSplit(s string, sep string) []string {
	if s == "" {
		return []string{}
	}

	return strings.Split(s, sep)
}
