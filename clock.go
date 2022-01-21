package gridon

import (
	"time"
)

// newClock - clockの取得
func newClock() IClock {
	return &clock{}
}

type IClock interface {
	Now() time.Time
	NextMinuteDuration(now time.Time) time.Duration
	NextAfternoonClosingDuration(now time.Time) time.Duration
	IsTradingTime(now time.Time) bool
}

type clock struct{}

func (c *clock) Now() time.Time {
	return time.Now()
}

// NextMinuteDuration - 次の0秒のタイミングまでのDurationを返す
func (c *clock) NextMinuteDuration(now time.Time) time.Duration {
	nextTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	return nextTime.Sub(now)
}

// NextAfternoonClosingDuration - 次の後場引け(15:00)までのDurationを返す
func (c *clock) NextAfternoonClosingDuration(now time.Time) time.Duration {
	today1500 := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.Local)
	if today1500.After(now) {
		return today1500.Sub(now)
	}
	return today1500.AddDate(0, 0, 1).Sub(now)
}

// IsTradingTime - 取引可能時刻かを返す
func (c *clock) IsTradingTime(now time.Time) bool {
	nowTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	morningStart := time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)
	morningEnd := time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)
	if !nowTime.Before(morningStart) && !nowTime.After(morningEnd) {
		return true
	}

	afternoonStart := time.Date(0, 1, 1, 12, 30, 0, 0, time.Local)
	afternoonEnd := time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)
	return !nowTime.Before(afternoonStart) && !nowTime.After(afternoonEnd)
}
