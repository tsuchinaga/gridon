package gridon

import (
	"reflect"
	"testing"
	"time"
)

type testClock struct {
	IClock
	Now1                time.Time
	NextMinuteDuration1 time.Duration
}

func (t *testClock) Now() time.Time                             { return t.Now1 }
func (t *testClock) NextMinuteDuration(time.Time) time.Duration { return t.NextMinuteDuration1 }

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

func Test_clock_NextMinuteDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  time.Time
		want1 time.Duration
	}{
		{name: "nowが00分なら1分との差、60秒が返される",
			arg1:  time.Date(2021, 12, 31, 8, 0, 0, 0, time.Local),
			want1: 1 * time.Minute},
		{name: "nowが00分30秒なら1分との差、30秒が返される",
			arg1:  time.Date(2021, 12, 31, 8, 0, 30, 0, time.Local),
			want1: 30 * time.Second},
		{name: "nowが59分30秒なら00分との差、30秒が返される",
			arg1:  time.Date(2021, 12, 31, 8, 59, 30, 0, time.Local),
			want1: 30 * time.Second},
		{name: "nowが59分59秒999999999なら00分との差、0.000000001秒が返される",
			arg1:  time.Date(2021, 12, 31, 8, 59, 59, 999999999, time.Local),
			want1: 1 * time.Nanosecond},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clock := &clock{}
			got1 := clock.NextMinuteDuration(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
