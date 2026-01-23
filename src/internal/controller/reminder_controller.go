package controller

import (
	"context"
	"log"
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	"github.com/dkim/slack-pr-reminder/src/internal/view"
)

// ReminderController orchestrates the PR reminder process.
type ReminderController struct {
	prRepo      model.PRRepository
	notifier    model.Notifier
	timeChecker model.TimeChecker
	slackView   *view.SlackView
	userMapping map[string]string
}

// NewReminderController creates a new ReminderController.
func NewReminderController(
	prRepo model.PRRepository,
	notifier model.Notifier,
	timeChecker model.TimeChecker,
	slackView *view.SlackView,
	userMapping map[string]string,
) *ReminderController {
	return &ReminderController{
		prRepo:      prRepo,
		notifier:    notifier,
		timeChecker: timeChecker,
		slackView:   slackView,
		userMapping: userMapping,
	}
}

// Run executes the reminder process.
func (c *ReminderController) Run(ctx context.Context) error {
	now := time.Now()

	// Check if it's business time
	if !c.timeChecker.IsBusinessTime(now) {
		log.Println("Not business time, skipping")
		return nil
	}

	// Fetch open PRs
	prs, err := c.prRepo.FetchOpenPRs(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d open PRs", len(prs))

	// Process each PR
	for _, pr := range prs {
		if err := c.processPR(ctx, pr, now); err != nil {
			log.Printf("Error processing PR #%d: %v", pr.Number, err)
		}
	}

	return nil
}

func (c *ReminderController) processPR(ctx context.Context, pr model.PR, now time.Time) error {
	// Check if it's time to send a reminder for this PR
	if !c.timeChecker.ShouldRemind(pr.CreatedAt, now) {
		return nil
	}

	log.Printf("Processing PR #%d: %s", pr.Number, pr.Title)

	// Check each assignee
	for _, assignee := range pr.Assignees {
		// TODO: テスト完了後にコメントを外すこと
		// Skip if assignee is the author
		// if assignee == pr.Author {
		// 	continue
		// }

		// Check if assignee has already reviewed
		hasReviewed, err := c.prRepo.HasReviewed(ctx, pr.Number, assignee)
		if err != nil {
			log.Printf("Error checking review status for %s: %v", assignee, err)
			continue
		}

		if hasReviewed {
			log.Printf("Assignee %s has already reviewed PR #%d", assignee, pr.Number)
			continue
		}

		// Get Slack ID for assignee
		slackID, ok := c.userMapping[assignee]
		if !ok {
			log.Printf("No Slack mapping found for GitHub user: %s", assignee)
			continue
		}

		// Generate and send reminder
		message := c.slackView.FormatReminder(slackID, pr, now)
		if err := c.notifier.SendReminder(ctx, slackID, message); err != nil {
			log.Printf("Error sending reminder to %s: %v", slackID, err)
			continue
		}
	}

	return nil
}
