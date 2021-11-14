package gridon

// newGridService - 新しいグリッドサービスの取得
func newGridService(clock IClock, tick ITick, kabusAPI IKabusAPI, orderService IOrderService) IGridService {
	return &gridService{
		clock:        clock,
		tick:         tick,
		kabusAPI:     kabusAPI,
		orderService: orderService,
	}
}

// IGridService - グリッドサービスのインターフェース
type IGridService interface {
	Leveling(strategy *Strategy) error
}

// gridService - グリッドサービス
type gridService struct {
	clock        IClock
	tick         ITick
	kabusAPI     IKabusAPI
	orderService IOrderService
}

// Leveling - グリッドの整地
func (s *gridService) Leveling(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// グリッド戦略が無効なら抜ける
	if !strategy.GridStrategy.IsRunnable(s.clock.Now()) {
		return nil
	}

	// 注文中の注文から各グリッドに乗っている数量を取得
	orders, err := s.orderService.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}

	// 最終約定価格と最終約定時刻から基準価格を取得
	// 基準価格が取得できない場合、現在値を取得して基準価格とする
	basePrice, err := s.getBasePrice(strategy)
	if err != nil {
		return err
	}

	// 基準価格から最大グリッド数より外にある注文を特定して取り消す
	upper := s.tick.TickAddedPrice(strategy.TickGroup, basePrice, strategy.GridStrategy.NumberOfGrids*strategy.GridStrategy.Width)
	lower := s.tick.TickAddedPrice(strategy.TickGroup, basePrice, -1*strategy.GridStrategy.NumberOfGrids*strategy.GridStrategy.Width)
	gridQuantities := make(map[float64]float64)
	for _, o := range orders {
		// 指値注文以外はスキップ
		if o.ExecutionType != ExecutionTypeLimit {
			continue
		}

		if lower <= o.Price && o.Price <= upper {
			gridQuantities[o.Price] += o.OrderQuantity - o.ContractQuantity
		} else {
			if err := s.orderService.Cancel(strategy, o.Code); err != nil {
				return err
			}
		}
	}

	// グリッドの中心から外に注文を確認していく
	for i := 1; i <= strategy.GridStrategy.NumberOfGrids; i++ {
		// upper
		{
			upper := s.tick.TickAddedPrice(strategy.TickGroup, basePrice, i*strategy.GridStrategy.Width)
			quantity := strategy.GridStrategy.Quantity - gridQuantities[upper]
			// 部分約定対策として、基準価格の隣の場合に限り基準価格に乗っている数量を減算する
			if i == 1 {
				quantity -= gridQuantities[basePrice]
			}

			// 注文数量があれば注文送信
			if quantity > 0 {
				if err := s.sendGridOrder(strategy, upper, basePrice, quantity); err != nil {
					return err
				}
			}
		}

		// lower
		{
			lower := s.tick.TickAddedPrice(strategy.TickGroup, basePrice, -1*i*strategy.GridStrategy.Width)
			quantity := strategy.GridStrategy.Quantity - gridQuantities[lower]
			// 部分約定対策として、基準価格の隣の場合に限り基準価格に乗っている数量を減算する
			if i == 1 {
				quantity -= gridQuantities[basePrice]
			}

			// 注文数量があれば注文送信
			if quantity > 0 {
				if err := s.sendGridOrder(strategy, lower, basePrice, quantity); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// getBasePrice - 戦略から基準価格を取り出す
// グリッド戦略の実行時刻範囲のうち、現在時刻と同じ範囲内の約定があればその価格を基準価格にし、
// なければ銘柄情報を取得して現在値を基準価格とする
func (s *gridService) getBasePrice(strategy *Strategy) (float64, error) {
	if strategy == nil {
		return 0, ErrNilArgument
	}

	if len(strategy.GridStrategy.TimeRanges) == 0 {
		return 0, ErrNotExistsTimeRange
	}

	now := s.clock.Now()
	for _, tr := range strategy.GridStrategy.TimeRanges {
		if tr.In(now) && tr.In(strategy.LastContractDateTime) {
			return strategy.LastContractPrice, nil
		}
	}

	symbol, err := s.kabusAPI.GetSymbol(strategy.SymbolCode, strategy.Exchange)
	if err != nil {
		return 0, err
	}
	return symbol.CurrentPrice, nil
}

// sendGridOrder - グリッド注文を作成し、送信する
func (s *gridService) sendGridOrder(strategy *Strategy, limitPrice float64, basePrice float64, quantity float64) error {
	if strategy == nil {
		return ErrNilArgument
	}

	var side Side
	if limitPrice < basePrice {
		side = SideBuy
	} else if basePrice < limitPrice {
		side = SideSell
	} else {
		return ErrUndecidableValue
	}
	if strategy.EntrySide == side {
		if err := s.orderService.EntryLimit(strategy.Code, limitPrice, quantity); err != nil {
			return err
		}
	} else if strategy.EntrySide == side.Turn() {
		if err := s.orderService.ExitLimit(strategy.Code, limitPrice, quantity, SortOrderNewest); err != nil {
			return err
		}
	}
	return nil
}
