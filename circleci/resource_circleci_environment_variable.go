package circleci

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
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
			"project": &schema.Schema{
				Description: "The name of the CircleCI project to create the variable in",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": &schema.Schema{
				Description: "The name of the environment variable",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"value": &schema.Schema{
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

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)
	envValue := d.Get("value").(string)

	exists, err := providerClient.EnvVarExists(projectName, envName)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("environment variable '%s' already exists for project '%s'", envName, projectName)
	}

	if _, err := providerClient.AddEnvVar(projectName, envName, envValue); err != nil {
		return err
	}

	d.SetId(envName)

	return resourceCircleCIEnvironmentVariableRead(d, m)
}

func resourceCircleCIEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	providerClient := m.(*ProviderClient)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	envVar, err := providerClient.GetEnvVar(projectName, envName)
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

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	err := providerClient.DeleteEnvVar(projectName, envName)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCircleCIEnvironmentVariableExists(d *schema.ResourceData, m interface{}) (bool, error) {
	providerClient := m.(*ProviderClient)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	envVar, err := providerClient.GetEnvVar(projectName, envName)
	if err != nil {
		return false, err
	}

	return bool(envVar.Value != ""), nil
}
