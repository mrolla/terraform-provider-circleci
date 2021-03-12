package circleci

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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

func resourceCircleCIContextEnvironmentVariableCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	variable := d.Get("variable").(string)
	context := d.Get("context_id").(string)
	value := d.Get("value").(string)

	if err := client.contexts.CreateEnvironmentVariable(context, variable, value); err != nil {
		return fmt.Errorf("error storing environment variable: %w", err)
	}

	d.SetId(variable)

	return resourceCircleCIContextEnvironmentVariableRead(d, m)
}

func resourceCircleCIContextEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	variable := d.Get("variable").(string)
	ctx := d.Get("context_id").(string)

	envs, err := client.EnvironmentVariables(ctx)

	var httpError *api.HTTPError
	if errors.As(err, httpError) && httpError.Code == 404 {
		d.SetId("")
		return nil
	}

	for _, env := range envs {
		if env.Variable == variable {
			return nil
		}
	}

	d.SetId("")

	return nil
}

func resourceCircleCIContextEnvironmentVariableDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	if err := client.contexts.DeleteEnvironmentVariable(d.Get("context_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error deleting environment variable: %w", err)
	}

	return nil
}

func resourceCircleCIContextEnvironmentVariableImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Client)

	value := os.Getenv("CIRCLECI_ENV_VALUE")
	if value == "" {
		return nil, errors.New("CIRCLECI_ENV_VALUE is required to import a context environment variable")
	}
	d.Set("value", value)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return nil, errors.New("importing context environment variables requires $organization/$context/$variable")
	}

	d.Set("variable", parts[2])
	d.SetId(parts[2])

	ctx, err := client.GetContextByIDOrName(parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.Set("context_id", ctx.ID)

	return []*schema.ResourceData{d}, nil
}
