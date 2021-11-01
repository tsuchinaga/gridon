package gridon

import "fmt"

// IOrderService - 注文サービスのインターフェース
type IOrderService interface {
	CancelAll(strategy *Strategy) error
}

// orderService - 注文サービス
type orderService struct {
	kabusAPI      IKabusAPI
	orderStore    IOrderStore
	positionStore IPositionStore
	clock         IClock
}

// CancelAll - 戦略に関連する全ての注文を取り消す
func (s *orderService) CancelAll(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
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
		// TODO 取消注文に失敗したログを残す？
	}
	return nil
}

// ExitAll - 戦略に関連する拘束されていないポジションを全てエグジットする
func (s *orderService) ExitAll(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
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
