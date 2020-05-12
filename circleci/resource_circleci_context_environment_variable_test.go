package circleci

import (
	"fmt"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

func testAccCheckCircleCIContextEnvironmentVariableExists(addr string, variable *api.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccOrgProvider.Meta().(*Client)

		resource, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("Not found: %s", addr)
		}
		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ctx, err := GetContextByID(client.graphql, client.organization, client.vcs, resource.Primary.Attributes["context_id"])
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
	client := testAccOrgProvider.Meta().(*Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "circleci_context_environment_variable" {
			continue
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := GetContextByID(client.graphql, client.organization, client.vcs, resource.Primary.Attributes["context_id"])
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