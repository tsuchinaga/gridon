package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testContractService struct {
	IContractService
	Confirm1              error
	ConfirmCount          int
	ConfirmHistory        []interface{}
	ConfirmGridEnd1       error
	ConfirmGridEndCount   int
	ConfirmGridEndHistory []interface{}
}

func (t *testContractService) Confirm(strategy *Strategy) error {
	t.ConfirmHistory = append(t.ConfirmHistory, strategy)
	t.ConfirmCount++
	return t.Confirm1
}
func (t *testContractService) ConfirmGridEnd(strategy *Strategy) error {
	t.ConfirmGridEndHistory = append(t.ConfirmGridEndHistory, strategy)
	t.ConfirmGridEndCount++
	return t.ConfirmGridEnd1
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
	clock := &testClock{}
	want1 := &contractService{
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
		clock:         clock,
	}
	got1 := newContractService(kabusAPI, strategyStore, orderStore, positionStore, clock)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_contractService_updateContractPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                           string
		strategyStore                  *testStrategyStore
		arg1                           *Strategy
		arg2                           float64
		arg3                           time.Time
		want1                          error
		wantSetContractPriceHistory    []interface{}
		wantSetMaxContractPriceHistory []interface{}
		wantSetMinContractPriceHistory []interface{}
	}{
		{name: "引数がnilならエラー",
			strategyStore: &testStrategyStore{},
			arg1:          nil,
			arg2:          2000,
			arg3:          time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:         ErrNilArgument},
		{name: "最終約定日時以前の約定情報なら何もしない",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local)},
			arg2:  2000,
			arg3:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1: nil},
		{name: "最終約定日時以降の約定情報なら何もしない約定情報を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    1999,
				LastContractDateTime: time.Date(2021, 12, 24, 13, 59, 0, 0, time.Local)},
			arg2:                        2000,
			arg3:                        time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                       nil,
			wantSetContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "約定情報更新でエラーがでたらエラーを返す",
			strategyStore: &testStrategyStore{SetContractPrice1: ErrUnknown},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    1999,
				LastContractDateTime: time.Date(2021, 12, 24, 13, 59, 0, 0, time.Local)},
			arg2:                        2000,
			arg3:                        time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                       ErrUnknown,
			wantSetContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "実行可能なグリッド戦略がなければ最小最大は更新しない",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: false}},
			arg2:  2000,
			arg3:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1: nil},
		{name: "最大約定日時が前日なら、保持している最大約定日時よりも安くても最大約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 23, 12, 30, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMaxContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最大約定日時が当日でも違うグリッド期間なら、保持している最大約定日時よりも安くても最大約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 11, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMaxContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最大約定日時が当日で同じグリッド期間で、さらに最大約定日時がゼロ値なら、最大約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     0,
				MaxContractDateTime:  time.Time{},
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMaxContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最大約定日時が当日で同じグリッド期間で、さらに最大約定値より約定値が高ければ、最大約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     1998,
				MaxContractDateTime:  time.Date(2021, 12, 24, 13, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMaxContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最大約定日時が当日で同じグリッド期間でも、最大約定値より約定値が安ければ、更新しない",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2002,
				MaxContractDateTime:  time.Date(2021, 12, 24, 13, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:  2000,
			arg3:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1: nil},
		{name: "最大約定を更新でエラーが出たらエラーを返す",
			strategyStore: &testStrategyStore{SetMaxContractPrice1: ErrUnknown},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     1998,
				MaxContractDateTime:  time.Date(2021, 12, 24, 13, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          ErrUnknown,
			wantSetMaxContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最小約定日時が前日なら、保持している最小約定日時よりも高くても最小約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 23, 12, 30, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMinContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最小約定日時が当日でも違うグリッド期間なら、保持している最小約定日時よりも高くても最小約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     1900,
				MinContractDateTime:  time.Date(2021, 12, 24, 11, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMinContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最小約定日時が当日で同じグリッド期間で、さらに最小約定日時がゼロ値なら、最小約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     0,
				MinContractDateTime:  time.Time{},
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMinContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最小約定日時が当日で同じグリッド期間で、さらに最小約定値より約定値が高ければ、最小約定を更新する",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     2002,
				MinContractDateTime:  time.Date(2021, 12, 24, 13, 59, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          nil,
			wantSetMinContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
		{name: "最小約定日時が当日で同じグリッド期間でも、最小約定値より約定値が安ければ、更新しない",
			strategyStore: &testStrategyStore{},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     1998,
				MinContractDateTime:  time.Date(2021, 12, 24, 13, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:  2000,
			arg3:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1: nil},
		{name: "最小約定を更新でエラーが出たらエラーを返す",
			strategyStore: &testStrategyStore{SetMinContractPrice1: ErrUnknown},
			arg1: &Strategy{
				Code:                 "strategy-code-001",
				LastContractPrice:    2000,
				LastContractDateTime: time.Date(2021, 12, 24, 14, 1, 0, 0, time.Local),
				MaxContractPrice:     2100,
				MaxContractDateTime:  time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
				MinContractPrice:     2002,
				MinContractDateTime:  time.Date(2021, 12, 24, 13, 0, 0, 0, time.Local),
				GridStrategy: GridStrategy{
					Runnable: true,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					}}},
			arg2:                           2000,
			arg3:                           time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local),
			want1:                          ErrUnknown,
			wantSetMinContractPriceHistory: []interface{}{"strategy-code-001", 2000.0, time.Date(2021, 12, 24, 14, 0, 0, 0, time.Local)}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{strategyStore: test.strategyStore}
			got1 := service.updateContractPrice(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) ||
				!reflect.DeepEqual(test.wantSetContractPriceHistory, test.strategyStore.SetContractPriceHistory) ||
				!reflect.DeepEqual(test.wantSetMaxContractPriceHistory, test.strategyStore.SetMaxContractPriceHistory) ||
				!reflect.DeepEqual(test.wantSetMinContractPriceHistory, test.strategyStore.SetMinContractPriceHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!reflect.DeepEqual(test.want1, got1),
					!reflect.DeepEqual(test.wantSetContractPriceHistory, test.strategyStore.SetContractPriceHistory),
					!reflect.DeepEqual(test.wantSetMaxContractPriceHistory, test.strategyStore.SetMaxContractPriceHistory),
					!reflect.DeepEqual(test.wantSetMinContractPriceHistory, test.strategyStore.SetMinContractPriceHistory),
					test.want1, test.wantSetContractPriceHistory, test.wantSetMaxContractPriceHistory, test.wantSetMinContractPriceHistory,
					got1, test.strategyStore.SetContractPriceHistory, test.strategyStore.SetMaxContractPriceHistory, test.strategyStore.SetMinContractPriceHistory)
			}
		})
	}
}

func Test_contractService_Confirm_issue29(t *testing.T) {
	t.Parallel()

	// 2022/01/07 14:05:00 の約定確認を再現する

	orderStore := &testOrderStore{
		GetActiveOrdersByStrategyCode1: []*Order{
			{
				Code:         "20220107A02N87954257",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2046,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 44, 59, 9540114000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    nil},
			{
				Code:         "20220107A02N87954259",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2050,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 45, 0, 2610443000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    []HoldPosition{{PositionCode: "E2022010701ZI6", Price: 2041, HoldQuantity: 3, ContractQuantity: 0, ReleaseQuantity: 0}}},
			{
				Code:         "20220107A02N87984149",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2047,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 59, 8, 0605613000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    nil},
			{
				Code:         "20220107A02N87984151",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2051,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 59, 8, 3556296000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    []HoldPosition{{PositionCode: "E2022010701WVD", Price: 2046, HoldQuantity: 3, ContractQuantity: 0, ReleaseQuantity: 0}}},
			{
				Code:         "20220107A02N87984381",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2048,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 14, 1, 54, 4805408000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    nil},
			{
				Code:         "20220107A02N87984383",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2052,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 7774265000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions:    []HoldPosition{{PositionCode: "E2022010701T4M", Price: 2050, HoldQuantity: 3, ContractQuantity: 0, ReleaseQuantity: 0}}},
		},
	}

	kabusAPI := &testKabusAPI{
		GetOrders1: []SecurityOrder{
			{
				Code:       "20220107A02N87948181",
				Status:     OrderStatusDone,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2049,
				OrderQuantity:    3,
				ContractQuantity: 3,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 30, 39, 989353000, time.Local),
				ContractDateTime: time.Date(2022, 1, 7, 14, 1, 46, 861981000, time.Local),
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{{OrderCode: "20220107A02N87948181", PositionCode: "E2022010702ZV5", Price: 2049, Quantity: 3, ContractDateTime: time.Date(2022, 1, 7, 14, 1, 46, 861981000, time.Local)}}},
			{
				Code:       "20220107A02N87948186",
				Status:     OrderStatusCanceled,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2045,
				OrderQuantity:    0,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 30, 40, 301361000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Date(2022, 1, 7, 14, 1, 54, 485076000, time.Local),
				Contracts:        nil},
			{
				Code:       "20220107A02N87954257",
				Status:     OrderStatusInOrder,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2046,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 45, 0, 133007000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil},
			{
				Code:       "20220107A02N87954259",
				Status:     OrderStatusInOrder,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2050,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 45, 0, 413815000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil},
			{
				Code:       "20220107A02N87984149",
				Status:     OrderStatusInOrder,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2047,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 59, 8, 264354000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil},
			{
				Code:       "20220107A02N87984151",
				Status:     OrderStatusInOrder,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2051,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 13, 59, 8, 545161000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil},
			{
				Code:       "20220107A02N87984381",
				Status:     OrderStatusDone,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeEntry, Side: SideBuy,
				Price:            2048,
				OrderQuantity:    3,
				ContractQuantity: 3,
				OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 679975000, time.Local),
				ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local),
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{{OrderCode: "20220107A02N87984381", PositionCode: "E202201070305Y", Price: 2048, Quantity: 3, ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local)}}},
			{
				Code:       "20220107A02N87984383",
				Status:     OrderStatusInOrder,
				SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
				TradeType: TradeTypeExit, Side: SideSell,
				Price:            2052,
				OrderQuantity:    3,
				ContractQuantity: 0,
				OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 976382000, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil},
		},
	}

	service := &contractService{
		orderStore:    orderStore,
		positionStore: &testPositionStore{},
		strategyStore: &testStrategyStore{},
		kabusAPI:      kabusAPI,
	}
	strategy := &Strategy{
		Code:                 "1475-buy",
		SymbolCode:           "1475",
		Exchange:             ExchangeToushou,
		Product:              ProductMargin,
		MarginTradeType:      MarginTradeTypeDay,
		EntrySide:            SideBuy,
		Cash:                 147339,
		BasePrice:            2049,
		BasePriceDateTime:    time.Date(2022, 1, 7, 14, 1, 46, 861981000, time.Local),
		LastContractPrice:    2049,
		LastContractDateTime: time.Date(2022, 1, 7, 14, 1, 46, 861981000, time.Local),
		MaxContractPrice:     2049,
		MaxContractDateTime:  time.Date(2022, 1, 7, 12, 30, 0, 125167000, time.Local),
		MinContractPrice:     2044,
		MinContractDateTime:  time.Date(2022, 1, 7, 12, 33, 23, 855114000, time.Local),
		TickGroup:            TickGroupOther,
		TradingUnit:          1,
		RebalanceStrategy: RebalanceStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
				time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
			},
		},
		GridStrategy: GridStrategy{
			Runnable:      true,
			Width:         1,
			Quantity:      3,
			NumberOfGrids: 3,
			TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
			},
			GridType: GridTypeDynamicMinMax,
			DynamicGridMinMax: DynamicGridMinMax{
				Divide:    6,
				Rounding:  RoundingFloor,
				Operation: OperationPlus,
			},
		},
		CancelStrategy: CancelStrategy{
			Runnable: true,
			Timings: []time.Time{
				time.Date(0, 1, 1, 11, 31, 0, 0, time.Local),
				time.Date(0, 1, 1, 14, 58, 0, 0, time.Local),
			},
		},
		ExitStrategy: ExitStrategy{
			Runnable: true,
			Conditions: []ExitCondition{
				{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
			},
		},
		Account: Account{
			Password:    "Password1234",
			AccountType: AccountTypeSpecific,
		},
	}
	got1 := service.Confirm(strategy)
	var want1 error = nil
	wantOrderStoreSaveHistory := []interface{}{&Order{
		Code:         "20220107A02N87984381",
		StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
		TradeType: TradeTypeEntry, Side: SideBuy,
		Status:           OrderStatusDone,
		Price:            2048,
		OrderQuantity:    3,
		ContractQuantity: 3,
		OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 679975000, time.Local),
		ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local),
		CancelDateTime:   time.Time{},
		Contracts:        []Contract{{OrderCode: "20220107A02N87984381", PositionCode: "E202201070305Y", Price: 2048, Quantity: 3, ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local)}},
		HoldPositions:    nil}}

	if !errors.Is(got1, nil) || !reflect.DeepEqual(wantOrderStoreSaveHistory, orderStore.SaveHistory) {
		t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), want1, wantOrderStoreSaveHistory, got1, orderStore.SaveHistory)
	}
}

