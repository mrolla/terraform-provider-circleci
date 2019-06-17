package circleci

import (
	"net/url"

	circleciapi "github.com/jszwedko/go-circleci"
)

// ProviderClient is a thin commodity wrapper on top of circleciapi
type ProviderClient struct {
	client  *circleciapi.Client
	vcsType string
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

// GetEnvVar get the environment variable with given name
// It returns an empty structure if no environment variable exists with that name
func (pv *ProviderClient) GetEnvVar(organization, projectName, envVarName string) (*circleciapi.EnvVar, error) {
	return pv.client.GetEnvVar(pv.vcsType, organization, projectName, envVarName)
}

// EnvVarExists check if environment variable exists with given name
func (pv *ProviderClient) EnvVarExists(organization, projectName, envVarName string) (bool, error) {
	envVar, err := pv.client.GetEnvVar(pv.vcsType, organization, projectName, envVarName)
	if err != nil {
		return false, err
	}
	return bool(envVar.Name != ""), nil
}

// AddEnvVar create an environment variable with given name and value
func (pv *ProviderClient) AddEnvVar(organization, projectName, envVarName, envVarValue string) (*circleciapi.EnvVar, error) {
	return pv.client.AddEnvVar(pv.vcsType, organization, projectName, envVarName, envVarValue)
}

// DeleteEnvVar delete the environment variable with given name
func (pv *ProviderClient) DeleteEnvVar(organization, projectName, envVarName string) error {
	return pv.client.DeleteEnvVar(pv.vcsType, organization, projectName, envVarName)
}
