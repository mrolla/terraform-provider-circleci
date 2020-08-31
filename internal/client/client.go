package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"go.uber.org/zap"
)

const (
	queryLimit = 100 // maximum that CircleCI allows
)

var (
	defaultBaseURL = &url.URL{Host: "circleci.com", Scheme: "https", Path: "/api/v1.1/"}
)

// APIError represents an error from CircleCI
type APIError struct {
	HTTPStatusCode int
	Message        string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.HTTPStatusCode, e.Message)
}

type option func(*Client)

func WithBaseURL(baseURL *url.URL) option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// Client is a CircleCI client
// Its zero value is a usable client for examining public CircleCI repositories
type Client struct {
	baseURL *url.URL     // CircleCI API endpoint (defaults to DefaultEndpoint)
	token   string       // CircleCI API token (needed for private repositories and mutative actions)
	c       *http.Client // HTTPClient to use for connecting to CircleCI (defaults to http.DefaultClient)

	l *zap.Logger // logger to send debug messages on (if enabled), defaults to logging to stderr with the standard flags
}

func NewClient(logger *zap.Logger, apiToken string, httpClient *http.Client, options ...option) *Client {
	c := &Client{
		l:       logger,
		baseURL: defaultBaseURL,
		token:   apiToken,
		c:       httpClient,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

func (c *Client) request(method, path string, responseStruct interface{}, params url.Values, bodyStruct interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Add("circle-token", c.token)

	u := c.baseURL.ResolveReference(&url.URL{Path: path, RawQuery: params.Encode()})

	c.l.Debug("building request", zap.Stringer("url", u))

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return err
	}

	if bodyStruct != nil {
		b, err := json.Marshal(bodyStruct)
		if err != nil {
			return err
		}

		req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	c.l.Debug("request dump", zap.Stringer("request", (*requestDumper)(req)))

	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.l.Debug("response debug", zap.Stringer("response", (*responseDumper)(resp)))

	if resp.StatusCode >= 300 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &APIError{HTTPStatusCode: resp.StatusCode, Message: "unable to parse response: %s"}
		}

		if len(body) > 0 {
			message := struct {
				Message string `json:"message"`
			}{}
			err = json.Unmarshal(body, &message)
			if err != nil {
				return &APIError{
					HTTPStatusCode: resp.StatusCode,
					Message:        fmt.Sprintf("unable to parse API response: %s", err),
				}
			}
			return &APIError{HTTPStatusCode: resp.StatusCode, Message: message.Message}
		}

		return &APIError{HTTPStatusCode: resp.StatusCode}
	}

	if responseStruct != nil {
		err = json.NewDecoder(resp.Body).Decode(responseStruct)
		if err != nil {
			return err
		}
	}

	return nil
}

