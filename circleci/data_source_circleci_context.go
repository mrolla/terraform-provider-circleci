package circleci

import (
	"errors"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func dataSourceCircleCIContextRead(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)

	org := getOrganization(d, providerContext)
	if org == "" {
		return errors.New("organization is required")
	}

	ctx, err := client.GetContextByName(
		providerContext.GraphQLClient,
		org,
		providerContext.VCS,
		d.Get("name").(string),
	)
	if err != nil {
		if errors.Is(err, client.ErrContextNotFound) {
			d.SetId("")
			return nil
		}

		return err
	}

	d.SetId(ctx.ID)
	return nil
}
