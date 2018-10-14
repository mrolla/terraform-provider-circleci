package circleci

import (
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					// TODO(matteo): this is very naive, maybe storing the real value in
					// the state is a better approach
					return oldValue == censorValue(newValue)
				},
			},
		},
	}
}

func censorValue(value string) string {
	length := len(value)
	switch {
	case length <= 1:
		return "xxxx"
	case length == 2 || length == 3:
		return "xxxx" + value[length-1:]
	case length == 4 || length == 5:
		return "xxxx" + value[length-2:]
	case length == 6 || length == 7:
		return "xxxx" + value[length-3:]
	default:
		return "xxxx" + value[length-4:]
	}
	return value
}

func resourceCircleCIEnvironmentVariableCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)
	envValue := d.Get("value").(string)

	alreadyExists, err := client.EnvironmentVariableExists(projectName, envName)
	if err != nil {
		return err
	}

	if alreadyExists {
		return fmt.Errorf("Environment variable '%s' already exists for project '%s'.", envName, projectName)
	}

	if err := client.CreateEnvironmentVariable(projectName, envName, envValue); err != nil {
		return err
	}

	d.SetId(envName)

	return resourceCircleCIEnvironmentVariableRead(d, m)
}

func resourceCircleCIEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	envVar, err := client.GetEnvironmentVariable(projectName, envName)
	if err != nil {
		return err
	}

	if err := d.Set("name", envVar.Name); err != nil {
		return err
	}

	if err := d.Set("value", envVar.Value); err != nil {
		return err
	}

	return nil
}

func resourceCircleCIEnvironmentVariableDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	err := client.DeleteEnvironmentVariable(projectName, envName)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCircleCIEnvironmentVariableExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*Client)

	projectName := d.Get("project").(string)
	envName := d.Get("name").(string)

	return client.EnvironmentVariableExists(projectName, envName)
}