// Me returns information about the current user
func (c *Client) Me() (*User, error) {
	user := &User{}

	err := c.request("GET", "me", user, nil, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ListProjects returns the list of projects the user is watching
func (c *Client) ListProjects() ([]*Project, error) {
	projects := []*Project{}

	err := c.request("GET", "projects", &projects, nil, nil)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// EnableProject enables a project - generates a deploy SSH key used to checkout the Github repo.
// The Github user tied to the Circle API Token must have "admin" access to the repo.
func (c *Client) EnableProject(vcsType, account, repo string) error {
	return c.request("POST", fmt.Sprintf("project/%s/%s/%s/enable", vcsType, account, repo), nil, nil, nil)
}

// DisableProject disables a project
func (c *Client) DisableProject(account, repo string) error {
	return c.request("DELETE", fmt.Sprintf("project/%s/%s/enable", account, repo), nil, nil, nil)
}

// FollowProject follows a project
func (c *Client) FollowProject(vcsType, account, repo string) (*Project, error) {

	project := &Project{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/follow", vcsType, account, repo), project, nil, nil)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a specific project
// Returns nil of the project is not in the list of watched projects
func (c *Client) GetProject(account, repo string) (*Project, error) {
	projects, err := c.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if account == project.Username && repo == project.Reponame {
			return project, nil
		}
	}

	return nil, nil
}

func (c *Client) recentBuilds(path string, params url.Values, limit, offset int) ([]*Build, error) {
	allBuilds := []*Build{}

	if params == nil {
		params = url.Values{}
	}

	fetchAll := limit == -1
	for {
		builds := []*Build{}

		if fetchAll {
			limit = queryLimit + 1
		}

		l := limit
		if l > queryLimit {
			l = queryLimit
		}

		params.Set("limit", strconv.Itoa(l))
		params.Set("offset", strconv.Itoa(offset))

		err := c.request("GET", path, &builds, params, nil)
		if err != nil {
			return nil, err
		}
		allBuilds = append(allBuilds, builds...)

		offset += len(builds)
		limit -= len(builds)
		if len(builds) < queryLimit || limit == 0 {
			break
		}
	}

	return allBuilds, nil
}

// ListRecentBuilds fetches the list of recent builds for all repositories the user is watching
// If limit is -1, fetches all builds
func (c *Client) ListRecentBuilds(limit, offset int) ([]*Build, error) {
	return c.recentBuilds("recent-builds", nil, limit, offset)
}

// ListRecentBuildsForProject fetches the list of recent builds for the given repository
// The status and branch parameters are used to further filter results if non-empty
// If limit is -1, fetches all builds
func (c *Client) ListRecentBuildsForProject(vcsType, account, repo, branch, status string, limit, offset int) ([]*Build, error) {
	path := fmt.Sprintf("project/%s/%s/%s", vcsType, account, repo)
	if branch != "" {
		path = fmt.Sprintf("%s/tree/%s", path, branch)
	}

	params := url.Values{}
	if status != "" {
		params.Set("filter", status)
	}

	return c.recentBuilds(path, params, limit, offset)
}

// GetBuild fetches a given build by number
func (c *Client) GetBuild(vcsType, account, repo string, buildNum int) (*Build, error) {
	build := &Build{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/%d", vcsType, account, repo, buildNum), build, nil, nil)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// ListBuildArtifacts fetches the build artifacts for the given build
func (c *Client) ListBuildArtifacts(vcsType, account, repo string, buildNum int) ([]*Artifact, error) {
	artifacts := []*Artifact{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/%d/artifacts", vcsType, account, repo, buildNum), &artifacts, nil, nil)
	if err != nil {
		return nil, err
	}

	return artifacts, nil
}

// ListTestMetadata fetches the build metadata for the given build
func (c *Client) ListTestMetadata(vcsType, account, repo string, buildNum int) ([]*TestMetadata, error) {
	metadata := struct {
		Tests []*TestMetadata `json:"tests"`
	}{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/%d/tests", vcsType, account, repo, buildNum), &metadata, nil, nil)
	if err != nil {
		return nil, err
	}

	return metadata.Tests, nil
}

// AddSSHUser adds the user associated with the API token to the list of valid
// SSH users for a build.
//
// The API token being used must be a user API token
func (c *Client) AddSSHUser(account, repo string, buildNum int) (*Build, error) {
	build := &Build{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%d/ssh-users", account, repo, buildNum), build, nil, nil)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// Build triggers a new build for the given project on the given branch
// Returns the new build information
func (c *Client) Build(vcsType, account, repo, branch string) (*Build, error) {
	build := &Build{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/tree/%s", vcsType, account, repo, branch), build, nil, nil)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// RetryBuild triggers a retry of the specified build
// Returns the new build information
func (c *Client) RetryBuild(vcsType, account, repo string, buildNum int) (*Build, error) {
	build := &Build{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/%d/retry", vcsType, account, repo, buildNum), build, nil, nil)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// CancelBuild triggers a cancel of the specified build
// Returns the new build information
func (c *Client) CancelBuild(vcsType, account, repo string, buildNum int) (*Build, error) {
	build := &Build{}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/%d/cancel", vcsType, account, repo, buildNum), build, nil, nil)
	if err != nil {
		return nil, err
	}

	return build, nil
}

// ClearCache clears the cache of the specified project
// Returns the status returned by CircleCI
func (c *Client) ClearCache(vcsType, account, repo string) (string, error) {
	status := &struct {
		Status string `json:"status"`
	}{}

	err := c.request("DELETE", fmt.Sprintf("project/%s/%s/%s/build-cache", vcsType, account, repo), status, nil, nil)
	if err != nil {
		return "", err
	}

	return status.Status, nil
}

// AddEnvVar adds a new environment variable to the specified project
// Returns the added env var (the value will be masked)
func (c *Client) AddEnvVar(vcsType, account, repo, name, value string) (*EnvVar, error) {
	envVar := &EnvVar{}

	if !ValidateEnvVarName(name) {
		return nil, fmt.Errorf("environment variable name is not valid")
	}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/envvar", vcsType, account, repo), envVar, nil, &EnvVar{Name: name, Value: value})
	if err != nil {
		return nil, err
	}

	return envVar, nil
}

// GetEnvVar get an environment variable from a specified project
// Returns the environment variable (the value will be masked).
// If an environment variable with that name does not exists, it returns an empty environment variable
func (c *Client) GetEnvVar(vcsType, account, repo, name string) (*EnvVar, error) {
	envVar := EnvVar{}
	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/envvar/%s", vcsType, account, repo, name), &envVar, nil, nil)
	if err != nil {

		typedErr, ok := err.(*APIError)
		if !ok {
			return nil, err
		}
		// if error is 404 and the message is {"message":"env var not found"}
		// we can assume the environment variable is not found and return an empty structure
		if typedErr.HTTPStatusCode == 404 && typedErr.Message == "env var not found" {
			return &envVar, nil
		}
		return nil, err

	}

	return &envVar, nil
}

// ListEnvVars list environment variable to the specified project
// Returns the env vars (the value will be masked)
func (c *Client) ListEnvVars(vcsType, account, repo string) ([]EnvVar, error) {
	envVar := []EnvVar{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/envvar", vcsType, account, repo), &envVar, nil, nil)
	if err != nil {
		return nil, err
	}

	return envVar, nil
}

// DeleteEnvVar deletes the specified environment variable from the project
func (c *Client) DeleteEnvVar(vcsType, account, repo, name string) error {
	return c.request("DELETE", fmt.Sprintf("project/%s/%s/%s/envvar/%s", vcsType, account, repo, name), nil, nil, nil)
}

// AddSSHKey adds a new SSH key to the project
func (c *Client) AddSSHKey(vcsType, account, repo, hostname, privateKey string) error {
	key := &struct {
		Hostname   string `json:"hostname"`
		PrivateKey string `json:"private_key"`
	}{hostname, privateKey}
	return c.request("POST", fmt.Sprintf("project/%s/%s/%s/ssh-key", vcsType, account, repo), nil, nil, key)
}

// GetActionOutputs fetches the output for the given action
// If the action has no output, returns nil
func (c *Client) GetActionOutputs(a *Action) ([]*Output, error) {
	if !a.HasOutput || a.OutputURL == "" {
		return nil, nil
	}

	req, err := http.NewRequest("GET", a.OutputURL, nil)
	if err != nil {
		return nil, err
	}

	c.l.Debug("request dump", zap.Stringer("request", (*requestDumper)(req)))

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	c.l.Debug("response dump", zap.Stringer("response", (*responseDumper)(resp)))

	output := []*Output{}
	if err = json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, err
	}

	return output, nil
}

// ListCheckoutKeys fetches the checkout keys associated with the given project
func (c *Client) ListCheckoutKeys(account, repo string) ([]*CheckoutKey, error) {
	checkoutKeys := []*CheckoutKey{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/checkout-key", account, repo), &checkoutKeys, nil, nil)
	if err != nil {
		return nil, err
	}

	return checkoutKeys, nil
}

// CreateCheckoutKey creates a new checkout key for a project
// Valid key types are currently deploy-key and github-user-key
//
// The github-user-key type requires that the API token being used be a user API token
func (c *Client) CreateCheckoutKey(vcsType, account, repo, keyType string) (*CheckoutKey, error) {
	checkoutKey := &CheckoutKey{}

	body := struct {
		KeyType string `json:"type"`
	}{KeyType: keyType}

	err := c.request("POST", fmt.Sprintf("project/%s/%s/%s/checkout-key", vcsType, account, repo), checkoutKey, nil, body)
	if err != nil {
		return nil, err
	}

	return checkoutKey, nil
}

// GetCheckoutKey fetches the checkout key for the given project by fingerprint
func (c *Client) GetCheckoutKey(vcsType, account, repo, fingerprint string) (*CheckoutKey, error) {
	checkoutKey := &CheckoutKey{}

	err := c.request("GET", fmt.Sprintf("project/%s/%s/%s/checkout-key/%s", vcsType, account, repo, fingerprint), &checkoutKey, nil, nil)
	if err != nil {
		return nil, err
	}

	return checkoutKey, nil
}

// DeleteCheckoutKey fetches the checkout key for the given project by fingerprint
func (c *Client) DeleteCheckoutKey(vcsType, account, repo, fingerprint string) error {
	return c.request("DELETE", fmt.Sprintf("project/%s/%s/%s/checkout-key/%s", vcsType, account, repo, fingerprint), nil, nil, nil)
}

// AddHerokuKey associates a Heroku key with the user's API token to allow
// CircleCI to deploy to Heroku on your behalf
//
// The API token being used must be a user API token
//
// NOTE: It doesn't look like there is currently a way to dissaccociate your
// Heroku key, so use with care
func (c *Client) AddHerokuKey(key string) error {
	body := struct {
		APIKey string `json:"apikey"`
	}{APIKey: key}

	return c.request("POST", "/user/heroku-key", nil, nil, body)
}

type requestDumper http.Request

func (req *requestDumper) String() string {
	out, _ := httputil.DumpRequest((*http.Request)(req), true)
	return string(out)
}

type responseDumper http.Response

func (req *responseDumper) String() string {
	out, _ := httputil.DumpResponse((*http.Response)(req), true)
	return string(out)
}
