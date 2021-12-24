package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testContractService struct {
	IContractService
	Confirm1       error
	ConfirmCount   int
	ConfirmHistory []interface{}
}

func (t *testContractService) Confirm(strategy *Strategy) error {
	t.ConfirmHistory = append(t.ConfirmHistory, strategy)
	t.ConfirmCount++
	return t.Confirm1
}

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
	}{
		{name: "引数がnilならエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{},
			arg1:          nil,
			arg2:          Contract{},
			want1:         ErrNilArgument},
		{name: "ポジションの登録に失敗したらエラー",
			positionStore: &testPositionStore{Save1: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
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
			}}},
		{name: "余力の登録に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}, AddStrategyCash1: ErrUnknown},
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
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
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
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1 * 2070.0 * 4.0}},
		{name: "戦略に今回の約定データよりも新しいデータが登録されていたら、約定情報の更新をしない",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{LastContractPrice: 2080, LastContractDateTime: time.Date(2021, 10, 28, 10, 1, 0, 0, time.Local)}},
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
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", -1 * 2070.0 * 4.0}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{positionStore: test.positionStore, strategyStore: test.strategyStore}
			got1 := service.entryContract(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.positionStore.SaveHistory) ||
				!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantSaveHistory, test.positionStore.SaveHistory),
					!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory),
					test.want1, test.wantSaveHistory, test.wantAddStrategyCashHistory,
					got1, test.positionStore.SaveHistory, test.strategyStore.AddStrategyCashHistory)
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
			wantOrder:                  nil},
		{name: "ポジションの返済約定登録に失敗したらエラー",
			positionStore: &testPositionStore{ExitContract1: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
			arg1: &Order{HoldPositions: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80},
				{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
				{PositionCode: "position-code-003", HoldQuantity: 100},
			}},
			arg2:                       Contract{Quantity: 100},
			want1:                      ErrUnknown,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0},
			wantAddStrategyCashHistory: nil,
			wantOrder: &Order{HoldPositions: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100},
				{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70},
				{PositionCode: "position-code-003", HoldQuantity: 100},
			}}},
		{name: "現金余力の更新に失敗したらエラー",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}, AddStrategyCash1: ErrUnknown},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80, Price: 2040},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2050},
					{PositionCode: "position-code-003", HoldQuantity: 100, Price: 2060},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100},
			want1:                      ErrUnknown,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 2070.0 * 20.0},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100, Price: 2040},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2050},
					{PositionCode: "position-code-003", HoldQuantity: 100, Price: 2060},
				}}},
		{name: "エラーがなかったらnilが返される",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80, Price: 2030},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2040},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2050},
					{PositionCode: "position-code-004", HoldQuantity: 100, Price: 2060},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100, ContractDateTime: time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			want1:                      nil,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0, "position-code-002", 30.0, "position-code-004", 50.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 2070.0 * 20.0, "strategy-code-001", 2070.0 * 30.0, "strategy-code-001", 2070.0 * 50.0},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100, Price: 2030},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, ContractQuantity: 30, Price: 2040},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2050},
					{PositionCode: "position-code-004", HoldQuantity: 100, ContractQuantity: 50, Price: 2060},
				}}},
		{name: "売り建て = 買いエグジットなら戻す資産の計算ロジックが違う",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideBuy,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80, Price: 2110},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2100},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2090},
					{PositionCode: "position-code-004", HoldQuantity: 100, Price: 2080},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 100, ContractDateTime: time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			want1:                      nil,
			wantExitContractHistory:    []interface{}{"position-code-001", 20.0, "position-code-002", 30.0, "position-code-004", 50.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", (2110.0*2 - 2070.0) * 20.0, "strategy-code-001", (2100.0*2 - 2070.0) * 30.0, "strategy-code-001", (2080.0*2 - 2070.0) * 50.0},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideBuy,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 100, Price: 2110},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, ContractQuantity: 30, Price: 2100},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2090},
					{PositionCode: "position-code-004", HoldQuantity: 100, Price: 2080, ContractQuantity: 50},
				}}},
		{name: "少量の約定時は全てのポジションをループすることなく必要最低限のチェックでループを抜けて処理を進められる",
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
			arg1: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 80, Price: 2110},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2100},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2090},
					{PositionCode: "position-code-004", HoldQuantity: 100, Price: 2080},
				}},
			arg2:                       Contract{Price: 2070, Quantity: 10, ContractDateTime: time.Date(2021, 9, 29, 10, 0, 0, 0, time.Local)},
			want1:                      nil,
			wantExitContractHistory:    []interface{}{"position-code-001", 10.0},
			wantAddStrategyCashHistory: []interface{}{"strategy-code-001", 2070 * 10.0},
			wantOrder: &Order{
				StrategyCode: "strategy-code-001",
				Side:         SideSell,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 90, Price: 2110},
					{PositionCode: "position-code-002", HoldQuantity: 100, ReleaseQuantity: 70, Price: 2100},
					{PositionCode: "position-code-003", HoldQuantity: 100, ReleaseQuantity: 100, Price: 2090},
					{PositionCode: "position-code-004", HoldQuantity: 100, Price: 2080},
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
				!reflect.DeepEqual(test.wantOrder, test.arg1) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantExitContractHistory, test.positionStore.ExitContractHistory),
					!reflect.DeepEqual(test.wantAddStrategyCashHistory, test.strategyStore.AddStrategyCashHistory),
					!reflect.DeepEqual(test.wantOrder, test.arg1),
					test.want1, test.wantExitContractHistory, test.wantAddStrategyCashHistory, test.wantOrder,
					got1, test.positionStore.ExitContractHistory, test.strategyStore.AddStrategyCashHistory, test.arg1)
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
		wantGetActiveOrdersByStrategyCodeCount int
		wantGetOrdersHistory                   []interface{}
		wantSaveHistory                        []interface{}
		wantSetContractPriceHistory            []interface{}
	}{
		{name: "引数がnilならerror",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "storeから注文一覧が取れなければerror",
			kabusAPI:                               &testKabusAPI{GetOrders1: []SecurityOrder{}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  ErrUnknown,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "storeからの注文一覧が空なら何もせずに終わり",
			kabusAPI:                               &testKabusAPI{GetOrders1: []SecurityOrder{}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001"},
			want1:                                  nil,
			wantGetActiveOrdersByStrategyCodeCount: 1},
		{name: "kabusapiから注文一覧が取れなければerror",
			kabusAPI:                               &testKabusAPI{GetOrders2: ErrUnknown},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  ErrUnknown,
			wantGetActiveOrdersByStrategyCodeCount: 1,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}}},
		{name: "kabusapiから返ってきた注文が空なら何もしない",
			kabusAPI:                               &testKabusAPI{GetOrders1: []SecurityOrder{}},
			orderStore:                             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}}},
			positionStore:                          &testPositionStore{},
			strategyStore:                          &testStrategyStore{},
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  ErrUnknown,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  ErrUnknown,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  ErrUnknown,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
			wantGetActiveOrdersByStrategyCodeCount: 1,
			wantSetContractPriceHistory:            []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  ErrUnknown,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			}},
			wantSetContractPriceHistory: []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475"},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Time{}},
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
			}},
			wantSetContractPriceHistory: []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
		{name: "約定日時が得られたら1分前以降に更新された注文を取得する",
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475", LastContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Date(2021, 10, 29, 9, 59, 0, 0, time.Local)},
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
		{name: "戦略で保持している約定日時よりも古い約定日時なら約定情報を反映しない",
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
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475", LastContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)},
			want1:                                  nil,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Date(2021, 10, 29, 9, 59, 0, 0, time.Local)},
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
		{name: "約定情報登録に失敗したらエラー",
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
			strategyStore:                          &testStrategyStore{SetContractPrice1: ErrUnknown},
			arg1:                                   &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475", LastContractDateTime: time.Date(2021, 10, 29, 9, 59, 0, 0, time.Local)},
			want1:                                  ErrUnknown,
			wantGetOrdersHistory:                   []interface{}{ProductMargin, "1475", time.Date(2021, 10, 29, 9, 58, 0, 0, time.Local)},
			wantGetActiveOrdersByStrategyCodeCount: 1,
			wantSetContractPriceHistory:            []interface{}{"strategy-code-001", 2070.0, time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)}},
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
				!reflect.DeepEqual(test.wantGetOrdersHistory, test.kabusAPI.GetOrdersHistory) ||
				!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeCount, test.orderStore.GetActiveOrdersByStrategyCodeCount) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory) ||
				!reflect.DeepEqual(test.wantSetContractPriceHistory, test.strategyStore.SetContractPriceHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantGetOrdersHistory, test.kabusAPI.GetOrdersHistory),
					!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeCount, test.orderStore.GetActiveOrdersByStrategyCodeCount),
					!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory),
					!reflect.DeepEqual(test.wantSetContractPriceHistory, test.strategyStore.SetContractPriceHistory),
					test.want1, test.wantGetOrdersHistory, test.wantGetActiveOrdersByStrategyCodeCount, test.wantSaveHistory, test.wantSetContractPriceHistory,
					got1, test.kabusAPI.GetOrdersHistory, test.orderStore.GetActiveOrdersByStrategyCodeCount, test.orderStore.SaveHistory, test.strategyStore.SetContractPriceHistory)
			}
		})
	}
}

func Test_newContractService(t *testing.T) {
	t.Parallel()
	kabusAPI := &testKabusAPI{}
	strategyStore := &strategyStore{}
	orderStore := &orderStore{}
	positionStore := &positionStore{}
	want1 := &contractService{
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
	}
	got1 := newContractService(kabusAPI, strategyStore, orderStore, positionStore)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}
