package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_contractService_releaseHoldPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		positionStore    *testPositionStore
		arg1             *Order
		want1            error
		wantReleaseCount int
		wantOrder        *Order
	}{
		{name: "HoldPositionがnilなら何もせず終了",
			positionStore:    &testPositionStore{},
			arg1:             &Order{HoldPositions: nil},
			want1:            nil,
			wantReleaseCount: 0,
			wantOrder:        &Order{HoldPositions: nil}},
		{name: "HoldPositionが空配列なら何もせず終了",
			positionStore:    &testPositionStore{},
			arg1:             &Order{HoldPositions: []HoldPosition{}},
			want1:            nil,
			wantReleaseCount: 0,
			wantOrder:        &Order{HoldPositions: []HoldPosition{}}},
		{name: "HoldPositionに有効な数量がなければ何もせず終了",
			positionStore: &testPositionStore{},
			arg1: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100, ContractQuantity: 100},
				{HoldQuantity: 100, ReleaseQuantity: 100},
			}},
			want1:            nil,
			wantReleaseCount: 0,
			wantOrder: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100, ContractQuantity: 100},
				{HoldQuantity: 100, ReleaseQuantity: 100},
			}}},
		{name: "HoldPositionのReleaseに失敗したらerror",
			positionStore: &testPositionStore{Release1: ErrUnknown},
			arg1: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100},
				{HoldQuantity: 100},
			}},
			want1:            ErrUnknown,
			wantReleaseCount: 1,
			wantOrder: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100},
				{HoldQuantity: 100},
			}}},
		{name: "HoldPositionの返せるポジションの数だけReleaseを叩き、orderを更新する",
			positionStore: &testPositionStore{Release1: nil},
			arg1: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100, ContractQuantity: 50},
				{HoldQuantity: 100, ReleaseQuantity: 30},
				{HoldQuantity: 100, ContractQuantity: 100},
				{HoldQuantity: 100},
			}},
			want1:            nil,
			wantReleaseCount: 3,
			wantOrder: &Order{HoldPositions: []HoldPosition{
				{HoldQuantity: 100, ContractQuantity: 50, ReleaseQuantity: 50},
				{HoldQuantity: 100, ReleaseQuantity: 100},
				{HoldQuantity: 100, ContractQuantity: 100},
				{HoldQuantity: 100, ReleaseQuantity: 100},
			}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{positionStore: test.positionStore}
			got1 := service.releaseHoldPositions(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) ||
				!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount) ||
				!reflect.DeepEqual(test.wantOrder, test.arg1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantReleaseCount, test.wantOrder,
					got1, test.positionStore.ReleaseCount, test.arg1)
			}
		})
	}
}

func Test_contractService_entryContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		positionStore              *testPositionStore
		strategyStore              *testStrategyStore
		arg1                       *Order
		arg2                       Contract
		want1                      error
		wantSaveHistory            []interface{}
		wantAddStrategyCashHistory []interface{}
		wantSetContractHistory     []interface{}
	}{
		{name: "引数がnilならエラー",
			positionStore:          &testPositionStore{},
			strategyStore:          &testStrategyStore{},
			arg1:                   nil,
			arg2:                   Contract{},
			want1:                  ErrNilArgument,
			wantSetContractHistory: nil},
		{name: "ポジションの登録に失敗したらエラー",
			positionStore: &testPositionStore{Save1: ErrUnknown},
			strategyStore: &testStrategyStore{},
			arg1:          &Order{Code: "order-code-001", StrategyCode: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou, Side: SideBuy, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay},
			arg2:          Contract{PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)},
			want1:         ErrUnknown,
			wantSaveHistory: []interface{}{&Position{
				Code:             "position-code-001",
				StrategyCode:     "strategy-code-001",
				OrderCode:        "order-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Side:             SideBuy,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				Price:            2070,
				OwnedQuantity:    4,
				HoldQuantity:     0,
				ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
			}},
			wantSetContractHistory: nil},
		{name: "戦略の最終約定情報保存に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{SetContract1: ErrUnknown},
			arg1:          &Order{Code: "order-code-001", StrategyCode: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou, Side: SideBuy, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay},
			arg2:          Contract{PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)},
			want1:         ErrUnknown,
			wantSaveHistory: []interface{}{&Position{
				Code:             "position-code-001",
				StrategyCode:     "strategy-code-001",
				OrderCode:        "order-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Side:             SideBuy,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				Price:            2070,
				OwnedQuantity:    4,
				HoldQuantity:     0,
				ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
			}},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1 * 2070.0 * 4.0},
			wantSetContractHistory:     []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)}},
		{name: "余力の登録に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{AddStrategyCash1: ErrUnknown},
			arg1:          &Order{Code: "order-code-001", StrategyCode: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou, Side: SideBuy, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay},
			arg2:          Contract{PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)},
			want1:         ErrUnknown,
			wantSaveHistory: []interface{}{&Position{
				Code:             "position-code-001",
				StrategyCode:     "strategy-code-001",
				OrderCode:        "order-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Side:             SideBuy,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				Price:            2070,
				OwnedQuantity:    4,
				HoldQuantity:     0,
				ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
			}},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1 * 2070.0 * 4.0}},
		{name: "最後までエラーなく処理できたらnilを返す",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{},
			arg1:          &Order{Code: "order-code-001", StrategyCode: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou, Side: SideBuy, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay},
			arg2:          Contract{PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)},
			want1:         nil,
			wantSaveHistory: []interface{}{&Position{
				Code:             "position-code-001",
				StrategyCode:     "strategy-code-001",
				OrderCode:        "order-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Side:             SideBuy,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				Price:            2070,
				OwnedQuantity:    4,
				HoldQuantity:     0,
				ContractDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
			}},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1 * 2070.0 * 4.0},
			wantSetContractHistory:     []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{positionStore: test.positionStore, strategyStore: test.strategyStore}
			got1 := service.entryContract(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.positionStore.SaveHistory) ||
				!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory) ||
				!reflect.DeepEqual(test.wantSetContractHistory, test.strategyStore.SetContractHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantSaveHistory, test.positionStore.SaveHistory),
					!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory),
					!reflect.DeepEqual(test.wantSetContractHistory, test.strategyStore.SetContractHistory),
					test.want1, test.wantSaveHistory, test.wantAddStrategyCashHistory, test.wantSetContractHistory,
					got1, test.positionStore.SaveHistory, test.strategyStore.AddStrategyCashHistory, test.strategyStore.SetContractHistory)
			}
		})
	}
}

