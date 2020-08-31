package circleci

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccCircleCIEnvironmentVariableOrganizationNotSet(t *testing.T) {
	project := "TEST_" + acctest.RandString(8)
	envName := "TEST_" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccNoOrgProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test"),
				ExpectError: regexp.MustCompile("organization has not been set for environment variable .*"),
			},
		},
	})
}

func TestAccCircleCIEnvironmentVariableCreateThenUpdateProviderOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	resourceName := "circleci_environment_variable." + envName
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCircleCIEnvironmentVariableProviderOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				Config: testAccCircleCIEnvironmentVariableConfigProviderOrg(project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test-again")),
				),
			},
		},
	})
}

func TestAccCircleCIEnvironmentVariableCreateThenUpdateResourceOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	envName := "TEST_" + acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccNoOrgProviders,
		CheckDestroy: testAccCircleCIEnvironmentVariableResourceOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				Config: testAccCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test-again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test-again")),
				),
			},
		},
	})
}

func TestAccCircleCIEnvironmentVariableCreateAlreadyExists(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	envValue := acctest.RandString(8)
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		Providers: testAccOrgProviders,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy: testAccCircleCIEnvironmentVariableProviderOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIEnvironmentVariableConfigProviderOrg(project, envName, envValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString(envValue)),
				),
			},
			{
				Config:      testAccCircleCIEnvironmentVariableConfigIdentical(project, envName, envValue),
				ExpectError: regexp.MustCompile("already exists"),
			},
		},
	})
}

func TestAccCircleCIEnvironmentVariableImportProviderOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	resourceName := "circleci_environment_variable." + envName
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCircleCIEnvironmentVariableProviderOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				ResourceName:      fmt.Sprintf("circleci_environment_variable.%s", envName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"value",
				},
			},
		},
	})
}

func TestAccCircleCIEnvironmentVariableImportResourceOrg(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	envName := "TEST_" + acctest.RandString(8)

	resourceName := "circleci_environment_variable." + envName

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccNoOrgProviders,
		CheckDestroy: testAccCircleCIEnvironmentVariableResourceOrgCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIEnvironmentVariableConfigResourceOrg(organization, project, envName, "value-for-the-test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s.%s.%s", organization, project, envName)),
					resource.TestCheckResourceAttr(resourceName, "project", project),
					resource.TestCheckResourceAttr(resourceName, "name", envName),
					resource.TestCheckResourceAttr(resourceName, "value", hashString("value-for-the-test")),
				),
			},
			{
				ResourceName:      fmt.Sprintf("circleci_environment_variable.%s", envName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"value",
				},
			},
		},
	})
}

func TestParseEnvironmentVariableId(t *testing.T) {
	organization := acctest.RandString(8)
	envName := acctest.RandString(8)
	projectNames := []string{
		"TEST_" + acctest.RandString(8),
		"TEST-" + acctest.RandString(8),
		"TEST." + acctest.RandString(8),
		"TEST_" + acctest.RandString(8) + "." + acctest.RandString(8),
		"TEST-" + acctest.RandString(8) + "." + acctest.RandString(8),
		"TEST." + acctest.RandString(8) + "." + acctest.RandString(8),
	}

	for _, name := range projectNames {
		expectedId := fmt.Sprintf("%s.%s.%s", organization, name, envName)
		actualOrganization, actualProjectName, actualEnvName := parseEnvironmentVariableId(expectedId)
		assert.Equal(t, organization, actualOrganization)
		assert.Equal(t, name, actualProjectName)
		assert.Equal(t, envName, actualEnvName)
	}
}

func testCircleCIEnvironmentVariableResourceOrgStateDataV0(organization, project, name string) map[string]interface{} {
	return map[string]interface{}{
		"id":           name,
		"name":         name,
		"project":      project,
		"organization": organization,
	}
}

