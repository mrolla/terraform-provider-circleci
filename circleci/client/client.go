package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/settings"

	"github.com/mrolla/terraform-provider-circleci/circleci/client/rest"
)

// Client provides access to the CircleCI REST API
// It uses upstream client functionality where possible and defines its own methods as needed
type Client struct {
	contexts     *api.ContextRestClient
	schedules    *api.ScheduleRestClient
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

// New initializes a client object for the provider
func New(config Config) (*Client, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	rootURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	cfg := settings.Config{
		Host:         rootURL,
		RestEndpoint: u.Path,
		Token:        config.Token,
		HTTPClient:   http.DefaultClient,
	}

	contexts, err := api.NewContextRestClient(cfg)
	if err != nil {
		return nil, err
	}

	schedules, err := api.NewScheduleRestClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		rest:      rest.New(rootURL, u.Path, config.Token),
		contexts:  contexts,
		schedules: schedules,

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
