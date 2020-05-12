package circleci

import (
	"errors"
	"fmt"
	"time"

	"github.com/CircleCI-Public/circleci-cli/api"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCircleCIContextEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIContextEnvironmentVariableCreate,
		Read:   resourceCircleCIContextEnvironmentVariableRead,
		Delete: resourceCircleCIContextEnvironmentVariableDelete,
		Exists: resourceCircleCIContextEnvironmentVariableExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The value that will be set for the environment variable. This will be displayed as plain text in a plan.",
			},
			"sensitive_value": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"value"},
				Description:   "The value that will be set for the environment variable. This will be hidden as sensitive during a plan.",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the context where the environment variable is defined",
			},
		},
	}
}

func resourceCircleCIContextEnvironmentVariableCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	variable := d.Get("variable").(string)
	context := d.Get("context_id").(string)

	var value string
	if v, ok := d.GetOk("sensitive_value"); ok {
		value = v.(string)
	} else {
		value = d.Get("value").(string)
	}

	if value == "" {
		return errors.New("one of 'value' or 'sensitive_value' is required")
	}

	if err := api.StoreEnvironmentVariable(client.graphql, context, variable, value); err != nil {
		return fmt.Errorf("error storing environment variable: %v", err)
	}

	return resourceCircleCIContextEnvironmentVariableRead(d, m)
}

func resourceCircleCIContextEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	variable := d.Get("variable").(string)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	res, err := api.ListContexts(client.graphql, org, client.vcs)
	if err != nil {
		return fmt.Errorf("error listing contexts: %v", err)
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.ID == d.Id() {
			for _, env := range context.Node.Resources {
				if env.Variable == variable {
					d.SetId(env.Variable)
					return nil
				}
			}
		}
	}

	d.SetId("")
	return nil
}

func resourceCircleCIContextEnvironmentVariableDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	if err := api.DeleteEnvironmentVariable(client.graphql, d.Get("context_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error deleting environment variable: %v", err)
	}

	return nil
}

func resourceCircleCIContextEnvironmentVariableExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Client)
	variable := d.Get("variable").(string)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return false, err
	}

	res, err := api.ListContexts(client.graphql, org, client.vcs)
	if err != nil {
		return false, fmt.Errorf("error listing contexts: %v", err)
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.ID == d.Id() {
			for _, env := range context.Node.Resources {
				if env.Variable == variable {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
