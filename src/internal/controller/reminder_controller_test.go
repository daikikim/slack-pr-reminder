package controller

import (
	"context"
	"testing"
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	"github.com/dkim/slack-pr-reminder/src/internal/view"
)

// MockPRRepository implements model.PRRepository for testing.
type MockPRRepository struct {
	prs      []model.PR
	reviews  map[int]map[string]bool // prNumber -> username -> hasReviewed
	fetchErr error
}

func (m *MockPRRepository) FetchOpenPRs(ctx context.Context) ([]model.PR, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return m.prs, nil
}

func (m *MockPRRepository) HasReviewed(ctx context.Context, prNumber int, username string) (bool, error) {
	if m.reviews == nil {
		return false, nil
	}
	if prReviews, ok := m.reviews[prNumber]; ok {
		return prReviews[username], nil
	}
	return false, nil
}

// MockNotifier implements model.Notifier for testing.
type MockNotifier struct {
	sentMessages []sentMessage
}

type sentMessage struct {
	slackID string
	message string
}

func (m *MockNotifier) SendReminder(ctx context.Context, slackID string, message string) error {
	m.sentMessages = append(m.sentMessages, sentMessage{slackID: slackID, message: message})
	return nil
}

// MockTimeChecker implements model.TimeChecker for testing.
type MockTimeChecker struct {
	isBusinessTime bool
	shouldRemind   bool
}

func (m *MockTimeChecker) IsBusinessTime(t time.Time) bool {
	return m.isBusinessTime
}

func (m *MockTimeChecker) ShouldRemind(createdAt time.Time, now time.Time) bool {
	return m.shouldRemind
}

func TestReminderController_Run_NotBusinessTime(t *testing.T) {
	mockRepo := &MockPRRepository{}
	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: false}
	slackView := view.NewSlackView()

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		map[string]string{},
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected no messages sent during non-business time, got %d", len(mockNotifier.sentMessages))
	}
}

func TestReminderController_Run_SendsReminders(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    123,
				Title:     "Test PR",
				URL:       "https://github.com/owner/repo/pull/123",
				Author:    "author",
				Assignees: []string{"reviewer1", "reviewer2"},
				CreatedAt: createdAt,
			},
		},
		reviews: map[int]map[string]bool{
			123: {"reviewer1": true}, // reviewer1 has reviewed
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"reviewer1": "U11111111",
		"reviewer2": "U22222222",
	}

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	// Should only send to reviewer2 (reviewer1 has already reviewed)
	if len(mockNotifier.sentMessages) != 1 {
		t.Errorf("Expected 1 message sent, got %d", len(mockNotifier.sentMessages))
	}

	if len(mockNotifier.sentMessages) > 0 && mockNotifier.sentMessages[0].slackID != "U22222222" {
		t.Errorf("Expected message to U22222222, got %s", mockNotifier.sentMessages[0].slackID)
	}
}

func TestReminderController_Run_SkipsAuthorAsAssignee(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    123,
				Title:     "Test PR",
				URL:       "https://github.com/owner/repo/pull/123",
				Author:    "author",
				Assignees: []string{"author", "reviewer1"}, // author is also assignee
				CreatedAt: createdAt,
			},
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"author":    "U00000000",
		"reviewer1": "U11111111",
	}

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	// Should only send to reviewer1 (author is skipped)
	if len(mockNotifier.sentMessages) != 1 {
		t.Errorf("Expected 1 message sent, got %d", len(mockNotifier.sentMessages))
	}

	if len(mockNotifier.sentMessages) > 0 && mockNotifier.sentMessages[0].slackID != "U11111111" {
		t.Errorf("Expected message to U11111111, got %s", mockNotifier.sentMessages[0].slackID)
	}
}

func TestReminderController_Run_NoReminderBeforeTime(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    123,
				Title:     "Test PR",
				URL:       "https://github.com/owner/repo/pull/123",
				Author:    "author",
				Assignees: []string{"reviewer1"},
				CreatedAt: createdAt,
			},
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: false} // Not time to remind
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"reviewer1": "U11111111",
	}

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected no messages sent when not reminder time, got %d", len(mockNotifier.sentMessages))
	}
}
