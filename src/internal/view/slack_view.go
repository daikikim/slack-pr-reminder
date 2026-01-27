package view

import (
	"fmt"
	"time"

	"strings"
)

// PendingReviewPR holds information for a PR pending review.
type PendingReviewPR struct {
	Title            string
	URL              string
	ElapsedTime      time.Duration
	ReviewerSlackIDs []string
}

// SlackView generates Slack messages for PRs.
type SlackView struct{}

// NewSlackView creates a new SlackView.
func NewSlackView() *SlackView {
	return &SlackView{}
}

// FormatBatchReviewReminder generates a single batch reminder message.
func (v *SlackView) FormatBatchReviewReminder(prs []PendingReviewPR) string {
	if len(prs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("PRのレビューをお願いします\n\n")

	for i, pr := range prs {
		elapsedStr := formatDuration(pr.ElapsedTime)

		// Format reviewers: <@ID1>, <@ID2>
		var reviewers []string
		for _, id := range pr.ReviewerSlackIDs {
			reviewers = append(reviewers, fmt.Sprintf("<@%s>", id))
		}
		reviewersStr := strings.Join(reviewers, ", ")

		sb.WriteString(fmt.Sprintf("%s\n・対象者：%s\n・経過時間：%s", pr.URL, reviewersStr, elapsedStr))

		if i < len(prs)-1 {
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%d時間", hours)
	}

	days := hours / 24
	remainingHours := hours % 24
	if remainingHours == 0 {
		return fmt.Sprintf("%d日", days)
	}
	return fmt.Sprintf("%d日%d時間", days, remainingHours)
}
