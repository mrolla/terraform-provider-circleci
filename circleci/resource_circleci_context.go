package circleci

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZymoticB/terraform-provider-circleci/internal/client"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCircleCIContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIContextCreate,
		Read:   resourceCircleCIContextRead,
		Delete: resourceCircleCIContextDelete,
		Exists: resourceCircleCIContextExists,
		Importer: &schema.ResourceImporter{
			State: resourceCircleCIContextImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the context",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The organization where the context will be created",
			},
		},
	}
}

func resourceCircleCIContextCreate(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	name := d.Get("name").(string)

	org := getOrganization(d, providerContext)
	if org == "" {
		return errors.New("organization is required")
	}

	if err := api.CreateContext(gqlClient, providerContext.VCS, org, name); err != nil {
		return fmt.Errorf("error creating context: %w", err)
	}

	ctx, err := client.GetContextByName(gqlClient, org, providerContext.VCS, name)
	if err != nil {
		return err
	}
	d.SetId(ctx.ID)

	return resourceCircleCIContextRead(d, meta)
}

func resourceCircleCIContextRead(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	org := getOrganization(d, providerContext)
	if org == "" {
		return errors.New("organization is required")
	}

	ctx, err := client.GetContextByID(gqlClient, org, providerContext.VCS, d.Id())
	if err != nil {
		return err
	}

	d.Set("name", ctx.Name)
	return nil
}

func resourceCircleCIContextDelete(d *schema.ResourceData, meta interface{}) error {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	if err := api.DeleteContext(gqlClient, d.Id()); err != nil {
		return fmt.Errorf("error deleting context: %w", err)
	}

	return nil
}

func resourceCircleCIContextExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	org := getOrganization(d, providerContext)
	if org == "" {
		return false, errors.New("organization is required")
	}

	_, err := client.GetContextByID(
		gqlClient,
		org,
		providerContext.VCS,
		d.Id(),
	)
	if err != nil {
		if errors.Is(err, client.ErrContextNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func resourceCircleCIContextImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	providerContext := meta.(ProviderContext)
	gqlClient := providerContext.GraphQLClient

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, errors.New("importing context requires $organization/$context")
	}

	ctx, err := client.GetContextByIDOrName(gqlClient, parts[0], providerContext.VCS, parts[1])
	if err != nil {
		return nil, err
	}
	d.SetId(ctx.ID)

	return []*schema.ResourceData{d}, nil
}
