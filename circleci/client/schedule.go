package client

import (
	"github.com/CircleCI-Public/circleci-cli/api"
)

func (c *Client) GetSchedule(id string) (*api.Schedule, error) {
	return c.schedules.ScheduleByID(id)
}

func (c *Client) CreateSchedule(organization, project, name, description string, timetable api.Timetable, useSchedulingSystem bool, parameters map[string]string) (*api.Schedule, error) {
	return c.schedules.CreateSchedule(c.vcs, organization, project, name, description, useSchedulingSystem, timetable, parameters)
}

func (c *Client) DeleteSchedule(id string) error {
	return c.schedules.DeleteSchedule(id)
}

func (c *Client) UpdateSchedule(id, name, description string, timetable api.Timetable, useSchedulingActor bool, parameters map[string]string) (*api.Schedule, error) {
	return c.schedules.UpdateSchedule(id, name, description, useSchedulingActor, timetable, parameters)
}
