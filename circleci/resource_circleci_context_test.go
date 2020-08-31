package circleci

import (
	"fmt"
	"testing"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCircleCIContext_basic(t *testing.T) {
	context := &api.CircleCIContext{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContext_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextExists("circleci_context.foo", context),
					testAccCheckCircleCIContextAttributes_basic(context),
					resource.TestCheckResourceAttr("circleci_context.foo", "name", "terraform-test"),
				),
			},
		},
	})
}

func TestAccCircleCIContext_update(t *testing.T) {
	context := &api.CircleCIContext{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContext_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextExists("circleci_context.foo", context),
					testAccCheckCircleCIContextAttributes_basic(context),
					resource.TestCheckResourceAttr("circleci_context.foo", "name", "terraform-test"),
				),
			},
			{
				Config: testAccCircleCIContext_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextExists("circleci_context.foo", context),
					testAccCheckCircleCIContextAttributes_update(context),
					resource.TestCheckResourceAttr("circleci_context.foo", "name", "terraform-test-updated"),
				),
			},
		},
	})
}

func TestAccCircleCIContext_import(t *testing.T) {
	context := &api.CircleCIContext{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContext_basic,
				Check:  testAccCheckCircleCIContextExists("circleci_context.foo", context),
			},
			{
				ResourceName: "circleci_context.foo",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf(
						"%s/%s",
						testAccOrgProvider.Meta().(ProviderContext).Org,
						context.ID,
					), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCircleCIContext_import_name(t *testing.T) {
	context := &api.CircleCIContext{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContext_basic,
				Check:  testAccCheckCircleCIContextExists("circleci_context.foo", context),
			},
			{
				ResourceName: "circleci_context.foo",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf(
						"%s/%s",
						testAccOrgProvider.Meta().(ProviderContext).Org,
						"terraform-test",
					), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCircleCIContextExists(addr string, context *api.CircleCIContext) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerContext := testAccOrgProvider.Meta().(ProviderContext)

		resource, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("Not found: %s", addr)
		}
		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ctx, err := client.GetContextByID(
			providerContext.GraphQLClient,
			providerContext.Org,
			providerContext.VCS,
			resource.Primary.ID,
		)
		if err != nil {
			return fmt.Errorf("error getting context: %w", err)
		}

		*context = *ctx

		return nil
	}
}

func testAccCheckCircleCIContextDestroy(s *terraform.State) error {
	providerContext := testAccOrgProvider.Meta().(ProviderContext)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "circleci_context" {
			continue
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := client.GetContextByID(
			providerContext.GraphQLClient,
			providerContext.Org,
			providerContext.VCS,
			resource.Primary.ID,
		)
		if err == nil {
			return fmt.Errorf("Context %s still exists: %w", resource.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckCircleCIContextAttributes_basic(context *api.CircleCIContext) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if context.Name != "terraform-test" {
			return fmt.Errorf("Unexpected context name: %s", context.Name)
		}

		return nil
	}
}

func testAccCheckCircleCIContextAttributes_update(context *api.CircleCIContext) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if context.Name != "terraform-test-updated" {
			return fmt.Errorf("Unexpected context name: %s", context.Name)
		}

		return nil
	}
}

const testAccCircleCIContext_basic = `
resource "circleci_context" "foo" {
	name = "terraform-test"
}
`

const testAccCircleCIContext_update = `
resource "circleci_context" "foo" {
	name = "terraform-test-updated"
}
`
