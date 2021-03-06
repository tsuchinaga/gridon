package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"gitlab.com/tsuchinaga/kabus-grpc-server/kabuspb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	EntryMarket1                         error
	EntryMarketCount                     int
	EntryMarketHistory                   []interface{}
	ExitMarket1                          error
	ExitMarketCount                      int
	ExitMarketHistory                    []interface{}
	CancelAll1                           error
	CancelAllCount                       int
	CancelAllHistory                     []interface{}
	ExitAll1                             error
	ExitAllCount                         int
	ExitAllHistory                       []interface{}
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
func (t *testOrderService) EntryMarket(strategyCode string, quantity float64) error {
	t.EntryMarketHistory = append(t.EntryMarketHistory, strategyCode)
	t.EntryMarketHistory = append(t.EntryMarketHistory, quantity)
	t.EntryMarketCount++
	return t.EntryMarket1
}
func (t *testOrderService) ExitMarket(strategyCode string, quantity float64, sortOrder SortOrder) error {
	t.ExitMarketHistory = append(t.ExitMarketHistory, strategyCode)
	t.ExitMarketHistory = append(t.ExitMarketHistory, quantity)
	t.ExitMarketHistory = append(t.ExitMarketHistory, sortOrder)
	t.ExitMarketCount++
	return t.ExitMarket1
}
func (t *testOrderService) CancelAll(strategy *Strategy) error {
	t.CancelAllHistory = append(t.CancelAllHistory, strategy)
	t.CancelAllCount++
	return t.CancelAll1
}
func (t *testOrderService) ExitAll(strategy *Strategy) error {
	t.ExitAllHistory = append(t.ExitAllHistory, strategy)
	t.ExitAllCount++
	return t.ExitAll1
}

