package gridon

import "math"

// newRebalanceService - 新しいリバランスサービスの取得
func newRebalanceService(clock IClock, kabusAPI IKabusAPI, positionStore IPositionStore, orderService IOrderService) IRebalanceService {
	return &rebalanceService{
		clock:         clock,
		kabusAPI:      kabusAPI,
		positionStore: positionStore,
		orderService:  orderService,
	}
}

// IRebalanceService - リバランスサービスのインターフェース
type IRebalanceService interface {
	Rebalance(strategy *Strategy) error
}

// rebalanceService - リバランスサービス
type rebalanceService struct {
	clock         IClock
	kabusAPI      IKabusAPI
	positionStore IPositionStore
	orderService  IOrderService
}

// Rebalance - リバランスの実行
func (s *rebalanceService) Rebalance(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}
	if !strategy.IsRunnable() {
		return nil
	}

	if !strategy.RebalanceStrategy.IsRunnable(s.clock.Now()) {
		return nil
	}

	symbol, err := s.kabusAPI.GetSymbol(strategy.SymbolCode, strategy.Exchange)
	if err != nil {
		return err
	}
	price := (symbol.AskPrice + symbol.BidPrice) / 2

	positionValue, err := s.positionValue(strategy.Code, price)
	if err != nil {
		return err
	}

	q := s.rebalanceQuantity(strategy.Cash, price, positionValue, symbol.TradingUnit)
	switch {
	case q < 0:
		if err := s.orderService.ExitMarket(strategy.Code, q*-1, SortOrderLatest); err != nil {
			return err
		}
	case q > 0:
		if err := s.orderService.EntryMarket(strategy.Code, q); err != nil {
			return err
		}
	}

	return nil
}

// positionValue - ポジションの評価額の計算
func (s *rebalanceService) positionValue(strategyCode string, price float64) (float64, error) {
	positions, err := s.positionStore.GetActivePositionsByStrategyCode(strategyCode)
	if err != nil {
		return 0, err
	}

	var val float64
	for _, p := range positions {
		switch p.Side {
		case SideBuy: // 買いポジション
			val += price * p.OwnedQuantity
		case SideSell: // 売りポジション
			val += (p.Price + p.Price - price) * p.OwnedQuantity
		}
	}
	return val, nil
}

// rebalanceQuantity - リバランス調整数量
// 負の値なら売り、正の値なら買い
func (s *rebalanceService) rebalanceQuantity(cash float64, price float64, positionValue float64, tradeUnit float64) float64 {
	// ゼロ除算はできないので、必須の情報がなければ判断できないため0枚を返す
	if price <= 0 || tradeUnit <= 0 {
		return 0
	}

	q := (cash - positionValue) / 2 / price / tradeUnit // 現金と評価額の差 / 2 で中央値との差異を出し、その差異の中でどれだけ売買できるかを計算する
	return math.Round(q) * tradeUnit
}
