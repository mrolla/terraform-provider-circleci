package client

import (
	"errors"
	"fmt"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/google/uuid"
)

var (
	ErrContextNotFound = errors.New("context not found")
)

func (c *Client) GetContext(id string) (*api.Context, error) {
	req, err := c.rest.NewRequest("GET", &url.URL{Path: fmt.Sprintf("context/%s", id)}, nil)
	if err != nil {
		return nil, err
	}

	ctx := &api.Context{}

	_, err := c.rest.DoRequest(req, ctx)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (c *Client) GetContextByIDOrName(id, org string) (*api.Context, error) {
	if _, uuidErr := uuid.Parse(id); uuidErr != nil {
		return client.GetContext(id)
	} else {
		return client.contexts.ContextByName(client.vcs, org, id)
	}
}

type CreateContextRequest struct {
	Name  string `json:"name"`
	Owner *ContextOwner `json:"owner"`
}

type ContextOwner struct {
	Slug string `json:"slug"`
	Type string `json:"type"`
}

func (c *Client) CreateContext(org, name string) (*api.Context, error) {
	req, err := c.rest.NewRequest("POST", &url.URL{Path: fmt.Sprintf("context/%s", id)}, &CreateContextRequest{
		Name: name,
		Owner: &ContextOwner{
			Slug: fmt.Sprintf("%s/%s", c.vcs, org),
			Type: "organization",
		}
	})
	if err != nil {
		return nil, err
	}

	ctx := &api.Context{}
	_, err := c.rest.DoRequest(req, ctx)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}
