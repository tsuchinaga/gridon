package gridon

import (
	"fmt"
	"math"
	"sort"
)

// newOrderService - 新しい注文サービスの取得
func newOrderService(clock IClock, kabusAPI IKabusAPI, strategyStore IStrategyStore, orderStore IOrderStore, positionStore IPositionStore) IOrderService {
	return &orderService{
		clock:         clock,
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
	}
}

// IOrderService - 注文サービスのインターフェース
type IOrderService interface {
	GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error)
	EntryLimit(strategyCode string, price float64, quantity float64) error
	ExitLimit(strategyCode string, price float64, quantity float64, sortOrder SortOrder) error
	EntryMarket(strategyCode string, quantity float64) error
	ExitMarket(strategyCode string, quantity float64, sortOrder SortOrder) error
	Cancel(strategy *Strategy, orderCode string) error
	CancelAll(strategy *Strategy) error
	ExitAll(strategy *Strategy) error
}

// orderService - 注文サービス
type orderService struct {
	clock         IClock
	kabusAPI      IKabusAPI
	strategyStore IStrategyStore
	orderStore    IOrderStore
	positionStore IPositionStore
}

// GetActiveOrdersByStrategyCode - 戦略を指定して有効な注文を取り出す
func (s *orderService) GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error) {
	return s.orderStore.GetActiveOrdersByStrategyCode(strategyCode)
}

// EntryLimit - エントリーの指値注文
func (s *orderService) EntryLimit(strategyCode string, price float64, quantity float64) error {
	strategy, err := s.strategyStore.GetByCode(strategyCode)
	if err != nil {
		return err
	}

	check, err := s.checkEntryCash(strategyCode, strategy.Cash, price, quantity)
	if err != nil {
		return err
	}
	if !check {
		return ErrNotEnoughCash
	}

	order := &Order{
		StrategyCode:    strategy.Code,
		SymbolCode:      strategy.SymbolCode,
		Exchange:        strategy.Exchange,
		Status:          OrderStatusInOrder,
		Product:         strategy.Product,
		MarginTradeType: strategy.MarginTradeType,
		TradeType:       TradeTypeEntry,
		Side:            strategy.EntrySide,
		ExecutionType:   ExecutionTypeLimit,
		Price:           price,
		OrderQuantity:   quantity,
		AccountType:     strategy.Account.AccountType,
		OrderDateTime:   s.clock.Now(),
	}

	return s.sendOrder(strategy, order)
}

// ExitLimit - エグジットの指値注文
func (s *orderService) ExitLimit(strategyCode string, price float64, quantity float64, sortOrder SortOrder) error {
	strategy, err := s.strategyStore.GetByCode(strategyCode)
	if err != nil {
		return err
	}

	order := &Order{
		StrategyCode:    strategy.Code,
		SymbolCode:      strategy.SymbolCode,
		Exchange:        strategy.Exchange,
		Status:          OrderStatusInOrder,
		Product:         strategy.Product,
		MarginTradeType: strategy.MarginTradeType,
		TradeType:       TradeTypeExit,
		Side:            strategy.EntrySide.Turn(),
		ExecutionType:   ExecutionTypeLimit,
		Price:           price,
		OrderQuantity:   quantity,
		AccountType:     strategy.Account.AccountType,
		OrderDateTime:   s.clock.Now(),
	}

	hp, err := s.holdPositions(strategyCode, quantity, sortOrder)
	if err != nil {
		return err
	}
	order.HoldPositions = hp

	return s.sendOrder(strategy, order)
}

// Cancel - 指定した注文を取り消す
func (s *orderService) Cancel(strategy *Strategy, orderCode string) error {
	if strategy == nil {
		return ErrNilArgument
	}

	res, err := s.kabusAPI.CancelOrder(strategy.Account.Password, orderCode)
	if err != nil {
		return err
	}
	if !res.Result {
		return fmt.Errorf("result=%+v, orderCode=%+v: %w", res, orderCode, ErrCancelCondition)
	}

	return nil
}

