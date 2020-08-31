package client

import (
	"time"
)

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
