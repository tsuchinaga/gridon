package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testGridService struct {
	IGridService
	Leveling1       error
	LevelingCount   int
	LevelingHistory []interface{}
}

func (t *testGridService) Leveling(strategy *Strategy) error {
	t.LevelingHistory = append(t.LevelingHistory, strategy)
	t.LevelingCount++
	return t.Leveling1
}

func Test_gridService_getBasePrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		clock         *testClock
		kabusAPI      *testKabusAPI
		strategyStore *testStrategyStore
		arg1          *Strategy
		want1         float64
		want2         error
	}{
		{name: "引数がnilならエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			arg1:          nil,
			want1:         0,
			want2:         ErrNilArgument},
		{name: "グリッド戦略の実行時刻が指定されていなければデータ取得は無駄なのでエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			arg1:          &Strategy{},
			want1:         0,
			want2:         ErrNotExistsTimeRange},
		{name: "現在時刻が09:00:00で、最終約定日時が09:00:00なら、最終約定価格を返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 2100,
			want2: nil},
		{name: "現在時刻が09:00:01で、最終約定日時が09:00:00なら、最終約定価格を返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 9, 0, 1, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 2100,
			want2: nil},
		{name: "現在時刻が14:59:59で、最終約定日時が14:59:59なら、最終約定価格を返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 59, 59, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 14, 59, 59, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 2100,
			want2: nil},
		{name: "現在時刻が09:00:00で、最終約定日時が前日なら、銘柄の現在値を返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{CurrentPrice: 2105, CurrentPriceDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)}},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 1, 15, 0, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 2105,
			want2: nil},
		{name: "銘柄の現在値を取れなかったらエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 15, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol2: ErrUnknown},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 15, 0, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 0,
			want2: ErrUnknown},
		{name: "現在時刻が12:30:00で、最終約定日時が現在の時間範囲でないなら、銘柄の現在値を返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 12, 30, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{CurrentPrice: 2105, CurrentPriceDateTime: time.Date(2021, 11, 2, 12, 30, 0, 0, time.Local)}},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 11, 25, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 2105,
			want2: nil},
		{name: "現在時刻が12:30:00で、最終約定日時が現在の時間範囲でなく、現在値の時刻も時間ないなら、エラーを返す",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 12, 30, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{CurrentPrice: 2105, CurrentPriceDateTime: time.Date(2021, 11, 2, 11, 30, 0, 0, time.Local)}},
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 11, 25, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 0,
			want2: ErrCannotGetBasePrice},
		{name: "現在時刻が12:30:00で、最終約定日時が現在の時間範囲でなく、現在値の時刻が時間内でも、戦略に保存しようとして失敗したらエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 12, 30, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{CurrentPrice: 2105, CurrentPriceDateTime: time.Date(2021, 11, 2, 12, 30, 0, 0, time.Local)}},
			strategyStore: &testStrategyStore{SetBasePrice1: ErrUnknown},
			arg1: &Strategy{
				BasePriceDateTime: time.Date(2021, 11, 2, 11, 25, 0, 0, time.Local),
				BasePrice:         2100,
				GridStrategy: GridStrategy{TimeRanges: []TimeRange{
					{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
					{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local)},
				}}},
			want1: 0,
			want2: ErrUnknown},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &gridService{clock: test.clock, kabusAPI: test.kabusAPI, strategyStore: test.strategyStore}
			got1, got2 := service.getBasePrice(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_gridService_sendGridOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		orderService          *testOrderService
		arg1                  *Strategy
		arg2                  float64
		arg3                  float64
		arg4                  float64
		want1                 error
		wantEntryLimitHistory []interface{}
		wantExitLimitHistory  []interface{}
	}{
		{name: "引数がnilならエラー",
			orderService: &testOrderService{},
			arg1:         nil,
			want1:        ErrNilArgument},
		{name: "limitPriceとbasePriceが一致したらエラー",
			orderService: &testOrderService{},
			arg1:         &Strategy{},
			arg2:         2000.0,
			arg3:         2000.0,
			want1:        ErrUndecidableValue},
		{name: "entrySideが買いで、limitPriceがbasePrice未満の場合、entryを叩くが、エラーが返されたらエラー",
			orderService:          &testOrderService{EntryLimit1: ErrUnknown},
			arg1:                  &Strategy{Code: "strategy-code-001", EntrySide: SideBuy},
			arg2:                  2000.0,
			arg3:                  2100.0,
			arg4:                  4,
			want1:                 ErrUnknown,
			wantEntryLimitHistory: []interface{}{"strategy-code-001", 2000.0, 4.0}},
		{name: "entrySideが買いで、limitPriceがbasePrice未満の場合、entryを叩き、エラーがなければエラーなし",
			orderService:          &testOrderService{EntryLimit1: nil},
			arg1:                  &Strategy{Code: "strategy-code-001", EntrySide: SideBuy},
			arg2:                  2000.0,
			arg3:                  2100.0,
			arg4:                  4,
			want1:                 nil,
			wantEntryLimitHistory: []interface{}{"strategy-code-001", 2000.0, 4.0}},
		{name: "entrySideが買いで、limitPriceがbasePriceより大きい場合、exitを叩くが、エラーが返されたらエラー",
			orderService:         &testOrderService{ExitLimit1: ErrUnknown},
			arg1:                 &Strategy{Code: "strategy-code-001", EntrySide: SideBuy},
			arg2:                 2100.0,
			arg3:                 2000.0,
			arg4:                 4,
			want1:                ErrUnknown,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2100.0, 4.0, SortOrderNewest}},
		{name: "entrySideが買いで、limitPriceがbasePriceより大きい場合、exitを叩き、エラーがなければエラーなし",
			orderService:         &testOrderService{ExitLimit1: nil},
			arg1:                 &Strategy{Code: "strategy-code-001", EntrySide: SideBuy},
			arg2:                 2100.0,
			arg3:                 2000.0,
			arg4:                 4,
			want1:                nil,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2100.0, 4.0, SortOrderNewest}},
		{name: "entrySideが売りで、limitPriceがbasePriceより大きい場合、entryを叩くが、エラーが返されたらエラー",
			orderService:          &testOrderService{EntryLimit1: ErrUnknown},
			arg1:                  &Strategy{Code: "strategy-code-001", EntrySide: SideSell},
			arg2:                  2100.0,
			arg3:                  2000.0,
			arg4:                  4,
			want1:                 ErrUnknown,
			wantEntryLimitHistory: []interface{}{"strategy-code-001", 2100.0, 4.0}},
		{name: "entrySideが売りで、limitPriceがbasePriceより大きい場合、entryを叩き、エラーがなければエラーなし",
			orderService:          &testOrderService{EntryLimit1: nil},
			arg1:                  &Strategy{Code: "strategy-code-001", EntrySide: SideSell},
			arg2:                  2100.0,
			arg3:                  2000.0,
			arg4:                  4,
			want1:                 nil,
			wantEntryLimitHistory: []interface{}{"strategy-code-001", 2100.0, 4.0}},
		{name: "entrySideが売りで、limitPriceがbasePrice未満の場合、entryを叩くが、エラーが返されたらエラー",
			orderService:         &testOrderService{ExitLimit1: ErrUnknown},
			arg1:                 &Strategy{Code: "strategy-code-001", EntrySide: SideSell},
			arg2:                 2000.0,
			arg3:                 2100.0,
			arg4:                 4,
			want1:                ErrUnknown,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2000.0, 4.0, SortOrderNewest}},
		{name: "entrySideが売りで、limitPriceがbasePrice未満の場合、entryを叩き、エラーがなければエラーなし",
			orderService:         &testOrderService{ExitLimit1: nil},
			arg1:                 &Strategy{Code: "strategy-code-001", EntrySide: SideSell},
			arg2:                 2000.0,
			arg3:                 2100.0,
			arg4:                 4,
			want1:                nil,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2000.0, 4.0, SortOrderNewest}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &gridService{orderService: test.orderService}
			got1 := service.sendGridOrder(test.arg1, test.arg2, test.arg3, test.arg4)
			if !reflect.DeepEqual(test.want1, got1) ||
				!reflect.DeepEqual(test.wantEntryLimitHistory, test.orderService.EntryLimitHistory) ||
				!reflect.DeepEqual(test.wantExitLimitHistory, test.orderService.ExitLimitHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					!reflect.DeepEqual(test.want1, got1),
					!reflect.DeepEqual(test.wantEntryLimitHistory, test.orderService.EntryLimitHistory),
					!reflect.DeepEqual(test.wantExitLimitHistory, test.orderService.ExitLimitHistory),
					test.want1, test.wantEntryLimitHistory, test.wantExitLimitHistory,
					got1, test.orderService.EntryLimitHistory, test.orderService.ExitLimitHistory)
			}
		})
	}
}

