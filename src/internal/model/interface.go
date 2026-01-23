package model

import (
	"context"
	"time"
)

// PRRepository defines the interface for fetching pull requests.
type PRRepository interface {
	FetchOpenPRs(ctx context.Context) ([]PR, error)
	HasReviewed(ctx context.Context, prNumber int, username string) (bool, error)
	IsMerged(ctx context.Context, prNumber int) (bool, error)
}

// Notifier defines the interface for sending notifications.
type Notifier interface {
	SendReminder(ctx context.Context, slackID string, message string) error
}

// ConfigLoader defines the interface for loading configuration.
type ConfigLoader interface {
	Load() (*Config, error)
}

// TimeChecker defines the interface for time-based checks.
type TimeChecker interface {
	IsBusinessTime(t time.Time) bool
	ShouldRemind(createdAt time.Time, now time.Time) bool
}
