package circleci

import (
	"fmt"
	"net/url"

	circleciapi "github.com/jszwedko/go-circleci"
)

// ProviderClient is a thin commodity wrapper on top of circleciapi
type ProviderClient struct {
	client       *circleciapi.Client
	vcsType      string
	organization *string
}

// NewConfig initialize circleci API client and returns a new config object
func NewConfig(token, vscType, baseURL string) (*ProviderClient, error) {
	parsedUrl, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &ProviderClient{
		client: &circleciapi.Client{
			BaseURL: parsedUrl,
			Token:   token,
		},
		vcsType: vscType,
	}, nil
}

// NewOrganizationConfig initialize circleci API with an organization client and returns a new config object
func NewOrganizationConfig(token, vscType, organization, baseURL string) (*ProviderClient, error) {
	parsedUrl, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &ProviderClient{
		client: &circleciapi.Client{
			BaseURL: parsedUrl,
			Token:   token,
		},
		organization: &organization,
		vcsType:      vscType,
	}, nil
}

// GetEnvVar get the environment variable with given name
// It returns an empty structure if no environment variable exists with that name
func (pv *ProviderClient) GetEnvVar(organization *string, projectName, envVarName string) (*circleciapi.EnvVar, error) {
	org, err := pv.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return nil, err
	}

	return pv.client.GetEnvVar(pv.vcsType, *org, projectName, envVarName)
}

// EnvVarExists check if environment variable exists with given name
func (pv *ProviderClient) EnvVarExists(organization *string, projectName, envVarName string) (bool, error) {
	org, err := pv.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return false, err
	}

	envVar, err := pv.client.GetEnvVar(pv.vcsType, *org, projectName, envVarName)
	if err != nil {
		return false, err
	}
	return bool(envVar.Name != ""), nil
}

// AddEnvVar create an environment variable with given name and value
func (pv *ProviderClient) AddEnvVar(organization *string, projectName, envVarName, envVarValue string) (*circleciapi.EnvVar, error) {
	org, err := pv.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return nil, err
	}

	return pv.client.AddEnvVar(pv.vcsType, *org, projectName, envVarName, envVarValue)
}

// DeleteEnvVar delete the environment variable with given name
func (pv *ProviderClient) DeleteEnvVar(organization *string, projectName, envVarName string) error {
	org, err := pv.validateOrganization(organization, projectName, envVarName)
	if err != nil {
		return err
	}

	return pv.client.DeleteEnvVar(pv.vcsType, *org, projectName, envVarName)
}

func (pv *ProviderClient) validateOrganization(organization *string, projectName, envVarName string) (*string, error) {
	if organization == nil && pv.organization == nil {
		return nil, fmt.Errorf("organization has not been set for environment variable %s in project %s", projectName, envVarName)
	}

	if organization != nil {
		return organization, nil
	}

	return pv.organization, nil

}
