package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/settings"

	"github.com/mrolla/terraform-provider-circleci/circleci/client/rest"
)

// Client provides access to the CircleCI REST API
// It uses upstream client functionality where possible and defines its own methods as needed
type Client struct {
	contexts     *api.ContextRestClient
	rest         *rest.Client
	vcs          string
	organization string
}

// Config configures a Client
type Config struct {
	URL   string
	Token string

	VCS          string
	Organization string
}

// retryTransport wraps an http.RoundTripper and adds retry logic for rate limiting
type retryTransport struct {
	transport http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	const (
		maxRetries     = 5
		initialBackoff = 1 * time.Second
		maxBackoff     = 32 * time.Second
	)

	var resp *http.Response
	var err error
	var bodyBytes []byte

	// Save the request body if it exists
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Restore the request body for each attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = t.transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// Check for rate limiting (429)
		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			// Close the response body before retrying
			resp.Body.Close()

			// Calculate exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt))) * initialBackoff
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			time.Sleep(backoff)
			continue
		}

		// Not a rate limit error or max retries reached
		break
	}

	return resp, err
}

// newHTTPClientWithRetry creates an HTTP client with retry logic for rate limiting
func newHTTPClientWithRetry() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &retryTransport{
			transport: http.DefaultTransport,
		},
	}
}

// New initializes a client object for the provider
func New(config Config) (*Client, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	rootURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	contexts, err := api.NewContextRestClient(settings.Config{
		Host:         rootURL,
		RestEndpoint: u.Path,
		Token:        config.Token,
		HTTPClient:   newHTTPClientWithRetry(),
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		rest:     rest.New(rootURL, u.Path, config.Token),
		contexts: contexts,

		vcs:          config.VCS,
		organization: config.Organization,
	}, nil
}

// Organization returns the organization for a request. If an organization is provided,
// that is returned. Next, an organization configured in the provider is returned.
// If neither are set, an error is returned.
func (c *Client) Organization(org string) (string, error) {
	if org != "" {
		return org, nil
	}

	if c.organization != "" {
		return c.organization, nil
	}

	return "", errors.New("organization is required")
}

// Slug returns a project slug, including the VCS, organization, and project names
func (c *Client) Slug(org, project string) (string, error) {
	o, err := c.Organization(org)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", c.vcs, o, project), nil
}

func isNotFound(err error) bool {
	var httpError *rest.HTTPError
	if errors.As(err, &httpError) && httpError.Code == 404 {
		return true
	}

	return false
}
