package circleci

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCircleCIContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIContextCreate,
		Read:   resourceCircleCIContextRead,
		Delete: resourceCircleCIContextDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCircleCIContextImport,
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

	ctx, err := client.CreateContext(org, name)
	if err != nil {
		return fmt.Errorf("error creating context: %w", err)
	}

	d.SetId(ctx.ID)
	return resourceCircleCIContextRead(d, m)
}

func resourceCircleCIContextRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	ctx, err := client.GetContext(d.Id())
	if err != nil {
		var httpError *api.HTTPError
		if errors.Is(err, httpError) && httpError.Code == 404 {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", ctx.Name)
	return nil
}

func resourceCircleCIContextDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	if err := client.contexts.DeleteContext(d.Id()); err != nil {
		return fmt.Errorf("error deleting context: %w", err)
	}

	return nil
}

func resourceCircleCIContextImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*Client)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, errors.New("importing context requires $organization/$context")
	}

	ctx, err := client.GetContextByIDOrName(parts...)
	if err != nil {
		return nil, err
	}

	d.SetId(ctx.ID)

	return []*schema.ResourceData{d}, nil
}
