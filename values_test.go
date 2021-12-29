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

func Test_GridStrategy_IsRunnable(t *testing.T) {
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
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 8, 59, 59, 0, time.Local),
			want1: false},
		{name: "ちょうど開始時刻ならtrue",
			gridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local),
			want1: true},
		{name: "開始時刻以降ならtrue",
			gridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 9, 0, 1, 0, time.Local),
			want1: true},
		{name: "終了時刻以前ならtrue",
			gridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 14, 54, 59, 0, time.Local),
			want1: true},
		{name: "ちょうど終了時刻ならfalse",
			gridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "終了時刻以降ならfalse",
			gridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 5, 14, 55, 1, 0, time.Local),
			want1: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.gridStrategy.IsRunnable(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_RebalanceStrategy_IsRunnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		rebalanceStrategy RebalanceStrategy
		arg1              time.Time
		want1             bool
	}{
		{name: "実行不可ならfalse",
			rebalanceStrategy: RebalanceStrategy{Runnable: false},
			arg1:              time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1:             false},
		{name: "実行可能でも、実行タイミングがnilならfalse",
			rebalanceStrategy: RebalanceStrategy{
				Runnable: true,
				Timings:  nil},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングが空配列ならfalse",
			rebalanceStrategy: RebalanceStrategy{
				Runnable: true,
				Timings:  []time.Time{}},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分がなければfalse",
			rebalanceStrategy: RebalanceStrategy{
				Runnable: true,
				Timings: []time.Time{
					time.Date(0, 1, 1, 11, 25, 0, 0, time.Local),
					time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			arg1:  time.Date(2021, 11, 10, 10, 0, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分があればtrue",
			rebalanceStrategy: RebalanceStrategy{
				Runnable: true,
				Timings: []time.Time{
					time.Date(0, 1, 1, 11, 25, 0, 0, time.Local),
					time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			arg1:  time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local),
			want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.rebalanceStrategy.IsRunnable(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_ExitStrategy_IsRunnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		exitStrategy ExitStrategy
		arg1         time.Time
		want1        bool
	}{
		{name: "実行不可ならfalse",
			exitStrategy: ExitStrategy{Runnable: false},
			arg1:         time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1:        false},
		{name: "実行可能でも、実行タイミングがnilならfalse",
			exitStrategy: ExitStrategy{
				Runnable:   true,
				Conditions: nil},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングが空配列ならfalse",
			exitStrategy: ExitStrategy{
				Runnable:   true,
				Conditions: []ExitCondition{}},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分がなければfalse",
			exitStrategy: ExitStrategy{
				Runnable: true,
				Conditions: []ExitCondition{
					{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 25, 0, 0, time.Local)},
					{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 10, 10, 0, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分があればtrue",
			exitStrategy: ExitStrategy{
				Runnable: true,
				Conditions: []ExitCondition{
					{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 25, 0, 0, time.Local)},
					{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local),
			want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.exitStrategy.IsRunnable(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_CancelStrategy_IsRunnable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		cancelStrategy CancelStrategy
		arg1           time.Time
		want1          bool
	}{
		{name: "実行不可ならfalse",
			cancelStrategy: CancelStrategy{Runnable: false},
			arg1:           time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1:          false},
		{name: "実行可能でも、実行タイミングがnilならfalse",
			cancelStrategy: CancelStrategy{
				Runnable: true,
				Timings:  nil},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングが空配列ならfalse",
			cancelStrategy: CancelStrategy{
				Runnable: true,
				Timings:  []time.Time{}},
			arg1:  time.Date(0, 1, 1, 14, 55, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分がなければfalse",
			cancelStrategy: CancelStrategy{
				Runnable: true,
				Timings: []time.Time{
					time.Date(0, 1, 1, 11, 25, 0, 0, time.Local),
					time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			arg1:  time.Date(2021, 11, 10, 10, 0, 0, 0, time.Local),
			want1: false},
		{name: "実行可能でも、実行タイミングに引数と同じ時分があればtrue",
			cancelStrategy: CancelStrategy{
				Runnable: true,
				Timings: []time.Time{
					time.Date(0, 1, 1, 11, 25, 0, 0, time.Local),
					time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			arg1:  time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local),
			want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.cancelStrategy.IsRunnable(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_TimeRange_In(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		timeRange TimeRange
		arg1      time.Time
		want1     bool
	}{
		{name: "年月日は無視される",
			timeRange: TimeRange{
				Start: time.Date(2021, 11, 11, 9, 0, 0, 0, time.Local),
				End:   time.Date(2021, 10, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(2022, 1, 1, 10, 0, 0, 0, time.Local),
			want1: true},
		{name: "ゼロ値はfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 0, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)},
			arg1:  time.Time{},
			want1: false},
		{name: "start < endのとき、引数 < startならfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 8, 59, 59, 999999999, time.Local),
			want1: false},
		{name: "start < endのとき、start == 引数ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
			want1: true},
		{name: "start < endのとき、start < 引数 < endならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want1: true},
		{name: "start < endのとき、引数 == endならfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 11, 30, 0, 0, time.Local),
			want1: false},
		{name: "start < endのとき、end < 引数ならfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 12, 0, 0, 0, time.Local),
			want1: false},
		{name: "end < startのとき、引数 == 00:00:00ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 0, 0, 0, 0, time.Local),
			want1: true},
		{name: "end < startのとき、00:00:00 < 引数 < endならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 8, 0, 0, 0, time.Local),
			want1: true},
		{name: "end < startのとき、引数 == endならfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
			want1: false},
		{name: "end < startのとき、end < 引数 < startならfalse",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want1: false},
		{name: "end < startのとき、引数 == startならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
			want1: true},
		{name: "end < startのとき、start < 引数 < 23:59:59ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 16, 0, 0, 0, time.Local),
			want1: true},
		{name: "end < startのとき、引数 == 23:59:59ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 23, 59, 59, 999999999, time.Local),
			want1: true},
		{name: "start == endのとき、引数 == 00:00:00ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 0, 0, 0, 0, time.Local),
			want1: true},
		{name: "start == endのとき、引数 == start == endならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
			want1: true},
		{name: "start == endのとき、引数 == 23:59:59ならtrue",
			timeRange: TimeRange{
				Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				End:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
			arg1:  time.Date(0, 1, 1, 23, 59, 59, 999999999, time.Local),
			want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.timeRange.In(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_ExitStrategy_ExecutionType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		exitStrategy ExitStrategy
		arg1         time.Time
		want1        ExecutionType
	}{
		{name: "実行可能な戦略でなければunspecified",
			exitStrategy: ExitStrategy{Runnable: false},
			arg1:         time.Date(2021, 11, 25, 14, 59, 0, 0, time.Local),
			want1:        ExecutionTypeUnspecified},
		{name: "実行可能時間自体がなければunspecified",
			exitStrategy: ExitStrategy{
				Runnable:   true,
				Conditions: nil},
			arg1:  time.Date(2021, 11, 25, 14, 59, 0, 0, time.Local),
			want1: ExecutionTypeUnspecified},
		{name: "実行可能時間でなければunspecified",
			exitStrategy: ExitStrategy{
				Runnable: true,
				Conditions: []ExitCondition{
					{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 25, 0, 0, time.Local)},
					{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 25, 14, 59, 0, 0, time.Local),
			want1: ExecutionTypeUnspecified},
		{name: "実行可能時間の執行条件を返す",
			exitStrategy: ExitStrategy{
				Runnable: true,
				Conditions: []ExitCondition{
					{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 29, 0, 0, time.Local)},
					{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)}}},
			arg1:  time.Date(2021, 11, 25, 14, 59, 0, 0, time.Local),
			want1: ExecutionTypeMarketAfternoonClose},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.exitStrategy.ExecutionType(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_DynamicGridMinMax_Width(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		dynamicGridMinMax DynamicGridMinMax
		arg1              int
		arg2              int
		want1             int
	}{
		{name: "divideが0で加算ならwidthが返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    0,
				Rounding:  RoundingFloor,
				Operation: OperationPlus,
			},
			arg1:  2,
			arg2:  10,
			want1: 2},
		{name: "divideが1で加算ならwidth + diffが返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    1,
				Rounding:  RoundingFloor,
				Operation: OperationPlus,
			},
			arg1:  2,
			arg2:  10,
			want1: 12},
		{name: "divideが2で加算ならwidth + diff / 2が返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    2,
				Rounding:  RoundingFloor,
				Operation: OperationPlus,
			},
			arg1:  2,
			arg2:  10,
			want1: 7},
		{name: "divideが0で積算ならwidthが返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    0,
				Rounding:  RoundingFloor,
				Operation: OperationMultiple,
			},
			arg1:  2,
			arg2:  10,
			want1: 2},
		{name: "divideが1で積算ならwidth x diffが返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    1,
				Rounding:  RoundingFloor,
				Operation: OperationMultiple,
			},
			arg1:  2,
			arg2:  10,
			want1: 20},
		{name: "divideが2で積算ならwidth x diff / 2が返される",
			dynamicGridMinMax: DynamicGridMinMax{
				Divide:    2,
				Rounding:  RoundingFloor,
				Operation: OperationMultiple,
			},
			arg1:  2,
			arg2:  10,
			want1: 10},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.dynamicGridMinMax.Width(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
