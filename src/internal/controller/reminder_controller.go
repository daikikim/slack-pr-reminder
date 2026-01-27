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
	prRepo       model.PRRepository
	notifier     model.Notifier
	timeChecker  model.TimeChecker
	slackView    *view.SlackView
	userMapping  map[string]string
	slackChannel string
}

// NewReminderController creates a new ReminderController.
func NewReminderController(
	prRepo model.PRRepository,
	notifier model.Notifier,
	timeChecker model.TimeChecker,
	slackView *view.SlackView,
	userMapping map[string]string,
	slackChannel string,
) *ReminderController {
	return &ReminderController{
		prRepo:       prRepo,
		notifier:     notifier,
		timeChecker:  timeChecker,
		slackView:    slackView,
		userMapping:  userMapping,
		slackChannel: slackChannel,
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

	// Collect all pending reviews
	var pendingPRs []view.PendingReviewPR

	// Process each PR
	log.Printf("[CONTROLLER] Processing %d PR(s)", len(prs))
	for i, pr := range prs {
		log.Printf("[CONTROLLER] Processing PR %d/%d: #%d - %s", i+1, len(prs), pr.Number, pr.Title)
		pending, err := c.processPR(ctx, pr, now)
		if err != nil {
			log.Printf("[CONTROLLER] Error processing PR #%d: %v", pr.Number, err)
			continue
		}
		if pending != nil {
			pendingPRs = append(pendingPRs, *pending)
		}
	}

	// Send batch reminder if there are pending reviews
	if len(pendingPRs) > 0 {
		log.Printf("[CONTROLLER] Found %d PRs pending review, sending batch reminder", len(pendingPRs))
		message := c.slackView.FormatBatchReviewReminder(pendingPRs)

		if err := c.notifier.SendReminder(ctx, c.slackChannel, message); err != nil {
			log.Printf("[CONTROLLER] Error sending batch reminder: %v", err)
		}
	}

	log.Printf("[CONTROLLER] Finished processing all PRs")
	return nil
}

func (c *ReminderController) processPR(ctx context.Context, pr model.PR, now time.Time) (*view.PendingReviewPR, error) {
	log.Printf("[PR #%d] Starting processing - Title: %s, Author: %s, Created: %s",
		pr.Number, pr.Title, pr.Author, pr.CreatedAt.Format("2006-01-02 15:04:05"))

	// Check if it's time to send a reminder for this PR
	elapsed := now.Sub(pr.CreatedAt)
	log.Printf("[PR #%d] Time elapsed since creation: %s", pr.Number, elapsed)
	log.Printf("[PR #%d] Checking if reminder should be sent", pr.Number)
	if !c.timeChecker.ShouldRemind(pr.CreatedAt, now) {
		log.Printf("[PR #%d] Not time to send reminder yet (elapsed: %s)", pr.Number, elapsed)
		return nil, nil
	}
	log.Printf("[PR #%d] Reminder timing check passed", pr.Number)

	// PRがマージされているかチェック
	log.Printf("[PR #%d] Checking if PR is merged", pr.Number)
	isMerged, err := c.prRepo.IsMerged(ctx, pr.Number)
	if err != nil {
		log.Printf("[PR #%d] Error checking merge status: %v", pr.Number, err)
		return nil, err
	}
	if isMerged {
		log.Printf("[PR #%d] PR is already merged, skipping reminder", pr.Number)
		return nil, nil
	}

	// レビュー状況を取得
	reviewStatuses, err := c.prRepo.GetReviewStatuses(ctx, pr.Number)
	if err != nil {
		log.Printf("[PR #%d] Error getting review statuses: %v", pr.Number, err)
		return nil, err
	}

	var reviewerSlackIDs []string

	// Check each assignee
	for _, assignee := range pr.Assignees {
		// Skip if assignee is the author
		if assignee == pr.Author {
			continue
		}

		status := reviewStatuses[assignee]
		log.Printf("[PR #%d] Assignee %s status: '%s'", pr.Number, assignee, status)

		// 未レビュー(statusが空)の場合のみリマインド対象
		if status != "" {
			log.Printf("[PR #%d] Assignee %s has already reviewed (status: %s)", pr.Number, assignee, status)
			continue
		}

		// Get Slack ID for assignee
		slackID, ok := c.userMapping[assignee]
		if !ok {
			log.Printf("[PR #%d] No Slack mapping found for GitHub user: %s", pr.Number, assignee)
			continue
		}

		reviewerSlackIDs = append(reviewerSlackIDs, slackID)
	}

	if len(reviewerSlackIDs) > 0 {
		log.Printf("[PR #%d] Found %d unreviewed assignees", pr.Number, len(reviewerSlackIDs))
		return &view.PendingReviewPR{
			Title:            pr.Title,
			URL:              pr.URL,
			ElapsedTime:      elapsed,
			ReviewerSlackIDs: reviewerSlackIDs,
		}, nil
	}

	log.Printf("[PR #%d] No unreviewed assignees found (or all approved/reviewed)", pr.Number)
	return nil, nil
}
