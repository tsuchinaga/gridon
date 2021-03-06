package gridon

import (
	"math"
	"time"
)

// newContractService - 新しい約定管理サービスの取得
func newContractService(kabusAPI IKabusAPI, strategyStore IStrategyStore, orderStore IOrderStore, positionStore IPositionStore, clock IClock) IContractService {
	return &contractService{
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
		clock:         clock,
	}
}

// IContractService - 約定管理サービスのインターフェース
type IContractService interface {
	Confirm(strategy *Strategy) error
	ConfirmGridEnd(strategy *Strategy) error
}

// contractService - 約定管理サービス
type contractService struct {
	kabusAPI      IKabusAPI
	strategyStore IStrategyStore
	orderStore    IOrderStore
	positionStore IPositionStore
	clock         IClock
}

// Confirm - 約定確認
func (s *contractService) Confirm(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// 手元にある注文中の注文の一覧を取得
	//   注文中のデータがなければ約定確認をスキップする
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}
	if len(orders) < 1 {
		return nil
	}

	// kabusapiから最終確認以降に変更された注文の一覧を取得
	// 約定日時より少し前にキャンセルが入る可能性があるから、1分の猶予を持つようにする
	updateDateTime := strategy.LastContractDateTime
	if !updateDateTime.IsZero() {
		updateDateTime = updateDateTime.Add(-1 * time.Minute)
	}

	securityOrders, err := s.kabusAPI.GetOrders(strategy.Product, strategy.SymbolCode, updateDateTime)
	if err != nil {
		return err
	}

	return s.confirm(strategy, orders, securityOrders)
}

// ConfirmGridEnd - グリッド終了時の約定確認
func (s *contractService) ConfirmGridEnd(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// グリッド戦略が動かないならチェックしない
	if !strategy.GridStrategy.Runnable {
		return nil
	}

	// グリッドの終了タイミングかをチェックし、終了タイミングならそのグリッドの開始時刻を持っておく
	// グリッドの終了タイミングでなければ終了
	now := s.clock.Now()
	var updateDateTime time.Time
	for _, timeRange := range strategy.GridStrategy.TimeRanges {
		tr := TimeRange{Start: timeRange.End, End: timeRange.End.Add(1 * time.Minute)}
		if tr.In(now) {
			updateDateTime = time.Date(now.Year(), now.Month(), now.Day(), timeRange.Start.Hour(), timeRange.Start.Minute(), timeRange.Start.Second(), timeRange.Start.Nanosecond(), timeRange.Start.Location())
			break
		}
	}
	if updateDateTime.IsZero() {
		return nil
	}

	// 手元にある注文中の注文の一覧を取得
	//   注文中のデータがなければ約定確認をスキップする
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}
	if len(orders) < 1 {
		return nil
	}

	securityOrders, err := s.kabusAPI.GetOrders(strategy.Product, strategy.SymbolCode, updateDateTime)
	if err != nil {
		return err
	}

	return s.confirm(strategy, orders, securityOrders)
}

// confirm - 渡された注文情報を使って約定確認を行なう
func (s *contractService) confirm(strategy *Strategy, storeOrders []*Order, securityOrders []SecurityOrder) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// 両者の差異を確認し、注文やポジションに反映する
	//   新しく増えた約定
	//     エントリーなら注文の更新、ポジションの作成、現金余力の更新
	//     エグジットなら注文の更新、ポジションの更新、現金余力の更新
	//   ステータスが取消になった
	//     エントリーなら注文の更新
	//     エグジットなら注文の更新(拘束ポジション情報も)、拘束ポジションの解放
	for _, so := range securityOrders {
		for _, o := range storeOrders {
			if so.Code != o.Code || o.IsEqualSecurityOrder(so) { // 違う注文か、同じ注文で内容が一致しているならスキップ
				continue
			}

			newContracts := o.ContractDiff(so)
			for _, c := range newContracts {
				switch o.TradeType {
				case TradeTypeEntry:
					if err := s.entryContract(o, c); err != nil {
						return err
					}
				case TradeTypeExit:
					if err := s.exitContract(o, c); err != nil {
						return err
					}
				}

				// 戦略の最終約定情報より新しい約定情報であれば更新
				if strategy.LastContractDateTime.Before(c.ContractDateTime) {
					if err := s.updateContractPrice(strategy, c.Price, c.ContractDateTime); err != nil {
						return err
					}
				}
			}

			// エグジット注文が取消されたら拘束していたポジションを解放する
			if so.Status == OrderStatusCanceled && so.TradeType == TradeTypeExit {
				if err := s.releaseHoldPositions(o); err != nil {
					return err
				}
			}

			// 注文の更新
			o.ReflectSecurityOrder(so)
			if err := s.orderStore.Save(o); err != nil {
				return err
			}
		}
	}
	return nil
}

