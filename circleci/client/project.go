package client

type ProjectEnvironmentVariable struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) HasProjectEnvironmentVariable(org, project, name string) (bool, error) {
	slug, err := c.Slug(org, project)
	if err != nil {
		return false, err
	}

	u := &url.URL{
		Path: fmt.Sprintf("project/%s/envvar/%s", slug, name)
	}

	req, err := c.rest.NewRequest("GET", u, nil)
	if err != nil {
		return false, err
	}

	_, err := c.rest.DoRequest(req, nil)
	if err != nil {
		var httpError *api.HTTPError
		if errors.As(err, httpError) && httpError.Code == 404 {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *Client) CreateProjectEnvironmentVariable(org, project, name, value string) error {
	slug, err := c.Slug(org, project)
	if err != nil {
		return err
	}
	
	u := &url.URL{
		Path: fmt.Sprintf("project/%s/envvar", slug)
	}

	req, err := c.rest.NewRequest("POST", u, &ProjectEnvironmentVariable{
		Name: name,
		Value: value,
	})

	if err != nil {
		return err
	}

	_, err := c.rest.DoRequest(req, nil)
	return err
}

func (c *Client) CreateProjectEnvironmentVariable(org, project, name, value string) error {
	slug, err := c.Slug(org, project)
	if err != nil {
		return err
	}
	
	u := &url.URL{
		Path: fmt.Sprintf("project/%s/envvar", slug)
	}

	req, err := c.rest.NewRequest("POST", u, &ProjectEnvironmentVariable{
		Name: name,
		Value: value,
	})

	if err != nil {
		return err
	}

	_, err := c.rest.DoRequest(req, nil)
	return err
}

func (c *Client) DeleteProjectEnvironmentVariable(org, project, name string) error {
	slug, err := c.Slug(org, project)
	if err != nil {
		return err
	}
	
	u := &url.URL{
		Path: fmt.Sprintf("project/%s/envvar", slug)
	}

	req, err := c.rest.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	_, err := c.rest.DoRequest(req, nil)
	return err
}