func Test_contractService_exitContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		positionStore              *testPositionStore
		strategyStore              *testStrategyStore
		arg1                       *Order
		arg2                       Contract
		want1                      error
		wantExitContractHistory    []interface{}
		wantAddStrategyCashHistory []interface{}
		wantSetContractHistory     []interface{}
		wantOrder                  *Order
	}{
		{name: "引数がnilならエラー",
			positionStore:              &testPositionStore{},
			strategyStore:              &testStrategyStore{},
			arg1:                       nil,
			arg2:                       Contract{},
			want1:                      ErrNilArgument,
			wantExitContractHistory:    nil,
			wantAddStrategyCashHistory: nil,
			wantSetContractHistory:     nil,
			wantOrder:                  nil},
		{name: "ポジションの返済約定登録に失敗したらエラー",
			positionStore: &testPositionStore{ExitContract1: ErrUnknown},
			strategyStore: &testStrategyStore{},
			arg1: &Order{HoldPositions: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
				{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
				{PositionCode: "position-code-003", HoldQuantity: 100},
			}},
			arg2:                       Contract{Quantity: 100},
			want1:                      ErrUnknown,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0},
			wantAddStrategyCashHistory: nil,
			wantSetContractHistory:     nil,
			wantOrder: &Order{HoldPositions: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
				{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
				{PositionCode: "position-code-003", HoldQuantity: 100},
			}}},
		{name: "現金余力の更新に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{AddStrategyCash1: ErrUnknown},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
					{PositionCode: "position-code-003", HoldQuantity: 100},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100},
			want1:                      ErrUnknown,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0, "position-code-002", 30.0, "position-code-003", 50.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 207_000.0},
			wantSetContractHistory:     nil,
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 30, ReleaseQuantity: 70},
					{PositionCode: "position-code-003", HoldQuantity: 100, ContractQuantity: 50},
				}}},
		{name: "戦略の最終約定情報の更新に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{SetContract1: ErrUnknown},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
					{PositionCode: "position-code-003", HoldQuantity: 100},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100, ContractDateTime: time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			want1:                      ErrUnknown,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0, "position-code-002", 30.0, "position-code-003", 50.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 207_000.0},
			wantSetContractHistory:     []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 30, ReleaseQuantity: 70},
					{PositionCode: "position-code-003", HoldQuantity: 100, ContractQuantity: 50},
				}}},
		{name: "エラーがなかったらnilが返される",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 100},
					{PositionCode: "position-code-003", HoldQuantity: 100},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100, ContractDateTime: time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			want1:                      nil,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0, "position-code-002", 30.0, "position-code-003", 50.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 207_000.0},
			wantSetContractHistory:     []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 30, ReleaseQuantity: 70},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 100},
					{PositionCode: "position-code-003", HoldQuantity: 100, ContractQuantity: 50},
				}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{positionStore: test.positionStore, strategyStore: test.strategyStore}
			got1 := service.exitContract(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantExitContractHistory, test.positionStore.ExitContractHistory) ||
				!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory) ||
				!reflect.DeepEqual(test.wantSetContractHistory, test.strategyStore.SetContractHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantExitContractHistory, test.positionStore.ExitContractHistory),
					!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory),
					!reflect.DeepEqual(test.wantSetContractHistory, test.strategyStore.SetContractHistory),
					test.want1, test.wantExitContractHistory, test.wantAddStrategyCashHistory, test.wantSetContractHistory,
					got1, test.positionStore.ExitContractHistory, test.strategyStore.AddStrategyCashHistory, test.strategyStore.SetContractHistory)
			}
		})
	}
}