func Test_gridService_Leveling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		clock                 *testClock
		orderService          *testOrderService
		kabusAPI              *testKabusAPI
		strategyStore         *testStrategyStore
		tick                  ITick
		arg1                  *Strategy
		want1                 error
		wantCancelHistory     []interface{}
		wantEntryLimitHistory []interface{}
		wantExitLimitHistory  []interface{}
	}{
		{name: "引数がnilならエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService:  &testOrderService{},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "戦略が実行不可なら何もせず終了",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService:  &testOrderService{},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1:          &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{Runnable: false}},
			want1:         nil},
		{name: "取引時間でないなら何もせずに終了",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 8, 0, 0, 0, time.Local),
				IsTradingTime1: false},
			orderService:  &testOrderService{},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1:          &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{Runnable: false}},
			want1:         nil},
		{name: "注文一覧の取得に失敗したらエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService:  &testOrderService{GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1: ErrUnknown},
		{name: "基準価格取得に失敗したらエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService:  &testOrderService{GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:      &testKabusAPI{GetSymbol2: ErrUnknown},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1: ErrUnknown},
		{name: "widthが0ならエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService:  &testOrderService{GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable: true,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1: ErrZeroGridWidth},
		{name: "グリッドの範囲外の注文の取消で失敗したらエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2112, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2088, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit}},
				Cancel1: ErrUnknown},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 5,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			}},
			want1: ErrUnknown,
			wantCancelHistory: []interface{}{
				&Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 5,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
				"order-code-001"}},
		{name: "グリッドに乗っていない注文の取消",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2101, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2099, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-003", Price: 2102, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-004", Price: 2098, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-005", Price: 2112, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-006", Price: 2088, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit}}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 5,
				TimeRanges:    []TimeRange{{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1: nil,
			wantCancelHistory: []interface{}{
				&Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 5,
					TimeRanges:    []TimeRange{{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
				"order-code-001",
				&Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 5,
					TimeRanges:    []TimeRange{{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
				"order-code-002",
				&Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 5,
					TimeRanges:    []TimeRange{{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
				"order-code-005",
				&Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 5,
					TimeRanges:    []TimeRange{{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
				"order-code-006"}},
		{name: "グリッドの範囲外の指値注文がなければ取消は実行しない",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2102, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2098, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-003", OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeMarket}}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 1,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			}},
			want1:             nil,
			wantCancelHistory: nil},
		{name: "グリッド本数が0本の場合、何もせずに終了",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 0,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			}},
			want1:             nil,
			wantCancelHistory: nil},
		{name: "乗せたいグリッドにすでに注文があり、不足分がなければ、注文しない",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2102, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2098, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-003", Price: 2104, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-004", Price: 2096, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-005", Price: 2106, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-006", Price: 2094, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-007", Price: 2108, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-008", Price: 2092, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-009", Price: 2110, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-010", Price: 2090, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit}}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 5,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			}},
			want1: nil},
		{name: "グリッド本数が0本の場合、何もせずに終了",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{
				Runnable:      true,
				Width:         2,
				Quantity:      4,
				NumberOfGrids: 0,
				TimeRanges: []TimeRange{{
					Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
					End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
			}},
			want1: nil},
		{name: "グリッドに全く注文がなければ、各グリッドに必要数の注文をのせる",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{
				Code:      "strategy-code-001",
				EntrySide: SideBuy,
				GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 2,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
			want1: nil,
			wantEntryLimitHistory: []interface{}{
				"strategy-code-001", 2098.0, 4.0,
				"strategy-code-001", 2096.0, 4.0,
			},
			wantExitLimitHistory: []interface{}{
				"strategy-code-001", 2102.0, 4.0, SortOrderNewest,
				"strategy-code-001", 2104.0, 4.0, SortOrderNewest,
			}},
		{name: "エントリー注文でエラーがでたらエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{},
				EntryLimit1:                    ErrUnknown},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{
				Code:      "strategy-code-001",
				EntrySide: SideBuy,
				GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 2,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
			want1:                 ErrUnknown,
			wantEntryLimitHistory: []interface{}{"strategy-code-001", 2098.0, 4.0},
			wantExitLimitHistory:  []interface{}{"strategy-code-001", 2102.0, 4.0, SortOrderNewest}},
		{name: "エグジット注文でエラーがでたらエラー",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{},
				ExitLimit1:                     ErrUnknown},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{
				Code:      "strategy-code-001",
				EntrySide: SideBuy,
				GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 2,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
			want1:                ErrUnknown,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2102.0, 4.0, SortOrderNewest}},
		{name: "乗せたいグリッドにすでに注文があれば、不足分だけを注文する",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2102, OrderQuantity: 2, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2098, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-003", Price: 2104, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-004", Price: 2096, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
				},
				ExitLimit1: ErrUnknown},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{
				Code:      "strategy-code-001",
				EntrySide: SideBuy,
				GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 2,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
			want1:                ErrUnknown,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2102.0, 2.0, SortOrderNewest}},
		{name: "基準価格のグリッドに注文が残っていたら、隣のグリッドに乗せる数を減らす",
			clock: &testClock{
				Now1:           time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local),
				IsTradingTime1: true},
			orderService: &testOrderService{
				GetActiveOrdersByStrategyCode1: []*Order{
					{Code: "order-code-001", Price: 2100, OrderQuantity: 1, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-002", Price: 2102, OrderQuantity: 2, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-003", Price: 2098, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-004", Price: 2104, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
					{Code: "order-code-005", Price: 2096, OrderQuantity: 4, ContractQuantity: 0, ExecutionType: ExecutionTypeLimit},
				},
				ExitLimit1: ErrUnknown},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, CurrentPriceDateTime: time.Date(2021, 11, 5, 9, 0, 0, 0, time.Local), BidPrice: 2101, AskPrice: 2099}},
			strategyStore: &testStrategyStore{},
			tick:          &tick{},
			arg1: &Strategy{
				Code:      "strategy-code-001",
				EntrySide: SideBuy,
				GridStrategy: GridStrategy{
					Runnable:      true,
					Width:         2,
					Quantity:      4,
					NumberOfGrids: 2,
					TimeRanges: []TimeRange{{
						Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
						End:   time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}},
				}},
			want1:                ErrUnknown,
			wantExitLimitHistory: []interface{}{"strategy-code-001", 2102.0, 1.0, SortOrderNewest}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &gridService{
				clock:         test.clock,
				tick:          test.tick,
				kabusAPI:      test.kabusAPI,
				strategyStore: test.strategyStore,
				orderService:  test.orderService,
			}
			got1 := service.Leveling(test.arg1)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantCancelHistory, test.orderService.CancelHistory) ||
				!reflect.DeepEqual(test.wantEntryLimitHistory, test.orderService.EntryLimitHistory) ||
				!reflect.DeepEqual(test.wantExitLimitHistory, test.orderService.ExitLimitHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantCancelHistory, test.orderService.CancelHistory),
					!reflect.DeepEqual(test.wantEntryLimitHistory, test.orderService.EntryLimitHistory),
					!reflect.DeepEqual(test.wantExitLimitHistory, test.orderService.ExitLimitHistory),
					test.want1, test.wantCancelHistory, test.wantEntryLimitHistory, test.wantExitLimitHistory,
					got1, test.orderService.CancelHistory, test.orderService.EntryLimitHistory, test.orderService.ExitLimitHistory)
			}
		})
	}
}