func Test_orderService_CancelAll(t *testing.T) {
	t.Parallel()

	simpleSt := status.New(codes.Internal, "simple status")

	notRequestErrSt := status.New(codes.Internal, "not request status")
	notRequestErrStDt, _ := notRequestErrSt.WithDetails(&kabuspb.Order{})

	requestErr47St := status.New(codes.Internal, "request 47 status")
	requestErr47StDt, _ := requestErr47St.WithDetails(&kabuspb.RequestError{Code: 47})

	requestErr2St := status.New(codes.Internal, "request 2 status")
	requestErr2StDt, _ := requestErr2St.WithDetails(&kabuspb.RequestError{Code: 2})

	tests := []struct {
		name                   string
		clock                  *testClock
		orderStore             *testOrderStore
		kabusAPI               *testKabusAPI
		logger                 *testLogger
		arg1                   *Strategy
		want1                  error
		wantCancelOrderHistory []interface{}
	}{
		{name: "?????????nil???????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{},
			kabusAPI:               &testKabusAPI{},
			logger:                 &testLogger{},
			arg1:                   nil,
			want1:                  ErrNilArgument,
			wantCancelOrderHistory: nil},
		{name: "???????????????????????????????????????????????????nil",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusAPI:               &testKabusAPI{},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", CancelStrategy: CancelStrategy{Runnable: false}},
			want1:                  nil,
			wantCancelOrderHistory: nil},
		{name: "?????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusAPI:               &testKabusAPI{},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: nil},
		{name: "???????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:               &testKabusAPI{},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  nil,
			wantCancelOrderHistory: nil},
		{name: "????????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: ErrUnknown},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "?????????????????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: false, ResultCode: -1}},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001", "Password1234", "order-code-002", "Password1234", "order-code-003"}},
		{name: "???????????????????????????????????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: simpleSt.Err()},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  simpleSt.Err(),
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "????????????????????????????????????????????????kabuspb.RequestError?????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: notRequestErrStDt.Err()},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  notRequestErrStDt.Err(),
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "???????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: requestErr47StDt.Err()},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001", "Password1234", "order-code-002", "Password1234", "order-code-003"}},
		{name: "?????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????",
			clock:                  &testClock{Now1: time.Date(2021, 11, 10, 14, 55, 0, 0, time.Local)},
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: requestErr2StDt.Err()},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}, CancelStrategy: CancelStrategy{Runnable: true, Timings: []time.Time{time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
			want1:                  requestErr2StDt.Err(),
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{clock: test.clock, kabusAPI: test.kabusAPI, orderStore: test.orderStore, logger: test.logger}
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
		{name: "?????????nil??????error",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          nil,
			want1:         ErrNilArgument},
		{name: "???????????????????????????????????????????????????nil",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 10, 0, 0, 0, time.Local)},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          &Strategy{},
			want1:         nil},
		{name: "????????????????????????????????????????????????error",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          &Strategy{ExitStrategy: ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1:         ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1},
		{name: "?????????????????????????????????????????????nil",
			clock:         &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{}},
			orderStore:    &testOrderStore{},
			kabusAPI:      &testKabusAPI{},
			arg1:          &Strategy{Code: "strategy-code-001", ExitStrategy: ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1:         nil,
			wantGetActivePositionsByStrategyCodeCount: 1},
		{name: "??????????????????hold??????????????????release??????error",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, Price: 102},
				},
				Hold1: ErrUnknown},
			orderStore: &testOrderStore{},
			kabusAPI:   &testKabusAPI{},
			arg1:       &Strategy{Code: "strategy-code-001", ExitStrategy: ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}}},
			want1:      ErrUnknown,
			wantGetActivePositionsByStrategyCodeCount: 1,
			wantHoldCount: 1},
		{name: "?????????????????????????????????release??????error",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, Price: 102},
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
				ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific}},
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
					ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific}},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarketAfternoonClose,
					Price:            0,
					OrderQuantity:    600,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
						{PositionCode: "position-code-002", HoldQuantity: 200, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
						{PositionCode: "position-code-003", HoldQuantity: 300, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
					},
				},
			}},
		{name: "??????????????????????????????error",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0, Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100, Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150, Price: 102},
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
				ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific}},
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
					ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific}},
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
					ExecutionType:    ExecutionTypeMarketAfternoonClose,
					Price:            0,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
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
				ExecutionType:    ExecutionTypeMarketAfternoonClose,
				Price:            0,
				OrderQuantity:    350,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
					{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
				},
			}}},
		{name: "?????????????????????????????????nil",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0, Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100, Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150, Price: 102},
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
				ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific}},
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
					ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific}},
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
					ExecutionType:    ExecutionTypeMarketAfternoonClose,
					Price:            0,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					Contracts:        nil,
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
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
				ExecutionType:    ExecutionTypeMarketAfternoonClose,
				Price:            0,
				OrderQuantity:    350,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        nil,
				HoldPositions: []HoldPosition{
					{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
					{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
					{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
				},
			}}},
		{name: "?????????????????????????????????????????????release???????????????",
			clock: &testClock{Now1: time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local)},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, HoldQuantity: 0, Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 200, HoldQuantity: 100, Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 300, HoldQuantity: 150, Price: 102},
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
				ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
				Account: Account{
					Password:    "Password1234",
					AccountType: AccountTypeSpecific}},
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
					ExitStrategy:    ExitStrategy{Runnable: true, Conditions: []ExitCondition{{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 55, 0, 0, time.Local)}}},
					Account: Account{
						Password:    "Password1234",
						AccountType: AccountTypeSpecific}},
				&Order{
					StrategyCode:     "strategy-code-001",
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Status:           OrderStatusInOrder,
					Product:          ProductMargin,
					MarginTradeType:  MarginTradeTypeDay,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					ExecutionType:    ExecutionTypeMarketAfternoonClose,
					OrderQuantity:    350,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					OrderDateTime:    time.Date(2021, 11, 1, 14, 55, 0, 0, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Time{},
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-001", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 100},
						{PositionCode: "position-code-002", HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0, Price: 101},
						{PositionCode: "position-code-003", HoldQuantity: 150, ContractQuantity: 0, ReleaseQuantity: 0, Price: 102},
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
		{name: "store???error???????????????error?????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			want1:      nil,
			want2:      ErrUnknown},
		{name: "store???nil, nil???????????????nil, nil?????????",
			orderStore: &testOrderStore{},
			want1:      nil,
			want2:      nil},
		{name: "store???????????????????????????????????????",
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
		name                 string
		kabusAPI             *testKabusAPI
		orderStore           *testOrderStore
		strategyStore        *testStrategyStore
		clock                IClock
		arg1                 string
		arg2                 float64
		arg3                 float64
		want1                error
		wantSendOrderHistory []interface{}
		wantSaveHistory      []interface{}
	}{
		{name: "?????????????????????????????????????????????",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			clock:         &testClock{},
			arg1:          "strategy-code-001",
			want1:         ErrNoData},
		{name: "?????????????????????????????????????????????",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Cash: 8_000}},
			clock:         &testClock{},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrUnknown},
		{name: "????????????????????????????????????",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Cash: 8_000}},
			clock:         &testClock{},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrNotEnoughCash},
		{name: "SendOrder??????????????????error",
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
		{name: "????????????????????????error",
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
		{name: "????????????????????????????????????????????????error",
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
		{name: "???????????????????????????????????????????????????????????????????????????nil",
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
			}}},
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
				!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					!errors.Is(got1, test.want1),
					!reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory),
					!reflect.DeepEqual(test.wantSaveHistory, test.orderStore.SaveHistory),
					test.want1, test.wantSendOrderHistory, test.wantSaveHistory,
					got1, test.kabusAPI.SendOrderHistory, test.orderStore.SaveHistory)
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
		{name: "?????????????????????????????????????????????",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			arg1:          "strategy-code-001",
			want1:         ErrNoData},
		{name: "?????????????????????????????????????????????????????????",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			want1:         ErrUnknown},
		{name: "??????????????????0?????????????????????????????????",
			clock:         &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrNotEnoughPosition},
		{name: "???????????????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}},
				Hold1: ErrUnknown},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:          "strategy-code-001",
			arg2:          2100,
			arg3:          4,
			want1:         ErrUnknown,
			wantHoldCount: 1},
		{name: "?????????????????????????????????????????????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
			strategyStore:    &testStrategyStore{GetByCode1: &Strategy{Code: "strategy-code-001"}},
			arg1:             "strategy-code-001",
			arg2:             2100,
			arg3:             200,
			want1:            ErrNotEnoughPosition,
			wantHoldCount:    3,
			wantReleaseCount: 3},
		{name: "??????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder2: ErrUnknown},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
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
						{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
						{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
					},
				},
			},
			wantHoldCount:    2,
			wantReleaseCount: 2},
		{name: "?????????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: 4, OrderCode: ""}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
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
						{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
						{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
					},
				},
			},
			wantHoldCount:    2,
			wantReleaseCount: 2},
		{name: "??????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{Save1: ErrUnknown},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
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
						{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
						{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
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
					{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
					{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
				}}},
			wantHoldCount: 2},
		{name: "??????????????????????????????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
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
						{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
						{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
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
					{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
					{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
				}}},
			wantHoldCount: 2},
		{name: "????????????????????????????????????????????????????????????",
			clock:      &testClock{Now1: time.Date(2021, 11, 2, 14, 0, 0, 0, time.Local)},
			kabusAPI:   &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 120, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 4, ContractDateTime: time.Date(2021, 11, 2, 9, 1, 0, 0, time.Local), Price: 101},
					{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 2, 0, 0, time.Local), Price: 102},
					{Code: "position-code-004", StrategyCode: "strategy-code-001", OwnedQuantity: 2, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 2, 9, 3, 0, 0, time.Local), Price: 103}}},
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
					HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4, Price: 100}},
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
				HoldPositions:    []HoldPosition{{PositionCode: "position-code-001", HoldQuantity: 4, Price: 100}}}},
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

	requestErr43St := status.New(codes.Internal, "request 43 status")
	requestErr43StDt, _ := requestErr43St.WithDetails(&kabuspb.RequestError{Code: 43})

	tests := []struct {
		name                   string
		kabusAPI               *testKabusAPI
		logger                 *testLogger
		arg1                   *Strategy
		arg2                   string
		want1                  error
		wantCancelOrderHistory []interface{}
		wantWarningCount       int
	}{
		{name: "?????????nil???????????????",
			kabusAPI: &testKabusAPI{},
			logger:   &testLogger{},
			arg1:     nil,
			arg2:     "order-code-001",
			want1:    ErrNilArgument},
		{name: "???????????????????????????????????????",
			kabusAPI:               &testKabusAPI{CancelOrder2: ErrUnknown},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "???????????????????????????????????????????????????????????????????????????nil?????????",
			kabusAPI:               &testKabusAPI{CancelOrder2: requestErr43StDt.Err()},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"},
			wantWarningCount:       1},
		{name: "?????????????????????????????????",
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: false, ResultCode: 4}},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  ErrCancelCondition,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "????????????????????????nil",
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "cancel-order-code-001"}},
			logger:                 &testLogger{},
			arg1:                   &Strategy{Account: Account{Password: "Password1234"}},
			arg2:                   "order-code-001",
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{kabusAPI: test.kabusAPI, logger: test.logger}
			got1 := service.Cancel(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantWarningCount, got1, test.logger.WarningCount)
			}
		})
	}
}

