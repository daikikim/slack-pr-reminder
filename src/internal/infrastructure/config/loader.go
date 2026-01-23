package config

import (
	"fmt"
	"os"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	"gopkg.in/yaml.v3"
)

// Loader implements model.ConfigLoader.
type Loader struct {
	configPath string
}

// NewLoader creates a new Loader.
func NewLoader(configPath string) *Loader {
	return &Loader{configPath: configPath}
}

// Load reads the configuration from the YAML file.
func (l *Loader) Load() (*model.Config, error) {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// GetGitHubToken returns the GitHub token from environment variable.
func GetGitHubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

// GetSlackToken returns the Slack token from environment variable.
func GetSlackToken() string {
	return os.Getenv("SLACK_TOKEN")
}
