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

func TestCircleCIEnvironmentVariableOrganizationNotSet(t *testing.T) {
	project := "TEST_" + acctest.RandString(8)
	envName := "TEST_" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheck(t)
		},
		Providers:    testProviders,
		CheckDestroy: testCircleCIEnvironmentVariableCheckDestroy,
		IsUnitTest:   true,
		Steps: []resource.TestStep{
			{
				Config:      testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test"),
				ExpectError: regexp.MustCompile("organization has not been set for environment variable .*"),
			},
		},
	})
}

func TestCircleCIEnvironmentVariableCreateThenUpdateProviderOrg(t *testing.T) {
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
			{
				Config: testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", "value-for-the-test"),
				),
			},
			{
				Config: testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", "value-for-the-test-again"),
				),
			},
		},
	})
}

func TestCircleCIEnvironmentVariableCreateThenUpdateResourceOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	organization := "ORG_" + acctest.RandString(8)
	envName := "TEST_" + acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheck(t)
		},
		Providers:    testProviders,
		CheckDestroy: testCircleCIEnvironmentVariableCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", "value-for-the-test"),
				),
			},
			{
				Config: testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test-again"),
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
	organization := "ORG_" + acctest.RandString(8)
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
			{
				Config: testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, envValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", envValue),
				),
			},
			{
				Config:      testCircleCIEnvironmentVariableConfigIdentical(organization, project, envName, envValue),
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

		organization := rs.Primary.Attributes["organization"]
		envVar, err := providerClient.GetEnvVar(&organization, rs.Primary.Attributes["project"], rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if envVar.Name != "" {
			return errors.New("Environment variable should have been destroyed")
		}
	}

	return nil
}

func testCircleCIEnvironmentVariableConfigProviderOrg(project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}`, project, name, value)
}

func testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  organization = "%[4]s"
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}`, project, name, value, organization)
}

func testCircleCIEnvironmentVariableConfigIdentical(organization, project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  organization = "%[4]s"
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}

resource "circleci_environment_variable" "%[2]s_2" {
  organization = "%[4]s"
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}`, project, name, value, organization)
}
