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
	log.Println("=== PR Reminder Tool Started ===")

	// Parse CLI flags
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no actual Slack messages sent)")
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	envFile := flag.String("env", ".env", "Path to .env file (optional)")
	flag.Parse()

	log.Printf("[STEP 1] CLI flags parsed - dry-run: %v, config: %s, env: %s", *dryRun, *configPath, *envFile)

	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("[STEP 2] Note: .env file not loaded (%s): %v", *envFile, err)
	} else {
		log.Printf("[STEP 2] .env file loaded successfully: %s", *envFile)
	}

	// Load configuration
	log.Printf("[STEP 3] Loading configuration from: %s", *configPath)
	loader := config.NewLoader(*configPath)
	cfg, err := loader.Load()
	if err != nil {
		log.Fatalf("[STEP 3] Failed to load config: %v", err)
	}
	log.Printf("[STEP 3] Configuration loaded - Repository: %s, Channel: %s, User mappings: %d",
		cfg.GitHub.Repository, cfg.Slack.Channel, len(cfg.Slack.Mapping))

	// Get tokens from environment
	log.Printf("[STEP 4] Checking environment variables")
	githubToken := config.GetGitHubToken()
	if githubToken == "" {
		log.Fatal("[STEP 4] GITHUB_TOKEN environment variable is required")
	}
	log.Printf("[STEP 4] GITHUB_TOKEN found (length: %d)", len(githubToken))

	slackToken := config.GetSlackToken()
	if slackToken == "" && !*dryRun {
		log.Fatal("[STEP 4] SLACK_TOKEN environment variable is required (or use -dry-run)")
	}
	if *dryRun {
		log.Printf("[STEP 4] DRY-RUN mode: SLACK_TOKEN not required")
	} else {
		log.Printf("[STEP 4] SLACK_TOKEN found (length: %d)", len(slackToken))
	}

	// Initialize dependencies
	log.Printf("[STEP 5] Initializing GitHub client for repository: %s", cfg.GitHub.Repository)
	githubClient := github.NewClient(githubToken, cfg.GitHub.Repository)
	log.Printf("[STEP 5] GitHub client initialized")

	log.Printf("[STEP 6] Initializing Slack client - Channel: %s, DryRun: %v", cfg.Slack.Channel, *dryRun)
	slackClient := slack.NewClient(slackToken, cfg.Slack.Channel, *dryRun)
	log.Printf("[STEP 6] Slack client initialized")

	log.Printf("[STEP 7] Initializing time checker - Timezone: %s, Business hours: %s-%s",
		cfg.Schedule.Timezone, cfg.Schedule.BusinessHours.Start, cfg.Schedule.BusinessHours.End)
	timeChecker, err := businessTime.NewBusinessTimer(cfg.Schedule)
	if err != nil {
		log.Fatalf("[STEP 7] Failed to initialize time checker: %v", err)
	}
	log.Printf("[STEP 7] Time checker initialized")

	slackView := view.NewSlackView()
	log.Printf("[STEP 8] Slack view initialized")

	// Create controller
	log.Printf("[STEP 9] Creating reminder controller")
	ctrl := controller.NewReminderController(
		githubClient,
		slackClient,
		timeChecker,
		slackView,
		cfg.Slack.Mapping,
	)
	log.Printf("[STEP 9] Reminder controller created")

	// Run
	log.Printf("[STEP 10] Starting reminder process")
	ctx := context.Background()
	if err := ctrl.Run(ctx); err != nil {
		log.Fatalf("[STEP 10] Error running reminder: %v", err)
		os.Exit(1)
	}

	log.Println("=== PR Reminder Tool Completed Successfully ===")
}
