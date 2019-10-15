package circleci

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testProvider *schema.Provider
var testProviders map[string]terraform.ResourceProvider

var resourceOrgTestProvider *schema.Provider
var resourceOrgTestProviders map[string]terraform.ResourceProvider

func init() {
	resourceOrgTestProvider = Provider().(*schema.Provider)
	resourceOrgTestProviders = map[string]terraform.ResourceProvider{
		"circleci": resourceOrgTestProvider,
	}

	testProvider = Provider().(*schema.Provider)
	testProvider.Schema["organization"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		DefaultFunc: schema.EnvDefaultFunc("TEST_CIRCLECI_ORGANIZATION", nil),
		Description: "The CircleCI organization.",
	}
	testProviders = map[string]terraform.ResourceProvider{
		"circleci": testProvider,
	}
}

func testPreCheck(t *testing.T) {

	if v := os.Getenv("CIRCLECI_TOKEN"); v == "" {
		t.Fatal("CIRCLECI_TOKEN must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_VCS_TYPE"); v == "" {
		t.Fatal("CIRCLECI_VCS_TYPE must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_PROJECT"); v == "" {
		t.Fatal("CIRCLECI_PROJECT must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_ORGANIZATION"); v != "" {
		t.Fatal("For testing purposes do not set CIRCLECI_ORGANIZATION instead set TEST_CIRCLECI_ORGANIZATION for acceptance tests")
	}

	if v := os.Getenv("TEST_CIRCLECI_ORGANIZATION"); v == "" {
		t.Fatal("TEST_CIRCLECI_ORGANIZATION must be set for acceptance tests")
	}
}