// entryContract - エントリー注文の約定
// ポジションの登録、現金余力の更新
func (s *contractService) entryContract(order *Order, contract Contract) error {
	if order == nil {
		return ErrNilArgument
	}

	// ポジションの登録
	err := s.positionStore.Save(&Position{
		Code:             contract.PositionCode,
		StrategyCode:     order.StrategyCode,
		OrderCode:        order.Code,
		SymbolCode:       order.SymbolCode,
		Exchange:         order.Exchange,
		Side:             order.Side,
		Product:          order.Product,
		MarginTradeType:  order.MarginTradeType,
		Price:            contract.Price,
		OwnedQuantity:    contract.Quantity,
		HoldQuantity:     0,
		ContractDateTime: contract.ContractDateTime,
	})
	if err != nil {
		return err
	}

	// 現金余力の更新
	if err := s.strategyStore.AddStrategyCash(order.StrategyCode, -1*contract.Price*contract.Quantity); err != nil {
		return err
	}

	return nil
}

// exitContract - エグジット注文の約定
// ポジションの更新、損益の登録、現金余力の更新
// 引数の注文に副作用がある
func (s *contractService) exitContract(order *Order, contract Contract) error {
	if order == nil {
		return ErrNilArgument
	}

	// 注文が拘束しているポジションの更新
	q := contract.Quantity
	for i, hp := range order.HoldPositions {
		// 約定数量がなくなったら抜ける
		if q <= 0 {
			break
		}

		// 拘束残量のないポジションはスキップ
		lq := hp.LeaveQuantity()
		if lq <= 0 {
			continue
		}

		// 拘束ポジションのうち、約定した数量の特定
		cq := math.Min(lq, q)
		q -= cq

		order.HoldPositions[i].ContractQuantity += cq

		// ポジションの更新
		if err := s.positionStore.ExitContract(hp.PositionCode, cq); err != nil {
			return err
		}

		// 現金余力の更新
		//   エントリー時の約定値 + 損益 を反映する
		//   買いエントリーの場合: 約定値 x 数量 を加算
		//   売りエントリーの場合: (エントリー約定値 + エントリー約定値 - エグジット約定値) x 数量 を加算
		var price float64
		switch order.Side {
		case SideSell: // 売りエグジット = 買いエントリー
			price = contract.Price * cq
		case SideBuy: // 買いエグジット = 売りエントリー
			price = (hp.Price + hp.Price - contract.Price) * cq
		}
		if err := s.strategyStore.AddStrategyCash(order.StrategyCode, price); err != nil {
			return err
		}
	}

	return nil
}

// releaseHoldPositions - 注文に拘束されているポジションを解放する
// 引数の注文に副作用がある
func (s *contractService) releaseHoldPositions(order *Order) error {
	// 注文の拘束しているポジションを確認し、解放できるポジションを解放する
	for i, hp := range order.HoldPositions {
		leave := hp.LeaveQuantity()
		if leave <= 0 {
			continue
		}

		if err := s.positionStore.Release(hp.PositionCode, leave); err != nil {
			return err
		}
		order.HoldPositions[i].ReleaseQuantity += leave
	}
	return nil
}

// updateContractPrice - 戦略の約定情報を更新する
func (s *contractService) updateContractPrice(strategy *Strategy, contractPrice float64, contractDateTime time.Time) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// 戦略の最終約定情報より新しい約定情報であれば更新
	if strategy.LastContractDateTime.Before(contractDateTime) {
		if err := s.strategyStore.SetContractPrice(strategy.Code, contractPrice, contractDateTime); err != nil {
			return err
		}
	}

	// 実行可能なグリッド戦略がなければ終了
	if !strategy.GridStrategy.IsRunnable(contractDateTime) {
		return nil
	}

	for _, tr := range strategy.GridStrategy.TimeRanges {
		// 約定日時がグリッド時刻範囲外ならスキップ
		if !tr.In(contractDateTime) {
			continue
		}

		// 約定日の当グリッド期間の開始時刻
		currentStart := time.Date(
			contractDateTime.Year(),
			contractDateTime.Month(),
			contractDateTime.Day(),
			tr.Start.Hour(),
			tr.Start.Minute(),
			tr.Start.Second(),
			tr.Start.Nanosecond(),
			contractDateTime.Location())
		var err error

		// 最大約定日時が当グリッド時間の始まりより前か、最大約定日時がグリッド時間範囲外なら無条件で更新
		// 最大約定日時が同一グリッド時間範囲であって、新しい約定値のほうが高ければ更新
		if strategy.MaxContractDateTime.Before(currentStart) || !tr.In(strategy.MaxContractDateTime) {
			err = s.strategyStore.SetMaxContractPrice(strategy.Code, contractPrice, contractDateTime)
		} else if strategy.MaxContractDateTime.IsZero() || strategy.MaxContractPrice < contractPrice {
			err = s.strategyStore.SetMaxContractPrice(strategy.Code, contractPrice, contractDateTime)
		}
		if err != nil {
			return err
		}

		// 最小約定日時が当グリッド時間の始まりより前か、最小約定日時がグリッド時間範囲外なら無条件で更新
		// 最小約定日時が同一グリッド時間範囲であって、新しい約定値のほうが高ければ更新
		if strategy.MinContractDateTime.Before(currentStart) || !tr.In(strategy.MinContractDateTime) {
			err = s.strategyStore.SetMinContractPrice(strategy.Code, contractPrice, contractDateTime)
		} else if strategy.MinContractDateTime.IsZero() || strategy.MinContractPrice > contractPrice {
			err = s.strategyStore.SetMinContractPrice(strategy.Code, contractPrice, contractDateTime)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
