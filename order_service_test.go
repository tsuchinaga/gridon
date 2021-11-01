package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_orderService_CancelAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		orderStore             *testOrderStore
		kabusAPI               *testKabusAPI
		arg1                   *Strategy
		want1                  error
		wantCancelOrderHistory []interface{}
	}{
		{name: "引数がnilならエラー",
			orderStore:             &testOrderStore{},
			kabusAPI:               &testKabusAPI{},
			arg1:                   nil,
			want1:                  ErrNilArgument,
			wantCancelOrderHistory: nil},
		{name: "注文一覧取得に失敗したらエラー",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusAPI:               &testKabusAPI{},
			arg1:                   &Strategy{Code: "strategy-code-001"},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: nil},
		{name: "注文一覧が空なら何もしない",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:               &testKabusAPI{},
			arg1:                   &Strategy{Code: "strategy-code-001"},
			want1:                  nil,
			wantCancelOrderHistory: nil},
		{name: "取消注文の実行に失敗したらエラー",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: ErrUnknown},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "取消注文の実行結果が失敗でもエラーなし",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: false, ResultCode: -1}},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}},
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001", "Password1234", "order-code-002", "Password1234", "order-code-003"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{kabusAPI: test.kabusAPI, orderStore: test.orderStore}
			got1 := service.CancelAll(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantCancelOrderHistory, test.kabusAPI.CancelOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantCancelOrderHistory, got1, test.kabusAPI.CancelOrderHistory)
			}
		})
	}
}

func Test_ExitAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                      string
		clock                                     IClock
		positionStore                             *testPositionStore
		orderStore                                *testOrderStore
		kabusAPI                                  *testKabusAPI
		arg1                                      *Strategy
		want1                                     error
		wantGetActivePositionsByStrategyCodeCount int
		wantHoldCount                             int
		wantReleaseCount                          int
		wantSendOrderHistory                      []interface{}
		wantOrderSave                             []interface{}
	}{
		{name: "引数がnilならerror",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "ポジションの取り出しに失敗したらerror",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          &Strategy{},
			want1:         ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1},
		{name: "取り出したポジションがなければnil",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{}},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          &Strategy{Code: "strategy-code-001"},
			want1:         nil,
			wantGetActivePositionsByStrategyCodeCount: 1},
		{name: "ポジションのholdに失敗したらreleaseしてerror",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300},
				},
				Hold1: ErrUnknown},
			orderStore: &testOrderStore{},
			kabusAPI:   &testKabusAPI{},
			arg1:       &Strategy{Code: "strategy-code-001"},
			want1:      ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount: 1},
		{name: "注文の送信に失敗したらreleaseしてerror",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300},
				}},
			orderStore: &testOrderStore{},
			kabusAPI:   &testKabusAPI{SendOrder2: ErrUnknown},
			arg1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific,
				},
			},
			want1: ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount:    3,
			wantReleaseCount: 3,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific,
					},
				},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarket,
					Price:            0,
					OrderQuantity:    600,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-002", HoldQuantity: 200, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-003", HoldQuantity: 300, ContractQuantity: 0, ReleaseQuantity: 0},
					},
				},
			}},
		{name: "注文保存に失敗したらerror",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150},
				}},
			orderStore: &testOrderStore{Save1: ErrUnknown},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			arg1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific,
				},
			},
			want1: ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount:    3,
			wantReleaseCount: 0,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific,
					},
				},
				&Order{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarket,
					Price:            0,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0},
					},
				},
			},
			wantOrderSave: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				ExecutionType:    ExecutionTypeMarket,
				Price:            0,
				OrderQuantity:    350,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
					{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0},
				},
			}}},
		{name: "注文の保存に成功したらnil",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150},
				}},
			orderStore: &testOrderStore{Save1: nil},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			arg1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific,
				},
			},
			want1: nil,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount:    3,
			wantReleaseCount: 0,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific,
					},
				},
				&Order{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarket,
					Price:            0,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0},
					},
				},
			},
			wantOrderSave: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				ExecutionType:    ExecutionTypeMarket,
				Price:            0,
				OrderQuantity:    350,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
					{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0},
				},
			}}},
		{name: "注文コードが正常終了でなければreleaseしてエラー",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150},
				}},
			orderStore: &testOrderStore{Save1: nil},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: -1, OrderCode: ""}},
			arg1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific,
				},
			},
			want1: ErrOrderCondition,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount:    3,
			wantReleaseCount: 3,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific,
					},
				},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarket,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{kabusAPI: test.kabusAPI, orderStore: test.orderStore, positionStore: test.positionStore, clock: test.clock}
			got1 := service.ExitAll(test.arg1)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantGetActivePositionsByStrategyCodeCount, test.positionStore.GetActivePositionsByStrategyCodeCount) ||
				!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount) ||
				!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount) ||
				!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory) ||
				!reflect.DeepEqual(test.wantOrderSave, test.orderStore.SaveHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantGetActivePositionsByStrategyCodeCount, test.positionStore.GetActivePositionsByStrategyCodeCount),
					!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount),
					!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount),
					!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory),
					!reflect.DeepEqual(test.wantOrderSave, test.orderStore.SaveHistory),
					test.want1, test.wantGetActivePositionsByStrategyCodeCount, test.wantHoldCount, test.wantReleaseCount, test.wantSendOrderHistory, test.wantOrderSave,
					got1, test.positionStore.GetActivePositionsByStrategyCodeCount, test.positionStore.HoldCount, test.positionStore.ReleaseCount, test.kabusAPI.SendOrderHistory, test.orderStore.SaveHistory)
			}
		})
	}
}
