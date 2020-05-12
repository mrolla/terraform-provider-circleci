package circleci

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccCircleCIContextDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccOrgProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContextDataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.circleci_context.foo", "name", "terraform-test"),
				),
			},
		},
	})
}

const testAccCircleCIContextDataSource = `
resource "circleci_context" "foo" {
  name = "terraform-test"
}

data "circleci_context" "foo" {
  name = circleci_context.foo.name
}
`