// CancelAll - 戦略に関連する全ての注文を取り消す
func (s *orderService) CancelAll(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	if !strategy.ExitStrategy.IsRunnable(s.clock.Now()) {
		return nil
	}

	// 有効な注文を取り出す
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}

	// キャンセルに流す
	for _, o := range orders {
		_, err := s.kabusAPI.CancelOrder(strategy.Account.Password, o.Code)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExitAll - 戦略に関連する拘束されていないポジションを全てエグジットする
func (s *orderService) ExitAll(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	if !strategy.ExitStrategy.IsRunnable(s.clock.Now()) {
		return nil
	}

	// 保有中のポジションを取り出す
	positions, err := s.positionStore.GetActivePositionsByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}

	// 返済すべきポジションがなければ何もしない
	if len(positions) <= 0 {
		return nil
	}

	// エグジット注文を流す
	order := &Order{
		StrategyCode:    strategy.Code,
		SymbolCode:      strategy.SymbolCode,
		Exchange:        strategy.Exchange,
		Status:          OrderStatusInOrder,
		Product:         strategy.Product,
		MarginTradeType: strategy.MarginTradeType,
		TradeType:       TradeTypeExit,
		Side:            strategy.EntrySide.Turn(),
		ExecutionType:   ExecutionTypeMarket,
		Price:           0,
		OrderQuantity:   0,
		AccountType:     strategy.Account.AccountType,
		OrderDateTime:   s.clock.Now(),
		HoldPositions:   []HoldPosition{},
	}
	for _, p := range positions {
		leave := p.LeaveQuantity()
		order.OrderQuantity += leave
		if err := s.positionStore.Hold(p.Code, leave); err != nil {
			// 拘束したポジションを解放する
			// ただし、解放の処理でエラーでたら対応できない
			for _, hp := range order.HoldPositions {
				_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
			}
			return err
		}
		order.HoldPositions = append(order.HoldPositions, HoldPosition{PositionCode: p.Code, HoldQuantity: leave})
	}

	// 注文の送信
	res, err := s.kabusAPI.SendOrder(strategy, order)
	if err != nil {
		// 拘束したポジションを解放する
		// ただし、解放の処理でエラーでたら対応できない
		for _, hp := range order.HoldPositions {
			_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
		}
		return err
	}
	if res.Result {
		order.Code = res.OrderCode
		if err := s.orderStore.Save(order); err != nil {
			return fmt.Errorf("order=%+v: %w", order, err)
		}
	} else {
		// 拘束したポジションを解放する
		// ただし、解放の処理でエラーでたら対応できない
		for _, hp := range order.HoldPositions {
			_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
		}
		return fmt.Errorf("result=%+v, order=%+v: %w", res, order, ErrOrderCondition)
	}

	return nil
}

// EntryMarket - エントリーの成行注文
func (s *orderService) EntryMarket(strategyCode string, quantity float64) error {
	strategy, err := s.strategyStore.GetByCode(strategyCode)
	if err != nil {
		return err
	}

	order := &Order{
		StrategyCode:    strategy.Code,
		SymbolCode:      strategy.SymbolCode,
		Exchange:        strategy.Exchange,
		Status:          OrderStatusInOrder,
		Product:         strategy.Product,
		MarginTradeType: strategy.MarginTradeType,
		TradeType:       TradeTypeEntry,
		Side:            strategy.EntrySide,
		ExecutionType:   ExecutionTypeMarket,
		OrderQuantity:   quantity,
		AccountType:     strategy.Account.AccountType,
		OrderDateTime:   s.clock.Now(),
	}

	return s.sendOrder(strategy, order)
}

