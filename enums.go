package gridon

import "math"

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
	ExecutionTypeUnspecified          ExecutionType = ""                       // 未指定
	ExecutionTypeMarket               ExecutionType = "market"                 // 成行
	ExecutionTypeMarketMorningClose   ExecutionType = "market_morning_close"   // 前場引成
	ExecutionTypeMarketAfternoonClose ExecutionType = "market_afternoon_close" // 後場引成
	ExecutionTypeLimit                ExecutionType = "limit"                  // 指値
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

// GridType - グリッド戦略種別
type GridType string

const (
	GridTypeUnspecified   GridType = ""
	GridTypeStatic        GridType = "static"  // 静的、動的な変更なし
	GridTypeDynamicMinMax GridType = "min_max" // 最小・最大約定値からの動的グリッド
)

// Rounding - 端数処理
type Rounding string

const (
	RoundingUnspecified Rounding = ""      // 未指定, そのまま
	RoundingFloor       Rounding = "floor" // 切り捨て
	RoundingRound       Rounding = "round" // 四捨五入
	RoundingCeil        Rounding = "ceil"  // 切り上げ
)

// Calc - 端数処理を計算した結果を返す
func (e Rounding) Calc(v float64) float64 {
	switch e {
	case RoundingFloor:
		return math.Floor(v)
	case RoundingRound:
		return math.Round(v)
	case RoundingCeil:
		return math.Ceil(v)
	default:
		return v
	}
}

// Operation - 演算子
type Operation string

const (
	OperationUnspecified Operation = ""  // 未指定
	OperationPlus        Operation = "+" // 加算
	OperationMinus       Operation = "-" // 減算
	OperationMultiple    Operation = "*" // 積算
	OperationDived       Operation = "/" // 除算
)

// Calc - 演算子の計算結果を返す
func (e Operation) Calc(a, b float64) float64 {
	switch e {
	case OperationPlus:
		return a + b
	case OperationMinus:
		return a - b
	case OperationMultiple:
		return a * b
	case OperationDived:
		return a / b
	default:
		return a + b
	}
}
