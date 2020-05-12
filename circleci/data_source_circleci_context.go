package circleci

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceCircleCIContext() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCircleCIContextRead,

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
				Description: "The organization where the context is defined",
			},
		},
	}
}

func dataSourceCircleCIContextRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	org, err := client.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	ctx, err := GetContextByName(
		client.graphql,
		org,
		client.vcs,
		d.Get("name").(string),
	)
	if err != nil {
		if errors.Is(err, ErrContextNotFound) {
			d.SetId("")
			return nil
		}

		return err
	}

	d.SetId(ctx.ID)
	return nil
}