func Test_orderService_checkEntryCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		orderStore *testOrderStore
		arg1       string
		arg2       float64
		arg3       float64
		arg4       float64
		want1      bool
		want2      error
	}{
		{name: "?????????????????????????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			arg1:       "strategy-code-001",
			arg2:       100_000,
			arg3:       2_100,
			arg4:       4,
			want1:      false,
			want2:      ErrUnknown},
		{name: "??????????????????????????????????????????????????????1???1???????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{Price: 100_000, OrderQuantity: 1}}},
			arg1:  "strategy-code-001",
			arg2:  100_000,
			arg3:  1,
			arg4:  1,
			want1: false,
			want2: nil},
		{name: "???????????????????????? + ??????????????????????????????????????????????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{Price: 95_001, OrderQuantity: 1}}},
			arg1:  "strategy-code-001",
			arg2:  100_000,
			arg3:  5_000,
			arg4:  1,
			want1: false,
			want2: nil},
		{name: "???????????????????????? + ?????????????????????????????????????????????????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{Price: 90_000, OrderQuantity: 1}}},
			arg1:  "strategy-code-001",
			arg2:  100_000,
			arg3:  2_000,
			arg4:  5,
			want1: true,
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{orderStore: test.orderStore}
			got1, got2 := service.checkEntryCash(test.arg1, test.arg2, test.arg3, test.arg4)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_orderService_holdPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		positionStore    *testPositionStore
		arg1             string
		arg2             float64
		arg3             SortOrder
		want1            []HoldPosition
		want2            error
		wantHoldCount    int
		wantReleaseCount int
	}{
		{name: "???????????????????????????????????????????????????",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			arg1:          "strategy-code-001",
			arg2:          4,
			arg3:          SortOrderNewest,
			want1:         nil,
			want2:         ErrUnknown},
		{name: "????????????????????????????????????????????????????????????Hold??????",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local), Price: 100},
				{Code: "position-code-002", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 1, 0, time.Local), Price: 101},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 2, 0, time.Local), Price: 102},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 3, 0, time.Local), Price: 103},
				{Code: "position-code-005", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 4, 0, time.Local), Price: 104}}},
			arg1: "strategy-code-001",
			arg2: 8,
			arg3: SortOrderNewest,
			want1: []HoldPosition{
				{PositionCode: "position-code-005", HoldQuantity: 4, Price: 104},
				{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
				{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102}},
			want2:         nil,
			wantHoldCount: 3},
		{name: "??????????????????????????????????????????????????????Hold??????",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local), Price: 100},
				{Code: "position-code-002", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 1, 0, time.Local), Price: 101},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 2, 0, time.Local), Price: 102},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 3, 0, time.Local), Price: 103},
				{Code: "position-code-005", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 4, 0, time.Local), Price: 104}}},
			arg1: "strategy-code-001",
			arg2: 16,
			arg3: SortOrderLatest,
			want1: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 4, Price: 100},
				{PositionCode: "position-code-002", HoldQuantity: 2, Price: 101},
				{PositionCode: "position-code-003", HoldQuantity: 4, Price: 102},
				{PositionCode: "position-code-004", HoldQuantity: 2, Price: 103},
				{PositionCode: "position-code-005", HoldQuantity: 4, Price: 104}},
			want2:         nil,
			wantHoldCount: 5},
		{name: "??????????????????????????????Hold??????????????????????????????????????????Release???????????????",
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local), Price: 100},
				{Code: "position-code-002", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 1, 0, time.Local), Price: 101},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 2, 0, time.Local), Price: 102},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 3, 0, time.Local), Price: 103},
				{Code: "position-code-005", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 4, 0, time.Local), Price: 104}}},
			arg1:             "strategy-code-001",
			arg2:             17,
			arg3:             SortOrderNewest,
			want1:            nil,
			want2:            ErrNotEnoughPosition,
			wantHoldCount:    5,
			wantReleaseCount: 5},
		{name: "Hold????????????????????????????????????Hold??????????????????Release???????????????",
			positionStore: &testPositionStore{
				GetActivePositionsByStrategyCode1: []*Position{
					{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 0, 0, time.Local), Price: 100},
					{Code: "position-code-002", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 1, 0, time.Local), Price: 101},
					{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 2, 0, time.Local), Price: 102},
					{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 2, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 3, 0, time.Local), Price: 103},
					{Code: "position-code-005", OwnedQuantity: 4, HoldQuantity: 0, ContractDateTime: time.Date(2021, 11, 5, 10, 0, 4, 0, time.Local), Price: 104}},
				Hold1: ErrUnknown},
			arg1:             "strategy-code-001",
			arg2:             17,
			arg3:             SortOrderNewest,
			want1:            nil,
			want2:            ErrUnknown,
			wantHoldCount:    1,
			wantReleaseCount: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{positionStore: test.positionStore}
			got1, got2 := service.holdPositions(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(test.want2, got2) ||
				!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount) ||
				!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount) {
				t.Errorf("%s error\nresult: %+v, %+v, %+v, %+v\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					!reflect.DeepEqual(test.want1, got1),
					!errors.Is(test.want2, got2),
					!reflect.DeepEqual(test.wantHoldCount, test.positionStore.HoldCount),
					!reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount),
					test.want1, test.want2, test.wantHoldCount, test.wantReleaseCount,
					got1, got2, test.positionStore.HoldCount, test.positionStore.ReleaseCount)
			}
		})
	}
}

