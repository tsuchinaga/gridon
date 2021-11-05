package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testOrderService struct {
	IOrderService
	EntryLimit1                          error
	EntryLimitCount                      int
	EntryLimitHistory                    []interface{}
	ExitLimit1                           error
	ExitLimitCount                       int
	ExitLimitHistory                     []interface{}
	GetActiveOrdersByStrategyCode1       []*Order
	GetActiveOrdersByStrategyCode2       error
	GetActiveOrdersByStrategyCodeCount   int
	GetActiveOrdersByStrategyCodeHistory []interface{}
	Cancel1                              error
	CancelCount                          int
	CancelHistory                        []interface{}
}

func (t *testOrderService) GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error) {
	t.GetActiveOrdersByStrategyCodeHistory = append(t.GetActiveOrdersByStrategyCodeHistory, strategyCode)
	t.GetActiveOrdersByStrategyCodeCount++
	return t.GetActiveOrdersByStrategyCode1, t.GetActiveOrdersByStrategyCode2
}
func (t *testOrderService) EntryLimit(strategyCode string, price float64, quantity float64) error {
	t.EntryLimitHistory = append(t.EntryLimitHistory, strategyCode)
	t.EntryLimitHistory = append(t.EntryLimitHistory, price)
	t.EntryLimitHistory = append(t.EntryLimitHistory, quantity)
	t.EntryLimitCount++
	return t.EntryLimit1
}
func (t *testOrderService) ExitLimit(strategyCode string, price float64, quantity float64, sortOrder SortOrder) error {
	t.ExitLimitHistory = append(t.ExitLimitHistory, strategyCode)
	t.ExitLimitHistory = append(t.ExitLimitHistory, price)
	t.ExitLimitHistory = append(t.ExitLimitHistory, quantity)
	t.ExitLimitHistory = append(t.ExitLimitHistory, sortOrder)
	t.ExitLimitCount++
	return t.ExitLimit1
}
func (t *testOrderService) Cancel(strategy *Strategy, orderCode string) error {
	t.CancelHistory = append(t.CancelHistory, strategy)
	t.CancelHistory = append(t.CancelHistory, orderCode)
	t.CancelCount++
	return t.Cancel1
}

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

func Test_orderService_GetActiveOrdersByStrategyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		orderStore *testOrderStore
		arg1       string
		want1      []*Order
		want2      error
	}{
		{name: "storeがerrorを返したらerrorを返す",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			want1:      nil,
			want2:      ErrUnknown},
		{name: "storeがnil, nilを返したらnil, nilを返す",
			orderStore: &testOrderStore{},
			want1:      nil,
			want2:      nil},
		{name: "storeが配列を返したら配列を返す",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			want1:      []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}},
			want2:      nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{orderStore: test.orderStore}
			got1, got2 := service.GetActiveOrdersByStrategyCode(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_orderService_EntryLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		kabusAPI                   *testKabusAPI
		orderStore                 *testOrderStore
		strategyStore              *testStrategyStore
		clock                      IClock
		arg1                       string
		arg2                       float64
		arg3                       float64
		want1                      error
		wantSendOrderHistory       []interface{}
		wantSaveHistory            []interface{}
		wantAddStrategyCashHistory []interface{}
	}{
		{name: "指定された戦略がなければエラー",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			clock:         &testClock{},
			arg1:          "strategy-code-001",
			want1:         ErrNoData},
		{name: "余力が足りないならエラー",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Cash: 8_000}},
			clock:         &testClock{},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrNotEnoughCash},
		{name: "SendOrderに失敗したらerror",
			kabusAPI:   &testKabusAPI{SendOrder2: ErrUnknown},
			orderStore: &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            10_000,
				GridStrategy:    GridStrategy{},
				Account:         Account{AccountType: AccountTypeSpecific}}},
			clock: &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			want1: ErrUnknown,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				},
			}},
		{name: "注文に失敗したらerror",
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: 4}},
			orderStore: &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            10_000,
				GridStrategy:    GridStrategy{},
				Account:         Account{AccountType: AccountTypeSpecific}}},
			clock: &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			want1: ErrOrderCondition,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				},
			}},
		{name: "注文に成功しても保存に失敗したらerror",
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{Save1: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            10_000,
				GridStrategy:    GridStrategy{},
				Account:         Account{AccountType: AccountTypeSpecific}}},
			clock: &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			want1: ErrUnknown,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				&Order{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
			}}},
		{name: "注文に成功して保存に成功しても余力更新に失敗したらerror",
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			strategyStore: &testStrategyStore{
				GetByCode1: &Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				AddStrategyCash1: ErrUnknown},
			clock: &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			want1: ErrUnknown,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				&Order{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
			}},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1.0 * 2100 * 4}},
		{name: "注文に成功して保存に成功して余力更新にも成功したらnil",
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            10_000,
				GridStrategy:    GridStrategy{},
				Account:         Account{AccountType: AccountTypeSpecific}}},
			clock: &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			want1: nil,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            10_000,
					GridStrategy:    GridStrategy{},
					Account:         Account{AccountType: AccountTypeSpecific},
				},
				&Order{
					Code:             "order-code-001",
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
			}},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1.0 * 2100 * 4}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{
				kabusAPI:      test.kabusAPI,
				orderStore:    test.orderStore,
				strategyStore: test.strategyStore,
				clock:         test.clock,
			}
			got1 := service.EntryLimit(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory) ||
				!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory),
					!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory),
					!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory),
					test.want1, test.wantSendOrderHistory, test.wantSaveHistory, test.wantAddStrategyCashHistory,
					got1, test.kabusAPI.SendOrderHistory, test.orderStore.SaveHistory, test.strategyStore.AddStrategyCashHistory)
			}
		})
	}
}