func Test_contractService_ConfirmGridEnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                     string
		clock                                    *testClock
		orderStore                               *testOrderStore
		kabusapi                                 *testKabusAPI
		strategyStore                            *testStrategyStore
		positionStore                            *testPositionStore
		arg1                                     *Strategy
		want1                                    error
		wantOrderStoreSaveHistory                []interface{}
		wantGetActiveOrdersByStrategyCodeHistory []interface{}
		wantKabusApiGetOrdersHistory             []interface{}
	}{
		{name: "引数がnilならエラー",
			clock:         &testClock{},
			orderStore:    &testOrderStore{},
			kabusapi:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "グリッド戦略が有効でなければ何もせずに終了",
			clock:         &testClock{},
			orderStore:    &testOrderStore{},
			kabusapi:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{GridStrategy: GridStrategy{Runnable: false}},
			want1:         nil},
		{name: "グリッドの終了タイミングから1分以内でなければ何もせずに終了",
			clock:         &testClock{Now1: time.Date(2022, 1, 7, 14, 0, 0, 0, time.Local)},
			orderStore:    &testOrderStore{},
			kabusapi:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1: &Strategy{GridStrategy: GridStrategy{Runnable: true, TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)}}}},
			want1: nil},
		{name: "注文中の注文がなければ何もせずに終了",
			clock: &testClock{Now1: time.Date(2022, 1, 7, 14, 58, 10, 0, time.Local)},
			orderStore: &testOrderStore{
				GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusapi:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{Runnable: true, TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)}}}},
			want1:                                    nil,
			wantGetActiveOrdersByStrategyCodeHistory: []interface{}{"strategy-code-001"}},
		{name: "注文中の取得でエラーがあればエラーを返す",
			clock: &testClock{Now1: time.Date(2022, 1, 7, 14, 58, 10, 0, time.Local)},
			orderStore: &testOrderStore{
				GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusapi:      &testKabusAPI{},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1: &Strategy{Code: "strategy-code-001", GridStrategy: GridStrategy{Runnable: true, TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)}}}},
			want1:                                    ErrUnknown,
			wantGetActiveOrdersByStrategyCodeHistory: []interface{}{"strategy-code-001"}},
		{name: "証券会社の注文取得でエラーがあればエラーを返す",
			clock: &testClock{Now1: time.Date(2022, 1, 7, 14, 58, 10, 0, time.Local)},
			orderStore: &testOrderStore{
				GetActiveOrdersByStrategyCode1: []*Order{
					{
						Code:         "20220107A02N87984381",
						StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
						TradeType: TradeTypeEntry, Side: SideBuy,
						Price:            2048,
						OrderQuantity:    3,
						ContractQuantity: 0,
						OrderDateTime:    time.Date(2022, 1, 7, 14, 1, 54, 4805408000, time.Local),
						ContractDateTime: time.Time{},
						CancelDateTime:   time.Time{},
						Contracts:        nil,
						HoldPositions:    nil}}},
			kabusapi:      &testKabusAPI{GetOrders2: ErrUnknown},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1: &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475", GridStrategy: GridStrategy{Runnable: true, TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)}}}},
			want1:                                    ErrUnknown,
			wantGetActiveOrdersByStrategyCodeHistory: []interface{}{"strategy-code-001"},
			wantKabusApiGetOrdersHistory:             []interface{}{ProductMargin, "1475", time.Date(2022, 1, 7, 12, 30, 0, 0, time.Local)}},
		{name: "グリッドの終了タイミングから1分以内であれば、storeの注文一覧と、グリッド開始時刻以降に更新された注文を集めて約定確認をする",
			clock: &testClock{Now1: time.Date(2022, 1, 7, 14, 58, 10, 0, time.Local)},
			orderStore: &testOrderStore{
				GetActiveOrdersByStrategyCode1: []*Order{
					{
						Code:         "20220107A02N87984381",
						StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Status: OrderStatusInOrder, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
						TradeType: TradeTypeEntry, Side: SideBuy,
						Price:            2048,
						OrderQuantity:    3,
						ContractQuantity: 0,
						OrderDateTime:    time.Date(2022, 1, 7, 14, 1, 54, 4805408000, time.Local),
						ContractDateTime: time.Time{},
						CancelDateTime:   time.Time{},
						Contracts:        nil,
						HoldPositions:    nil}}},
			kabusapi: &testKabusAPI{GetOrders1: []SecurityOrder{
				{
					Code:       "20220107A02N87984381",
					Status:     OrderStatusDone,
					SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, AccountType: AccountTypeSpecific, ExpireDay: time.Date(2022, 1, 7, 0, 0, 0, 0, time.Local),
					TradeType: TradeTypeEntry, Side: SideBuy,
					Price:            2048,
					OrderQuantity:    3,
					ContractQuantity: 3,
					OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 679975000, time.Local),
					ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local),
					CancelDateTime:   time.Time{},
					Contracts:        []Contract{{OrderCode: "20220107A02N87984381", PositionCode: "E202201070305Y", Price: 2048, Quantity: 3, ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local)}}}}},
			strategyStore: &testStrategyStore{},
			positionStore: &testPositionStore{},
			arg1: &Strategy{Code: "strategy-code-001", Product: ProductMargin, SymbolCode: "1475", GridStrategy: GridStrategy{Runnable: true, TimeRanges: []TimeRange{
				{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local)},
				{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)}}}},
			want1:                                    nil,
			wantGetActiveOrdersByStrategyCodeHistory: []interface{}{"strategy-code-001"},
			wantKabusApiGetOrdersHistory:             []interface{}{ProductMargin, "1475", time.Date(2022, 1, 7, 12, 30, 0, 0, time.Local)},
			wantOrderStoreSaveHistory: []interface{}{&Order{
				Code:         "20220107A02N87984381",
				StrategyCode: "1475-buy", SymbolCode: "1475", Exchange: ExchangeToushou, Product: ProductMargin, MarginTradeType: MarginTradeTypeDay, ExecutionType: ExecutionTypeLimit, AccountType: AccountTypeSpecific,
				TradeType: TradeTypeEntry, Side: SideBuy,
				Status:           OrderStatusDone,
				Price:            2048,
				OrderQuantity:    3,
				ContractQuantity: 3,
				OrderDateTime:    time.Date(2022, 1, 7, 14, 2, 0, 679975000, time.Local),
				ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local),
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{{OrderCode: "20220107A02N87984381", PositionCode: "E202201070305Y", Price: 2048, Quantity: 3, ContractDateTime: time.Date(2022, 1, 7, 14, 3, 5, 816538000, time.Local)}},
				HoldPositions:    nil}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{
				kabusAPI:      test.kabusapi,
				strategyStore: test.strategyStore,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
				clock:         test.clock,
			}
			got1 := service.ConfirmGridEnd(test.arg1)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeHistory, test.orderStore.GetActiveOrdersByStrategyCodeHistory) ||
				!reflect.DeepEqual(test.wantKabusApiGetOrdersHistory, test.kabusapi.GetOrdersHistory) ||
				!reflect.DeepEqual(test.wantOrderStoreSaveHistory, test.orderStore.SaveHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantGetActiveOrdersByStrategyCodeHistory, test.orderStore.GetActiveOrdersByStrategyCodeHistory),
					!reflect.DeepEqual(test.wantKabusApiGetOrdersHistory, test.kabusapi.GetOrdersHistory),
					!reflect.DeepEqual(test.wantOrderStoreSaveHistory, test.orderStore.SaveHistory),
					test.want1, test.wantGetActiveOrdersByStrategyCodeHistory, test.wantKabusApiGetOrdersHistory, test.wantOrderStoreSaveHistory,
					got1, test.orderStore.GetActiveOrdersByStrategyCodeHistory, test.kabusapi.GetOrdersHistory, test.orderStore.SaveHistory)
			}
		})
	}
}

func Test_contractService_confirm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  *Strategy
		arg2  []*Order
		arg3  []SecurityOrder
		want1 error
	}{
		{name: "引数がnilならエラー",
			arg1:  nil,
			arg2:  nil,
			arg3:  nil,
			want1: ErrNilArgument},
		// 上記以外のエラーは他のconfirm処理から叩かれることでテストできている
		// confirm自体に処理を追加した場合は、ここにテストを追加することが望ましい
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &contractService{}
			got1 := service.confirm(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got1, test.want1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