// ExitMarket - エグジットの成行注文
func (s *orderService) ExitMarket(strategyCode string, quantity float64, sortOrder SortOrder) error {
	strategy, err := s.strategyStore.GetByCode(strategyCode)
	if err != nil {
		return err
	}

	order := &Order{
		StrategyCode:    strategy.Code,
		SymbolCode:      strategy.SymbolCode,
		Exchange:        strategy.Exchange,
		Status:          OrderStatusInOrder,
		Product:         strategy.Product,
		MarginTradeType: strategy.MarginTradeType,
		TradeType:       TradeTypeExit,
		Side:            strategy.EntrySide.Turn(),
		ExecutionType:   ExecutionTypeMarket,
		OrderQuantity:   quantity,
		AccountType:     strategy.Account.AccountType,
		OrderDateTime:   s.clock.Now(),
	}
	hp, err := s.holdPositions(strategyCode, quantity, sortOrder)
	if err != nil {
		return err
	}
	order.HoldPositions = hp

	return s.sendOrder(strategy, order)
}

// checkEntryCash - エントリーするために必要な現金があるか
func (s *orderService) checkEntryCash(strategyCode string, cash float64, limitPrice float64, orderQuantity float64) (bool, error) {
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategyCode)
	if err != nil {
		return false, err
	}
	var totalLimitOrderPrice float64
	for _, o := range orders {
		totalLimitOrderPrice += o.Price * (o.OrderQuantity - o.ContractQuantity)
	}

	if cash < totalLimitOrderPrice+limitPrice*orderQuantity {
		return false, nil
	}

	return true, nil
}

// holdPositions - 注文に必要なポジションを拘束する
func (s *orderService) holdPositions(strategyCode string, quantity float64, sortOrder SortOrder) ([]HoldPosition, error) {
	positions, err := s.positionStore.GetActivePositionsByStrategyCode(strategyCode)
	if err != nil {
		return nil, err
	}

	// 並び順を変更し、古いのから返すか、新しいのから返すかを管理する
	switch sortOrder {
	case SortOrderNewest:
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].ContractDateTime.After(positions[j].ContractDateTime)
		})
	case SortOrderLatest:
		sort.Slice(positions, func(i, j int) bool {
			return positions[i].ContractDateTime.Before(positions[j].ContractDateTime)
		})
	}

	var hp []HoldPosition
	q := quantity
	for _, p := range positions {
		hq := math.Min(q, p.LeaveQuantity())
		if hq <= 0 {
			continue
		}
		if err := s.positionStore.Hold(p.Code, hq); err != nil {
			// 拘束したポジションを解放する
			// ただし、解放の処理でエラーでたら対応できない
			for _, hp := range hp {
				_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
			}
			return nil, err
		}
		hp = append(hp, HoldPosition{PositionCode: p.Code, HoldQuantity: hq})
		q -= hq

		// 必要数拘束したところで抜ける
		if q <= 0 {
			break
		}
	}

	// 必要数を拘束できないならエラー
	if q > 0 {
		// 拘束したポジションを解放する
		// ただし、解放の処理でエラーでたら対応できない
		for _, hp := range hp {
			_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
		}
		return nil, ErrNotEnoughPosition
	}

	return hp, nil
}

// sendOrder - 注文の送信から保存までの処理
func (s *orderService) sendOrder(strategy *Strategy, order *Order) error {
	if strategy == nil || order == nil {
		return ErrNilArgument
	}

	res, err := s.kabusAPI.SendOrder(strategy, order)
	if err != nil {
		// 拘束したポジションを解放する
		// ただし、解放の処理でエラーでたら対応できない
		for _, hp := range order.HoldPositions {
			_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
		}
		return err
	}

	if res.Result {
		order.Code = res.OrderCode
		if err := s.orderStore.Save(order); err != nil {
			return fmt.Errorf("order=%+v: %w", order, err)
		}
	} else {
		// 拘束したポジションを解放する
		// ただし、解放の処理でエラーでたら対応できない
		for _, hp := range order.HoldPositions {
			_ = s.positionStore.Release(hp.PositionCode, hp.HoldQuantity)
		}
		return fmt.Errorf("result=%+v, order=%+v: %w", res, order, ErrOrderCondition)
	}

	return nil
}
