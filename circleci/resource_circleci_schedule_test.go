package circleci

import (
	"fmt"
	"os"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	client "github.com/mrolla/terraform-provider-circleci/circleci/client"
)

func TestAccCircleCISchedule_basic(t *testing.T) {
	schedule := &api.Schedule{}
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	project := os.Getenv("CIRCLECI_PROJECT")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCISchedule_basic(organization, project, "terraform_test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIScheduleExists("circleci_schedule.terraform_test", schedule),
					testAccCheckCircleCIScheduleAttributes_basic(schedule),
					resource.TestCheckResourceAttr("circleci_schedule.terraform_test", "name", "terraform_test"),
				),
			},
		},
	})
}

func testAccCheckCircleCIScheduleExists(addr string, schedule *api.Schedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := testAccOrgProvider.Meta().(*client.Client)

		resource, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("Not found: %s", addr)
		}
		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		ctx, err := c.GetSchedule(resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting schedule: %w", err)
		}

		*schedule = *ctx

		return nil
	}
}

func testAccCheckCircleCIScheduleDestroy(s *terraform.State) error {
	c := testAccOrgProvider.Meta().(*client.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "circleci_schedule" {
			continue
		}

		if resource.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := c.GetSchedule(resource.Primary.ID)
		if err == nil {
			return fmt.Errorf("Schedule %s still exists: %w", resource.Primary.ID, err)
		}
	}

	return nil
}

func TestAccCircleCISchedule_update(t *testing.T) {
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	project := os.Getenv("CIRCLECI_PROJECT")
	schedule := &api.Schedule{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCISchedule_basic(organization, project, "terraform_test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIScheduleExists("circleci_schedule.terraform_test", schedule),
					testAccCheckCircleCIScheduleAttributes_basic(schedule),
					resource.TestCheckResourceAttr("circleci_schedule.terraform_test", "name", "terraform_test"),
				),
			},
			{
				Config: testAccCircleCISchedule_update(organization, project, "updated_name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCircleCIScheduleExists("circleci_schedule.updated_name", schedule),
					testAccCheckCircleCIScheduleAttributes_update(schedule),
					resource.TestCheckResourceAttr("circleci_schedule.updated_name", "name", "updated_name"),
				),
			},
		},
	})
}

func TestAccCircleCISchedule_import(t *testing.T) {
	organization := os.Getenv("TEST_CIRCLECI_ORGANIZATION")
	project := os.Getenv("CIRCLECI_PROJECT")
	schedule := &api.Schedule{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccOrgProviders,
		CheckDestroy: testAccCheckCircleCIScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCISchedule_basic(organization, project, "terraform_test"),
				Check:  testAccCheckCircleCIScheduleExists("circleci_schedule.terraform_test", schedule),
			},
			{
				ResourceName: "circleci_schedule.terraform_test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return schedule.ID, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCircleCIScheduleAttributes_basic(schedule *api.Schedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if schedule.Name != "terraform_test" {
			return fmt.Errorf("Unexpected schedule name: %s", schedule.Name)
		}

		if schedule.Description != "A terraform test schedule" {
			return fmt.Errorf("Unexpected schedule description: %s", schedule.Description)
		}

		if schedule.Actor.ID != "d9b3fcaa-6032-405a-8c75-40079ce33c3e" {
			return fmt.Errorf("Unexpected schedule actor ID: %s", schedule.Actor.ID)
		}

		if schedule.Timetable.PerHour != 1 {
			return fmt.Errorf("Unexpected schedule per hour: %d", schedule.Timetable.PerHour)
		}

		if len(schedule.Timetable.HoursOfDay) != 1 || schedule.Timetable.HoursOfDay[0] != 14 {
			return fmt.Errorf("Unexpected schedule hours of day: %v", schedule.Timetable.HoursOfDay)
		}

		if len(schedule.Timetable.DaysOfWeek) != 1 || schedule.Timetable.DaysOfWeek[0] != "MON" {
			return fmt.Errorf("Unexpected schedule days of week: %v", schedule.Timetable.DaysOfWeek)
		}

		if len(schedule.Parameters) != 2 || schedule.Parameters["foo"] != "bar" || schedule.Parameters["branch"] != "master" {
			return fmt.Errorf("Unxepected schedule parameters: %v", schedule.Parameters)
		}

		return nil
	}
}

func testAccCircleCISchedule_basic(organization, project, name string) string {
	const template = `
resource "circleci_schedule" "%[3]s" {
    organization = "%[1]s"
    project = "%[2]s"
	name = "%[3]s"
    description = "A terraform test schedule"
    per_hour = 1
    hours_of_day = [14]
    days_of_week = ["MON"]
    use_scheduling_system = true
    parameters = {
        foo = "bar"
        branch = "master"
    }
}
`
	return fmt.Sprintf(template, organization, project, name)
}

func testAccCheckCircleCIScheduleAttributes_update(schedule *api.Schedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if schedule.Name != "updated_name" {
			return fmt.Errorf("Unexpected schedule name: %s", schedule.Name)
		}

		if schedule.Description != "An updated terraform test schedule" {
			return fmt.Errorf("Unexpected schedule description: %s", schedule.Description)
		}

		if schedule.Actor.ID == "d9b3fcaa-6032-405a-8c75-40079ce33c3e" {
			return fmt.Errorf("Unexpected schedule actor ID: %s", schedule.Actor.ID)
		}

		if schedule.Timetable.PerHour != 2 {
			return fmt.Errorf("Unexpected schedule per hour: %d", schedule.Timetable.PerHour)
		}

		if len(schedule.Timetable.HoursOfDay) != 1 || schedule.Timetable.HoursOfDay[0] != 19 {
			return fmt.Errorf("Unexpected schedule hours of day: %v", schedule.Timetable.HoursOfDay)
		}

		if len(schedule.Timetable.DaysOfWeek) != 1 || schedule.Timetable.DaysOfWeek[0] != "TUE" {
			return fmt.Errorf("Unexpected schedule days of week: %v", schedule.Timetable.DaysOfWeek)
		}

		if len(schedule.Parameters) != 1 || schedule.Parameters["branch"] != "main" {
			return fmt.Errorf("Unxepected schedule parameters: %v", schedule.Parameters)
		}

		return nil
	}
}

func testAccCircleCISchedule_update(organization, project, name string) string {
	const template = `
resource "circleci_schedule" "%[3]s" {
    organization = "%[1]s"
    project = "%[2]s"
	name = "%[3]s"
    description = "An updated terraform test schedule"
    per_hour = 2
    hours_of_day = [19]
    days_of_week = ["TUE"]
    use_scheduling_system = false
    parameters = {
        branch = "main"
    }
}
`
	return fmt.Sprintf(template, organization, project, name)
}