func Test_orderService_ExitLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		clock                IClock
		kabusAPI             *testKabusAPI
		orderStore           *testOrderStore
		positionStore        *testPositionStore
		strategyStore        *testStrategyStore
		arg1                 string
		arg2                 float64
		arg3                 float64
		arg4                 SortOrder
		want1                error
		wantHoldCount        int
		wantReleaseCount     int
		wantSendOrderHistory []interface{}
		wantSaveHistory      []interface{}
	}{
		{name: "指定された戦略がなければエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			arg1:          "strategy-code-001",
			want1:         ErrNoData},
		{name: "ポジション一覧の取得に失敗したらエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			want1:         ErrUnknown},
		{name: "ポジションが0件なら何もせずにエラー",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrNotEnoughPosition},
		{name: "ポジションの拘束に失敗したらエラー",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}},
				Hold1: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrUnknown,
			wantHoldCount: 1},
		{name: "拘束したい数量に対してポジションが足りない場合はエラー",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore:    &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:             "strategy-code-001",
			arg2:             2100,
			arg3:             200,
			want1:            ErrNotEnoughPosition,
			wantHoldCount:    4,
			wantReleaseCount: 4},
		{name: "注文の送信に失敗したらエラー",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder2: ErrUnknown},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account:         Account{AccountType: AccountTypeSpecific}}},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			arg4:  SortOrderNewest,
			want1: ErrUnknown,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account:         Account{AccountType: AccountTypeSpecific},
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
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-004", HoldQuantity: 2},
						{PositionCode: "position-code-003", HoldQuantity: 2},
					},
				},
			},
			wantHoldCount: 2},
		{name: "注文に失敗したら解放してエラー",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: 4, OrderCode: ""}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account:         Account{AccountType: AccountTypeSpecific}}},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			arg4:  SortOrderNewest,
			want1: ErrOrderCondition,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account:         Account{AccountType: AccountTypeSpecific},
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
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-004", HoldQuantity: 2},
						{PositionCode: "position-code-003", HoldQuantity: 2},
					},
				},
			},
			wantHoldCount:    2,
			wantReleaseCount: 2},
		{name: "注文の保存に失敗したらエラー",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{Save1: ErrUnknown},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account:         Account{AccountType: AccountTypeSpecific}}},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			arg4:  SortOrderNewest,
			want1: ErrUnknown,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account:         Account{AccountType: AccountTypeSpecific},
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
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-004", HoldQuantity: 2},
						{PositionCode: "position-code-003", HoldQuantity: 2},
					},
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-004", HoldQuantity: 2},
					{PositionCode: "position-code-003", HoldQuantity: 2},
				}}},
			wantHoldCount: 2},
		{name: "新しいもの順なら新しいポジションから拘束する",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account:         Account{AccountType: AccountTypeSpecific}}},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			arg4:  SortOrderNewest,
			want1: nil,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account:         Account{AccountType: AccountTypeSpecific},
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
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-004", HoldQuantity: 2},
						{PositionCode: "position-code-003", HoldQuantity: 2},
					},
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-004", HoldQuantity: 2},
					{PositionCode: "position-code-003", HoldQuantity: 2},
				}}},
			wantHoldCount: 2},
		{name: "古いもの順なら古いポジションから拘束する",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local)},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local)},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local)},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local)}}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Account:         Account{AccountType: AccountTypeSpecific}}},
			arg1:  "strategy-code-001",
			arg2:  2100,
			arg3:  4,
			arg4:  SortOrderLatest,
			want1: nil,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Account:         Account{AccountType: AccountTypeSpecific},
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
					ExecutionType:    ExecutionTypeLimit,
					Price:            2100,
					OrderQuantity:    4,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4}},
				},
			},
			wantSaveHistory: []interface{}{&Order{
				Code:             "order-code-001",
				StrategyCode:     "strategy-code-001",
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Status:           OrderStatusInOrder,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				ExecutionType:    ExecutionTypeLimit,
				Price:            2100,
				OrderQuantity:    4,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4}}}},
			wantHoldCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{
				kabusAPI:      test.kabusAPI,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
				strategyStore: test.strategyStore,
				clock:         test.clock,
			}
			got1 := service.ExitLimit(test.arg1, test.arg2, test.arg3, test.arg4)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory) ||
				!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount) ||
				!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory),
					!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory),
					!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount),
					!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount),
					test.want1, test.wantSendOrderHistory, test.wantSaveHistory, test.wantHoldCount, test.wantReleaseCount,
					got1, test.kabusAPI.SendOrderHistory, test.orderStore.SaveHistory, test.positionStore.HoldCount, test.positionStore.ReleaseCount)
			}
		})
	}
}

func Test_orderService_Cancel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		kabusAPI               *testKabusAPI
		arg1                   *Strategy
		arg2                   string
		want1                  error
		wantCancelOrderHistory []interface{}
	}{
		{name: "引数がnilならエラー",
			kabusAPI: &testKabusAPI{},
			arg1:     nil,
			arg2:     "order-code-001",
			want1:    ErrNilArgument},
		{name: "取消送信に失敗したらエラー",
			kabusAPI:               &testKabusAPI{CancelOrder2: ErrUnknown},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "取消に失敗したらエラー",
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: false, ResultCode: 4}},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  ErrCancelCondition,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "取消に成功したらnil",
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "cancel-order-code-001"}},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{kabusAPI: test.kabusAPI}
			got1 := service.Cancel(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
