package circleci

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	client "github.com/mrolla/terraform-provider-circleci/circleci/client"
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
	c := m.(*client.Client)

	org := d.Get("organization").(string)
	name := d.Get("name").(string)

	ctx, err := c.GetContextByName(org, name)
	if err != nil {
		return err
	}

	d.SetId(ctx.ID)
	return nil
}
