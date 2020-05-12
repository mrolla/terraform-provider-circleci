package circleci

import (
	"fmt"
	"time"

	"github.com/CircleCI-Public/circleci-cli/api"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCircleCIContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIContextCreate,
		Read:   resourceCircleCIContextRead,
		Delete: resourceCircleCIContextDelete,
		Exists: resourceCircleCIContextExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the context",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The organization where the context will be created",
			},
		},
	}
}

func resourceCircleCIContextCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	name := d.Get("name").(string)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	if err := api.CreateContext(client.graphql, client.vcs, org, name); err != nil {
		return fmt.Errorf("error creating context: %v", err)
	}

	return resourceCircleCIContextRead(d, m)
}

func resourceCircleCIContextRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

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
			d.Set("name", context.Node.Name)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceCircleCIContextDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	if err := api.DeleteContext(client.graphql, d.Id()); err != nil {
		return fmt.Errorf("error deleting context: %v", err)
	}

	return nil
}

func resourceCircleCIContextExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Client)
	name := d.Get("name").(string)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return false, err
	}

	res, err := api.ListContexts(client.graphql, org, client.vcs)
	if err != nil {
		return false, fmt.Errorf("error listing contexts: %v", err)
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.Name == name {
			return true, nil
		}
	}

	return false, nil
}