func Test_orderService_sendOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		kabusAPI         *testKabusAPI
		orderStore       *testOrderStore
		positionStore    *testPositionStore
		arg1             *Strategy
		arg2             *Order
		want1            error
		wantReleaseCount int
	}{
		{name: "arg1???nil???????????????",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          nil,
			arg2:          &Order{},
			want1:         ErrNilArgument},
		{name: "arg2???nil???????????????",
			kabusAPI:      &testKabusAPI{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2:          nil,
			want1:         ErrNilArgument},
		{name: "validation???error?????????????????????",
			kabusAPI:   &testKabusAPI{},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideSell, OwnedQuantity: 50},
			}},
			arg1:  &Strategy{},
			arg2:  &Order{TradeType: TradeTypeEntry, Side: SideSell, OrderQuantity: 1},
			want1: ErrShortSellingRestriction},
		{name: "???????????????????????????????????????",
			kabusAPI:      &testKabusAPI{SendOrder2: ErrUnknown},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2:          &Order{},
			want1:         ErrUnknown},
		{name: "?????????????????????????????????????????????????????????",
			kabusAPI:      &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore:    &testOrderStore{Save1: ErrUnknown},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2:          &Order{},
			want1:         ErrUnknown},
		{name: "?????????????????????????????????????????????nil",
			kabusAPI:      &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			orderStore:    &testOrderStore{Save1: nil},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2:          &Order{},
			want1:         nil},
		{name: "????????????????????????Release???????????????",
			kabusAPI:      &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: 4, OrderCode: ""}},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2: &Order{HoldPositions: []HoldPosition{
				{PositionCode: "position-code-001", HoldQuantity: 4},
				{PositionCode: "position-code-002", HoldQuantity: 4},
				{PositionCode: "position-code-003", HoldQuantity: 4}}},
			want1:            ErrOrderCondition,
			wantReleaseCount: 3},
		{name: "????????????????????????Hold???????????????????????????????????????Release?????????????????????",
			kabusAPI:         &testKabusAPI{SendOrder1: OrderResult{Result: false, ResultCode: 4, OrderCode: ""}},
			orderStore:       &testOrderStore{},
			positionStore:    &testPositionStore{},
			arg1:             &Strategy{},
			arg2:             &Order{HoldPositions: nil},
			want1:            ErrOrderCondition,
			wantReleaseCount: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{
				kabusAPI:      test.kabusAPI,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
			}
			got1 := service.sendOrder(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantReleaseCount, test.positionStore.ReleaseCount) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantReleaseCount, got1, test.positionStore.ReleaseCount)
			}
		})
	}
}

