package circleci

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	client "github.com/mrolla/terraform-provider-circleci/circleci/client"
)

func resourceCircleCIEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIEnvironmentVariableCreate,
		Read:   resourceCircleCIEnvironmentVariableRead,
		Delete: resourceCircleCIEnvironmentVariableDelete,
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
				Optional:    true,
				ForceNew:    true,
			},
			"project": {
				Description: "The name of the CircleCI project to create the variable in",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "The name of the environment variable",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEnvironmentVariableNameFunc,
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceCircleCIEnvironmentVariableResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceCircleCIEnvironmentVariableUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceCircleCIEnvironmentVariableResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"organization": {
				Description: "The CircleCI organization.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"project": {
				Description: "The name of the CircleCI project to create the variable in",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "The name of the environment variable",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEnvironmentVariableNameFunc,
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

func resourceCircleCIEnvironmentVariableUpgradeV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	var organization string
	rawOrg := rawState["organization"]
	if rawOrg != nil && rawOrg.(string) != "" {
		organization = rawState["organization"].(string)
	} else {
		c := meta.(*client.Client)

		org, err := c.Organization(organization)
		if err != nil {
			return nil, err
		}

		organization = org
	}

	rawState["id"] = generateId(organization, rawState["project"].(string), rawState["name"].(string))

	return rawState, nil
}

// hashString do a sha256 checksum, encode it in base64 and return it as string
// The choice of sha256 for checksum is arbitrary.
func hashString(str string) string {
	hash := sha256.Sum256([]byte(str))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func resourceCircleCIEnvironmentVariableCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	organization, err := c.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)

	if err := c.CreateProjectEnvironmentVariable(organization, project, name, value); err != nil {
		return fmt.Errorf("failed to create environment variable: %w", err)
	}

	d.SetId(generateId(organization, project, name))
	return resourceCircleCIEnvironmentVariableRead(d, m)
}

func resourceCircleCIEnvironmentVariableRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	// If we don't have a project name we're doing an import. Parse it from the ID.
	if _, ok := d.GetOk("name"); !ok {
		if err := setOrgProjectNameFromEnvironmentVariableId(d); err != nil {
			return err
		}
	}

	organization, err := c.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	name := d.Get("name").(string)

	has, err := c.HasProjectEnvironmentVariable(organization, project, name)
	if err != nil {
		return fmt.Errorf("failed to get project environment variable: %w", err)
	}

	if !has {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceCircleCIEnvironmentVariableDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	organization, err := c.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	name := d.Get("name").(string)

	err = c.DeleteProjectEnvironmentVariable(organization, project, name)
	if err != nil {
		return fmt.Errorf("failed to delete project environment variable: %w", err)
	}

	d.SetId("")

	return nil
}

func setOrgProjectNameFromEnvironmentVariableId(d *schema.ResourceData) error {
	organization, projectName, envName := parseEnvironmentVariableId(d.Id())
	// Validate that he have values for all the ID segments. This should be at least 3
	if organization == "" || projectName == "" || envName == "" {
		return fmt.Errorf("error calculating circle_ci_environment_variable. Please make sure the ID is in the form ORGANIZATION.PROJECTNAME.VARNAME (i.e. foo.bar.my_var)")
	}

	_ = d.Set("organization", organization)
	_ = d.Set("project", projectName)
	_ = d.Set("name", envName)
	return nil
}

func parseEnvironmentVariableId(id string) (organization, projectName, envName string) {
	parts := strings.Split(id, ".")

	if len(parts) >= 3 {
		organization = parts[0]
		projectName = strings.Join(parts[1:len(parts)-1], ".")
		envName = parts[len(parts)-1]
	}

	return organization, projectName, envName
}

func generateId(organization, projectName, envName string) string {
	vars := []string{
		organization,
		projectName,
		envName,
	}
	return strings.Join(vars, ".")
}
