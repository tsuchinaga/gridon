package gridon

import (
	"encoding/json"
	"time"
)

// Strategy - 戦略
type Strategy struct {
	Code                 string            // 戦略コード
	SymbolCode           string            // 銘柄コード
	Exchange             Exchange          // 市場
	Product              Product           // 商品種別
	MarginTradeType      MarginTradeType   // 信用取引区分
	EntrySide            Side              // エントリー方向
	Cash                 float64           // 運用中現金
	BasePrice            float64           // 基準価格
	BasePriceDateTime    time.Time         // 基準価格日時
	LastContractPrice    float64           // 最終約定価格
	LastContractDateTime time.Time         // 最終約定価格日時
	MaxContractPrice     float64           // 最大約定価格
	MaxContractDateTime  time.Time         // 最大約定価格日時
	MinContractPrice     float64           // 最小約定価格
	MinContractDateTime  time.Time         // 最小約定価格日時
	TickGroup            TickGroup         // 呼値グループ
	TradingUnit          float64           // 売買単位
	RebalanceStrategy    RebalanceStrategy // リバランス戦略
	GridStrategy         GridStrategy      // グリッド戦略
	CancelStrategy       CancelStrategy    // 全取消戦略
	ExitStrategy         ExitStrategy      // 全エグジット戦略
	Account              Account           // 口座情報
	Runnable             bool              // 実行可能かどうか
}

func (e *Strategy) String() string {
	if b, err := json.Marshal(e); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

func (e *Strategy) IsRunnable() bool {
	return e.Runnable
}

// Order - 注文
type Order struct {
	Code             string          // 注文コード
	StrategyCode     string          // 戦略コード
	SymbolCode       string          // 銘柄コード
	Exchange         Exchange        // 市場
	Status           OrderStatus     // 注文状態
	Product          Product         // 商品種別
	MarginTradeType  MarginTradeType // 信用取引区分
	TradeType        TradeType       // 取引種別
	Side             Side            // 方向
	ExecutionType    ExecutionType   // 執行条件
	Price            float64         // 指値価格
	OrderQuantity    float64         // 注文数量
	ContractQuantity float64         // 約定数量
	AccountType      AccountType     // 口座種別
	OrderDateTime    time.Time       // 注文日時
	ContractDateTime time.Time       // 約定日時
	CancelDateTime   time.Time       // 取消日時
	Contracts        []Contract      // 約定
	HoldPositions    []HoldPosition  // エグジットのために拘束ポジション
}

func (e *Order) String() string {
	if b, err := json.Marshal(e); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

// IsActive - 有効な注文か (更新される可能性のある注文)
func (e *Order) IsActive() bool {
	return e.Status == OrderStatusInOrder
}

// IsEqualSecurityOrder - 証券会社の注文と一致しているか
func (e *Order) IsEqualSecurityOrder(securityOrder SecurityOrder) bool {
	return e.Status == securityOrder.Status &&
		e.ContractQuantity == securityOrder.ContractQuantity &&
		e.ContractDateTime.Equal(securityOrder.ContractDateTime) &&
		e.CancelDateTime.Equal(securityOrder.CancelDateTime)
}

// ReflectSecurityOrder - 証券会社の注文を反映する
func (e *Order) ReflectSecurityOrder(securityOrder SecurityOrder) {
	e.Status = securityOrder.Status
	e.ContractQuantity = securityOrder.ContractQuantity
	e.OrderDateTime = securityOrder.OrderDateTime
	e.ContractDateTime = securityOrder.ContractDateTime
	e.CancelDateTime = securityOrder.CancelDateTime
	e.Contracts = securityOrder.Contracts
}

// HasContract - 注文が引数の約定を持っているかどうか
func (e *Order) HasContract(contract Contract) bool {
	for _, c := range e.Contracts {
		if c.PositionCode == contract.PositionCode {
			return true
		}
	}
	return false
}

// ContractDiff - 約定情報の差を返す (SecurityOrderにあってOrderにない約定)
func (e *Order) ContractDiff(securityOrder SecurityOrder) []Contract {
	contracts := make([]Contract, 0)
	for _, c := range securityOrder.Contracts {
		if !e.HasContract(c) {
			contracts = append(contracts, c)
		}
	}

	return contracts
}

// Position - ポジション
type Position struct {
	Code             string          // ポジションコード
	StrategyCode     string          // 戦略コード
	OrderCode        string          // 注文コード
	SymbolCode       string          // 銘柄コード
	Exchange         Exchange        // 市場
	Side             Side            // 売買方向
	Product          Product         // 商品種別
	MarginTradeType  MarginTradeType // 信用取引区分
	Price            float64         // 約定値
	OwnedQuantity    float64         // 保有数量
	HoldQuantity     float64         // 拘束数両
	ContractDateTime time.Time       // 約定日時
}

func (e *Position) String() string {
	if b, err := json.Marshal(e); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

// IsActive - 有効なポジションか (更新される可能性のあるポジション)
func (e *Position) IsActive() bool {
	return e.OwnedQuantity > 0
}

// LeaveQuantity - 残量
func (e *Position) LeaveQuantity() float64 {
	return e.OwnedQuantity - e.HoldQuantity
}

// FourPrice - 四本値
type FourPrice struct {
	SymbolCode string    // 銘柄コード
	Exchange   Exchange  // 市場
	DateTime   time.Time // 日付
	Open       float64   // 始値
	High       float64   // 高値
	Low        float64   // 安値
	Close      float64   // 終値
}
