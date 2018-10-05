package circleci

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vcs_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"circleci_environment_variable": resourceCircleCIEnvironmentVariable(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	token := d.Get("api_token").(string)
	vcsType := d.Get("vcs_type").(string)
	organization := d.Get("organization").(string)

	return NewClient(token, vcsType, organization)
}
