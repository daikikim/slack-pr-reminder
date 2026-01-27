package time

import (
	"log"
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
	log.Printf("[TIME] Checking business time - Local time: %s, Weekday: %s",
		localTime.Format("2006-01-02 15:04:05 MST"), localTime.Weekday().String())

	// Check weekend
	if bt.isWeekend(localTime) {
		log.Printf("[TIME] Not business time: Weekend (%s)", localTime.Weekday().String())
		return false
	}
	log.Printf("[TIME] Weekend check passed")

	// Check holiday
	if bt.isHoliday(localTime) {
		log.Printf("[TIME] Not business time: Holiday")
		return false
	}
	log.Printf("[TIME] Holiday check passed")

	// Check new year break
	if bt.isNewYearBreak(localTime) {
		log.Printf("[TIME] Not business time: New year break")
		return false
	}
	log.Printf("[TIME] New year break check passed")

	// Check business hours
	isWithinHours := bt.isWithinBusinessHours(localTime)
	if !isWithinHours {
		log.Printf("[TIME] Not business time: Outside business hours (current: %s, range: %s-%s)",
			localTime.Format("15:04"), bt.cfg.BusinessHours.Start, bt.cfg.BusinessHours.End)
		return false
	}
	log.Printf("[TIME] Business hours check passed (current: %s, range: %s-%s)",
		localTime.Format("15:04"), bt.cfg.BusinessHours.Start, bt.cfg.BusinessHours.End)
	log.Printf("[TIME] All business time checks passed - IS BUSINESS TIME")
	return true
}

// ShouldRemind checks if a reminder should be sent based on PR creation time.
// Reminders are sent every hour after the first hour since creation.
func (bt *BusinessTimer) ShouldRemind(createdAt time.Time, now time.Time) bool {
	elapsed := now.Sub(createdAt)
	hours := elapsed.Hours()
	fractionalHour := hours - float64(int(hours))
	minutesPastHour := fractionalHour * 60

	log.Printf("[TIME] ShouldRemind check - Created: %s, Now: %s, Elapsed: %s (%.2f hours, %.1f minutes past hour)",
		createdAt.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"), elapsed, hours, minutesPastHour)

	// Must be at least 1 hour since creation
	if elapsed < time.Hour {
		log.Printf("[TIME] ShouldRemind: false (elapsed %s < 1 hour)", elapsed)
		return false
	}

	log.Printf("[TIME] ShouldRemind: true (elapsed %s >= 1 hour)", elapsed)
	return true
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
