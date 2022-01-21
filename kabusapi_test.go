package gridon

import (
	"context"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc"

	"gitlab.com/tsuchinaga/kabus-grpc-server/kabuspb"
)

type testKabusAPI struct {
	IKabusAPI
	GetOrders1          []SecurityOrder
	GetOrders2          error
	GetOrdersCount      int
	GetOrdersHistory    []interface{}
	CancelOrder1        OrderResult
	CancelOrder2        error
	CancelOrderHistory  []interface{}
	SendOrder1          OrderResult
	SendOrder2          error
	SendOrderCount      int
	SendOrderHistory    []interface{}
	GetSymbol1          *Symbol
	GetSymbol2          error
	GetSymbolCount      int
	GetSymbolHistory    []interface{}
	GetFourPrice1       *FourPrice
	GetFourPrice2       error
	GetFourPriceCount   int
	GetFourPriceHistory []interface{}
}

func (t *testKabusAPI) GetSymbol(symbolCode string, exchange Exchange) (*Symbol, error) {
	t.GetSymbolHistory = append(t.GetSymbolHistory, symbolCode)
	t.GetSymbolHistory = append(t.GetSymbolHistory, exchange)
	t.GetSymbolCount++
	return t.GetSymbol1, t.GetSymbol2
}
func (t *testKabusAPI) GetOrders(product Product, symbolCode string, updateDateTime time.Time) ([]SecurityOrder, error) {
	t.GetOrdersHistory = append(t.GetOrdersHistory, product)
	t.GetOrdersHistory = append(t.GetOrdersHistory, symbolCode)
	t.GetOrdersHistory = append(t.GetOrdersHistory, updateDateTime)
	t.GetOrdersCount++
	return t.GetOrders1, t.GetOrders2
}
func (t *testKabusAPI) CancelOrder(orderPassword string, orderCode string) (OrderResult, error) {
	t.CancelOrderHistory = append(t.CancelOrderHistory, orderPassword)
	t.CancelOrderHistory = append(t.CancelOrderHistory, orderCode)
	return t.CancelOrder1, t.CancelOrder2
}
func (t *testKabusAPI) SendOrder(strategy *Strategy, order *Order) (OrderResult, error) {
	t.SendOrderHistory = append(t.SendOrderHistory, strategy)
	t.SendOrderHistory = append(t.SendOrderHistory, order)
	t.SendOrderCount++
	return t.SendOrder1, t.SendOrder2
}
func (t *testKabusAPI) GetFourPrice(symbolCode string, exchange Exchange) (*FourPrice, error) {
	t.GetFourPriceCount++
	t.GetFourPriceHistory = append(t.GetFourPriceHistory, symbolCode)
	t.GetFourPriceHistory = append(t.GetFourPriceHistory, exchange)
	return t.GetFourPrice1, t.GetFourPrice2
}

type testKabusServiceClient struct {
	GetBoard1              *kabuspb.Board
	GetBoard2              error
	GetSymbol1             *kabuspb.Symbol
	GetSymbol2             error
	GetOrders1             *kabuspb.Orders
	GetOrders2             error
	CancelOrder1           *kabuspb.OrderResponse
	CancelOrder2           error
	CancelOrderHistory     []interface{}
	SendStockOrder1        *kabuspb.OrderResponse
	SendStockOrder2        error
	SendStockOrderHistory  []interface{}
	SendMarginOrder1       *kabuspb.OrderResponse
	SendMarginOrder2       error
	SendMarginOrderHistory []interface{}
	kabuspb.KabusServiceClient
}

func (t *testKabusServiceClient) GetBoard(context.Context, *kabuspb.GetBoardRequest, ...grpc.CallOption) (*kabuspb.Board, error) {
	return t.GetBoard1, t.GetBoard2
}
func (t *testKabusServiceClient) GetSymbol(context.Context, *kabuspb.GetSymbolRequest, ...grpc.CallOption) (*kabuspb.Symbol, error) {
	return t.GetSymbol1, t.GetSymbol2
}
func (t *testKabusServiceClient) GetOrders(context.Context, *kabuspb.GetOrdersRequest, ...grpc.CallOption) (*kabuspb.Orders, error) {
	return t.GetOrders1, t.GetOrders2
}
func (t *testKabusServiceClient) CancelOrder(_ context.Context, in *kabuspb.CancelOrderRequest, _ ...grpc.CallOption) (*kabuspb.OrderResponse, error) {
	t.CancelOrderHistory = append(t.CancelOrderHistory, in)
	return t.CancelOrder1, t.CancelOrder2
}
func (t *testKabusServiceClient) SendStockOrder(_ context.Context, in *kabuspb.SendStockOrderRequest, _ ...grpc.CallOption) (*kabuspb.OrderResponse, error) {
	t.SendStockOrderHistory = append(t.SendStockOrderHistory, in)
	return t.SendStockOrder1, t.SendStockOrder2
}
func (t *testKabusServiceClient) SendMarginOrder(_ context.Context, in *kabuspb.SendMarginOrderRequest, _ ...grpc.CallOption) (*kabuspb.OrderResponse, error) {
	t.SendMarginOrderHistory = append(t.SendMarginOrderHistory, in)
	return t.SendMarginOrder1, t.SendMarginOrder2
}

