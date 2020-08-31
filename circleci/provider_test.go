package circleci

import (
	"net/url"
	"os"
	"testing"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"go.uber.org/zap"
)

var testAccNoOrgProvider *schema.Provider
var testAccNoOrgProviders map[string]terraform.ResourceProvider

var testAccOrgProvider *schema.Provider
var testAccOrgProviders map[string]terraform.ResourceProvider

func init() {
	testAccNoOrgProvider = Provider().(*schema.Provider)
	testAccNoOrgProviders = map[string]terraform.ResourceProvider{
		"circleci": testAccNoOrgProvider,
	}

	testAccOrgProvider = Provider().(*schema.Provider)
	testAccOrgProvider.Schema["organization"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		DefaultFunc: schema.EnvDefaultFunc("TEST_CIRCLECI_ORGANIZATION", nil),
		Description: "The CircleCI organization.",
	}
	testAccOrgProviders = map[string]terraform.ResourceProvider{
		"circleci": testAccOrgProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("CIRCLECI_TOKEN"); v == "" {
		t.Fatal("CIRCLECI_TOKEN must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_VCS_TYPE"); v == "" {
		t.Fatal("CIRCLECI_VCS_TYPE must be set for acceptance tests")
	}

	if v := os.Getenv("CIRCLECI_ORGANIZATION"); v != "" {
		t.Fatal("For testing purposes do not set CIRCLECI_ORGANIZATION instead set TEST_CIRCLECI_ORGANIZATION for acceptance tests")
	}

	if v := os.Getenv("TEST_CIRCLECI_ORGANIZATION"); v == "" {
		t.Fatal("TEST_CIRCLECI_ORGANIZATION must be set for acceptance tests")
	}
}

func testContext(organization string) ProviderContext {
	baseURL, err := url.Parse(_testBaseURL)
	if err != nil {
		panic(err)
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.OutputPaths = []string{
		"/tmp/test.txt",
	}
	logger, _ := logCfg.Build()

	logger.Error("test")

	providerClient := client.NewClient(
		zap.NewNop(),
		_testCCIToken,
		_testHTTPClient,
		client.WithBaseURL(baseURL),
	)

	return ProviderContext{
		Client: providerClient,
		VCS:    "github",
		Org:    organization,
		Logger: logger,
	}
}
