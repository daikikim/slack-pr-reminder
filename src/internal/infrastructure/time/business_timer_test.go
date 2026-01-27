package time

import (
	"testing"
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
)

func TestIsBusinessTime(t *testing.T) {
	cfg := model.ScheduleConfig{
		BusinessHours: model.BusinessHoursConfig{
			Start: "10:00",
			End:   "19:00",
		},
		Timezone: "Asia/Tokyo",
		Holidays: []string{"2024-01-04"},
		NewYearBreak: model.NewYearBreakConfig{
			Start: "12-29",
			End:   "01-03",
		},
	}

	bt, err := NewBusinessTimer(cfg)
	if err != nil {
		t.Fatalf("Failed to create BusinessTimer: %v", err)
	}

	jst, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "weekday during business hours",
			time:     time.Date(2024, 3, 13, 14, 0, 0, 0, jst), // Wednesday 14:00
			expected: true,
		},
		{
			name:     "weekday before business hours",
			time:     time.Date(2024, 3, 13, 9, 0, 0, 0, jst), // Wednesday 09:00
			expected: false,
		},
		{
			name:     "weekday after business hours",
			time:     time.Date(2024, 3, 13, 19, 0, 0, 0, jst), // Wednesday 19:00
			expected: false,
		},
		{
			name:     "Saturday",
			time:     time.Date(2024, 3, 16, 14, 0, 0, 0, jst), // Saturday 14:00
			expected: false,
		},
		{
			name:     "Sunday",
			time:     time.Date(2024, 3, 17, 14, 0, 0, 0, jst), // Sunday 14:00
			expected: false,
		},
		{
			name:     "configured holiday",
			time:     time.Date(2024, 1, 4, 14, 0, 0, 0, jst), // Configured holiday
			expected: false,
		},
		{
			name:     "new year break (Dec 29)",
			time:     time.Date(2024, 12, 29, 14, 0, 0, 0, jst),
			expected: false,
		},
		{
			name:     "new year break (Jan 1)",
			time:     time.Date(2024, 1, 1, 14, 0, 0, 0, jst),
			expected: false,
		},
		{
			name:     "start of business hours",
			time:     time.Date(2024, 3, 13, 10, 0, 0, 0, jst), // Wednesday 10:00
			expected: true,
		},
		{
			name:     "end of business hours minus 1 min",
			time:     time.Date(2024, 3, 13, 18, 59, 0, 0, jst), // Wednesday 18:59
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bt.IsBusinessTime(tt.time)
			if result != tt.expected {
				t.Errorf("IsBusinessTime(%v) = %v, expected %v", tt.time, result, tt.expected)
			}
		})
	}
}

func TestShouldRemind(t *testing.T) {
	cfg := model.ScheduleConfig{
		BusinessHours: model.BusinessHoursConfig{
			Start: "10:00",
			End:   "19:00",
		},
		Timezone: "Asia/Tokyo",
	}

	bt, err := NewBusinessTimer(cfg)
	if err != nil {
		t.Fatalf("Failed to create BusinessTimer: %v", err)
	}

	baseTime := time.Date(2024, 3, 13, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		createdAt time.Time
		now       time.Time
		expected  bool
	}{
		{
			name:      "less than 1 hour",
			createdAt: baseTime,
			now:       baseTime.Add(30 * time.Minute),
			expected:  false,
		},
		{
			name:      "exactly 1 hour",
			createdAt: baseTime,
			now:       baseTime.Add(1 * time.Hour),
			expected:  true,
		},
		{
			name:      "1 hour and 3 minutes",
			createdAt: baseTime,
			now:       baseTime.Add(1*time.Hour + 3*time.Minute),
			expected:  true,
		},
		{
			name:      "1 hour and 6 minutes (now valid)",
			createdAt: baseTime,
			now:       baseTime.Add(1*time.Hour + 6*time.Minute),
			expected:  true,
		},
		{
			name:      "2 hours exactly",
			createdAt: baseTime,
			now:       baseTime.Add(2 * time.Hour),
			expected:  true,
		},
		{
			name:      "1 hour 30 minutes (now valid)",
			createdAt: baseTime,
			now:       baseTime.Add(1*time.Hour + 30*time.Minute),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bt.ShouldRemind(tt.createdAt, tt.now)
			if result != tt.expected {
				t.Errorf("ShouldRemind() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsNewYearBreak(t *testing.T) {
	cfg := model.ScheduleConfig{
		BusinessHours: model.BusinessHoursConfig{
			Start: "10:00",
			End:   "19:00",
		},
		Timezone: "Asia/Tokyo",
		NewYearBreak: model.NewYearBreakConfig{
			Start: "12-29",
			End:   "01-03",
		},
	}

	bt, err := NewBusinessTimer(cfg)
	if err != nil {
		t.Fatalf("Failed to create BusinessTimer: %v", err)
	}

	jst, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "Dec 28 (before break)",
			time:     time.Date(2024, 12, 28, 12, 0, 0, 0, jst),
			expected: false,
		},
		{
			name:     "Dec 29 (start of break)",
			time:     time.Date(2024, 12, 29, 12, 0, 0, 0, jst),
			expected: true,
		},
		{
			name:     "Dec 31",
			time:     time.Date(2024, 12, 31, 12, 0, 0, 0, jst),
			expected: true,
		},
		{
			name:     "Jan 1",
			time:     time.Date(2025, 1, 1, 12, 0, 0, 0, jst),
			expected: true,
		},
		{
			name:     "Jan 3 (end of break)",
			time:     time.Date(2025, 1, 3, 12, 0, 0, 0, jst),
			expected: true,
		},
		{
			name:     "Jan 4 (after break)",
			time:     time.Date(2025, 1, 4, 12, 0, 0, 0, jst),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bt.isNewYearBreak(tt.time)
			if result != tt.expected {
				t.Errorf("isNewYearBreak(%v) = %v, expected %v", tt.time, result, tt.expected)
			}
		})
	}
}
