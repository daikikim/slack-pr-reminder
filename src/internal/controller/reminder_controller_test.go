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
	prs            []model.PR
	reviewStatuses map[int]map[string]string // prNumber -> username -> status
	merged         map[int]bool              // prNumber -> isMerged
	fetchErr       error
}

func (m *MockPRRepository) FetchOpenPRs(ctx context.Context) ([]model.PR, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return m.prs, nil
}

func (m *MockPRRepository) GetReviewStatuses(ctx context.Context, prNumber int) (map[string]string, error) {
	if m.reviewStatuses == nil {
		return make(map[string]string), nil
	}
	if statuses, ok := m.reviewStatuses[prNumber]; ok {
		return statuses, nil
	}
	return make(map[string]string), nil
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
		"test-channel",
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected no messages sent during non-business time, got %d", len(mockNotifier.sentMessages))
	}
}

func TestReminderController_Run_SendsBatchReminders(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    101,
				Title:     "PR 1",
				URL:       "http://github.com/repo/pr/101",
				Author:    "author1",
				Assignees: []string{"reviewer1"},
				CreatedAt: createdAt,
			},
			{
				ID:        2,
				Number:    102,
				Title:     "PR 2",
				URL:       "http://github.com/repo/pr/102",
				Author:    "author2",
				Assignees: []string{"reviewer2"},
				CreatedAt: createdAt,
			},
		},
		reviewStatuses: map[int]map[string]string{
			101: {"reviewer1": ""}, // Not reviewed
			102: {"reviewer2": ""}, // Not reviewed
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"reviewer1": "U1",
		"reviewer2": "U2",
	}

	testChannel := "test-channel"

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
		testChannel,
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	// Should send 1 batch message
	if len(mockNotifier.sentMessages) != 1 {
		t.Fatalf("Expected 1 message sent, got %d", len(mockNotifier.sentMessages))
	}

	msg := mockNotifier.sentMessages[0]
	if msg.slackID != testChannel {
		t.Errorf("Expected message to be sent to %s, got %s", testChannel, msg.slackID)
	}

	// Verify message content
	// Should contain both PRs and mentions
	if !strings.Contains(msg.message, "http://github.com/repo/pr/101") || !strings.Contains(msg.message, "<@U1>") {
		t.Errorf("Message does not contain PR 1 details: %s", msg.message)
	}
	if !strings.Contains(msg.message, "http://github.com/repo/pr/102") || !strings.Contains(msg.message, "<@U2>") {
		t.Errorf("Message does not contain PR 2 details: %s", msg.message)
	}
}

func TestReminderController_Run_SkipsReviewedPRs(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    101,
				Title:     "PR 1",
				URL:       "http://github.com/repo/pr/101",
				Author:    "author1",
				Assignees: []string{"reviewer1"},
				CreatedAt: createdAt,
			},
		},
		reviewStatuses: map[int]map[string]string{
			101: {"reviewer1": "APPROVED"}, // Approved
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"reviewer1": "U1",
	}

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
		"test-channel",
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	// Should send NO messages
	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected 0 messages sent, got %d", len(mockNotifier.sentMessages))
	}
}

func TestReminderController_Run_SkipsMergedPRs(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-2 * time.Hour)

	mockRepo := &MockPRRepository{
		prs: []model.PR{
			{
				ID:        1,
				Number:    101,
				Title:     "PR 1",
				URL:       "http://github.com/repo/pr/101",
				Author:    "author1",
				Assignees: []string{"reviewer1"},
				CreatedAt: createdAt,
			},
		},
		reviewStatuses: map[int]map[string]string{
			101: {"reviewer1": ""}, // Not reviewed but merged
		},
		merged: map[int]bool{
			101: true, // Merged
		},
	}

	mockNotifier := &MockNotifier{}
	mockTimeChecker := &MockTimeChecker{isBusinessTime: true, shouldRemind: true}
	slackView := view.NewSlackView()

	userMapping := map[string]string{
		"reviewer1": "U1",
	}

	ctrl := NewReminderController(
		mockRepo,
		mockNotifier,
		mockTimeChecker,
		slackView,
		userMapping,
		"test-channel",
	)

	err := ctrl.Run(context.Background())
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	// Should send NO messages
	if len(mockNotifier.sentMessages) != 0 {
		t.Errorf("Expected 0 messages sent, got %d", len(mockNotifier.sentMessages))
	}
}
