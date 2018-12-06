package circleci

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestCircleCIEnvironmentVariableCreateThenUpdate(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheck(t)
		},
		Providers:    testProviders,
		CheckDestroy: testCircleCIEnvironmentVariableCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testCircleCIEnvironmentVariableConfig(project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", "value-for-the-test"),
				),
			},
			resource.TestStep{
				Config: testCircleCIEnvironmentVariableConfig(project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", "value-for-the-test-again"),
				),
			},
		},
	})
}

func TestCircleCIEnvironmentVariableCreateAlreadyExists(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	envValue := acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		Providers: testProviders,
		PreCheck: func() {
			testPreCheck(t)
		},
		CheckDestroy: testCircleCIEnvironmentVariableCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testCircleCIEnvironmentVariableConfig(project, envName, envValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", envValue),
				),
			},
			resource.TestStep{
				Config:      testCircleCIEnvironmentVariableConfigIdentical(project, envName, envValue),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})
}

func testCircleCIEnvironmentVariableCheckDestroy(s *terraform.State) error {
	providerClient := testProvider.Meta().(*ProviderClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circleci_environment_variable" {
			continue
		}

		envVar, err := providerClient.GetEnvVar(rs.Primary.Attributes["project"], rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if envVar.Name != "" {
			return errors.New("Environment variable should have been destroyed")
		}
	}

	return nil
}

func testCircleCIEnvironmentVariableConfig(project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  project = "%[1]s"
  name    = "%[2]s"
  value   = "%[3]s"
}`, project, name, value)
}

func testCircleCIEnvironmentVariableConfigIdentical(project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  project = "%[1]s"
  name    = "%[2]s"
  value   = "%[3]s"
}

resource "circleci_environment_variable" "%[2]s_2" {
  project = "%[1]s"
  name    = "%[2]s"
  value   = "%[3]s"
}`, project, name, value)
}
