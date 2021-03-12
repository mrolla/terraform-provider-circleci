package circleci

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

// GetEnvVar get the environment variable with given name
// It returns an empty structure if no environment variable exists with that name
func (c *Client) GetEnvVar(organization string, projectName, envVarName string) (*restclient.EnvVar, error) {
	org, err := c.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return nil, err
	}

	return c.rest.GetEnvVar(c.vcs, org, projectName, envVarName)
}

// EnvVarExists check if environment variable exists with given name
func (c *Client) EnvVarExists(organization string, projectName, envVarName string) (bool, error) {
	org, err := c.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return false, err
	}

	envVar, err := c.rest.GetEnvVar(c.vcs, org, projectName, envVarName)
	if err != nil {
		return false, err
	}
	return bool(envVar.Name != ""), nil
}

// AddEnvVar create an environment variable with given name and value
func (c *Client) AddEnvVar(organization string, projectName, envVarName, envVarValue string) (*restclient.EnvVar, error) {
	org, err := c.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return nil, err
	}

	return c.rest.AddEnvVar(c.vcs, org, projectName, envVarName, envVarValue)
}

// DeleteEnvVar delete the environment variable with given name
func (c *Client) DeleteEnvVar(organization string, projectName, envVarName string) error {
	org, err := c.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return err
	}

	return c.rest.DeleteEnvVar(c.vcs, org, projectName, envVarName)
}

func (c *Client) validateOrganization(organization string, projectName, envVarName string) (string, error) {
	if organization == "" && c.organization == "" {
		return "", fmt.Errorf("organization has not been set for environment variable %s in project %s", projectName, envVarName)
	}

	if organization != "" {
		return organization, nil
	}

	return c.organization, nil

}