func Test_kabusAPI_exchangeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Exchange
		want kabuspb.Exchange
	}{
		{name: "未指定 を変換できる", arg1: ExchangeUnspecified, want: kabuspb.Exchange_EXCHANGE_UNSPECIFIED},
		{name: "東証 を変換できる", arg1: ExchangeToushou, want: kabuspb.Exchange_EXCHANGE_TOUSHOU},
		{name: "名証 を変換できる", arg1: ExchangeMeishou, want: kabuspb.Exchange_EXCHANGE_MEISHOU},
		{name: "福証 を変換できる", arg1: ExchangeFukushou, want: kabuspb.Exchange_EXCHANGE_FUKUSHOU},
		{name: "札証 を変換できる", arg1: ExchangeSatsushou, want: kabuspb.Exchange_EXCHANGE_SATSUSHOU},
		{name: "SOR を変換できる", arg1: ExchangeSOR, want: kabuspb.Exchange_EXCHANGE_UNSPECIFIED},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.exchangeTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_exchangeFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 kabuspb.Exchange
		want Exchange
	}{
		{name: "未指定 を変換できる", arg1: kabuspb.Exchange_EXCHANGE_UNSPECIFIED, want: ExchangeUnspecified},
		{name: "東証 を変換できる", arg1: kabuspb.Exchange_EXCHANGE_TOUSHOU, want: ExchangeToushou},
		{name: "名証 を変換できる", arg1: kabuspb.Exchange_EXCHANGE_MEISHOU, want: ExchangeMeishou},
		{name: "福証 を変換できる", arg1: kabuspb.Exchange_EXCHANGE_FUKUSHOU, want: ExchangeFukushou},
		{name: "札証 を変換できる", arg1: kabuspb.Exchange_EXCHANGE_SATSUSHOU, want: ExchangeSatsushou},
		{name: "Exchange(9) を変換できる", arg1: kabuspb.Exchange(9), want: ExchangeSOR},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.exchangeFrom(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_GetSymbol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		kabusServiceClient *testKabusServiceClient
		arg1               string
		arg2               Exchange
		want1              *Symbol
		want2              error
	}{
		{name: "symbol取得に失敗したらエラー",
			kabusServiceClient: &testKabusServiceClient{GetSymbol2: ErrUnknown},
			arg1:               "1475",
			arg2:               ExchangeToushou,
			want2:              ErrUnknown},
		{name: "board取得に失敗したらエラー",
			kabusServiceClient: &testKabusServiceClient{GetSymbol1: &kabuspb.Symbol{}, GetBoard2: ErrUnknown},
			arg1:               "1475",
			arg2:               ExchangeToushou,
			want2:              ErrUnknown},
		{name: "symbolもboardも取得できたら情報を返す",
			kabusServiceClient: &testKabusServiceClient{
				GetSymbol1: &kabuspb.Symbol{Code: "1475", Exchange: kabuspb.Exchange_EXCHANGE_TOUSHOU, TradingUnit: 1, UpperLimit: 2576, LowerLimit: 1576},
				GetBoard1:  &kabuspb.Board{CurrentPrice: 2076, CurrentPriceTime: timestamppb.New(time.Date(2021, 11, 19, 10, 0, 0, 0, time.Local)), CalculationPrice: 2076, BidPrice: 2075, AskPrice: 2077}},
			arg1: "1475",
			arg2: ExchangeToushou,
			want1: &Symbol{
				Code:                 "1475",
				Exchange:             ExchangeToushou,
				TradingUnit:          1,
				CurrentPrice:         2076,
				CurrentPriceDateTime: time.Date(2021, 11, 19, 10, 0, 0, 0, time.Local),
				BidPrice:             2075,
				AskPrice:             2077,
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusapi := &kabusAPI{kabucom: test.kabusServiceClient}
			got1, got2 := kabusapi.GetSymbol(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_kabusAPI_productTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		product Product
		want1   kabuspb.Product
	}{
		{name: "未指定 を変換できる", product: ProductUnspecified, want1: kabuspb.Product_PRODUCT_UNSPECIFIED},
		{name: "現物 を変換できる", product: ProductStock, want1: kabuspb.Product_PRODUCT_STOCK},
		{name: "信用 を変換できる", product: ProductMargin, want1: kabuspb.Product_PRODUCT_MARGIN},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusapi := &kabusAPI{}
			got1 := kabusapi.productTo(test.product)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_productFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		product kabuspb.Product
		want1   Product
	}{
		{name: "未指定 を変換できる", product: kabuspb.Product_PRODUCT_UNSPECIFIED, want1: ProductUnspecified},
		{name: "現物 を変換できる", product: kabuspb.Product_PRODUCT_STOCK, want1: ProductStock},
		{name: "信用 を変換できる", product: kabuspb.Product_PRODUCT_MARGIN, want1: ProductMargin},
		{name: "先物 を変換できる", product: kabuspb.Product_PRODUCT_FUTURE, want1: ProductUnspecified},
		{name: "オプション を変換できる", product: kabuspb.Product_PRODUCT_OPTION, want1: ProductUnspecified},
		{name: "全部 を変換できる", product: kabuspb.Product_PRODUCT_ALL, want1: ProductUnspecified},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusapi := &kabusAPI{}
			got1 := kabusapi.productFrom(test.product)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_orderStatusFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  kabuspb.RecordType
		arg2  float64
		arg3  float64
		want1 OrderStatus
	}{
		{name: "未指定 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_UNSPECIFIED, arg2: 0, arg3: 0, want1: OrderStatusUnspecified},
		{name: "受付 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_RECEIVE, arg2: 0, arg3: 0, want1: OrderStatusInOrder},
		{name: "繰越 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_CARRIED, arg2: 0, arg3: 0, want1: OrderStatusInOrder},
		{name: "期限切れ を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_EXPIRED, arg2: 0, arg3: 0, want1: OrderStatusCanceled},
		{name: "発注 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_ORDERED, arg2: 0, arg3: 0, want1: OrderStatusInOrder},
		{name: "訂正 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_MODIFIED, arg2: 0, arg3: 0, want1: OrderStatusInOrder},
		{name: "取消 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_CANCELED, arg2: 0, arg3: 0, want1: OrderStatusCanceled},
		{name: "失効 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_REVOCATION, arg2: 0, arg3: 0, want1: OrderStatusCanceled},
		{name: "部分約定 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, arg2: 100, arg3: 10, want1: OrderStatusInOrder},
		{name: "全約定 を変換できる", arg1: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, arg2: 100, arg3: 100, want1: OrderStatusDone},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.orderStatusFrom(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_marginTradeTypeFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  kabuspb.MarginTradeType
		want1 MarginTradeType
	}{
		{name: "未指定 を変換できる", arg1: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_UNSPECIFIED, want1: MarginTradeTypeUnspecified},
		{name: "制度 を変換できる", arg1: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_SYSTEM, want1: MarginTradeTypeSystem},
		{name: "長期 を変換できる", arg1: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_LONG, want1: MarginTradeTypeLong},
		{name: "デイトレ を変換できる", arg1: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY, want1: MarginTradeTypeDay},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got := kabusAPI.marginTradeTypeFrom(test.arg1)
			if !reflect.DeepEqual(test.want1, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got)
			}
		})
	}
}

func Test_kabusAPI_tradeTypeFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  kabuspb.Product
		arg2  kabuspb.Side
		arg3  kabuspb.TradeType
		want1 TradeType
	}{
		{name: "製品の 未指定 を変換できる",
			arg1:  kabuspb.Product_PRODUCT_UNSPECIFIED,
			arg2:  kabuspb.Side_SIDE_UNSPECIFIED,
			arg3:  kabuspb.TradeType_TRADE_TYPE_UNSPECIFIED,
			want1: TradeTypeUnspecified},
		{name: "現物の買い を変換できる",
			arg1:  kabuspb.Product_PRODUCT_STOCK,
			arg2:  kabuspb.Side_SIDE_BUY,
			arg3:  kabuspb.TradeType_TRADE_TYPE_UNSPECIFIED,
			want1: TradeTypeEntry},
		{name: "現物の売り を変換できる",
			arg1:  kabuspb.Product_PRODUCT_STOCK,
			arg2:  kabuspb.Side_SIDE_SELL,
			arg3:  kabuspb.TradeType_TRADE_TYPE_UNSPECIFIED,
			want1: TradeTypeExit},
		{name: "信用のエントリー を変換できる",
			arg1:  kabuspb.Product_PRODUCT_MARGIN,
			arg2:  kabuspb.Side_SIDE_SELL,
			arg3:  kabuspb.TradeType_TRADE_TYPE_ENTRY,
			want1: TradeTypeEntry},
		{name: "信用のエグジット を変換できる",
			arg1:  kabuspb.Product_PRODUCT_MARGIN,
			arg2:  kabuspb.Side_SIDE_BUY,
			arg3:  kabuspb.TradeType_TRADE_TYPE_EXIT,
			want1: TradeTypeExit},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.tradeTypeFrom(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot1: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_sideFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  kabuspb.Side
		want1 Side
	}{
		{name: "未指定 を変換できる", arg1: kabuspb.Side_SIDE_UNSPECIFIED, want1: SideUnspecified},
		{name: "買い を変換できる", arg1: kabuspb.Side_SIDE_BUY, want1: SideBuy},
		{name: "売り を変換できる", arg1: kabuspb.Side_SIDE_SELL, want1: SideSell},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.sideFrom(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot1: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_accountTypeFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  kabuspb.AccountType
		want1 AccountType
	}{
		{name: "未指定 を変換できる", arg1: kabuspb.AccountType_ACCOUNT_TYPE_UNSPECIFIED, want1: AccountTypeUnspecified},
		{name: "一般 を変換できる", arg1: kabuspb.AccountType_ACCOUNT_TYPE_GENERAL, want1: AccountTypeGeneral},
		{name: "特定 を変換できる", arg1: kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC, want1: AccountTypeSpecific},
		{name: "法人 を変換できる", arg1: kabuspb.AccountType_ACCOUNT_TYPE_CORPORATION, want1: AccountTypeCorporation},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.accountTypeFrom(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot1: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_contractFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  string
		arg2  *kabuspb.OrderDetail
		want1 Contract
	}{
		{name: "詳細を約定に変換できる",
			arg1: "order-code",
			arg2: &kabuspb.OrderDetail{
				SequenceNumber: 5,
				Id:             "20211026E02N24435416",
				RecordType:     kabuspb.RecordType_RECORD_TYPE_CONTRACTED,
				ExchangeId:     "507",
				State:          kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED,
				TransactTime:   &timestamppb.Timestamp{Seconds: 1635210383, Nanos: 643125000},
				Price:          2076,
				Quantity:       4,
				ExecutionId:    "E2021102601AWG",
				ExecutionDay:   &timestamppb.Timestamp{Seconds: 1635210383, Nanos: 643125000},
				DeliveryDay:    &timestamppb.Timestamp{Seconds: 1635346800},
			},
			want1: Contract{
				OrderCode:        "order-code",
				PositionCode:     "E2021102601AWG",
				Price:            2076,
				Quantity:         4,
				ContractDateTime: time.Unix(1635210383, 643125000),
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.contractFrom(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusAPI_securityOrderFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 kabuspb.Product
		arg2 *kabuspb.Order
		want SecurityOrder
	}{
		{name: "未約定の注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_STOCK,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N21800824",
				State:              kabuspb.State_STATE_PROCESSED,
				OrderState:         kabuspb.OrderState_ORDER_STATE_PROCESSED,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2048,
				OrderQuantity:      4,
				CumulativeQuantity: 0,
				Side:               kabuspb.Side_SIDE_SELL,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N21800824", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N21800825", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yz_9mfG4Qheq9Fk0DfyMUwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860821, Nanos: 148000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N21800824",
				Status:           OrderStatusInOrder,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductStock,
				MarginTradeType:  MarginTradeTypeUnspecified,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				Price:            2048,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 20, 785189700, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{},
			}},
		{name: "部分約定した注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_MARGIN,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N21800824",
				State:              kabuspb.State_STATE_PROCESSED,
				OrderState:         kabuspb.OrderState_ORDER_STATE_PROCESSED,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2048,
				OrderQuantity:      4,
				CumulativeQuantity: 2,
				Side:               kabuspb.Side_SIDE_SELL,
				TradeType:          kabuspb.TradeType_TRADE_TYPE_ENTRY,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				MarginTradeType:    kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY,
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N21800824", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N21800825", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yz_9mfG4Qheq9Fk0DfyMUwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860821, Nanos: 148000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 5, Id: "20211022E02N21811093", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "155", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, Price: 2048, Quantity: 2, ExecutionId: "E2021102200BJ6", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N21800824",
				Status:           OrderStatusInOrder,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideSell,
				Price:            2048,
				OrderQuantity:    4,
				ContractQuantity: 2,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 20, 785189700, time.Local),
				ContractDateTime: time.Date(2021, 10, 22, 9, 1, 25, 534000000, time.Local),
				CancelDateTime:   time.Time{},
				Contracts: []Contract{
					{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200BJ6", Price: 2048, Quantity: 2, ContractDateTime: time.Date(2021, 10, 22, 9, 1, 25, 534000000, time.Local)},
				}}},
		{name: "約定した注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_STOCK,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N21800824",
				State:              kabuspb.State_STATE_DONE,
				OrderState:         kabuspb.OrderState_ORDER_STATE_DONE,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2048,
				OrderQuantity:      4,
				CumulativeQuantity: 4,
				Side:               kabuspb.Side_SIDE_SELL,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N21800824", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N21800825", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yz_9mfG4Qheq9Fk0DfyMUwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860821, Nanos: 148000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 5, Id: "20211022E02N21811093", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "155", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, Price: 2048, Quantity: 1, ExecutionId: "E2021102200BJ6", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 6, Id: "20211022E02N21840437", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "212", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634861186, Nanos: 575000000}, Price: 2048, Quantity: 2, ExecutionId: "E2021102200GSG", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634861186, Nanos: 575000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 7, Id: "20211022E02N21887185", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "357", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634861930, Nanos: 190000000}, Price: 2048, Quantity: 1, ExecutionId: "E2021102200QR3", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634861930, Nanos: 190000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N21800824",
				Status:           OrderStatusDone,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductStock,
				MarginTradeType:  MarginTradeTypeUnspecified,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				Price:            2048,
				OrderQuantity:    4,
				ContractQuantity: 4,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 20, 785189700, time.Local),
				ContractDateTime: time.Date(2021, 10, 22, 9, 18, 50, 190000000, time.Local),
				CancelDateTime:   time.Time{},
				Contracts: []Contract{
					{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200BJ6", Price: 2048, Quantity: 1, ContractDateTime: time.Date(2021, 10, 22, 9, 1, 25, 534000000, time.Local)},
					{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200GSG", Price: 2048, Quantity: 2, ContractDateTime: time.Date(2021, 10, 22, 9, 6, 26, 575000000, time.Local)},
					{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200QR3", Price: 2048, Quantity: 1, ContractDateTime: time.Date(2021, 10, 22, 9, 18, 50, 190000000, time.Local)},
				}}},
		{name: "取消した注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_STOCK,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N21800958",
				State:              kabuspb.State_STATE_DONE,
				OrderState:         kabuspb.OrderState_ORDER_STATE_DONE,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860857, Nanos: 836139700},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2040,
				OrderQuantity:      4,
				CumulativeQuantity: 0,
				Side:               kabuspb.Side_SIDE_BUY,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				DeliveryType:       kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N21800958", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860857, Nanos: 836139700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2040, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N21800959", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yq_ru7r3T5uNzG0iV74qtwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860858, Nanos: 199000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2040, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 7, Id: "20211022D02N21970739", RecordType: kabuspb.RecordType_RECORD_TYPE_CANCELED, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634864453, Nanos: 160000000}, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N21800958",
				Status:           OrderStatusCanceled,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductStock,
				MarginTradeType:  MarginTradeTypeUnspecified,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				Price:            2040,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 57, 836139700, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Date(2021, 10, 22, 10, 0, 53, 160000000, time.Local),
				Contracts:        []Contract{},
			}},
		{name: "有効期限切れになった注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_MARGIN,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N22093740",
				State:              kabuspb.State_STATE_DONE,
				OrderState:         kabuspb.OrderState_ORDER_STATE_DONE,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 169160500},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2066,
				OrderQuantity:      4,
				CumulativeQuantity: 0,
				Side:               kabuspb.Side_SIDE_SELL,
				TradeType:          kabuspb.TradeType_TRADE_TYPE_ENTRY,
				MarginTradeType:    kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				DeliveryType:       kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N22093740", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 169160500}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2066, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N22093741", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "IHef9QTRSMenXGOWevVoqAA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 513000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2066, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 7, Id: "20211022D02N22313573", RecordType: kabuspb.RecordType_RECORD_TYPE_CANCELED, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_ERROR, TransactTime: &timestamppb.Timestamp{Seconds: 1634882580, Nanos: 42000000}, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 8, Id: "20211022F02N22353972", RecordType: kabuspb.RecordType_RECORD_TYPE_REVOCATION, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634885102, Nanos: 840154200}, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 9, Id: "20211022G02N22353974", RecordType: kabuspb.RecordType_RECORD_TYPE_EXPIRED, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634885168, Nanos: 843006600}, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N22093740",
				Status:           OrderStatusCanceled,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideSell,
				Price:            2066,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 12, 30, 9, 169160500, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Date(2021, 10, 22, 15, 46, 8, 843006600, time.Local),
				Contracts:        []Contract{},
			}},
		{name: "後場後の失効待ちの注文を変換できる",
			arg1: kabuspb.Product_PRODUCT_STOCK,
			arg2: &kabuspb.Order{
				Id:                 "20211022A02N22093740",
				State:              kabuspb.State_STATE_PROCESSED,
				OrderState:         kabuspb.OrderState_ORDER_STATE_PROCESSED,
				OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
				ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 169160500},
				SymbolCode:         "1475",
				SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
				Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
				ExchangeName:       "東証ETF/ETN",
				Price:              2066,
				OrderQuantity:      4,
				CumulativeQuantity: 0,
				Side:               kabuspb.Side_SIDE_SELL,
				AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				DeliveryType:       kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
				Details: []*kabuspb.OrderDetail{
					{SequenceNumber: 1, Id: "20211022A02N22093740", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 169160500}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2066, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					{SequenceNumber: 4, Id: "20211022B02N22093741", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "IHef9QTRSMenXGOWevVoqAA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634873409, Nanos: 513000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2066, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
				},
			},
			want: SecurityOrder{
				Code:             "20211022A02N22093740",
				Status:           OrderStatusInOrder,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductStock,
				MarginTradeType:  MarginTradeTypeUnspecified,
				TradeType:        TradeTypeExit,
				Side:             SideSell,
				Price:            2066,
				OrderQuantity:    4,
				ContractQuantity: 0,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 22, 12, 30, 9, 169160500, time.Local),
				ContractDateTime: time.Time{},
				CancelDateTime:   time.Time{},
				Contracts:        []Contract{},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got := kabusAPI.securityOrderFrom(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_GetOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		kabusServiceClient *testKabusServiceClient
		arg1               Product
		arg2               string
		arg3               time.Time
		want1              []SecurityOrder
		want2              error
	}{
		{name: "errが返されたらerrを返す",
			kabusServiceClient: &testKabusServiceClient{GetOrders2: ErrUnknown},
			arg1:               ProductStock,
			arg2:               "1475",
			arg3:               time.Date(2021, 10, 22, 10, 0, 0, 0, time.Local),
			want2:              ErrUnknown},
		{name: "ordersが空なら空配列を返す",
			kabusServiceClient: &testKabusServiceClient{GetOrders1: &kabuspb.Orders{Orders: []*kabuspb.Order{}}},
			arg1:               ProductStock,
			arg2:               "1475",
			arg3:               time.Date(2021, 10, 22, 10, 0, 0, 0, time.Local),
			want1:              []SecurityOrder{}},
		{name: "ordersの中身を変換して返す",
			kabusServiceClient: &testKabusServiceClient{GetOrders1: &kabuspb.Orders{Orders: []*kabuspb.Order{
				{
					Id:                 "20211022A02N21800824",
					State:              kabuspb.State_STATE_DONE,
					OrderState:         kabuspb.OrderState_ORDER_STATE_DONE,
					OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
					ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700},
					SymbolCode:         "1475",
					SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
					Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
					ExchangeName:       "東証ETF/ETN",
					Price:              2048,
					OrderQuantity:      4,
					Side:               kabuspb.Side_SIDE_BUY,
					CumulativeQuantity: 4,
					AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
					ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
					Details: []*kabuspb.OrderDetail{
						{SequenceNumber: 1, Id: "20211022A02N21800824", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860820, Nanos: 785189700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 4, Id: "20211022B02N21800825", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yz_9mfG4Qheq9Fk0DfyMUwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860821, Nanos: 148000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2048, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 5, Id: "20211022E02N21811093", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "155", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, Price: 2048, Quantity: 1, ExecutionId: "E2021102200BJ6", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634860885, Nanos: 534000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 6, Id: "20211022E02N21840437", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "212", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634861186, Nanos: 575000000}, Price: 2048, Quantity: 2, ExecutionId: "E2021102200GSG", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634861186, Nanos: 575000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 7, Id: "20211022E02N21887185", RecordType: kabuspb.RecordType_RECORD_TYPE_CONTRACTED, ExchangeId: "357", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634861930, Nanos: 190000000}, Price: 2048, Quantity: 1, ExecutionId: "E2021102200QR3", ExecutionDay: &timestamppb.Timestamp{Seconds: 1634861930, Nanos: 190000000}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					},
				}, {
					Id:                 "20211022A02N21800958",
					State:              kabuspb.State_STATE_DONE,
					OrderState:         kabuspb.OrderState_ORDER_STATE_DONE,
					OrderType:          kabuspb.OrderType_ORDER_TYPE_ZARABA,
					ReceiveTime:        &timestamppb.Timestamp{Seconds: 1634860857, Nanos: 836139700},
					SymbolCode:         "1475",
					SymbolName:         "ｉシェアーズ・コア　ＴＯＰＩＸ　ＥＴＦ",
					Exchange:           kabuspb.OrderExchange_ORDER_EXCHANGE_TOUSHOU,
					ExchangeName:       "東証ETF/ETN",
					Price:              2040,
					OrderQuantity:      4,
					CumulativeQuantity: 0,
					Side:               kabuspb.Side_SIDE_SELL,
					AccountType:        kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
					DeliveryType:       kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
					ExpireDay:          &timestamppb.Timestamp{Seconds: 1634828400},
					Details: []*kabuspb.OrderDetail{
						{SequenceNumber: 1, Id: "20211022A02N21800958", RecordType: kabuspb.RecordType_RECORD_TYPE_RECEIVE, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860857, Nanos: 836139700}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2040, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 4, Id: "20211022B02N21800959", RecordType: kabuspb.RecordType_RECORD_TYPE_ORDERED, ExchangeId: "Yq_ru7r3T5uNzG0iV74qtwA", State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634860858, Nanos: 199000000}, OrderType: kabuspb.OrderType_ORDER_TYPE_ZARABA, Price: 2040, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
						{SequenceNumber: 7, Id: "20211022D02N21970739", RecordType: kabuspb.RecordType_RECORD_TYPE_CANCELED, State: kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED, TransactTime: &timestamppb.Timestamp{Seconds: 1634864453, Nanos: 160000000}, Quantity: 4, ExecutionDay: &timestamppb.Timestamp{Seconds: -62135596800}, DeliveryDay: &timestamppb.Timestamp{Seconds: 1635174000}},
					},
				}}}},
			arg1: ProductStock,
			arg2: "1475",
			arg3: time.Date(2021, 10, 22, 10, 0, 0, 0, time.Local),
			want1: []SecurityOrder{
				{
					Code:             "20211022A02N21800824",
					Status:           OrderStatusDone,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductStock,
					MarginTradeType:  MarginTradeTypeUnspecified,
					TradeType:        TradeTypeEntry,
					Side:             SideBuy,
					Price:            2048,
					OrderQuantity:    4,
					ContractQuantity: 4,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 20, 785189700, time.Local),
					ContractDateTime: time.Date(2021, 10, 22, 9, 18, 50, 190000000, time.Local),
					CancelDateTime:   time.Time{},
					Contracts: []Contract{
						{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200BJ6", Price: 2048, Quantity: 1, ContractDateTime: time.Date(2021, 10, 22, 9, 1, 25, 534000000, time.Local)},
						{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200GSG", Price: 2048, Quantity: 2, ContractDateTime: time.Date(2021, 10, 22, 9, 6, 26, 575000000, time.Local)},
						{OrderCode: "20211022A02N21800824", PositionCode: "E2021102200QR3", Price: 2048, Quantity: 1, ContractDateTime: time.Date(2021, 10, 22, 9, 18, 50, 190000000, time.Local)},
					},
				}, {
					Code:             "20211022A02N21800958",
					Status:           OrderStatusCanceled,
					SymbolCode:       "1475",
					Exchange:         ExchangeToushou,
					Product:          ProductStock,
					MarginTradeType:  MarginTradeTypeUnspecified,
					TradeType:        TradeTypeExit,
					Side:             SideSell,
					Price:            2040,
					OrderQuantity:    4,
					ContractQuantity: 0,
					AccountType:      AccountTypeSpecific,
					ExpireDay:        time.Date(2021, 10, 22, 0, 0, 0, 0, time.Local),
					OrderDateTime:    time.Date(2021, 10, 22, 9, 0, 57, 836139700, time.Local),
					ContractDateTime: time.Time{},
					CancelDateTime:   time.Date(2021, 10, 22, 10, 0, 53, 160000000, time.Local),
					Contracts:        []Contract{},
				}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{kabucom: test.kabusServiceClient}
			got1, got2 := kabusAPI.GetOrders(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_kabusAPI_GetSymbol_Execute(t *testing.T) {
	t.Skip("実際にAPIを叩くテストのため、通常はスキップ")
	t.Parallel()

	conn, err := grpc.DialContext(context.Background(), "localhost:18082", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	kabusAPI := &kabusAPI{kabucom: kabuspb.NewKabusServiceClient(conn)}
	symbol, err := kabusAPI.GetSymbol("1475", ExchangeToushou)
	t.Logf("symbol: %+v, err: %+v\n", symbol, err)
}

func Test_kabusAPI_GetOrders_Execute(t *testing.T) {
	t.Skip("実際にAPIを叩くテストのため、通常はスキップ")
	t.Parallel()

	conn, err := grpc.DialContext(context.Background(), "localhost:18082", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	kabusAPI := &kabusAPI{kabucom: kabuspb.NewKabusServiceClient(conn)}
	orders, err := kabusAPI.GetOrders(ProductMargin, "1475", time.Date(2021, 12, 17, 10, 0, 0, 0, time.Local))
	t.Logf("orders: %+v, err: %+v\n", orders, err)
}

func Test_kabusAPI_CancelOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		kabusServiceClient     *testKabusServiceClient
		arg1                   string
		arg2                   string
		want1                  OrderResult
		want2                  error
		wantCancelOrderHistory []interface{}
	}{
		{name: "取消注文に失敗したらエラー",
			kabusServiceClient:     &testKabusServiceClient{CancelOrder2: ErrUnknown},
			arg1:                   "Password1234",
			arg2:                   "order-code-001",
			want1:                  OrderResult{},
			want2:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{&kabuspb.CancelOrderRequest{Password: "Password1234", OrderId: "order-code-001", IsVirtual: false}}},
		{name: "取消注文の結果を詰めて返す",
			kabusServiceClient:     &testKabusServiceClient{CancelOrder1: &kabuspb.OrderResponse{ResultCode: 0, OrderId: "cancel-order-code"}},
			arg1:                   "Password1234",
			arg2:                   "order-code-001",
			want1:                  OrderResult{Result: true, ResultCode: 0, OrderCode: "cancel-order-code"},
			want2:                  nil,
			wantCancelOrderHistory: []interface{}{&kabuspb.CancelOrderRequest{Password: "Password1234", OrderId: "order-code-001", IsVirtual: false}}},
		{name: "実行結果がエラーならresultがfalseになる",
			kabusServiceClient:     &testKabusServiceClient{CancelOrder1: &kabuspb.OrderResponse{ResultCode: -1, OrderId: "cancel-order-code"}},
			arg1:                   "Password1234",
			arg2:                   "order-code-001",
			want1:                  OrderResult{Result: false, ResultCode: -1, OrderCode: "cancel-order-code"},
			want2:                  nil,
			wantCancelOrderHistory: []interface{}{&kabuspb.CancelOrderRequest{Password: "Password1234", OrderId: "order-code-001", IsVirtual: false}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{kabucom: test.kabusServiceClient}
			got1, got2 := kabusAPI.CancelOrder(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) || !reflect.DeepEqual(test.wantCancelOrderHistory, test.kabusServiceClient.CancelOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantCancelOrderHistory,
					got1, got2, test.kabusServiceClient.CancelOrderHistory)
			}
		})
	}
}

func Test_kabusAPI_sideTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Side
		want kabuspb.Side
	}{
		{name: "未指定 を変換できる", arg1: SideUnspecified, want: kabuspb.Side_SIDE_UNSPECIFIED},
		{name: "買い を変換できる", arg1: SideBuy, want: kabuspb.Side_SIDE_BUY},
		{name: "売り を変換できる", arg1: SideSell, want: kabuspb.Side_SIDE_SELL},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.sideTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_accountTypeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 AccountType
		want kabuspb.AccountType
	}{
		{name: "未指定 を変換できる", arg1: AccountTypeUnspecified, want: kabuspb.AccountType_ACCOUNT_TYPE_UNSPECIFIED},
		{name: "一般 を変換できる", arg1: AccountTypeGeneral, want: kabuspb.AccountType_ACCOUNT_TYPE_GENERAL},
		{name: "特定 を変換できる", arg1: AccountTypeSpecific, want: kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC},
		{name: "法人 を変換できる", arg1: AccountTypeCorporation, want: kabuspb.AccountType_ACCOUNT_TYPE_CORPORATION},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.accountTypeTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_orderTypeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 ExecutionType
		want kabuspb.StockOrderType
	}{
		{name: "未指定 を変換できる", arg1: ExecutionTypeUnspecified, want: kabuspb.StockOrderType_STOCK_ORDER_TYPE_UNSPECIFIED},
		{name: "成行 を変換できる", arg1: ExecutionTypeMarket, want: kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO},
		{name: "前場引成 を変換できる", arg1: ExecutionTypeMarketMorningClose, want: kabuspb.StockOrderType_STOCK_ORDER_TYPE_MOMC},
		{name: "後場引成 を変換できる", arg1: ExecutionTypeMarketAfternoonClose, want: kabuspb.StockOrderType_STOCK_ORDER_TYPE_MOAC},
		{name: "指値 を変換できる", arg1: ExecutionTypeLimit, want: kabuspb.StockOrderType_STOCK_ORDER_TYPE_LO},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.orderTypeTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_tradeTypeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 TradeType
		want kabuspb.TradeType
	}{
		{name: "未指定 を変換できる", arg1: TradeTypeUnspecified, want: kabuspb.TradeType_TRADE_TYPE_UNSPECIFIED},
		{name: "エントリー を変換できる", arg1: TradeTypeEntry, want: kabuspb.TradeType_TRADE_TYPE_ENTRY},
		{name: "エグジット を変換できる", arg1: TradeTypeExit, want: kabuspb.TradeType_TRADE_TYPE_EXIT},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.tradeTypeTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_marginTradeTypeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 MarginTradeType
		want kabuspb.MarginTradeType
	}{
		{name: "未指定 を変換できる", arg1: MarginTradeTypeUnspecified, want: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_UNSPECIFIED},
		{name: "制度 を変換できる", arg1: MarginTradeTypeSystem, want: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_SYSTEM},
		{name: "長期 を変換できる", arg1: MarginTradeTypeLong, want: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_LONG},
		{name: "デイトレ を変換できる", arg1: MarginTradeTypeDay, want: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.marginTradeTypeTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_closePositionsTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 []HoldPosition
		want []*kabuspb.ClosePosition
	}{
		{name: "nilならnilを返す", arg1: nil, want: nil},
		{name: "空配列なら空配列を返す", arg1: []HoldPosition{}, want: []*kabuspb.ClosePosition{}},
		{name: "要素があれば置き換えて返す", arg1: []HoldPosition{
			{PositionCode: "POSITION-CODE-001", HoldQuantity: 100},
			{PositionCode: "POSITION-CODE-002", HoldQuantity: 150},
			{PositionCode: "POSITION-CODE-003", HoldQuantity: 300},
		}, want: []*kabuspb.ClosePosition{
			{ExecutionId: "POSITION-CODE-001", Quantity: 100},
			{ExecutionId: "POSITION-CODE-002", Quantity: 150},
			{ExecutionId: "POSITION-CODE-003", Quantity: 300},
		}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabus := &kabusAPI{}
			got := kabus.closePositionsTo(test.arg1)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_kabusAPI_SendOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		kabusServiceClient         *testKabusServiceClient
		arg1                       *Strategy
		arg2                       *Order
		want1                      OrderResult
		want2                      error
		wantSendStockOrderHistory  []interface{}
		wantSendMarginOrderHistory []interface{}
	}{
		{name: "strategyがnilならnil error",
			kabusServiceClient: &testKabusServiceClient{},
			arg1:               nil,
			arg2:               &Order{},
			want2:              ErrNilArgument},
		{name: "orderがnilならnil error",
			kabusServiceClient: &testKabusServiceClient{},
			arg1:               &Strategy{},
			arg2:               nil,
			want2:              ErrNilArgument},
		{name: "現物注文でエラーが出たらエラーを返す",
			kabusServiceClient: &testKabusServiceClient{
				SendStockOrder1: nil,
				SendStockOrder2: ErrUnknown,
			},
			arg1: &Strategy{
				Code:            "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductStock,
				MarginTradeType: MarginTradeTypeUnspecified,
				Account:         Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:  "strategy-1475",
				SymbolCode:    "1475",
				Exchange:      ExchangeToushou,
				Product:       ProductStock,
				ExecutionType: ExecutionTypeMarket,
				Side:          SideBuy,
				TradeType:     TradeTypeEntry,
				OrderQuantity: 5,
				AccountType:   AccountTypeSpecific,
			},
			want2: ErrUnknown,
			wantSendStockOrderHistory: []interface{}{&kabuspb.SendStockOrderRequest{
				Password:     "Password1234",
				SymbolCode:   "1475",
				Exchange:     kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:         kabuspb.Side_SIDE_BUY,
				DeliveryType: kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				FundType:     kabuspb.FundType_FUND_TYPE_SUBSTITUTE_MARGIN,
				AccountType:  kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:     5,
				OrderType:    kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
		{name: "現物注文で注文に成功したら注文結果を返す",
			kabusServiceClient: &testKabusServiceClient{
				SendStockOrder1: &kabuspb.OrderResponse{ResultCode: 0, OrderId: "ORDER-ID-001"},
			},
			arg1: &Strategy{
				Code:       "strategy-1475",
				SymbolCode: "1475",
				Exchange:   ExchangeToushou,
				Product:    ProductStock,
				Account:    Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:  "strategy-1475",
				SymbolCode:    "1475",
				Exchange:      ExchangeToushou,
				Product:       ProductStock,
				ExecutionType: ExecutionTypeMarket,
				Side:          SideBuy,
				TradeType:     TradeTypeEntry,
				OrderQuantity: 5,
				AccountType:   AccountTypeSpecific,
			},
			want1: OrderResult{
				Result:     true,
				ResultCode: 0,
				OrderCode:  "ORDER-ID-001",
			},
			wantSendStockOrderHistory: []interface{}{&kabuspb.SendStockOrderRequest{
				Password:     "Password1234",
				SymbolCode:   "1475",
				Exchange:     kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:         kabuspb.Side_SIDE_BUY,
				DeliveryType: kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				FundType:     kabuspb.FundType_FUND_TYPE_SUBSTITUTE_MARGIN,
				AccountType:  kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:     5,
				OrderType:    kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
		{name: "現物注文で注文に失敗したら注文結果を返す",
			kabusServiceClient: &testKabusServiceClient{
				SendStockOrder1: &kabuspb.OrderResponse{ResultCode: 4, OrderId: ""},
			},
			arg1: &Strategy{
				Code:       "strategy-1475",
				SymbolCode: "1475",
				Exchange:   ExchangeToushou,
				Product:    ProductStock,
				Account:    Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:  "strategy-1475",
				SymbolCode:    "1475",
				Exchange:      ExchangeToushou,
				Product:       ProductStock,
				ExecutionType: ExecutionTypeMarket,
				Side:          SideBuy,
				TradeType:     TradeTypeEntry,
				OrderQuantity: 5,
				AccountType:   AccountTypeSpecific,
			},
			want1: OrderResult{
				Result:     false,
				ResultCode: 4,
				OrderCode:  "",
			},
			wantSendStockOrderHistory: []interface{}{&kabuspb.SendStockOrderRequest{
				Password:     "Password1234",
				SymbolCode:   "1475",
				Exchange:     kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:         kabuspb.Side_SIDE_BUY,
				DeliveryType: kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				FundType:     kabuspb.FundType_FUND_TYPE_SUBSTITUTE_MARGIN,
				AccountType:  kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:     5,
				OrderType:    kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
		{name: "信用注文でエラーが出たらエラーを返す",
			kabusServiceClient: &testKabusServiceClient{
				SendMarginOrder1: nil,
				SendMarginOrder2: ErrUnknown,
			},
			arg1: &Strategy{
				Code:            "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				Account:         Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:    "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				ExecutionType:   ExecutionTypeMarket,
				Side:            SideBuy,
				TradeType:       TradeTypeEntry,
				OrderQuantity:   5,
				AccountType:     AccountTypeSpecific,
			},
			want2: ErrUnknown,
			wantSendMarginOrderHistory: []interface{}{&kabuspb.SendMarginOrderRequest{
				Password:        "Password1234",
				SymbolCode:      "1475",
				Exchange:        kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:            kabuspb.Side_SIDE_BUY,
				TradeType:       kabuspb.TradeType_TRADE_TYPE_ENTRY,
				MarginTradeType: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY,
				DeliveryType:    kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				AccountType:     kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:        5,
				OrderType:       kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
		{name: "現物注文で注文に成功したら注文結果を返す",
			kabusServiceClient: &testKabusServiceClient{
				SendMarginOrder1: &kabuspb.OrderResponse{ResultCode: 0, OrderId: "ORDER-ID-001"},
			},
			arg1: &Strategy{
				Code:            "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				Account:         Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:    "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				ExecutionType:   ExecutionTypeMarket,
				Side:            SideBuy,
				TradeType:       TradeTypeEntry,
				OrderQuantity:   5,
				AccountType:     AccountTypeSpecific,
			},
			want1: OrderResult{
				Result:     true,
				ResultCode: 0,
				OrderCode:  "ORDER-ID-001",
			},
			wantSendMarginOrderHistory: []interface{}{&kabuspb.SendMarginOrderRequest{
				Password:        "Password1234",
				SymbolCode:      "1475",
				Exchange:        kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:            kabuspb.Side_SIDE_BUY,
				TradeType:       kabuspb.TradeType_TRADE_TYPE_ENTRY,
				MarginTradeType: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY,
				DeliveryType:    kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				AccountType:     kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:        5,
				OrderType:       kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
		{name: "現物注文で注文に失敗したら注文結果を返す",
			kabusServiceClient: &testKabusServiceClient{
				SendMarginOrder1: &kabuspb.OrderResponse{ResultCode: 4, OrderId: ""},
			},
			arg1: &Strategy{
				Code:            "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				Account:         Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			},
			arg2: &Order{
				StrategyCode:    "strategy-1475",
				SymbolCode:      "1475",
				Exchange:        ExchangeToushou,
				Product:         ProductMargin,
				MarginTradeType: MarginTradeTypeDay,
				ExecutionType:   ExecutionTypeMarket,
				Side:            SideBuy,
				TradeType:       TradeTypeEntry,
				OrderQuantity:   5,
				AccountType:     AccountTypeSpecific,
			},
			want1: OrderResult{
				Result:     false,
				ResultCode: 4,
				OrderCode:  "",
			},
			wantSendMarginOrderHistory: []interface{}{&kabuspb.SendMarginOrderRequest{
				Password:        "Password1234",
				SymbolCode:      "1475",
				Exchange:        kabuspb.StockExchange_STOCK_EXCHANGE_TOUSHOU,
				Side:            kabuspb.Side_SIDE_BUY,
				TradeType:       kabuspb.TradeType_TRADE_TYPE_ENTRY,
				MarginTradeType: kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY,
				DeliveryType:    kabuspb.DeliveryType_DELIVERY_TYPE_CASH,
				AccountType:     kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC,
				Quantity:        5,
				OrderType:       kabuspb.StockOrderType_STOCK_ORDER_TYPE_MO,
			}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			api := &kabusAPI{kabucom: test.kabusServiceClient}
			got1, got2 := api.SendOrder(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantSendStockOrderHistory, test.kabusServiceClient.SendStockOrderHistory) ||
				!reflect.DeepEqual(test.wantSendMarginOrderHistory, test.kabusServiceClient.SendMarginOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantSendStockOrderHistory, test.wantSendMarginOrderHistory,
					got1, got2, test.kabusServiceClient.SendStockOrderHistory, test.kabusServiceClient.SendMarginOrderHistory)
			}
		})
	}
}

func Test_newKabusAPI(t *testing.T) {
	t.Parallel()
	kabucom := &testKabusServiceClient{}
	want1 := &kabusAPI{kabucom: kabucom}
	got1 := newKabusAPI(kabucom)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_kabusAPI_priceRangeGroupFrom(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  string
		want1 TickGroup
	}{
		{name: "未指定 を変換できる", arg1: "", want1: TickGroupUnspecified},
		{name: "その他テーブル を変換できる", arg1: "10000", want1: TickGroupOther},
		{name: "TOPIX100テーブル を変換できる", arg1: "10003", want1: TickGroupTopix100},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusAPI := &kabusAPI{}
			got1 := kabusAPI.priceRangeGroupFrom(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_kabusapi_CancelOrder_Execute(t *testing.T) {
	t.Skip("実際にAPIを叩くテストのため、通常はスキップ")
	t.Parallel()

	conn, err := grpc.DialContext(context.Background(), "localhost:18082", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	kabusAPI := &kabusAPI{kabucom: kabuspb.NewKabusServiceClient(conn)}
	res, err := kabusAPI.CancelOrder("Password1234", "20220111A02N89218703")
	if err != nil {
		if st, ok := status.FromError(err); ok { // grpcのエラーならハンドリング処理に入る
			t.Log("grpc error")
			for _, d := range st.Details() {
				t.Log(d)
				switch e := d.(type) {
				case *kabuspb.RequestError:
					t.Log("type is *kabuspb.RequestError", e.Code, e.Message)
				default:
					t.Log("type is default")
				}
			}
		} else {
			t.Fatal(err)
		}
	}
	t.Log(res)
}

func Test_kabusAPI_GetFourPrice_Execute(t *testing.T) {
	t.Skip("実際にAPIを叩くテストのため、通常はスキップ")
	t.Parallel()

	conn, err := grpc.DialContext(context.Background(), "localhost:18082", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	kabusAPI := &kabusAPI{kabucom: kabuspb.NewKabusServiceClient(conn)}
	symbol, err := kabusAPI.GetFourPrice("1699", ExchangeToushou)
	t.Logf("symbol: %+v, err: %+v\n", symbol, err)
}

func Test_kabusAPI_GetFourPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kabucom *testKabusServiceClient
		arg1    string
		arg2    Exchange
		want1   *FourPrice
		want2   error
	}{
		{name: "板情報の取得がエラーになったらエラーを返す",
			kabucom: &testKabusServiceClient{GetBoard2: ErrUnknown},
			arg1:    "1475",
			arg2:    ExchangeToushou,
			want1:   nil,
			want2:   ErrUnknown},
		{name: "板情報をFourPriceに詰めて返す",
			kabucom: &testKabusServiceClient{GetBoard1: &kabuspb.Board{
				SymbolCode:       "1475",
				Exchange:         kabuspb.Exchange_EXCHANGE_TOUSHOU,
				CurrentPrice:     238,
				CurrentPriceTime: timestamppb.New(time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local)),
				OpeningPrice:     241.9,
				HighPrice:        242.4,
				LowPrice:         238,
			}},
			arg1: "1475",
			arg2: ExchangeToushou,
			want1: &FourPrice{
				SymbolCode: "1475",
				Exchange:   ExchangeToushou,
				DateTime:   time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local),
				Open:       241.9,
				High:       242.4,
				Low:        238,
				Close:      238,
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kabusapi := &kabusAPI{kabucom: test.kabucom}
			got1, got2 := kabusapi.GetFourPrice(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}