func Test_orderService_EntryMarket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		clock                IClock
		kabusAPI             *testKabusAPI
		strategyStore        *testStrategyStore
		orderStore           *testOrderStore
		arg1                 string
		arg2                 float64
		want1                error
		wantSendOrderHistory []interface{}
	}{
		{name: "???????????????????????????????????????",
			clock:         &testClock{Now1: time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			arg1:          "strategy-code-001",
			arg2:          4.0,
			want1:         ErrNoData},
		{name: "?????????????????????????????????",
			clock:    &testClock{Now1: time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            100_000,
				Account:         Account{AccountType: AccountTypeSpecific},
			}},
			orderStore: &testOrderStore{},
			arg1:       "strategy-code-001",
			arg2:       4.0,
			want1:      nil,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            100_000,
					Account:         Account{AccountType: AccountTypeSpecific}},
				&Order{
					Code:            "order-code-001",
					StrategyCode:    "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Status:          OrderStatusInOrder,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					TradeType:       TradeTypeEntry,
					Side:            SideBuy,
					ExecutionType:   ExecutionTypeMarket,
					OrderQuantity:   4.0,
					AccountType:     AccountTypeSpecific,
					OrderDateTime:   time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local),
				},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{
				clock:         test.clock,
				kabusAPI:      test.kabusAPI,
				orderStore:    test.orderStore,
				strategyStore: test.strategyStore,
			}
			got1 := service.EntryMarket(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantSendOrderHistory, got1, test.kabusAPI.SendOrderHistory)
			}
		})
	}
}

