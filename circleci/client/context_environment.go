package client

import (
	"errors"

	"github.com/CircleCI-Public/circleci-cli/api"
)

// CreateOrUpdateContextEnvironmentVariable creates a new context environment variable
func (c *Client) CreateOrUpdateContextEnvironmentVariable(ctx, variable, value string) error {
	// CreateEnvironmentVariable calls PUT and can be used to update an existing variable with a matching context/name
	return c.contexts.CreateEnvironmentVariable(ctx, variable, value)
}

// ListContextEnvironmentVariables lists all environment variables for a given context
func (c *Client) ListContextEnvironmentVariables(ctx string) (*[]api.EnvironmentVariable, error) {
	return c.contexts.EnvironmentVariables(ctx)
}

// HasContextEnvironmentVariable lists all environment variables for a given context and checks whether the specified variable is defined.
// If either the context or the variable does not exist, it returns false.
func (c *Client) HasContextEnvironmentVariable(ctx, variable string) (bool, error) {
	if _, err := c.GetContext(ctx); err != nil {
		if errors.Is(err, ErrContextNotFound) {
			return false, nil
		}

		return false, err
	}

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
