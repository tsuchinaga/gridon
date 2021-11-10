package gridon

import "time"

// newClock - clockの取得
func newClock() IClock {
	return &clock{}
}

type IClock interface {
	Now() time.Time
}

type clock struct{}

func (c *clock) Now() time.Time {
	return time.Now()
}
