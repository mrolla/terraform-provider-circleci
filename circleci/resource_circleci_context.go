package circleci

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	client "github.com/mrolla/terraform-provider-circleci/circleci/client"
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
	c := m.(*client.Client)

	name := d.Get("name").(string)
	org := d.Get("organization").(string)

	ctx, err := c.CreateContext(org, name)
	if err != nil {
		return fmt.Errorf("error creating context: %w", err)
	}

	d.SetId(ctx.ID)
	return resourceCircleCIContextRead(d, m)
}

func resourceCircleCIContextRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	ctx, err := c.GetContext(d.Id())
	if err != nil {
		if errors.Is(err, client.ErrContextNotFound) {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", ctx.Name)
	return nil
}

func resourceCircleCIContextDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	if err := c.DeleteContext(d.Id()); err != nil {
		return fmt.Errorf("error deleting context: %w", err)
	}

	return nil
}

func resourceCircleCIContextImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, errors.New("importing context requires $organization/$context")
	}

	ctx, err := c.GetContextByIDOrName(parts[0], parts[1])
	if err != nil {
		return nil, err
	}

	d.SetId(ctx.ID)

	return []*schema.ResourceData{d}, nil
}
