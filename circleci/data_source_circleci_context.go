package circleci

import (
	"fmt"

	"github.com/CircleCI-Public/circleci-cli/api"

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

	res, err := api.ListContexts(client.graphql, org, client.vcs)
	if err != nil {
		return fmt.Errorf("error listing contexts: %v", err)
	}

	for _, context := range res.Organization.Contexts.Edges {
		if context.Node.Name == d.Get("name").(string) {
			d.SetId(context.Node.ID)
			return nil
		}
	}

	d.SetId("")
	return nil
}
