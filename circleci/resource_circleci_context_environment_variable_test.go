package circleci

import (
	"fmt"
	"os"
	"testing"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccCircleCIContextEnvironmentVariable_basic(t *testing.T) {
	variable := &api.Resource{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContextEnvironmentVariable_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextEnvironmentVariableExists("circleci_context_environment_variable.foo", variable),
					testAccCheckCircleCIContextEnvironmentVariableAttributes_basic(variable),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "variable", "VAR"),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "value", hashString("secret-value")),
					resource.TestCheckResourceAttrSet("circleci_context_environment_variable.foo", "context_id"),
				),
			},
		},
	})
}

func TestAccCircleCIContextEnvironmentVariable_update(t *testing.T) {
	variable := &api.Resource{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContextEnvironmentVariable_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextEnvironmentVariableExists("circleci_context_environment_variable.foo", variable),
					testAccCheckCircleCIContextEnvironmentVariableAttributes_basic(variable),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "variable", "VAR"),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "value", hashString("secret-value")),
					resource.TestCheckResourceAttrSet("circleci_context_environment_variable.foo", "context_id"),
				),
			},
			{
				Config: testAccCircleCIContextEnvironmentVariable_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIContextEnvironmentVariableExists("circleci_context_environment_variable.foo", variable),
					testAccCheckCircleCIContextEnvironmentVariableAttributes_update(variable),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "variable", "VAR_UPDATED"),
					resource.TestCheckResourceAttr("circleci_context_environment_variable.foo", "value", hashString("secret-value-updated")),
					resource.TestCheckResourceAttrSet("circleci_context_environment_variable.foo", "context_id"),
				),
			},
		},
	})
}

func TestAccCircleCIContextEnvironmentVariable_import(t *testing.T) {
	context := &api.CircleCIContext{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContextEnvironmentVariable_basic,
				Check:  testAccCheckCircleCIContextExists("circleci_context.foo", context),
			},
			{
				ResourceName: "circleci_context_environment_variable.foo",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf(
						"%s/%s/%s",
						testAccOrgProvider.Meta().(ProviderContext).Org,
						context.ID,
						"VAR",
					), nil
				},
				PreConfig: func() {
					os.Setenv("CIRCLECI_ENV_VALUE", "secret-value")
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if l := len(s); l != 1 {
						return fmt.Errorf("bad resource count, expected 1, got %d", l)
					}

					value := s[0].Attributes["value"]
					if value != "secret-value" {
						return fmt.Errorf("unexpected value, got: %s", value)
					}

					return nil
				},
			},
		},
	})
}

func TestAccCircleCIContextEnvironmentVariable_import_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIContextEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIContextEnvironmentVariable_basic,
			},
			{
				ResourceName: "circleci_context_environment_variable.foo",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf(
						"%s/%s/%s",
						testAccOrgProvider.Meta().(ProviderContext).Org,
						"terraform-test",
						"VAR",
					), nil
				},
				PreConfig: func() {
					os.Setenv("CIRCLECI_ENV_VALUE", "secret-value")
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func testAccCheckCircleCIContextEnvironmentVariableExists(addr string, variable *api.Resource) resource.TestCheckFunc {
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
			resource.Primary.Attributes["context_id"],
		)
		if err != nil {
			return fmt.Errorf("error getting context: %w", err)
		}

		for _, v := range ctx.Resources {
			if v.Variable == resource.Primary.Attributes["variable"] {
				*variable = v
				return nil
			}
		}

		return fmt.Errorf(
			"variable '%s' not found in context '%s'",
			resource.Primary.Attributes["variable"],
			resource.Primary.Attributes["context_id"],
		)
	}
}

func testAccCheckCircleCIContextEnvironmentVariableDestroy(s *terraform.State) error {
	providerContext := testAccOrgProvider.Meta().(ProviderContext)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "circleci_context_environment_variable" {
			continue
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := client.GetContextByID(
			providerContext.GraphQLClient,
			providerContext.Org,
			providerContext.VCS,
			resource.Primary.Attributes["context_id"],
		)
		if err == nil {
			return fmt.Errorf("Context still exists: %s", resource.Primary.Attributes["context_id"])
		}
	}

	return nil
}

func testAccCheckCircleCIContextEnvironmentVariableAttributes_basic(variable *api.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Variable != "VAR" {
			return fmt.Errorf("Unexpected variable: %s", variable.Variable)
		}

		return nil
	}
}

func testAccCheckCircleCIContextEnvironmentVariableAttributes_update(variable *api.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Variable != "VAR_UPDATED" {
			return fmt.Errorf("Unexpected variable: %s", variable.Variable)
		}

		return nil
	}
}

const testAccCircleCIContextEnvironmentVariable_basic = `
resource "circleci_context" "foo" {
	name = "terraform-test"
}

resource "circleci_context_environment_variable" "foo" {
	variable   = "VAR"
	value      = "secret-value"
	context_id = circleci_context.foo.id
}
`

const testAccCircleCIContextEnvironmentVariable_update = `
resource "circleci_context" "foo" {
	name = "terraform-test"
}

resource "circleci_context_environment_variable" "foo" {
	variable   = "VAR_UPDATED"
	value      = "secret-value-updated"
	context_id = circleci_context.foo.id
}
`
