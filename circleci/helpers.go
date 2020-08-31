package circleci

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func getOrganization(d *schema.ResourceData, providerContext ProviderContext) string {
	organization, ok := d.GetOk("organization")
	if ok {
		org := organization.(string)
		return org
	}

	return providerContext.Org
}
