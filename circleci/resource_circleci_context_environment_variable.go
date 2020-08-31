package circleci

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCircleCIContextEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIContextEnvironmentVariableCreate,
		Read:   resourceCircleCIContextEnvironmentVariableRead,
		Delete: resourceCircleCIContextEnvironmentVariableDelete,
		Exists: resourceCircleCIContextEnvironmentVariableExists,
		Importer: &schema.ResourceImporter{
			State: resourceCircleCIContextEnvironmentVariableImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"variable": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the environment variable",
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
				StateFunc: func(value interface{}) string {
					return hashString(value.(string))
				},
				Description: "The value that will be set for the environment variable.",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the context where the environment variable is defined",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The organization where the context is defined",
			},
		},
	}
}

func resourceCircleCIContextEnvironmentVariableCreate(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	variable := d.Get("variable").(string)
	context := d.Get("context_id").(string)
	value := d.Get("value").(string)

	if err := api.StoreEnvironmentVariable(gqlClient, context, variable, value); err != nil {
		return fmt.Errorf("error storing environment variable: %w", err)
	}

	d.SetId(variable)

	return resourceCircleCIContextEnvironmentVariableRead(d, meta)
}

func resourceCircleCIContextEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCircleCIContextEnvironmentVariableDelete(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	if err := api.DeleteEnvironmentVariable(gqlClient, d.Get("context_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error deleting environment variable: %w", err)
	}

	return nil
}

func resourceCircleCIContextEnvironmentVariableExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient
	variable := d.Get("variable").(string)

	org := getOrganization(d, providerContext)
	if org == "" {
		return false, errors.New("organization is required")
	}

	ctx, err := client.GetContextByID(
		gqlClient,
		org,
		providerContext.VCS,
		d.Get("context_id").(string),
	)
	if err != nil {
		if errors.Is(err, client.ErrContextNotFound) {
			return false, nil
		}

		return false, err
	}

	for _, env := range ctx.Resources {
		if env.Variable == variable {
			return true, nil
		}
	}

	return false, nil
}

func resourceCircleCIContextEnvironmentVariableImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	value := os.Getenv("CIRCLECI_ENV_VALUE")
	if value == "" {
		return nil, errors.New("CIRCLECI_ENV_VALUE is required to import a context environment variable")
	}
	d.Set("value", value)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, errors.New("importing context environtment variables requires $organization/$context/$variable")
	}

	d.Set("variable", parts[2])
	d.SetId(parts[2])

	ctx, err := client.GetContextByIDOrName(gqlClient, parts[0], providerContext.VCS, parts[1])
	if err != nil {
		return nil, err
	}
	d.Set("context_id", ctx.ID)

	return []*schema.ResourceData{d}, nil
}
