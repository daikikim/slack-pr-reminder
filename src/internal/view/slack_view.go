package view

import (
	"fmt"
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
)

// SlackView generates Slack messages for PRs.
type SlackView struct{}

// NewSlackView creates a new SlackView.
func NewSlackView() *SlackView {
	return &SlackView{}
}

// FormatReminder generates a reminder message for a PR.
func (v *SlackView) FormatReminder(slackID string, pr model.PR, now time.Time) string {
	elapsed := now.Sub(pr.CreatedAt)
	elapsedStr := formatDuration(elapsed)

	return fmt.Sprintf(
		"<@%s> PRのレビューをお願いします\n*%s*\n%s\n経過時間: %s",
		slackID,
		pr.Title,
		pr.URL,
		elapsedStr,
	)
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
