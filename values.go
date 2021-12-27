package gridon

import (
	"time"
)

// Symbol - 銘柄情報
type Symbol struct {
	Code                 string    // 銘柄コード
	Exchange             Exchange  // 市場
	TradingUnit          float64   // 売買単位
	CurrentPrice         float64   // 現在値
	CurrentPriceDateTime time.Time // 現在値日時
	BidPrice             float64   // 最良買い気配値
	AskPrice             float64   // 最良売り気配値
	TickGroup            TickGroup // 呼値グループ
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
	Price            float64 // 拘束したポジションの約定値 = エントリー約定値
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
	Runnable          bool              // 実行可能かどうか
	Width             int               // グリッド幅(tick数)
	Quantity          float64           // 1グリッドに乗せる数量
	NumberOfGrids     int               // 指値注文を入れておくグリッドの本数
	TimeRanges        []TimeRange       // 戦略動作時刻範囲
	GridType          GridType          // グリッド戦略種別
	DynamicGridMinMax DynamicGridMinMax // 最小・最大約定値からの動的グリッド
}

// IsRunnable - グリッド戦略が実行可能かどうか
func (v *GridStrategy) IsRunnable(now time.Time) bool {
	if !v.Runnable {
		return false
	}

	for _, tr := range v.TimeRanges {
		if tr.In(now) {
			return true
		}
	}

	return false
}

// DynamicGridMinMax - 最小・最大約定値からの動的グリッド
type DynamicGridMinMax struct {
	Rate      float64   // 最低・最大の差をどのくらい基準幅に加算するか。1ならそのまま、0.2なら差の1/5を加算
	Rounding  Rounding  // 端数処理
	Operation Operation // 演算子
}

// Width - 計算後グリッド幅
func (v *DynamicGridMinMax) Width(width, diff float64) float64 {
	return v.Operation.Calc(width, v.Rounding.Calc(diff*v.Rate))
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
	Runnable   bool            // 実行可能かどうか
	Conditions []ExitCondition // 全エグジット戦略の一覧
}

// ExitCondition - 全エグジット戦略の詳細設定
type ExitCondition struct {
	ExecutionType ExecutionType // 執行条件
	Timing        time.Time     // タイミング(時分)
}

// IsRunnable - 全エグジット戦略が実行可能かどうか
func (v *ExitStrategy) IsRunnable(now time.Time) bool {
	if !v.Runnable {
		return false
	}

	for _, ec := range v.Conditions {
		if now.Hour() == ec.Timing.Hour() && now.Minute() == ec.Timing.Minute() {
			return true
		}
	}
	return false
}

// ExecutionType - 指定時刻に行なうエグジット注文の執行条件
func (v *ExitStrategy) ExecutionType(now time.Time) ExecutionType {
	if !v.Runnable {
		return ExecutionTypeUnspecified
	}

	for _, ec := range v.Conditions {
		if now.Hour() == ec.Timing.Hour() && now.Minute() == ec.Timing.Minute() {
			return ec.ExecutionType
		}
	}
	return ExecutionTypeUnspecified
}

// CancelStrategy - 全取消戦略
type CancelStrategy struct {
	Runnable bool        // 実行可能かどうか
	Timings  []time.Time // タイミング(時分)の一覧
}

// IsRunnable - 全エグジット戦略が実行可能かどうか
func (v *CancelStrategy) IsRunnable(now time.Time) bool {
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

// TimeRange - 時間の範囲
type TimeRange struct {
	Start time.Time // 開始時刻
	End   time.Time // 終了時刻
}

// In - 引数の時刻が範囲内かどうか
func (v *TimeRange) In(target time.Time) bool {
	if target.IsZero() {
		return false
	}

	start := time.Date(0, 1, 1, v.Start.Hour(), v.Start.Minute(), v.Start.Second(), v.Start.Nanosecond(), v.Start.Location())
	end := time.Date(0, 1, 1, v.End.Hour(), v.End.Minute(), v.End.Second(), v.End.Nanosecond(), v.End.Location())
	t := time.Date(0, 1, 1, target.Hour(), target.Minute(), target.Second(), target.Nanosecond(), target.Location())

	if start.Before(end) {
		return !t.Before(start) && t.Before(end)
	} else {
		return !t.Before(start) || t.Before(end)
	}
}
