package gridon

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"gitlab.com/tsuchinaga/kabus-grpc-server/kabuspb"
)

// IKabusAPI - kabuステーションAPIのインターフェース
type IKabusAPI interface {
	GetSymbol(symbolCode string, exchange Exchange) (*Symbol, error)
	GetOrders(product Product, updateDateTime time.Time) ([]SecurityOrder, error)
	CancelOrder(orderPassword string, orderCode string) (OrderResult, error)
}

// kabusAPI - kabuステーションAPI
type kabusAPI struct {
	kabucom kabuspb.KabusServiceClient
}

func (k *kabusAPI) exchangeTo(exchange Exchange) kabuspb.Exchange {
	switch exchange {
	case ExchangeToushou:
		return kabuspb.Exchange_EXCHANGE_TOUSHOU
	case ExchangeMeishou:
		return kabuspb.Exchange_EXCHANGE_MEISHOU
	case ExchangeFukushou:
		return kabuspb.Exchange_EXCHANGE_FUKUSHOU
	case ExchangeSatsushou:
		return kabuspb.Exchange_EXCHANGE_SATSUSHOU
	}
	return kabuspb.Exchange_EXCHANGE_UNSPECIFIED
}

func (k *kabusAPI) exchangeFrom(exchange kabuspb.Exchange) Exchange {
	switch exchange {
	case kabuspb.Exchange_EXCHANGE_TOUSHOU:
		return ExchangeToushou
	case kabuspb.Exchange_EXCHANGE_MEISHOU:
		return ExchangeMeishou
	case kabuspb.Exchange_EXCHANGE_FUKUSHOU:
		return ExchangeFukushou
	case kabuspb.Exchange_EXCHANGE_SATSUSHOU:
		return ExchangeSatsushou
	case kabuspb.Exchange(9): // OrderExchange経由で入ってくることがある
		return ExchangeSOR
	}
	return ExchangeUnspecified
}

func (k *kabusAPI) productTo(product Product) kabuspb.Product {
	switch product {
	case ProductStock:
		return kabuspb.Product_PRODUCT_STOCK
	case ProductMargin:
		return kabuspb.Product_PRODUCT_MARGIN
	}
	return kabuspb.Product_PRODUCT_UNSPECIFIED
}

func (k *kabusAPI) productFrom(product kabuspb.Product) Product {
	switch product {
	case kabuspb.Product_PRODUCT_STOCK:
		return ProductStock
	case kabuspb.Product_PRODUCT_MARGIN:
		return ProductMargin
	}
	return ProductUnspecified
}

func (k *kabusAPI) orderStatusFrom(recordType kabuspb.RecordType, orderQuantity float64, contractQuantity float64) OrderStatus {
	switch recordType {
	case kabuspb.RecordType_RECORD_TYPE_RECEIVE, kabuspb.RecordType_RECORD_TYPE_CARRIED, kabuspb.RecordType_RECORD_TYPE_ORDERED, kabuspb.RecordType_RECORD_TYPE_MODIFIED:
		return OrderStatusInOrder
	case kabuspb.RecordType_RECORD_TYPE_EXPIRED, kabuspb.RecordType_RECORD_TYPE_CANCELED, kabuspb.RecordType_RECORD_TYPE_REVOCATION:
		return OrderStatusCanceled
	case kabuspb.RecordType_RECORD_TYPE_CONTRACTED:
		if orderQuantity == contractQuantity {
			return OrderStatusDone
		} else {
			return OrderStatusInOrder
		}
	}
	return OrderStatusUnspecified
}

func (k *kabusAPI) marginTradeTypeFrom(marginTradeType kabuspb.MarginTradeType) MarginTradeType {
	switch marginTradeType {
	case kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_SYSTEM:
		return MarginTradeTypeSystem
	case kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_LONG:
		return MarginTradeTypeLong
	case kabuspb.MarginTradeType_MARGIN_TRADE_TYPE_GENERAL_DAY:
		return MarginTradeTypeDay
	}
	return MarginTradeTypeUnspecified
}

func (k *kabusAPI) tradeTypeFrom(product kabuspb.Product, side kabuspb.Side, tradeType kabuspb.TradeType) TradeType {
	switch product {
	case kabuspb.Product_PRODUCT_STOCK:
		switch side {
		case kabuspb.Side_SIDE_BUY:
			return TradeTypeEntry
		case kabuspb.Side_SIDE_SELL:
			return TradeTypeExit
		}
	case kabuspb.Product_PRODUCT_MARGIN:
		switch tradeType {
		case kabuspb.TradeType_TRADE_TYPE_ENTRY:
			return TradeTypeEntry
		case kabuspb.TradeType_TRADE_TYPE_EXIT:
			return TradeTypeExit
		}
	}
	return TradeTypeUnspecified
}

func (k *kabusAPI) sideFrom(side kabuspb.Side) Side {
	switch side {
	case kabuspb.Side_SIDE_BUY:
		return SideBuy
	case kabuspb.Side_SIDE_SELL:
		return SideSell
	}
	return SideUnspecified
}

