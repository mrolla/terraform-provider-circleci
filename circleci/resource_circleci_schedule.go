package circleci

import (
	"fmt"
	"strings"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	client "github.com/mrolla/terraform-provider-circleci/circleci/client"
)

// NB Magic scheduled actor ID
const scheduledActorID = "d9b3fcaa-6032-405a-8c75-40079ce33c3e"

func resourceCircleCISchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceCircleCIScheduleCreate,
		Read:   resourceCircleCIScheduleRead,
		Delete: resourceCircleCIScheduleDelete,
		Update: resourceCircleCIScheduleUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceCircleCIScheduleImport,
		},
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:        schema.TypeString,
				Description: "The organization where the schedule will be created",
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == d.Get("organization").(string)
				},
			},
			"project": {
				Type:        schema.TypeString,
				Description: "The name of the CircleCI project to create the schedule in",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the schedule",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the schedule",
				Optional:    true,
			},
			"per_hour": {
				Type:        schema.TypeInt,
				Description: "How often per hour to trigger a pipeline",
				Required:    true,
			},
			"hours_of_day": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "Which hours of the day to trigger a pipeline",
				Required:    true,
			},
			"days_of_week": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Which days of the week (\"MON\" .. \"SUN\") to trigger a pipeline on",
				Required:    true,
			},
			"use_scheduling_system": {
				Type:        schema.TypeBool,
				Description: "Use the scheduled system actor for attribution",
				Required:    true,
			},
			"parameters": {
				Type:        schema.TypeMap,
				Description: "Pipeline parameters to pass to created pipelines",
				Optional:    true,
			},
		},
	}
}

func resourceCircleCIScheduleCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	organization, err := c.Organization(d.Get("organization").(string))
	if err != nil {
		return err
	}

	project := d.Get("project").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	useSchedulingSystem := d.Get("use_scheduling_system").(bool)

	parsedHours := d.Get("hours_of_day").([]interface{})
	var hoursOfDay []uint
	for _, hour := range parsedHours {
		hoursOfDay = append(hoursOfDay, uint(hour.(int)))
	}

	var exists = struct{}{}
	validDays := make(map[string]interface{})
	validDays["MON"] = exists
	validDays["TUE"] = exists
	validDays["WED"] = exists
	validDays["THU"] = exists
	validDays["FRI"] = exists
	validDays["SAT"] = exists
	validDays["SUN"] = exists

	parsedDays := d.Get("days_of_week").([]interface{})
	var daysOfWeek []string
	for _, day := range parsedDays {
		if validDays[day.(string)] == nil {
			return fmt.Errorf("Invalid day specified: %s", day)
		}
		daysOfWeek = append(daysOfWeek, day.(string))
	}

	timetable := api.Timetable{
		PerHour:    uint(d.Get("per_hour").(int)),
		HoursOfDay: hoursOfDay,
		DaysOfWeek: daysOfWeek,
	}

	parsedParams := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parsedParams {
		parameters[k] = v.(string)
	}

	schedule, err := c.CreateSchedule(organization, project, name, description, timetable, useSchedulingSystem, parameters)
	if err != nil {
		return fmt.Errorf("Failed to create schedule: %w", err)
	}

	d.SetId(schedule.ID)

	return resourceCircleCIScheduleRead(d, m)
}

func resourceCircleCIScheduleDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	if err := c.DeleteSchedule(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceCircleCIScheduleRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)
	id := d.Id()

	schedule, err := c.GetSchedule(id)
	if err != nil {
		return fmt.Errorf("Failed to read schedule: %s", id)
	}

	if schedule == nil {
		d.SetId("")
		return nil
	}

	_, organization, project, err := explodeProjectSlug(schedule.ProjectSlug)
	if err != nil {
		return err
	}

	d.Set("organization", organization)
	d.Set("project", project)
	d.Set("name", schedule.Name)
	d.Set("description", schedule.Description)
	d.Set("per_hour", schedule.Timetable.PerHour)
	d.Set("hours_of_day", schedule.Timetable.HoursOfDay)
	d.Set("days_of_week", schedule.Timetable.DaysOfWeek)
	d.Set("parameters", schedule.Parameters)

	if schedule.Actor.ID == scheduledActorID {
		d.Set("use_scheduling_system", true)
	} else {
		d.Set("use_scheduling_system", false)
	}

	return nil
}

func resourceCircleCIScheduleUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	id := d.Id()
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	attributionActor := d.Get("use_scheduling_system").(bool)

	parsedHours := d.Get("hours_of_day").([]interface{})
	var hoursOfDay []uint
	for _, hour := range parsedHours {
		hoursOfDay = append(hoursOfDay, uint(hour.(int)))
	}

	var exists = struct{}{}
	validDays := make(map[string]interface{})
	validDays["MON"] = exists
	validDays["TUE"] = exists
	validDays["WED"] = exists
	validDays["THU"] = exists
	validDays["FRI"] = exists
	validDays["SAT"] = exists
	validDays["SUN"] = exists

	parsedDays := d.Get("days_of_week").([]interface{})
	var daysOfWeek []string
	for _, day := range parsedDays {
		if validDays[day.(string)] == nil {
			return fmt.Errorf("Invalid day specified: %s", day)
		}
		daysOfWeek = append(daysOfWeek, day.(string))
	}

	timetable := api.Timetable{
		PerHour:    uint(d.Get("per_hour").(int)),
		HoursOfDay: hoursOfDay,
		DaysOfWeek: daysOfWeek,
	}

	parsedParams := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parsedParams {
		parameters[k] = v.(string)
	}

	_, err := c.UpdateSchedule(id, name, description, timetable, attributionActor, parameters)
	if err != nil {
		return fmt.Errorf("Failed to update schedule: %w", err)
	}

	return nil
}

func resourceCircleCIScheduleImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	schedule, err := c.GetSchedule(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(schedule.ID)

	return []*schema.ResourceData{d}, nil
}

func explodeProjectSlug(slug string) (string, string, string, error) {
	matches := strings.Split(slug, "/")

	if len(matches) != 3 {
		return "", "", "", fmt.Errorf("Extracting vcs, org, project from project-slug '%s' failed", slug)
	}
	return matches[0], matches[1], matches[2], nil
}
