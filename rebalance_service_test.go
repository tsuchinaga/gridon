package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testRebalanceService struct {
	IRebalanceService
	Rebalance1       error
	RebalanceCount   int
	RebalanceHistory []interface{}
}

func (t *testRebalanceService) Rebalance(strategy *Strategy) error {
	t.RebalanceHistory = append(t.RebalanceHistory, strategy)
	t.RebalanceCount++
	return t.Rebalance1
}

func Test_rebalanceService_rebalanceQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  float64 // cash
		arg2  float64 // price
		arg3  float64 // positionValue
		arg4  float64 // tradeUnit
		want1 float64
	}{
		{name: "第2引数が0なら何もせず0を返す", arg1: 100_000, arg2: 0, arg3: 0, arg4: 1, want1: 0},
		{name: "第4引数が0なら何もせず0を返す", arg1: 100_000, arg2: 2_000, arg3: 100_000, arg4: 0, want1: 0},
		{name: "arg1とarg2*arg3が同じ場合、調整が不要なので0", arg1: 100_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が 1 なら、0", arg1: 99_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が price/2 - 1 なら、0", arg1: 99_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が price/2 なら、0", arg1: 99_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が price/2 + 1 なら、0", arg1: 98_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が price - 1 なら、0", arg1: 98_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 < arg2*arg3 で、差が price なら、-1", arg1: 98_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price + 1 なら、-1", arg1: 97_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price*2 - 1 なら、-1", arg1: 96_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price*2 なら、-1", arg1: 96_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price*2 + 1 なら、-1", arg1: 95_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price*3 - 1 なら、-1", arg1: 94_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -1},
		{name: "arg1 < arg2*arg3 で、差が price*3 なら、-2", arg1: 94_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -2},
		{name: "arg1 < arg2*arg3 で、差が price*3 + 1 なら、-2", arg1: 93_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -2},
		{name: "arg1 < arg2*arg3 で、差が price*4 - 1 なら、-2", arg1: 92_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -2},
		{name: "arg1 < arg2*arg3 で、差が price*4 なら、-2", arg1: 92_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -2},
		{name: "arg1 < arg2*arg3 で、差が price*4 + 1 なら、-2", arg1: 91_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: -2},
		{name: "arg1 > arg2*arg3 で、差が 1 なら、0", arg1: 100_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 > arg2*arg3 で、差が price/2 - 1 なら、0", arg1: 100_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 > arg2*arg3 で、差が price/2 なら、0", arg1: 101_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 > arg2*arg3 で、差が price/2 + 1 なら、0", arg1: 101_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 > arg2*arg3 で、差が price - 1 なら、0", arg1: 101_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 0},
		{name: "arg1 > arg2*arg3 で、差が price なら、1", arg1: 102_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 > arg2*arg3 で、差が price + 1 なら、1", arg1: 102_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 > arg2*arg3 で、差が price*2 - 1 なら、1", arg1: 103_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 > arg2*arg3 で、差が price*2 なら、1", arg1: 104_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 < arg2*arg3 で、差が price*2 + 1 なら、1", arg1: 104_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 > arg2*arg3 で、差が price*3 - 1 なら、1", arg1: 105_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 1},
		{name: "arg1 > arg2*arg3 で、差が price*3 なら、2", arg1: 106_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 2},
		{name: "arg1 > arg2*arg3 で、差が price*3 + 1 なら、2", arg1: 106_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 2},
		{name: "arg1 > arg2*arg3 で、差が price*4 - 1 なら、2", arg1: 107_999, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 2},
		{name: "arg1 > arg2*arg3 で、差が price*4 なら、2", arg1: 108_000, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 2},
		{name: "arg1 > arg2*arg3 で、差が price*4 + 1 なら、2", arg1: 108_001, arg2: 2_000, arg3: 100_000, arg4: 1, want1: 2},
		{name: "tradeUnitが10でちょうどで割れる場合", arg1: 100_000, arg2: 200, arg3: 0, arg4: 10, want1: 250},
		{name: "tradeUnitが10でぎりぎり半分を超えない場合", arg1: 100_000, arg2: 204, arg3: 0, arg4: 10, want1: 250},
		{name: "tradeUnitが10でぎりぎり半分を超える場合", arg1: 100_000, arg2: 204.3, arg3: 0, arg4: 10, want1: 240},
	}

	for _, test := range tests {
		test := test
		service := &rebalanceService{}
		got1 := service.rebalanceQuantity(test.arg1, test.arg2, test.arg3, test.arg4)
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_rebalanceService_Rebalance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		clock                  *testClock
		kabusAPI               *testKabusAPI
		positionStore          *testPositionStore
		orderService           *testOrderService
		arg1                   *Strategy
		want1                  error
		wantEntryMarketHistory []interface{}
		wantExitMarketHistory  []interface{}
	}{
		{name: "引数がnilならエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			positionStore: &testPositionStore{},
			orderService:  &testOrderService{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "戦略が実行可能でないなら何もせずに終了",
			clock:         &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			positionStore: &testPositionStore{},
			orderService:  &testOrderService{},
			arg1:          &Strategy{Runnable: false},
			want1:         nil},
		{name: "戦略が実行可能でなければ何もせずに終了",
			clock:         &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol2: ErrUnknown},
			positionStore: &testPositionStore{},
			orderService:  &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				RebalanceStrategy: RebalanceStrategy{Runnable: false},
				Runnable:          true},
			want1: nil},
		{name: "銘柄取得に失敗したらエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol2: ErrUnknown},
			positionStore: &testPositionStore{},
			orderService:  &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1: ErrUnknown},
		{name: "ポジション一覧取得に失敗したらエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2100, BidPrice: 2101, AskPrice: 2099}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			orderService:  &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1: ErrUnknown},
		{name: "rebalanceの数量が0なら何もせず終了",
			clock:    &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2000, BidPrice: 2001, AskPrice: 1999}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 10},
				{Code: "position-code-002", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 15},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 25},
			}},
			orderService: &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				Cash:              100_000,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1: nil},
		{name: "rebalanceの数量が負の値ならExitを呼ぶ",
			clock:    &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2000, BidPrice: 2001, AskPrice: 1999}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 10},
				{Code: "position-code-002", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 15},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 25},
			}},
			orderService: &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				Cash:              50_000,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1:                 nil,
			wantExitMarketHistory: []interface{}{"strategy-code-001", 13.0, SortOrderLatest}},
		{name: "Exitで失敗したらエラー",
			clock:    &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2000, BidPrice: 2001, AskPrice: 1999}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 10},
				{Code: "position-code-002", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 15},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 25},
			}},
			orderService: &testOrderService{ExitMarket1: ErrUnknown},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				Cash:              50_000,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1:                 ErrUnknown,
			wantExitMarketHistory: []interface{}{"strategy-code-001", 13.0, SortOrderLatest}},
		{name: "rebalanceの数量が正の値ならEntryを呼ぶ",
			clock:    &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2000, BidPrice: 2001, AskPrice: 1999}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 10},
				{Code: "position-code-002", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 15},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 25},
			}},
			orderService: &testOrderService{},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				Cash:              150_000,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1:                  nil,
			wantEntryMarketHistory: []interface{}{"strategy-code-001", 13.0}},
		{name: "Entryで失敗したらエラー",
			clock:    &testClock{Now1: time.Date(2021, 11, 10, 8, 59, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{GetSymbol1: &Symbol{Code: "1475", Exchange: ExchangeToushou, TradingUnit: 1, CurrentPrice: 2000, BidPrice: 2001, AskPrice: 1999}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 10},
				{Code: "position-code-002", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 15},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", Side: SideBuy, Price: 2_000, OwnedQuantity: 25},
			}},
			orderService: &testOrderService{EntryMarket1: ErrUnknown},
			arg1: &Strategy{
				Code:              "strategy-code-001",
				SymbolCode:        "1475",
				Exchange:          ExchangeToushou,
				Cash:              150_000,
				RebalanceStrategy: RebalanceStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 8, 59, 0, 0, time.Local)}},
				Runnable:          true},
			want1:                  ErrUnknown,
			wantEntryMarketHistory: []interface{}{"strategy-code-001", 13.0}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &rebalanceService{clock: test.clock, kabusAPI: test.kabusAPI, positionStore: test.positionStore, orderService: test.orderService}
			got1 := service.Rebalance(test.arg1)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantEntryMarketHistory, test.orderService.EntryMarketHistory) ||
				!reflect.DeepEqual(test.wantExitMarketHistory, test.orderService.ExitMarketHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantEntryMarketHistory, test.orderService.EntryMarketHistory),
					!reflect.DeepEqual(test.wantExitMarketHistory, test.orderService.ExitMarketHistory),
					test.want1, test.wantEntryMarketHistory, test.wantExitMarketHistory,
					got1, test.orderService.EntryMarketHistory, test.orderService.ExitMarketHistory)
			}
		})
	}
}

