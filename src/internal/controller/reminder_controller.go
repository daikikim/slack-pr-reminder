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
	log.Printf("[CONTROLLER] Current time: %s (UTC: %s)", now.Format("2006-01-02 15:04:05 MST"), now.UTC().Format("2006-01-02 15:04:05 UTC"))

	// Check if it's business time
	log.Printf("[CONTROLLER] Checking if current time is business time")
	if !c.timeChecker.IsBusinessTime(now) {
		log.Println("[CONTROLLER] Not business time, skipping")
		return nil
	}
	log.Printf("[CONTROLLER] Business time check passed")

	// Fetch open PRs
	log.Printf("[CONTROLLER] Fetching open PRs from repository")
	prs, err := c.prRepo.FetchOpenPRs(ctx)
	if err != nil {
		log.Printf("[CONTROLLER] Error fetching PRs: %v", err)
		return err
	}

	log.Printf("[CONTROLLER] Found %d open PRs", len(prs))
	if len(prs) == 0 {
		log.Printf("[CONTROLLER] No open PRs found, exiting")
		return nil
	}

	// Process each PR
	log.Printf("[CONTROLLER] Processing %d PR(s)", len(prs))
	for i, pr := range prs {
		log.Printf("[CONTROLLER] Processing PR %d/%d: #%d - %s", i+1, len(prs), pr.Number, pr.Title)
		if err := c.processPR(ctx, pr, now); err != nil {
			log.Printf("[CONTROLLER] Error processing PR #%d: %v", pr.Number, err)
		}
	}

	log.Printf("[CONTROLLER] Finished processing all PRs")
	return nil
}

func (c *ReminderController) processPR(ctx context.Context, pr model.PR, now time.Time) error {
	log.Printf("[PR #%d] Starting processing - Title: %s, Author: %s, Created: %s",
		pr.Number, pr.Title, pr.Author, pr.CreatedAt.Format("2006-01-02 15:04:05"))

	// Check if it's time to send a reminder for this PR
	elapsed := now.Sub(pr.CreatedAt)
	log.Printf("[PR #%d] Time elapsed since creation: %s", pr.Number, elapsed)
	log.Printf("[PR #%d] Checking if reminder should be sent", pr.Number)
	if !c.timeChecker.ShouldRemind(pr.CreatedAt, now) {
		log.Printf("[PR #%d] Not time to send reminder yet (elapsed: %s, not at hour boundary)", pr.Number, elapsed)
		return nil
	}
	log.Printf("[PR #%d] Reminder timing check passed", pr.Number)

	log.Printf("[PR #%d] Processing PR #%d: %s", pr.Number, pr.Number, pr.Title)

	// TODO: テスト完了後にコメントを外すこと
	// 既存のPRレビュー依頼のリマインド処理（コメントアウト）
	/*
	// Check each assignee
	for _, assignee := range pr.Assignees {
		// Skip if assignee is the author
		if assignee == pr.Author {
			continue
		}

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
	*/

	// PRがマージされていなかったらPR作成者にリマインドを送る
	log.Printf("[PR #%d] Checking if PR is merged", pr.Number)
	isMerged, err := c.prRepo.IsMerged(ctx, pr.Number)
	if err != nil {
		log.Printf("[PR #%d] Error checking merge status: %v", pr.Number, err)
		return err
	}
	log.Printf("[PR #%d] Merge status: %v", pr.Number, isMerged)

	if !isMerged {
		log.Printf("[PR #%d] PR is not merged, checking Slack mapping for author: %s", pr.Number, pr.Author)
		// Get Slack ID for author
		slackID, ok := c.userMapping[pr.Author]
		if !ok {
			log.Printf("[PR #%d] No Slack mapping found for GitHub user (author): %s", pr.Number, pr.Author)
			log.Printf("[PR #%d] Available mappings: %v", pr.Number, c.userMapping)
			return nil
		}
		log.Printf("[PR #%d] Found Slack ID for author %s: %s", pr.Number, pr.Author, slackID)

		// Generate and send reminder to author
		log.Printf("[PR #%d] Generating reminder message", pr.Number)
		message := c.slackView.FormatAuthorReminder(slackID, pr, now)
		log.Printf("[PR #%d] Message generated (length: %d chars)", pr.Number, len(message))

		log.Printf("[PR #%d] Sending reminder to Slack", pr.Number)
		if err := c.notifier.SendReminder(ctx, slackID, message); err != nil {
			log.Printf("[PR #%d] Error sending reminder to author %s: %v", pr.Number, slackID, err)
			return err
		}

		log.Printf("[PR #%d] Successfully sent merge reminder to PR author: %s (Slack ID: %s)", pr.Number, pr.Author, slackID)
	} else {
		log.Printf("[PR #%d] PR is already merged, skipping reminder", pr.Number)
	}

	log.Printf("[PR #%d] Finished processing", pr.Number)
	return nil
}
