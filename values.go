package gridon

import "time"

// Symbol - 銘柄情報
type Symbol struct {
	Code         string   // 銘柄コード
	Exchange     Exchange // 市場
	TradingUnit  float64  // 売買単位
	CurrentPrice float64  // 現在値
	BidPrice     float64  // 最良買い気配値
	AskPrice     float64  // 最良売り気配値
}

// SecurityOrder - 証券会社の注文
type SecurityOrder struct {
	Code             string          // 注文コード
	Status           OrderStatus     // 注文状態
	SymbolCode       string          // 銘柄コード
	Exchange         Exchange        // 市場
	Product          Product         // 商品種別
	MarginTradeType  MarginTradeType // 信用取引区分
	TradeType        TradeType       // 取引種別
	Side             Side            // 方向
	Price            float64         // 指値価格
	OrderQuantity    float64         // 注文数量
	ContractQuantity float64         // 約定数量
	AccountType      AccountType     // 口座種別
	ExpireDay        time.Time       // 有効期限(年月日)
	OrderDateTime    time.Time       // 注文日時
	ContractDateTime time.Time       // 約定日時
	CancelDateTime   time.Time       // 取消日時
	Contracts        []Contract      // 約定
}

// Contract - 約定
type Contract struct {
	OrderCode        string    // 注文コード
	PositionCode     string    // ポジションコード (市場の約定のコードなので、同じポジションのエントリーとエグジットでもコードは一致しない)
	Price            float64   // 約定値
	Quantity         float64   // 約定数量
	ContractDateTime time.Time // 約定日時
}

// HoldPosition - 拘束ポジション
type HoldPosition struct {
	PositionCode     string  // ポジションコード
	HoldQuantity     float64 // 拘束した数量
	ContractQuantity float64 // 約定した数量
	ReleaseQuantity  float64 // 解放した数量
}

// LeaveQuantity - 残量
func (v *HoldPosition) LeaveQuantity() float64 {
	return v.HoldQuantity - v.ContractQuantity - v.ReleaseQuantity
}
