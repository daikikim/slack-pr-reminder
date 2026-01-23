package time

import (
	"time"

	"github.com/dkim/slack-pr-reminder/src/internal/model"
	holiday_jp "github.com/holiday-jp/holiday_jp-go"
)

// BusinessTimer implements model.TimeChecker.
type BusinessTimer struct {
	cfg      model.ScheduleConfig
	location *time.Location
}

// NewBusinessTimer creates a new BusinessTimer.
func NewBusinessTimer(cfg model.ScheduleConfig) (*BusinessTimer, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, err
	}
	return &BusinessTimer{cfg: cfg, location: loc}, nil
}

// IsBusinessTime checks if the given time is within business hours.
func (bt *BusinessTimer) IsBusinessTime(t time.Time) bool {
	localTime := t.In(bt.location)

	// Check weekend
	if bt.isWeekend(localTime) {
		return false
	}

	// Check holiday
	if bt.isHoliday(localTime) {
		return false
	}

	// Check new year break
	if bt.isNewYearBreak(localTime) {
		return false
	}

	// Check business hours
	return bt.isWithinBusinessHours(localTime)
}

// ShouldRemind checks if a reminder should be sent based on PR creation time.
// Reminders are sent every hour after the first hour since creation.
func (bt *BusinessTimer) ShouldRemind(createdAt time.Time, now time.Time) bool {
	elapsed := now.Sub(createdAt)

	// Must be at least 1 hour since creation
	if elapsed < time.Hour {
		return false
	}

	// Check if we're at an hour boundary (within a 5-minute window)
	hours := elapsed.Hours()
	fractionalHour := hours - float64(int(hours))

	// Allow a 5-minute window around the hour mark (0-5 minutes past)
	return fractionalHour < (5.0 / 60.0)
}

func (bt *BusinessTimer) isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (bt *BusinessTimer) isHoliday(t time.Time) bool {
	// Check Japanese public holidays
	if holiday_jp.IsHoliday(t) {
		return true
	}

	// Check configured holidays
	dateStr := t.Format("2006-01-02")
	for _, h := range bt.cfg.Holidays {
		if h == dateStr {
			return true
		}
	}

	return false
}

func (bt *BusinessTimer) isNewYearBreak(t time.Time) bool {
	if bt.cfg.NewYearBreak.Start == "" || bt.cfg.NewYearBreak.End == "" {
		return false
	}

	monthDay := t.Format("01-02")
	start := bt.cfg.NewYearBreak.Start
	end := bt.cfg.NewYearBreak.End

	// Handle year boundary (e.g., 12-29 to 01-03)
	if start > end {
		// Either in December or early January
		return monthDay >= start || monthDay <= end
	}

	return monthDay >= start && monthDay <= end
}

func (bt *BusinessTimer) isWithinBusinessHours(t time.Time) bool {
	timeStr := t.Format("15:04")
	return timeStr >= bt.cfg.BusinessHours.Start && timeStr < bt.cfg.BusinessHours.End
}