func Test_contractService_Confirm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kabusAPI                               *testKabusAPI
		orderStore                             *testOrderStore
		positionStore                          *testPositionStore
		strategyStore                          *testStrategyStore
		name                                   string
		arg1                                   *Strategy
		want1                                  error
		wantGetOrdersCount                     int
		wantGetActiveOrdersByStrategyCodeCount int
		wantSaveHistory                        []interface{}
	}{
		{name: "引数がnilならerror",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "kabusapiから注文一覧が取れなければerror",
			kabusAPI:           &testKabusAPI{GetOrders2: ErrUnknown},
			orderStore:         &testOrderStore{},
			positionStore:      &testPositionStore{},
			strategyStore:      &testStrategyStore{},
			arg1:               &Strategy{Code: "strategy-code-001"},
			want1:              ErrUnknown,
			wantGetOrdersCount: 1},
		{name: "storeから注文一覧が取れなければerror",
			kabusAPI:                               &testKabusAPI{GetOrders1: []SecurityOrder{}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "kabusapiから返ってきた注文が空なら何もしない",
			kabusAPI:                               &testKabusAPI{GetOrders1: []SecurityOrder{}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "storeから返ってきた注文が空なら何もしない",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{Code: "order-code-001"},
			}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "kabusapiとstoreの注文でコードが一致するものがなければ何もしない",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-003",
					Status:           OrderStatusInOrder,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    nil,
				},
			}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "kabusapiとstoreの注文でコードが一致しても更新するものがなければ何もしない",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusInOrder,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    nil,
				},
			}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "entryに失敗したらerror",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusDone,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    nil,
				},
			}},
			positionStore:                          &testPositionStore{Save1: ErrUnknown},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "exitに失敗したらerror",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusDone,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4, ContractQuantity: 0, ReleaseQuantity: 0}},
				},
			}},
			positionStore:                          &testPositionStore{ExitContract1: ErrUnknown},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "releaseに失敗したらerror",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusCanceled,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 2, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4, ContractQuantity: 0, ReleaseQuantity: 0}},
				},
			}},
			positionStore:                          &testPositionStore{Release1: ErrUnknown},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "注文の保存に失敗したらerror",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusDone,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				},
			}},
			orderStore: &testOrderStore{
				Save1: ErrUnknown,
				GetActiveOrdersByStrategyCode1: []*Order{
					{
						Code:             "order-code-001",
						StrategyCode:     "strategy-code-001",
						SymbolCode:       "1475",
						Exchange:         ExchangeToushou,
						Status:           OrderStatusInOrder,
						Product:          ProductMargin,
						MarginTradeType:  MarginTradeTypeDay,
						TradeType:        TradeTypeEntry,
						Side:             SideBuy,
						Price:            2070,
						OrderQuantity:    4,
						ContractQuantity: 0,
						AccountType:      AccountTypeSpecific,
						OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
						ContractDateTime: time.Time{},
						CancelDateTime:   time.Time{},
						Contracts:        nil,
						HoldPositions:    nil,
					},
				}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1,
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusDone,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				Price:            2070,
				OrderQuantity:    4,
				ContractQuantity: 4,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				HoldPositions:    nil,
			}}},
		{name: "すべて処理できたらerrorなく終了",
			kabusAPI: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:             "order-code-001",
					Status:           OrderStatusDone,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 29, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				},
			}},
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2070,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    nil,
				},
			}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetOrdersCount:                     1,
			wantGetActiveOrdersByStrategyCodeCount: 1,
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusDone,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				Price:            2070,
				OrderQuantity:    4,
				ContractQuantity: 4,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local),
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2070, Quantity: 4, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
				HoldPositions:    nil,
			}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{
				kabusAPI:      test.kabusAPI,
				strategyStore: test.strategyStore,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
			}
			got1 := service.Confirm(test.arg1)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantGetOrdersCount, test.kabusAPI.GetOrdersCount) ||
				!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeCount, test.orderStore.GetActiveOrdersByStrategyCodeCount) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantGetOrdersCount, test.kabusAPI.GetOrdersCount),
					!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeCount, test.orderStore.GetActiveOrdersByStrategyCodeCount),
					!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory),
					test.want1, test.wantGetOrdersCount, test.wantGetActiveOrdersByStrategyCodeCount, test.wantSaveHistory,
					got1, test.kabusAPI.GetOrdersCount, test.orderStore.GetActiveOrdersByStrategyCodeCount, test.orderStore.SaveHistory)
			}
		})
	}
}