func (k *kabusAPI) accountTypeFrom(accountType kabuspb.AccountType) AccountType {
	switch accountType {
	case kabuspb.AccountType_ACCOUNT_TYPE_GENERAL:
		return AccountTypeGeneral
	case kabuspb.AccountType_ACCOUNT_TYPE_SPECIFIC:
		return AccountTypeSpecific
	case kabuspb.AccountType_ACCOUNT_TYPE_CORPORATION:
		return AccountTypeCorporation
	}
	return AccountTypeUnspecified
}

func (k *kabusAPI) contractFrom(orderCode string, detail *kabuspb.OrderDetail) Contract {
	return Contract{
		OrderCode:        orderCode,
		PositionCode:     detail.ExecutionId,
		Price:            detail.Price,
		Quantity:         detail.Quantity,
		ContractDateTime: detail.ExecutionDay.AsTime().In(time.Local),
	}
}

func (k *kabusAPI) securityOrderFrom(product kabuspb.Product, order *kabuspb.Order) SecurityOrder {
	var lastRecordType kabuspb.RecordType
	var contractDateTime, cancelDateTime time.Time
	contracts := make([]Contract, 0)
	for _, d := range order.Details {
		// 処理済みの詳細でなければ無視する
		if d.State != kabuspb.OrderDetailState_ORDER_DETAIL_STATE_PROCESSED {
			continue
		}
		lastRecordType = d.RecordType // 最終レコード種別

		switch d.RecordType {
		case kabuspb.RecordType_RECORD_TYPE_EXPIRED, kabuspb.RecordType_RECORD_TYPE_CANCELED, kabuspb.RecordType_RECORD_TYPE_REVOCATION:
			cancelDateTime = d.TransactTime.AsTime().In(time.Local)
		case kabuspb.RecordType_RECORD_TYPE_CONTRACTED:
			contractDateTime = d.ExecutionDay.AsTime().In(time.Local)
			contracts = append(contracts, k.contractFrom(order.Id, d))
		}
	}

	return SecurityOrder{
		Code:             order.Id,
		Status:           k.orderStatusFrom(lastRecordType, order.OrderQuantity, order.CumulativeQuantity),
		SymbolCode:       order.SymbolCode,
		Exchange:         k.exchangeFrom(kabuspb.Exchange(order.Exchange)),
		Product:          k.productFrom(product),
		MarginTradeType:  k.marginTradeTypeFrom(order.MarginTradeType),
		TradeType:        k.tradeTypeFrom(product, order.Side, order.TradeType),
		Side:             k.sideFrom(order.Side),
		Price:            order.Price,
		OrderQuantity:    order.OrderQuantity,
		ContractQuantity: order.CumulativeQuantity,
		AccountType:      k.accountTypeFrom(order.AccountType),
		ExpireDay:        order.ExpireDay.AsTime().In(time.Local),
		OrderDateTime:    order.ReceiveTime.AsTime().In(time.Local),
		ContractDateTime: contractDateTime,
		CancelDateTime:   cancelDateTime,
		Contracts:        contracts,
	}
}

// GetSymbol - 銘柄情報の取得
func (k *kabusAPI) GetSymbol(symbolCode string, exchange Exchange) (*Symbol, error) {
	symbol, err := k.kabucom.GetSymbol(context.Background(), &kabuspb.GetSymbolRequest{SymbolCode: symbolCode, Exchange: k.exchangeTo(exchange)})
	if err != nil {
		return nil, err
	}
	board, err := k.kabucom.GetBoard(context.Background(), &kabuspb.GetBoardRequest{SymbolCode: symbolCode, Exchange: k.exchangeTo(exchange)})
	if err != nil {
		return nil, err
	}
	return &Symbol{
		Code:         symbol.Code,
		Exchange:     k.exchangeFrom(symbol.Exchange),
		TradingUnit:  symbol.TradingUnit,
		CurrentPrice: board.CurrentPrice,
		BidPrice:     board.BidPrice,
		AskPrice:     board.AskPrice,
	}, nil
}

// GetOrders - 注文一覧の取得
func (k *kabusAPI) GetOrders(product Product, updateDateTime time.Time) ([]SecurityOrder, error) {
	kabusProduct := k.productTo(product)
	res, err := k.kabucom.GetOrders(context.Background(), &kabuspb.GetOrdersRequest{
		Product:    kabusProduct,
		UpdateTime: timestamppb.New(updateDateTime),
		GetDetails: true,
	})
	if err != nil {
		return nil, err
	}

	result := make([]SecurityOrder, 0)
	for _, o := range res.Orders {
		result = append(result, k.securityOrderFrom(kabusProduct, o))
	}
	return result, nil
}

// CancelOrder - 注文の取消
func (k *kabusAPI) CancelOrder(orderPassword string, orderCode string) (OrderResult, error) {
	res, err := k.kabucom.CancelOrder(context.Background(), &kabuspb.CancelOrderRequest{Password: orderPassword, OrderId: orderCode})
	if err != nil {
		return OrderResult{}, err
	}
	return OrderResult{Result: res.ResultCode == 0, ResultCode: int(res.ResultCode)}, nil
}
