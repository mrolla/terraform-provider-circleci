package client

import (
	"fmt"
	"net/url"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/google/uuid"
)

// GetContext gets an existing context by its ID (UUID)
func (c *Client) GetContext(id string) (*api.Context, error) {
	req, err := c.rest.NewRequest("GET", &url.URL{Path: fmt.Sprintf("context/%s", id)}, nil)
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
