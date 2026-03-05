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

	reviewTargets := collectReviewTargets(pr.Assignees, pr.RequestedReviewers)
	log.Printf("[PR #%d] Review target summary - assignees: %d, requested reviewers: %d, unique targets: %d",
		pr.Number, len(pr.Assignees), len(pr.RequestedReviewers), len(reviewTargets))
	if len(reviewTargets) == 0 {
		log.Printf("[PR #%d] No review targets found from assignees/requested reviewers", pr.Number)
		return nil, nil
	}

	var reviewerSlackIDs []string
	reviewerSlackIDSet := make(map[string]struct{})

	// Check each review target
	for _, reviewer := range reviewTargets {
		// Skip if reviewer is the author
		if reviewer == pr.Author {
			log.Printf("[PR #%d] Skipping author from review targets: %s", pr.Number, reviewer)
			continue
		}

		status := reviewStatuses[reviewer]
		log.Printf("[PR #%d] Review target %s status: '%s'", pr.Number, reviewer, status)

		// 未レビュー(statusが空)の場合のみリマインド対象
		if status != "" {
			log.Printf("[PR #%d] Review target %s has already reviewed (status: %s)", pr.Number, reviewer, status)
			continue
		}

		// Get Slack ID for review target
		slackID, ok := c.userMapping[reviewer]
		if !ok {
			log.Printf("[PR #%d] No Slack mapping found for GitHub user: %s", pr.Number, reviewer)
			continue
		}

		if _, exists := reviewerSlackIDSet[slackID]; exists {
			continue
		}
		reviewerSlackIDSet[slackID] = struct{}{}
		reviewerSlackIDs = append(reviewerSlackIDs, slackID)
	}

	if len(reviewerSlackIDs) > 0 {
		log.Printf("[PR #%d] Found %d unreviewed reviewers", pr.Number, len(reviewerSlackIDs))
		return &view.PendingReviewPR{
			Title:            pr.Title,
			URL:              pr.URL,
			ElapsedTime:      elapsed,
			ReviewerSlackIDs: reviewerSlackIDs,
		}, nil
	}

	log.Printf("[PR #%d] No unreviewed reviewers found (or all approved/reviewed/mapped)", pr.Number)
	return nil, nil
}

func collectReviewTargets(assignees []string, requestedReviewers []string) []string {
	targetMap := make(map[string]struct{})
	var targets []string

	addTarget := func(username string) {
		if username == "" {
			return
		}
		if _, exists := targetMap[username]; exists {
			return
		}
		targetMap[username] = struct{}{}
		targets = append(targets, username)
	}

	for _, assignee := range assignees {
		addTarget(assignee)
	}

	for _, reviewer := range requestedReviewers {
		addTarget(reviewer)
	}

	return targets
}