func Test_newGridService(t *testing.T) {
	t.Parallel()
	clock := &testClock{}
	tick := &tick{}
	kabusAPI := &testKabusAPI{}
	orderService := &testOrderService{}
	strategyStore := &testStrategyStore{}
	want1 := &gridService{
		clock:         clock,
		tick:          tick,
		kabusAPI:      kabusAPI,
		orderService:  orderService,
		strategyStore: strategyStore,
	}
	got1 := newGridService(clock, tick, kabusAPI, orderService, strategyStore)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_gridService_width(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  *Strategy
		want1 int
		want2 error
	}{
		{name: "引数がnilならエラー",
			arg1:  nil,
			want1: 0,
			want2: ErrNilArgument},
		{name: "typeがunspecifiedならwidthをそのまま返す",
			arg1: &Strategy{
				GridStrategy: GridStrategy{
					Width:    2,
					GridType: GridTypeUnspecified,
				},
			},
			want1: 2,
			want2: nil},
		{name: "typeがstaticならwidthをそのまま返す",
			arg1: &Strategy{
				GridStrategy: GridStrategy{
					Width:    2,
					GridType: GridTypeStatic,
				},
			},
			want1: 2,
			want2: nil},
		{name: "typeがDynamicMinMaxならwidthを計算して返す",
			arg1: &Strategy{
				TickGroup:        TickGroupTopix100,
				MinContractPrice: 18125,
				MaxContractPrice: 18275,
				GridStrategy: GridStrategy{
					Width:    2,
					GridType: GridTypeDynamicMinMax,
					DynamicGridMinMax: DynamicGridMinMax{
						Divide:    5,
						Rounding:  RoundingCeil,
						Operation: OperationPlus,
					},
				},
			},
			want1: 8,
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &gridService{tick: &tick{}}
			got1, got2 := service.width(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}
