package pfsense

import (
	"errors"
	"fmt"
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
	DefaultURL            = "http://192.168.1.1"
	DefaultUsername       = "admin"
	clientName            = "go-pfsense"
	descriptionSeparator  = "|"
)

type Options struct {
	URL      *url.URL
	Username string
	Password string
}

type mutexes struct {
	Version                   sync.Mutex
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
		return errors.New("html not valid")
	}
	tokenKey := regexp.MustCompile(`var csrfMagicName = "([^"]+)";`)
	token := regexp.MustCompile(`var csrfMagicToken = "([^"]+)";`)
	tokenKeyMatches := tokenKey.FindStringSubmatch(head)
	tokenMatches := token.FindStringSubmatch(head)

	if len(tokenKeyMatches) < 1 {
		return errors.New("token key not found")
	}

	pf.tokenKey = tokenKeyMatches[1]

	if len(tokenMatches) < 1 {
		return errors.New("token not found")
	}

	pf.token = tokenMatches[1]

	return nil
}

func NewClient(opts *Options) (*Client, error) {
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
		return nil, errors.New("password required")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: DefaultTimeout,
			}).Dial,
			TLSHandshakeTimeout: DefaultTimeout,
		},
	}

	pf := &Client{
		Options:    opts,
		httpClient: client,
		mutexes:    &mutexes{},
	}

	u := pf.Options.URL.ResolveReference(&url.URL{Path: "/"})

	// get initial token
	resp, err := pf.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get homepage, %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	err = pf.updateToken(doc)
	if err != nil {
		return nil, err
	}

	// login
	resp, err = pf.httpClient.PostForm(u.String(), url.Values{
		pf.tokenKey:   {pf.token},
		"usernamefld": {pf.Options.Username},
		"passwordfld": {pf.Options.Password},
		"login":       {"Sign In"},
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	body := doc.FindMatcher(goquery.Single("body"))

	if body.Length() == 0 {
		return nil, errors.New("html not valid")
	}

	if strings.Contains(body.Text(), "Username or Password incorrect") {
		return nil, errors.New("login failed")
	}

	err = pf.updateToken(doc)

	if err != nil {
		return nil, err
	}

	return pf, nil
}

func parseDescription(rawDescription string) (string, string, error) {
	parts := strings.Split(rawDescription, descriptionSeparator)
	if len(parts) != 3 || parts[0] != clientName {
		return "", "", errors.New("not managed by client")
	}
	return parts[1], parts[2], nil
}

func formatDescription(id string, description string) string {
	parts := []string{clientName, id, description}
	return strings.Join(parts, descriptionSeparator)
}
