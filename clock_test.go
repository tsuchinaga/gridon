package gridon

import (
	"testing"
	"time"
)

type testClock struct {
	IClock
	Now1 time.Time
}

func (t *testClock) Now() time.Time {
	return t.Now1
}

func Test_clock_Now(t *testing.T) {
	t.Parallel()
	want := time.Now()
	clock := &clock{}
	got := clock.Now()
	if want.After(got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}