func Test_orderService_ExitMarket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		clock                IClock
		kabusAPI             *testKabusAPI
		strategyStore        *testStrategyStore
		orderStore           *testOrderStore
		positionStore        *testPositionStore
		arg1                 string
		arg2                 float64
		arg3                 SortOrder
		want1                error
		wantSendOrderHistory []interface{}
	}{
		{name: "???????????????????????????????????????",
			clock:         &testClock{Now1: time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local)},
			kabusAPI:      &testKabusAPI{},
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          "strategy-code-001",
			arg2:          4.0,
			arg3:          SortOrderLatest,
			want1:         ErrNoData},
		{name: "?????????????????????????????????",
			clock:    &testClock{Now1: time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            100_000,
				Account:         Account{AccountType: AccountTypeSpecific},
			}},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode2: ErrUnknown},
			arg1:          "strategy-code-001",
			arg2:          4.0,
			arg3:          SortOrderLatest,
			want1:         ErrUnknown},
		{name: "?????????????????????????????????",
			clock:    &testClock{Now1: time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local)},
			kabusAPI: &testKabusAPI{SendOrder1: OrderResult{Result: true, ResultCode: 0, OrderCode: "order-code-001"}},
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
				Code:            "strategy-code-001",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				EntrySide:       SideBuy,
				Cash:            100_000,
				Account:         Account{AccountType: AccountTypeSpecific},
			}},
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 4, Price: 100},
				{Code: "position-code-002", OwnedQuantity: 4, HoldQuantity: 3, Price: 101},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 2, Price: 102},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 1, Price: 103},
				{Code: "position-code-005", OwnedQuantity: 4, HoldQuantity: 2, Price: 104},
				{Code: "position-code-006", OwnedQuantity: 4, HoldQuantity: 3, Price: 105},
				{Code: "position-code-007", OwnedQuantity: 4, HoldQuantity: 4, Price: 106},
			}},
			arg1:  "strategy-code-001",
			arg2:  4.0,
			arg3:  SortOrderLatest,
			want1: nil,
			wantSendOrderHistory: []interface{}{
				&Strategy{
					Code:            "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					EntrySide:       SideBuy,
					Cash:            100_000,
					Account:         Account{AccountType: AccountTypeSpecific}},
				&Order{
					Code:            "order-code-001",
					StrategyCode:    "strategy-code-001",
					SymbolCode:      "1475",
					Exchange:        ExchangeToushou,
					Status:          OrderStatusInOrder,
					Product:         ProductMargin,
					MarginTradeType: MarginTradeTypeDay,
					TradeType:       TradeTypeExit,
					Side:            SideSell,
					ExecutionType:   ExecutionTypeMarket,
					OrderQuantity:   4.0,
					AccountType:     AccountTypeSpecific,
					OrderDateTime:   time.Date(2021, 11, 4, 10, 0, 0, 0, time.Local),
					HoldPositions: []HoldPosition{
						{PositionCode: "position-code-002", HoldQuantity: 1, Price: 101},
						{PositionCode: "position-code-003", HoldQuantity: 2, Price: 102},
						{PositionCode: "position-code-004", HoldQuantity: 1, Price: 103},
					},
				},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{
				clock:         test.clock,
				kabusAPI:      test.kabusAPI,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
				strategyStore: test.strategyStore,
			}
			got1 := service.ExitMarket(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantSendOrderHistory, test.kabusAPI.SendOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantSendOrderHistory, got1, test.kabusAPI.SendOrderHistory)
			}
		})
	}
}

