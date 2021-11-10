package gridon

import (
	"reflect"
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

func Test_newClock(t *testing.T) {
	t.Parallel()
	want1 := &clock{}
	got1 := newClock()
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}
