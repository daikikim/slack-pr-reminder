package controller

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	"github.com/dkim/slack-pr-reminder/src/internal/view"
)

// MockPRRepository implements model.PRRepository for testing.
type MockPRRepository struct {
	prs      []model.PR
	reviews  map[int]map[string]bool // prNumber -> username -> hasReviewed
	merged   map[int]bool              // prNumber -> isMerged
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

func (m *MockPRRepository) IsMerged(ctx context.Context, prNumber int) (bool, error) {
	if m.merged == nil {
		return false, nil
	}
	if isMerged, ok := m.merged[prNumber]; ok {
		return isMerged, nil
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

// TODO: テスト完了後にコメントを外すこと
// レビュー依頼のリマインド機能がコメントアウトされているため、このテストも一時的にコメントアウト
/*
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
*/

// TODO: テスト完了後にコメントを外すこと
// レビュー依頼のリマインド機能がコメントアウトされているため、このテストも一時的にコメントアウト
/*
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
*/

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

func TestReminderController_Run_SendsReminderToAuthorWhenNotMerged(t *testing.T) {
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
		merged: map[int]bool{
			123: false, // PR is not merged
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"author": "U00000000",
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

	// Should send reminder to author
	if len(mockNotifier.sentMessages) != 1 {
		t.Errorf("Expected 1 message sent to author, got %d", len(mockNotifier.sentMessages))
	}

	if len(mockNotifier.sentMessages) > 0 {
		if mockNotifier.sentMessages[0].slackID != "U00000000" {
			t.Errorf("Expected message to U00000000 (author), got %s", mockNotifier.sentMessages[0].slackID)
		}
		// Check message contains merge reminder text
		if !strings.Contains(mockNotifier.sentMessages[0].message, "マージ") {
			t.Errorf("Expected message to contain 'マージ', got: %s", mockNotifier.sentMessages[0].message)
		}
	}
}

func TestReminderController_Run_SkipsAuthorWhenPRIsMerged(t *testing.T) {
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
		merged: map[int]bool{
			123: true, // PR is merged
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"author": "U00000000",
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

	// Should not send reminder when PR is merged
	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected no messages sent when PR is merged, got %d", len(mockNotifier.sentMessages))
	}
}

func TestReminderController_Run_SkipsAuthorWhenNoSlackMapping(t *testing.T) {
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
		merged: map[int]bool{
			123: false, // PR is not merged
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	// No mapping for author
	userMapping := map[string]string{}

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

	// Should not send reminder when no Slack mapping exists
	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected no messages sent when no Slack mapping, got %d", len(mockNotifier.sentMessages))
	}
}