func Test_newOrderService(t *testing.T) {
	t.Parallel()
	clock := &testClock{}
	kabusAPI := &testKabusAPI{}
	strategyStore := &testStrategyStore{}
	orderStore := &testOrderStore{}
	positionStore := &testPositionStore{}
	logger := &testLogger{}
	want1 := &orderService{
		clock:         clock,
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
		logger:        logger,
	}
	got1 := newOrderService(clock, kabusAPI, strategyStore, orderStore, positionStore, logger)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_orderService_validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		orderStore    *testOrderStore
		positionStore *testPositionStore
		arg1          *Strategy
		arg2          *Order
		want1         error
	}{
		{name: "?????????nil???????????????",
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          nil,
			arg2:          &Order{},
			want1:         ErrNilArgument},
		{name: "?????????nil???????????????",
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			arg1:          &Strategy{},
			arg2:          nil,
			want1:         ErrNilArgument},
		{name: "?????????????????????????????????50???????????????????????????????????????????????????",
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
			}},
			arg1:  &Strategy{Code: "strategy-code-001", TradingUnit: 1},
			arg2:  &Order{StrategyCode: "strategy-code-001", Side: SideSell, TradeType: TradeTypeEntry, OrderQuantity: 1},
			want1: ErrShortSellingRestriction},
		{name: "????????????????????????50??????????????????????????????????????????????????????50????????????????????????????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
			}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
			}},
			arg1:  &Strategy{Code: "strategy-code-001", TradingUnit: 1},
			arg2:  &Order{StrategyCode: "strategy-code-001", Side: SideSell, TradeType: TradeTypeEntry, OrderQuantity: 1},
			want1: ErrShortSellingRestriction},
		{name: "????????????????????????50???????????????????????????????????????????????????50??????????????????????????????Exit???????????????????????????????????????",
			orderStore: &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeExit, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeExit, OrderQuantity: 2},
				{TradeType: TradeTypeEntry, OrderQuantity: 2},
				{TradeType: TradeTypeExit, OrderQuantity: 2},
			}},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
			}},
			arg1:  &Strategy{Code: "strategy-code-001", TradingUnit: 1},
			arg2:  &Order{StrategyCode: "strategy-code-001", Side: SideSell, TradeType: TradeTypeEntry, OrderQuantity: 1},
			want1: nil},
		{name: "???????????????????????????50???????????????????????????????????????????????????????????????????????????????????????",
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideBuy, OwnedQuantity: 10},
				{Side: SideBuy, OwnedQuantity: 10},
				{Side: SideBuy, OwnedQuantity: 10},
				{Side: SideBuy, OwnedQuantity: 10},
				{Side: SideBuy, OwnedQuantity: 10},
			}},
			arg1:  &Strategy{Code: "strategy-code-001", TradingUnit: 1},
			arg2:  &Order{StrategyCode: "strategy-code-001", Side: SideSell, TradeType: TradeTypeEntry, OrderQuantity: 1},
			want1: nil},
		{name: "?????????????????????????????????50????????????????????????????????????????????????????????????????????????????????????",
			orderStore: &testOrderStore{},
			positionStore: &testPositionStore{GetActivePositionsByStrategyCode1: []*Position{
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
				{Side: SideSell, OwnedQuantity: 10},
			}},
			arg1:  &Strategy{Code: "strategy-code-001", TradingUnit: 1},
			arg2:  &Order{StrategyCode: "strategy-code-001", Side: SideSell, TradeType: TradeTypeExit, OrderQuantity: 1},
			want1: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{orderStore: test.orderStore, positionStore: test.positionStore}
			got1 := service.validation(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_orderService_handleCancelOrderError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		logger *testLogger
		arg1   error
		arg2   string
		want1  error
	}{
		{name: "err???nil??????nil?????????",
			logger: &testLogger{},
			arg1:   nil,
			arg2:   "order-code-001",
			want1:  nil},
		// ?????????????????????????????????????????????????????????????????????????????????????????????
		// ???????????????????????????????????????????????????????????????
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{logger: test.logger}
			got1 := service.handleCancelOrderError(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
