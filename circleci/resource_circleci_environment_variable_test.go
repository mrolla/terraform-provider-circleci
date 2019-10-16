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
		Providers:  resourceOrgTestProviders,
		IsUnitTest: true,
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
		Providers:    providerOrgTestProviders,
		CheckDestroy: testCircleCIEnvironmentVariableProviderOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				Config: testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test-again")),
				),
			},
		},
	})
}

func TestCircleCIEnvironmentVariableCreateThenUpdateResourceOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	envName := "TEST_" + acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheck(t)
		},
		Providers:    resourceOrgTestProviders,
		CheckDestroy: testCircleCIEnvironmentVariableResourceOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				Config: testCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test-again")),
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
		Providers: providerOrgTestProviders,
		PreCheck: func() {
			testPreCheck(t)
		},
		CheckDestroy: testCircleCIEnvironmentVariableProviderOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCircleCIEnvironmentVariableConfigProviderOrg(project, envName, envValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString(envValue)),
				),
			},
			{
				Config:      testCircleCIEnvironmentVariableConfigIdentical(project, envName, envValue),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})
}

func testCircleCIEnvironmentVariableResourceOrgCheckDestroy(s *terraform.State) error {
	providerClient := resourceOrgTestProvider.Meta().(*ProviderClient)
	return testCircleCIEnvironmentVariableCheckDestroy(providerClient, s)
}

func testCircleCIEnvironmentVariableProviderOrgCheckDestroy(s *terraform.State) error {
	providerClient := providerOrgTestProvider.Meta().(*ProviderClient)
	return testCircleCIEnvironmentVariableCheckDestroy(providerClient, s)
}

func testCircleCIEnvironmentVariableCheckDestroy(providerClient *ProviderClient, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circleci_environment_variable" {
			continue
		}

		organization := rs.Primary.Attributes["organization"]
		if organization == "" {
			organization = providerClient.organization
		}

		envVar, err := providerClient.GetEnvVar(organization, rs.Primary.Attributes["project"], rs.Primary.Attributes["name"])
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

func testCircleCIEnvironmentVariableConfigIdentical(project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}

resource "circleci_environment_variable" "%[2]s_2" {
  project 	   = "%[1]s"
  name    	   = "%[2]s"
  value   	   = "%[3]s"
}`, project, name, value)
}
