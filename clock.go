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