func testCircleCIEnvironmentVariableNoOrgStateDataProviderOrgV0(project, name string) map[string]interface{} {
	return map[string]interface{}{
		"id":      name,
		"name":    name,
		"project": project,
	}
}

func testCircleCIEnvironmentVariableResourceOrgStateDataV1(organization, project, name string) map[string]interface{} {
	v0 := testCircleCIEnvironmentVariableResourceOrgStateDataV0(organization, project, name)
	return map[string]interface{}{
		"id":           fmt.Sprintf("%s.%s.%s", v0["organization"].(string), v0["project"].(string), v0["name"].(string)),
		"name":         v0["name"].(string),
		"project":      v0["project"].(string),
		"organization": v0["organization"].(string),
	}
}

func testCircleCIEnvironmentVariableNoOrgStateDataProviderOrgV1(organization, project, name string) map[string]interface{} {
	v0 := testCircleCIEnvironmentVariableNoOrgStateDataProviderOrgV0(project, name)
	return map[string]interface{}{
		"id":      fmt.Sprintf("%s.%s.%s", organization, v0["project"].(string), v0["name"].(string)),
		"name":    v0["name"].(string),
		"project": v0["project"].(string),
	}
}

func TestCircleCIEnvironmentVariableResourceOrgStateUpgradeV0(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")

	actual, err := resourceCircleCIEnvironmentVariableUpgradeV0(testCircleCIEnvironmentVariableResourceOrgStateDataV0(organization, project, envName), testContext(organization))
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	expected := testCircleCIEnvironmentVariableResourceOrgStateDataV1(organization, project, envName)
	assert.Equal(t, expected, actual)
}

func TestCircleCIEnvironmentVariableProviderOrgStateUpgradeV0(t *testing.T) {
	project := os.Getenv("CIRCLECI_PROJECT")
	envName := "TEST_" + acctest.RandString(8)
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")

	actual, err := resourceCircleCIEnvironmentVariableUpgradeV0(testCircleCIEnvironmentVariableNoOrgStateDataProviderOrgV0(project, envName), testContext(organization))
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	expected := testCircleCIEnvironmentVariableNoOrgStateDataProviderOrgV1(organization, project, envName)
	assert.Equal(t, expected, actual)
}

func testAccCircleCIEnvironmentVariableResourceOrgCheckDestroy(s *terraform.State) error {
	ctx := testAccNoOrgProvider.Meta().(ProviderContext)
	return testAccCircleCIEnvironmentVariableCheckDestroy(ctx, s)
}

func testAccCircleCIEnvironmentVariableProviderOrgCheckDestroy(s *terraform.State) error {
	ctx := testAccOrgProvider.Meta().(ProviderContext)
	return testAccCircleCIEnvironmentVariableCheckDestroy(ctx, s)
}

func testAccCircleCIEnvironmentVariableCheckDestroy(providerContext ProviderContext, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circleci_environment_variable" {
			continue
		}

		organization := rs.Primary.Attributes["organization"]
		if organization == "" {
			organization = providerContext.Org
		}

		envVar, err := providerContext.Client.GetEnvVar(providerContext.VCS, organization, rs.Primary.Attributes["project"], rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if envVar.Name != "" {
			return errors.New("Environment variable should have been destroyed")
		}
	}

	return nil
}

func testAccCircleCIEnvironmentVariableConfigProviderOrg(project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  project = "%[1]s"
  name    = "%[2]s"
  value   = "%[3]s"
}`, project, name, value)
}

func testAccCircleCIEnvironmentVariableConfigResourceOrg(organization, project, name, value string) string {
	return fmt.Sprintf(`
resource "circleci_environment_variable" "%[2]s" {
  organization = "%[4]s"
  project      = "%[1]s"
  name         = "%[2]s"
  value        = "%[3]s"
}`, project, name, value, organization)
}

func testAccCircleCIEnvironmentVariableConfigIdentical(project, name, value string) string {
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
