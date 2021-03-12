package client

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/google/uuid"
)

var (
	ErrContextNotFound = errors.New("context not found")
)

// GetContext gets an existing context by its ID (UUID)
func (c *Client) GetContext(id string) (*api.Context, error) {
	req, err := c.rest.NewRequest("GET", &url.URL{Path: fmt.Sprintf("context/%s", id)}, nil)
	if err != nil {
		return nil, err
	}

	ctx := &api.Context{}

	status, err := c.rest.DoRequest(req, ctx)
	if err != nil {
		if status == 404 {
			return nil, ErrContextNotFound
		}

		return nil, err
	}

	return ctx, nil
}

// GetContextByName gets an existing context by its name
func (c *Client) GetContextByName(name, org string) (*api.Context, error) {
	o, err := c.Organization(org)
	if err != nil {
		return nil, err
	}

	return c.contexts.ContextByName(c.vcs, o, name)
}

// GetContextByIDOrName gets a context by ID if a UUID is specified, and by name otherwise
func (c *Client) GetContextByIDOrName(id, org string) (*api.Context, error) {
	if _, uuidErr := uuid.Parse(id); uuidErr != nil {
		return c.GetContext(id)
	} else {
		return c.contexts.ContextByName(c.vcs, org, id)
	}
}

type createContextRequest struct {
	Name  string        `json:"name"`
	Owner *contextOwner `json:"owner"`
}

type contextOwner struct {
	Slug string `json:"slug"`
	Type string `json:"type"`
}

// CreateContext creates a new context and returns the created context object
func (c *Client) CreateContext(org, name string) (*api.Context, error) {
	org, err := c.Organization(org)
	if err != nil {
		return nil, err
	}

	req, err := c.rest.NewRequest("POST", &url.URL{Path: "context"}, &createContextRequest{
		Name: name,
		Owner: &contextOwner{
			Slug: fmt.Sprintf("%s/%s", c.vcs, org),
			Type: "organization",
		},
	})
	if err != nil {
		return nil, err
	}

	ctx := &api.Context{}
	_, err = c.rest.DoRequest(req, ctx)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (c *Client) DeleteContext(id string) error {
	return c.contexts.DeleteContext(id)
}

// CreateContextEnvironmentVariable creates a new context environment variable
func (c *Client) CreateContextEnvironmentVariable(ctx, variable, value string) error {
	return c.contexts.CreateEnvironmentVariable(ctx, variable, value)
}

// ListContextEnvironmentVariables lists all environment variables for a given context
func (c *Client) ListContextEnvironmentVariables(ctx string) (*[]api.EnvironmentVariable, error) {
	return c.contexts.EnvironmentVariables(ctx)
}

// HasContextEnvironmentVariable lists all environment variables for a given context and checks whether the specified variable is defined.
// If either the context or the variable does not exist, it returns false.
func (c *Client) HasContextEnvironmentVariable(ctx, variable string) (bool, error) {
	envs, err := c.ListContextEnvironmentVariables(ctx)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}

		return false, err
	}

	for _, env := range *envs {
		if env.Variable == variable {
			return true, nil
		}
	}

	return false, nil
}

// DeleteContextEnvironmentVariable deletes a context environment variable by context ID and name
func (c *Client) DeleteContextEnvironmentVariable(ctx, variable string) error {
	return c.contexts.DeleteEnvironmentVariable(ctx, variable)
}
