package gridon

import (
	"reflect"
	"testing"
	"time"
)

func Test_HoldPosition_LeaveQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		holdPosition HoldPosition
		want1        float64
	}{
		{name: "Hold - ExitContract - Releaseが返される 1", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0}, want1: 100},
		{name: "Hold - ExitContract - Releaseが返される 2", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 20, ReleaseQuantity: 0}, want1: 80},
		{name: "Hold - ExitContract - Releaseが返される 3", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 100, ReleaseQuantity: 0}, want1: 0},
		{name: "Hold - ExitContract - Releaseが返される 4", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 20}, want1: 80},
		{name: "Hold - ExitContract - Releaseが返される 5", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 100}, want1: 0},
		{name: "Hold - ExitContract - Releaseが返される 6", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 30, ReleaseQuantity: 70}, want1: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.holdPosition.LeaveQuantity()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_GridStrategy_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		gridStrategy GridStrategy
		arg1         time.Time
		want1        bool
	}{
		{name: "runnableでなければfalse",
			gridStrategy: GridStrategy{Runnable: false},
			arg1:         time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
			want1:        false},
		{name: "開始時刻以前ならfalse",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 8, 59, 59, 0, time.Local),
			want1: false},
		{name: "ちょうど開始時刻ならtrue",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local),
			want1: true},
		{name: "開始時刻以降ならtrue",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 9, 0, 1, 0, time.Local),
			want1: true},
		{name: "終了時刻以前ならtrue",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 14, 54, 59, 0, time.Local),
			want1: true},
		{name: "ちょうど終了時刻ならfalse",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "終了時刻以降ならfalse",
			gridStrategy: GridStrategy{
				Runnable:  true,
				StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				EndTime:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)},
			arg1:  time.Date(2021, 11, 5, 14, 55, 1, 0, time.Local),
			want1: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.gridStrategy.IsValid(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
