package circleci

import (
	"net/http"
	"net/url"
	"os"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	cciclient "github.com/CircleCI-Public/circleci-cli/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"go.uber.org/zap"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_TOKEN", nil),
				Description: "The token key for API operations.",
			},
			"vcs_type": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_VCS_TYPE", "github"),
				Description: "The VCS type for the organization.",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_ORGANIZATION", nil),
				Description: "The CircleCI organization.",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_URL", "https://circleci.com/api/v1.1/"),
				Description: "The URL of the Circle CI API.",
			},
			"graphql_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_GRAPHQL_URL", "https://circleci.com/graphql-unstable"),
				Description: "The URL of the CircleCI GraphQL API",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"circleci_environment_variable":         resourceCircleCIEnvironmentVariable(),
			"circleci_context":                      resourceCircleCIContext(),
			"circleci_context_environment_variable": resourceCircleCIContextEnvironmentVariable(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"circleci_context": dataSourceCircleCIContext(),
		},
		ConfigureFunc: providerConfigure,
	}
}

type ProviderContext struct {
	Client        *client.Client
	VCS           string
	Org           string
	GraphQLClient *cciclient.Client
	Logger        *zap.Logger
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	ctx := ProviderContext{
		Org: d.Get("organization").(string),
		VCS: d.Get("vcs_type").(string),
	}
	token := d.Get("api_token").(string)
	baseURL := d.Get("url").(string)
	graphqlURL := d.Get("graphql_url").(string)

	graphqlParsedURL, err := url.Parse(graphqlURL)
	if err != nil {
		return nil, err
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	logConfig := zap.NewProductionConfig()
	tfAccDebug := os.Getenv("TF_ACC_DEBUG")
	if tfAccDebug != "" {
		logConfig = zap.NewDevelopmentConfig()
	}

	logger, err := logConfig.Build()
	if err != nil {
		return nil, err
	}
	ctx.Logger = logger

	client := client.NewClient(
		logger,
		token,
		http.DefaultClient,
		client.WithBaseURL(parsedBaseURL),
	)
	ctx.Client = client

	graphqlClient := cciclient.NewClient(
		(&url.URL{Host: graphqlParsedURL.Host, Scheme: graphqlParsedURL.Scheme}).String(),
		graphqlParsedURL.Path,
		token,
		false,
	)
	ctx.GraphQLClient = graphqlClient

	return ctx, nil
}
