package client

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/rest"
)

// Client provides access to the CircleCI REST API
type Client struct {
	rest         *rest.Client
	contexts     *api.ContextRestClient
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

// NewClient initializes CircleCI API clients (REST and GraphQL) and returns a new client object
func NewClient(config Config) (*Client, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	return &Client{
		rest:     rest.New(u.Host, u.Path, config.Token),
		contexts: api.NewContextRestClient(u.Host, u.Path, u.Token),

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

func (c *Client) Slug(org, project string) (string, error) {
	o, err := c.Organization(org)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", c.vcs, o, project)
}
