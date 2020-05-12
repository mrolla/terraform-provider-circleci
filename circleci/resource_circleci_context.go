package circleci

import (
	"errors"
	"fmt"
	"strings"
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
			State: resourceCircleCIContextImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
		return fmt.Errorf("error creating context: %w", err)
	}

	return resourceCircleCIContextRead(d, m)
}

func resourceCircleCIContextRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	ctx, err := GetContextByID(client.graphql, org, client.vcs, d.Id())
	if err != nil {
		return err
	}

	d.Set("name", ctx.Name)
	return nil
}

func resourceCircleCIContextDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	if err := api.DeleteContext(client.graphql, d.Id()); err != nil {
		return fmt.Errorf("error deleting context: %w", err)
	}

	return nil
}

func resourceCircleCIContextExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Client)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return false, err
	}

	_, err = GetContextByID(
		client.graphql,
		org,
		client.vcs,
		d.Id(),
	)
	if err != nil {
		if errors.Is(err, ErrContextNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func resourceCircleCIContextImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Client)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, errors.New("importing context requires $organization/$context")
	}

	d.Set("organization", parts[0])

	ctx, err := GetContextByIDOrName(client.graphql, parts[0], client.vcs, parts[1])
	if err != nil {
		return nil, err
	}
	d.Set("name", ctx.Name)
	d.SetId(ctx.ID)

	return []*schema.ResourceData{d}, nil
}