func Test_newRebalanceService(t *testing.T) {
	t.Parallel()
	clock := &testClock{}
	kabusAPI := &testKabusAPI{}
	positionStore := &testPositionStore{}
	orderService := &testOrderService{}
	want1 := &rebalanceService{
		clock:         clock,
		kabusAPI:      kabusAPI,
		positionStore: positionStore,
		orderService:  orderService,
	}
	got1 := newRebalanceService(clock, kabusAPI, positionStore, orderService)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_rebalanceService_positionValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		positionStore *testPositionStore
		arg1          string
		arg2          float64
		want1         float64
		want2         error
	}{
		{name: "ポジション一覧取得に失敗したらエラー",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			arg1:          "strategy-code-001",
			arg2:          2000,
			want1:         0,
			want2:         ErrUnknown},
		{name: "ポジションがなければ0",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{}},
			arg1:          "strategy-code-001",
			arg2:          2000,
			want1:         0,
			want2:         nil},
		{name: "複数の買いポジションの評価額を計算",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", Side: SideBuy, OwnedQuantity: 10, Price: 2010},
				{Code: "position-code-002", Side: SideBuy, OwnedQuantity: 20, HoldQuantity: 10, Price: 2000},
				{Code: "position-code-003", Side: SideBuy, OwnedQuantity: 30, HoldQuantity: 30, Price: 1990},
			}},
			arg1:  "strategy-code-001",
			arg2:  2000,
			want1: 120_000,
			want2: nil},
		{name: "複数の売りポジションの評価額を計算",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", Side: SideSell, OwnedQuantity: 10, Price: 2010},
				{Code: "position-code-002", Side: SideSell, OwnedQuantity: 20, HoldQuantity: 10, Price: 2000},
				{Code: "position-code-003", Side: SideSell, OwnedQuantity: 30, HoldQuantity: 30, Price: 1990},
			}},
			arg1:  "strategy-code-001",
			arg2:  2000,
			want1: 119_600,
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &rebalanceService{positionStore: test.positionStore}
			got1, got2 := service.positionValue(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}
