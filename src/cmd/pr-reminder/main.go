package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/dkim/slack-pr-reminder/src/internal/controller"
	"github.com/dkim/slack-pr-reminder/src/internal/infrastructure/config"
	"github.com/dkim/slack-pr-reminder/src/internal/infrastructure/github"
	"github.com/dkim/slack-pr-reminder/src/internal/infrastructure/slack"
	businessTime "github.com/dkim/slack-pr-reminder/src/internal/infrastructure/time"
	"github.com/dkim/slack-pr-reminder/src/internal/view"
	"github.com/joho/godotenv"
)

func main() {
	// Parse CLI flags
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no actual Slack messages sent)")
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	envFile := flag.String("env", ".env", "Path to .env file (optional)")
	flag.Parse()

	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("Note: .env file not loaded (%s): %v", *envFile, err)
	}

	// Load configuration
	loader := config.NewLoader(*configPath)
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get tokens from environment
	githubToken := config.GetGitHubToken()
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	slackToken := config.GetSlackToken()
	if slackToken == "" && !*dryRun {
		log.Fatal("SLACK_TOKEN environment variable is required (or use -dry-run)")
	}

	// Initialize dependencies
	githubClient := github.NewClient(githubToken, cfg.GitHub.Repository)

	slackClient := slack.NewClient(slackToken, cfg.Slack.Channel, *dryRun)

	timeChecker, err := businessTime.NewBusinessTimer(cfg.Schedule)
	if err != nil {
		log.Fatalf("Failed to initialize time checker: %v", err)
	}

	slackView := view.NewSlackView()

	// Create controller
	ctrl := controller.NewReminderController(
		githubClient,
		slackClient,
		timeChecker,
		slackView,
		cfg.Slack.Mapping,
	)

	// Run
	ctx := context.Background()
	if err := ctrl.Run(ctx); err != nil {
		log.Fatalf("Error running reminder: %v", err)
		os.Exit(1)
	}

	log.Println("Reminder check completed")
}
