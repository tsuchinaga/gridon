package gridon

// Exchange - 市場
type Exchange string

const (
	ExchangeUnspecified Exchange = ""          // 未指定
	ExchangeToushou     Exchange = "toushou"   // 東証
	ExchangeMeishou     Exchange = "meishou"   // 名証
	ExchangeFukushou    Exchange = "fukushou"  // 福証
	ExchangeSatsushou   Exchange = "satsushou" // 札証
	ExchangeSOR         Exchange = "SOR"       // SOR
)

// Side - 売買方向
type Side string

const (
	SideUnspecified Side = ""     // 未指定
	SideBuy         Side = "buy"  // 買い
	SideSell        Side = "sell" // 売り
)

// Turn - 反対方向を取得する
func (e Side) Turn() Side {
	switch e {
	case SideBuy:
		return SideSell
	case SideSell:
		return SideBuy
	}
	return SideUnspecified
}

// TradeType - 取引種別
type TradeType string

const (
	TradeTypeUnspecified TradeType = ""      // 未指定
	TradeTypeEntry       TradeType = "entry" // エントリー
	TradeTypeExit        TradeType = "exit"  // エグジット
)

// OrderStatus - 注文状態
type OrderStatus string

const (
	OrderStatusUnspecified OrderStatus = ""         // 未指定
	OrderStatusInOrder     OrderStatus = "in_order" // 注文中
	OrderStatusDone        OrderStatus = "done"     // 約定済み
	OrderStatusCanceled    OrderStatus = "canceled" // 取消済み
)

// AccountType - 口座種別
type AccountType string

const (
	AccountTypeUnspecified AccountType = ""            // 未指定
	AccountTypeGeneral     AccountType = "general"     // 一般
	AccountTypeSpecific    AccountType = "specific"    // 特定
	AccountTypeCorporation AccountType = "corporation" // 特定
)

// Product - 商品種別
type Product string

const (
	ProductUnspecified Product = ""       // 未指定
	ProductStock       Product = "stock"  // 現物
	ProductMargin      Product = "margin" // 信用
)

// MarginTradeType - 信用の取引区分
type MarginTradeType string

const (
	MarginTradeTypeUnspecified MarginTradeType = ""       // 未指定
	MarginTradeTypeSystem      MarginTradeType = "system" // 制度
	MarginTradeTypeLong        MarginTradeType = "long"   // 一般長期
	MarginTradeTypeDay         MarginTradeType = "day"    // 一般デイトレ
)

// ExecutionType - 執行条件
type ExecutionType string

const (
	ExecutionTypeUnspecified ExecutionType = ""       // 未指定
	ExecutionTypeMarket      ExecutionType = "market" // 成行
	ExecutionTypeLimit       ExecutionType = "limit"  // 指値
)

// SortOrder - 並び順
type SortOrder string

const (
	SortOrderUnspecified SortOrder = "unspecified"
	SortOrderNewest      SortOrder = "newest"
	SortOrderLatest      SortOrder = "latest"
)

// TickGroup - 呼値グループ
type TickGroup string

const (
	TickGroupUnspecified TickGroup = ""
	TickGroupTopix100    TickGroup = "topix100"
	TickGroupOther       TickGroup = "other"
)
