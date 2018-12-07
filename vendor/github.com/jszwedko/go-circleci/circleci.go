package circleci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	queryLimit        = 100                         // maximum that CircleCI allows
	envVarNameValidRE = "^[[:alpha:]]+[[:word:]]*$" // source https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api
)

var (
	defaultBaseURL            = &url.URL{Host: "circleci.com", Scheme: "https", Path: "/api/v1.1/"}
	defaultLogger             = log.New(os.Stderr, "", log.LstdFlags)
	envVarNameValidCompiledRE *regexp.Regexp
)

func init() {
	envVarNameValidCompiledRE = regexp.MustCompile(envVarNameValidRE)
}

// Logger is a minimal interface for injecting custom logging logic for debug logs
type Logger interface {
	Printf(fmt string, args ...interface{})
}

// APIError represents an error from CircleCI
type APIError struct {
	HTTPStatusCode int
	Message        string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.HTTPStatusCode, e.Message)
}

// Client is a CircleCI client
// Its zero value is a usable client for examining public CircleCI repositories
type Client struct {
	BaseURL    *url.URL     // CircleCI API endpoint (defaults to DefaultEndpoint)
	Token      string       // CircleCI API token (needed for private repositories and mutative actions)
	HTTPClient *http.Client // HTTPClient to use for connecting to CircleCI (defaults to http.DefaultClient)

	Debug  bool   // debug logging enabled
	Logger Logger // logger to send debug messages on (if enabled), defaults to logging to stderr with the standard flags
}

func (c *Client) baseURL() *url.URL {
	if c.BaseURL == nil {
		return defaultBaseURL
	}

	return c.BaseURL
}

func (c *Client) client() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}

	return c.HTTPClient
}

func (c *Client) logger() Logger {
	if c.Logger == nil {
		return defaultLogger
	}

	return c.Logger
}

func (c *Client) debug(format string, args ...interface{}) {
	if c.Debug {
		c.logger().Printf(format, args...)
	}
}

func (c *Client) debugRequest(req *http.Request) {
	if c.Debug {
		out, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			c.debug("error debugging request %+v: %s", req, err)
		}
		c.debug("request:\n%+v", string(out))
	}
}

func (c *Client) debugResponse(resp *http.Response) {
	if c.Debug {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			c.debug("error debugging response %+v: %s", resp, err)
		}
		c.debug("response:\n%+v", string(out))
	}
}

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

