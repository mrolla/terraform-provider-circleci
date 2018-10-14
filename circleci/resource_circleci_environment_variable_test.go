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

func TestCensorValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1", "xxxx"},
		{"22", "xxxx2"},
		{"333", "xxxx3"},
		{"4444", "xxxx44"},
		{"55555", "xxxx55"},
		{"666666", "xxxx666"},
		{"7777777", "xxxx777"},
		{"88888888", "xxxx8888"},
	}

	for _, tt := range testCases {
		actual := censorValue(tt.input)
		if actual != tt.expected {
			t.Errorf("%s but expected %s", actual, tt.expected)
		}
	}
}

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
					resource.TestCheckResourceAttr(resourceName, "value", censorValue("value-for-the-test")),
				),
			},
			resource.TestStep{
				Config: testCircleCIEnvironmentVariableConfig(project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", censorValue("value-for-the-test-again")),
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
					resource.TestCheckResourceAttr(resourceName, "value", censorValue(envValue)),
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
	client := testProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circleci_environment_variable" {
			continue
		}

		exists, err := client.EnvironmentVariableExists(rs.Primary.Attributes["project"], rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if exists {
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
