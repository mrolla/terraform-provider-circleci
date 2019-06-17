package circleci

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/schema"

	circleciapi "github.com/jszwedko/go-circleci"
)

func resourceCircleCIEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIEnvironmentVariableCreate,
		Read:   resourceCircleCIEnvironmentVariableRead,
		Delete: resourceCircleCIEnvironmentVariableDelete,
		Exists: resourceCircleCIEnvironmentVariableExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"organization": {
				Description: "The CircleCI organization.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"project": {
				Description: "The name of the CircleCI project to create the variable in",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the environment variable",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(i interface{}, keyName string) (warnings []string, errors []error) {
					v, ok := i.(string)
					if !ok {
						return nil, []error{fmt.Errorf("expected type of %s to be string", keyName)}
					}
					if !circleciapi.ValidateEnvVarName(v) {
						return nil, []error{fmt.Errorf("environment variable name %s is not valid. See https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api", v)}
					}

					return nil, nil
				},
			},
			"value": {
				Description: "The value of the environment variable",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				StateFunc: func(value interface{}) string {
					/* To avoid storing the value of the environment variable in the state
					but still be able to know when the value change, we store a hash of the value.
					*/
					return hashString(value.(string))
				},
			},
		},
	}
}

// hashString do a sha256 checksum, encode it in base64 and return it as string
// The choice of sha256 for checksum is arbitrary.
func hashString(str string) string {
	hash := sha256.Sum256([]byte(str))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func resourceCircleCIEnvironmentVariableCreate(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*ProviderClient)

	organization := d.Get("organization").(string)
	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)
	envValue := d.Get("value").(string)

	exists, err := providerClient.EnvVarExists(organization, projectName, envName)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("environment variable '%s' already exists for project '%s'", envName, projectName)
	}

	if _, err := providerClient.AddEnvVar(organization, projectName, envName, envValue); err != nil {
		return err
	}

	d.SetId(envName)

	return resourceCircleCIEnvironmentVariableRead(d, m)
}

func resourceCircleCIEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*ProviderClient)

	organization := d.Get("organization").(string)
	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	envVar, err := providerClient.GetEnvVar(organization, projectName, envName)
	if err != nil {
		return err
	}

	if err := d.Set("name", envVar.Name); err != nil {
		return err
	}

	// environment variable value can only be set at creation since CircleCI API return hidden values : https://circleci.com/docs/api/#list-environment-variables
	// also it is better to avoid storing sensitive value in terraform state if possible.
	return nil
}

func resourceCircleCIEnvironmentVariableDelete(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*ProviderClient)

	organization := d.Get("organization").(string)
	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	err := providerClient.DeleteEnvVar(organization, projectName, envName)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCircleCIEnvironmentVariableExists(d *schema.ResourceData, m interface{}) (bool, error) {
	providerClient := m.(*ProviderClient)

	organization := d.Get("organization").(string)
	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	envVar, err := providerClient.GetEnvVar(organization, projectName, envName)
	if err != nil {
		return false, err
	}

	return bool(envVar.Value != ""), nil
}