func (c *Client) request(method, path string, responseStruct interface{}, params url.Values, bodyStruct interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Add("circle-token", c.Token)

	u := c.baseURL().ResolveReference(&url.URL{Path: path, RawQuery: params.Encode()})

	c.debug("building request for %s", u)

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return err
	}

	if bodyStruct != nil {
		b, err := json.Marshal(bodyStruct)
		if err != nil {
			return err
		}

		req.Body = nopCloser{bytes.NewBuffer(b)}
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	c.debugRequest(req)

	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.debugResponse(resp)

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

	c.debugRequest(req)

	resp, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	c.debugResponse(resp)

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

// ValidateEnvVarName check an environment variable name is valid according to https://circleci.com/docs/2.0/env-vars/#injecting-environment-variables-with-the-api
func ValidateEnvVarName(envVarName string) bool {
	return envVarNameValidCompiledRE.Match([]byte(envVarName))
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Artifact represents a build artifact
type Artifact struct {
	NodeIndex  int    `json:"node_index"`
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	URL        string `json:"url"`
}

// UserProject returns the selective project information included when querying
// for a User
type UserProject struct {
	Emails      string `json:"emails"`
	OnDashboard bool   `json:"on_dashboard"`
}

// User represents a CircleCI user
type User struct {
	Admin               bool                    `json:"admin"`
	AllEmails           []string                `json:"all_emails"`
	AvatarURL           string                  `json:"avatar_url"`
	BasicEmailPrefs     string                  `json:"basic_email_prefs"`
	Containers          int                     `json:"containers"`
	CreatedAt           time.Time               `json:"created_at"`
	DaysLeftInTrial     int                     `json:"days_left_in_trial"`
	GithubID            int                     `json:"github_id"`
	GithubOauthScopes   []string                `json:"github_oauth_scopes"`
	GravatarID          *string                 `json:"gravatar_id"`
	HerokuAPIKey        *string                 `json:"heroku_api_key"`
	LastViewedChangelog time.Time               `json:"last_viewed_changelog"`
	Login               string                  `json:"login"`
	Name                *string                 `json:"name"`
	Parallelism         int                     `json:"parallelism"`
	Plan                *string                 `json:"plan"`
	Projects            map[string]*UserProject `json:"projects"`
	SelectedEmail       *string                 `json:"selected_email"`
	SignInCount         int                     `json:"sign_in_count"`
	TrialEnd            time.Time               `json:"trial_end"`
}

// AWSConfig represents AWS configuration for a project
type AWSConfig struct {
	AWSKeypair *AWSKeypair `json:"keypair"`
}

// AWSKeypair represents the AWS access/secret key for a project
// SecretAccessKey will be a masked value
type AWSKeypair struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key_id"`
}

// BuildSummary represents the subset of build information returned with a Project
type BuildSummary struct {
	AddedAt     time.Time `json:"added_at"`
	BuildNum    int       `json:"build_num"`
	Outcome     string    `json:"outcome"`
	PushedAt    time.Time `json:"pushed_at"`
	Status      string    `json:"status"`
	VCSRevision string    `json:"vcs_revision"`
}

// Branch represents a repository branch
type Branch struct {
	LastSuccess   *BuildSummary   `json:"last_success"`
	PusherLogins  []string        `json:"pusher_logins"`
	RecentBuilds  []*BuildSummary `json:"recent_builds"`
	RunningBuilds []*BuildSummary `json:"running_builds"`
}

// PublicSSHKey represents the public part of an SSH key associated with a project
// PrivateKey will be a masked value
type PublicSSHKey struct {
	Hostname    string `json:"hostname"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
}

// Project represents information about a project
type Project struct {
	AWSConfig           AWSConfig         `json:"aws"`
	Branches            map[string]Branch `json:"branches"`
	CampfireNotifyPrefs *string           `json:"campfire_notify_prefs"`
	CampfireRoom        *string           `json:"campfire_room"`
	CampfireSubdomain   *string           `json:"campfire_subdomain"`
	CampfireToken       *string           `json:"campfire_token"`
	Compile             string            `json:"compile"`
	DefaultBranch       string            `json:"default_branch"`
	Dependencies        string            `json:"dependencies"`
	Extra               string            `json:"extra"`
	FeatureFlags        map[string]bool   `json:"feature_flags"`
	FlowdockAPIToken    *string           `json:"flowdock_api_token"`
	Followed            bool              `json:"followed"`
	HallNotifyPrefs     *string           `json:"hall_notify_prefs"`
	HallRoomAPIToken    *string           `json:"hall_room_api_token"`
	HasUsableKey        bool              `json:"has_usable_key"`
	HerokuDeployUser    *string           `json:"heroku_deploy_user"`
	HipchatAPIToken     *string           `json:"hipchat_api_token"`
	HipchatNotify       *bool             `json:"hipchat_notify"`
	HipchatNotifyPrefs  *string           `json:"hipchat_notify_prefs"`
	HipchatRoom         *string           `json:"hipchat_room"`
	IrcChannel          *string           `json:"irc_channel"`
	IrcKeyword          *string           `json:"irc_keyword"`
	IrcNotifyPrefs      *string           `json:"irc_notify_prefs"`
	IrcPassword         *string           `json:"irc_password"`
	IrcServer           *string           `json:"irc_server"`
	IrcUsername         *string           `json:"irc_username"`
	Parallel            int               `json:"parallel"`
	Reponame            string            `json:"reponame"`
	Setup               string            `json:"setup"`
	SlackAPIToken       *string           `json:"slack_api_token"`
	SlackChannel        *string           `json:"slack_channel"`
	SlackNotifyPrefs    *string           `json:"slack_notify_prefs"`
	SlackSubdomain      *string           `json:"slack_subdomain"`
	SlackWebhookURL     *string           `json:"slack_webhook_url"`
	SSHKeys             []*PublicSSHKey   `json:"ssh_keys"`
	Test                string            `json:"test"`
	Username            string            `json:"username"`
	VCSURL              string            `json:"vcs_url"`
}

// CommitDetails represents information about a commit returned with other
// structs
type CommitDetails struct {
	AuthorDate     *time.Time `json:"author_date"`
	AuthorEmail    string     `json:"author_email"`
	AuthorLogin    string     `json:"author_login"`
	AuthorName     string     `json:"author_name"`
	Body           string     `json:"body"`
	Commit         string     `json:"commit"`
	CommitURL      string     `json:"commit_url"`
	CommitterDate  *time.Time `json:"committer_date"`
	CommitterEmail string     `json:"committer_email"`
	CommitterLogin string     `json:"committer_login"`
	CommitterName  string     `json:"committer_name"`
	Subject        string     `json:"subject"`
}

// Message represents build messages
type Message struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Node represents the node a build was run on
type Node struct {
	ImageID      string `json:"image_id"`
	Port         int    `json:"port"`
	PublicIPAddr string `json:"public_ip_addr"`
	SSHEnabled   *bool  `json:"ssh_enabled"`
	Username     string `json:"username"`
}

// CircleYML represents the serialized CircleCI YML file for a given build
type CircleYML struct {
	String string `json:"string"`
}

// BuildStatus represents status information about the build
// Used when a short summary of previous builds is included
type BuildStatus struct {
	BuildTimeMillis int    `json:"build_time_millis"`
	Status          string `json:"status"`
	BuildNum        int    `json:"build_num"`
}

// BuildUser represents the user that triggered the build
type BuildUser struct {
	Email  *string `json:"email"`
	IsUser bool    `json:"is_user"`
	Login  string  `json:"login"`
	Name   *string `json:"name"`
}

// Build represents the details of a build
type Build struct {
	AllCommitDetails        []*CommitDetails  `json:"all_commit_details"`
	AuthorDate              *time.Time        `json:"author_date"`
	AuthorEmail             string            `json:"author_email"`
	AuthorName              string            `json:"author_name"`
	Body                    string            `json:"body"`
	Branch                  string            `json:"branch"`
	BuildNum                int               `json:"build_num"`
	BuildParameters         map[string]string `json:"build_parameters"`
	BuildTimeMillis         *int              `json:"build_time_millis"`
	BuildURL                string            `json:"build_url"`
	Canceled                bool              `json:"canceled"`
	CircleYML               *CircleYML        `json:"circle_yml"`
	CommitterDate           *time.Time        `json:"committer_date"`
	CommitterEmail          string            `json:"committer_email"`
	CommitterName           string            `json:"committer_name"`
	Compare                 *string           `json:"compare"`
	DontBuild               *string           `json:"dont_build"`
	Failed                  *bool             `json:"failed"`
	FeatureFlags            map[string]string `json:"feature_flags"`
	InfrastructureFail      bool              `json:"infrastructure_fail"`
	IsFirstGreenBuild       bool              `json:"is_first_green_build"`
	JobName                 *string           `json:"job_name"`
	Lifecycle               string            `json:"lifecycle"`
	Messages                []*Message        `json:"messages"`
	Node                    []*Node           `json:"node"`
	OSS                     bool              `json:"oss"`
	Outcome                 string            `json:"outcome"`
	Parallel                int               `json:"parallel"`
	Previous                *BuildStatus      `json:"previous"`
	PreviousSuccessfulBuild *BuildStatus      `json:"previous_successful_build"`
	QueuedAt                string            `json:"queued_at"`
	Reponame                string            `json:"reponame"`
	Retries                 []int             `json:"retries"`
	RetryOf                 *int              `json:"retry_of"`
	SSHEnabled              *bool             `json:"ssh_enabled"`
	SSHUsers                []*SSHUser        `json:"ssh_users"`
	StartTime               *time.Time        `json:"start_time"`
	Status                  string            `json:"status"`
	Steps                   []*Step           `json:"steps"`
	StopTime                *time.Time        `json:"stop_time"`
	Subject                 string            `json:"subject"`
	Timedout                bool              `json:"timedout"`
	UsageQueuedAt           string            `json:"usage_queued_at"`
	User                    *BuildUser        `json:"user"`
	Username                string            `json:"username"`
	VcsRevision             string            `json:"vcs_revision"`
	VCSURL                  string            `json:"vcs_url"`
	Why                     string            `json:"why"`
}

// Step represents an individual step in a build
// Will contain more than one action if the step was parallelized
type Step struct {
	Name    string    `json:"name"`
	Actions []*Action `json:"actions"`
}

// Action represents an individual action within a build step
type Action struct {
	BashCommand        *string    `json:"bash_command"`
	Canceled           *bool      `json:"canceled"`
	Continue           *string    `json:"continue"`
	EndTime            *time.Time `json:"end_time"`
	ExitCode           *int       `json:"exit_code"`
	Failed             *bool      `json:"failed"`
	HasOutput          bool       `json:"has_output"`
	Index              int        `json:"index"`
	InfrastructureFail *bool      `json:"infrastructure_fail"`
	Messages           []string   `json:"messages"`
	Name               string     `json:"name"`
	OutputURL          string     `json:"output_url"`
	Parallel           bool       `json:"parallel"`
	RunTimeMillis      int        `json:"run_time_millis"`
	StartTime          *time.Time `json:"start_time"`
	Status             string     `json:"status"`
	Step               int        `json:"step"`
	Timedout           *bool      `json:"timedout"`
	Truncated          bool       `json:"truncated"`
	Type               string     `json:"type"`
}

// TestMetadata represents metadata collected from the test run (e.g. JUnit output)
type TestMetadata struct {
	Classname  string  `json:"classname"`
	File       string  `json:"file"`
	Message    *string `json:"message"`
	Name       string  `json:"name"`
	Result     string  `json:"result"`
	RunTime    float64 `json:"run_time"`
	Source     string  `json:"source"`
	SourceType string  `json:"source_type"`
}

// Output represents the output of a given action
type Output struct {
	Type    string    `json:"type"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// SSHUser represents a user associated with an build with SSH enabled
type SSHUser struct {
	GithubID int    `json:"github_id"`
	Login    string `json:"login"`
}

// CheckoutKey represents an SSH checkout key for a project
type CheckoutKey struct {
	PublicKey   string    `json:"public_key"`
	Type        string    `json:"type"` // github-user-key or deploy-key
	Fingerprint string    `json:"fingerprint"`
	Login       *string   `json:"login"` // github username if this is a user key
	Preferred   bool      `json:"preferred"`
	Time        time.Time `json:"time"` // time key was created
}
