package client

import (
	"errors"
	"fmt"

	"github.com/CircleCI-Public/circleci-cli/api"
	graphql "github.com/CircleCI-Public/circleci-cli/client"
)

var (
	ErrContextNotFound = errors.New("context not found")
)

// GetContextByIDOrName lists all contexts and returns one with a matching ID or name, if found
func GetContextByIDOrName(client *graphql.Client, org string, vcs string, value string) (*api.CircleCIContext, error) {
	ctx, err := GetContextByID(client, org, vcs, value)
	if err != nil && !errors.Is(err, ErrContextNotFound) {
		return nil, err
	}
	if ctx != nil {
		return ctx, nil
	}

	ctx, err = GetContextByName(client, org, vcs, value)
	if err != nil && !errors.Is(err, ErrContextNotFound) {
		return nil, err
	}
	if ctx != nil {
		return ctx, nil
	}

	return nil, fmt.Errorf("%w: no context with ID or name '%s' in organization '%s'", ErrContextNotFound, value, org)
}

// GetContextByName lists all contexts and returns one with a matching name, if found
func GetContextByName(client *graphql.Client, org string, vcs string, name string) (*api.CircleCIContext, error) {
	res, err := api.ListContexts(client, org, vcs)
	if err != nil {
		return nil, err
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.Name == name {
			return &context.Node, nil
		}
	}

	return nil, fmt.Errorf("%w: no context with name '%s' in organization '%s'", ErrContextNotFound, name, org)
}

// GetContextByID lists all contexts and returns one with a matching ID, if found
func GetContextByID(client *graphql.Client, org string, vcs string, ID string) (*api.CircleCIContext, error) {
	res, err := api.ListContexts(client, org, vcs)
	if err != nil {
		return nil, err
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.ID == ID {
			return &context.Node, nil
		}
	}

	return nil, fmt.Errorf("%w: no context with ID '%s' in organization '%s'", ErrContextNotFound, ID, org)
}
