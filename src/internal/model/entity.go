package model

import "time"

// PR represents a GitHub Pull Request.
type PR struct {
	ID                 int64
	Number             int
	Title              string
	URL                string
	Author             string
	Assignees          []string
	RequestedReviewers []string
	CreatedAt          time.Time
}

// User represents a mapping between GitHub username and Slack ID.
type User struct {
	GitHubUsername string
	SlackID        string
}

// Config holds all configuration for the application.
type Config struct {
	GitHub   GitHubConfig   `yaml:"github"`
	Slack    SlackConfig    `yaml:"slack"`
	Schedule ScheduleConfig `yaml:"schedule"`
}

// GitHubConfig holds GitHub-related configuration.
type GitHubConfig struct {
	Repository string `yaml:"repository"`
}

// SlackConfig holds Slack-related configuration.
type SlackConfig struct {
	Channel string            `yaml:"channel"`
	Mapping map[string]string `yaml:"mapping"`
}

// ScheduleConfig holds scheduling configuration.
type ScheduleConfig struct {
	BusinessHours BusinessHoursConfig `yaml:"business_hours"`
	Timezone      string              `yaml:"timezone"`
	Holidays      []string            `yaml:"holidays"`
	NewYearBreak  NewYearBreakConfig  `yaml:"new_year_break"`
}

// BusinessHoursConfig defines the business hours window.
type BusinessHoursConfig struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// NewYearBreakConfig defines the new year break period.
type NewYearBreakConfig struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}
