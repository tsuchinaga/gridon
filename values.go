package gridon

import (
	"time"
)

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

// OrderResult - 注文結果
type OrderResult struct {
	Result     bool   // 結果
	ResultCode int    // 結果コード
	OrderCode  string // 注文コード
}

// Account - 口座情報
type Account struct {
	Password    string      // 注文パスワード
	AccountType AccountType // 口座種別
}

// GridStrategy - グリッド戦略
type GridStrategy struct {
	Runnable      bool      // 実行可能かどうか
	Width         int       // グリッド幅(tick数)
	Quantity      float64   // 1グリッドに乗せる数量
	NumberOfGrids int       // 指値注文を入れておくグリッドの本数
	StartTime     time.Time // 戦略動作開始時刻
	EndTime       time.Time // 戦略動作終了時刻
}

// IsRunnable - グリッド戦略が実行可能かどうか
func (v *GridStrategy) IsRunnable(now time.Time) bool {
	if !v.Runnable {
		return false
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), v.StartTime.Hour(), v.StartTime.Minute(), v.StartTime.Second(), v.StartTime.Nanosecond(), now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), v.EndTime.Hour(), v.EndTime.Minute(), v.EndTime.Second(), v.EndTime.Nanosecond(), now.Location())
	return !now.Before(start) && now.Before(end)
}

// RebalanceStrategy - リバランス戦略
type RebalanceStrategy struct {
	Runnable bool        // 実行可能かどうか
	Timings  []time.Time // タイミング(時分)の一覧
}

// IsRunnable - グリッド戦略が実行可能かどうか
func (v *RebalanceStrategy) IsRunnable(now time.Time) bool {
	if !v.Runnable {
		return false
	}

	for _, t := range v.Timings {
		if now.Hour() == t.Hour() && now.Minute() == t.Minute() {
			return true
		}
	}
	return false
}

// ExitStrategy - 全エグジット戦略
type ExitStrategy struct {
	Runnable bool        // 実行可能かどうか
	Timings  []time.Time // タイミング(時分)の一覧
}

// IsRunnable - 全エグジット戦略が実行可能かどうか
func (v *ExitStrategy) IsRunnable(now time.Time) bool {
	if !v.Runnable {
		return false
	}

	for _, t := range v.Timings {
		if now.Hour() == t.Hour() && now.Minute() == t.Minute() {
			return true
		}
	}
	return false
}
