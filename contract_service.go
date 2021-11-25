package gridon

// newContractService - 新しい約定管理サービスの取得
func newContractService(kabusAPI IKabusAPI, strategyStore IStrategyStore, orderStore IOrderStore, positionStore IPositionStore) IContractService {
	return &contractService{
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
		orderStore:    orderStore,
		positionStore: positionStore,
	}
}

// IContractService - 約定管理サービスのインターフェース
type IContractService interface {
	Confirm(strategy *Strategy) error
}

// contractService - 約定管理サービス
type contractService struct {
	kabusAPI      IKabusAPI
	strategyStore IStrategyStore
	orderStore    IOrderStore
	positionStore IPositionStore
}

// Confirm - 約定確認
func (s *contractService) Confirm(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// kabusapiから最終確認以降に変更された注文の一覧を取得
	securityOrders, err := s.kabusAPI.GetOrders(strategy.Product, strategy.LastContractDateTime)
	if err != nil {
		return err
	}

	// 手元にある注文中の注文の一覧を取得
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}

	// 両者の差異を確認し、注文やポジションに反映する
	//   新しく増えた約定
	//     エントリーなら注文の更新、ポジションの作成、現金余力の更新
	//     エグジットなら注文の更新、ポジションの更新、現金余力の更新
	//   ステータスが取消になった
	//     エントリーなら注文の更新
	//     エグジットなら注文の更新(拘束ポジション情報も)、拘束ポジションの解放
	for _, so := range securityOrders {
		for _, o := range orders {
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

	strategy, err := s.strategyStore.GetByCode(order.StrategyCode)
	if err != nil {
		return err
	}

	// ポジションの登録
	err = s.positionStore.Save(&Position{
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

	// 戦略の最終約定情報より新しい約定情報であれば更新
	if strategy.LastContractDateTime.Before(contract.ContractDateTime) {
		if err := s.strategyStore.SetContractPrice(order.StrategyCode, contract.Price, contract.ContractDateTime); err != nil {
			return err
		}
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

	strategy, err := s.strategyStore.GetByCode(order.StrategyCode)
	if err != nil {
		return err
	}

	// 注文が拘束しているポジションの更新
	q := contract.Quantity
	for i, hp := range order.HoldPositions {
		// 拘束残量のないポジションはスキップ
		lq := hp.LeaveQuantity()
		if lq <= 0 {
			continue
		}

		// 拘束ポジションのうち、約定した数量の特定
		cq := lq
		if q < lq {
			cq = q
		}
		q -= cq

		order.HoldPositions[i].ContractQuantity += cq

		// ポジションの更新
		if err := s.positionStore.ExitContract(hp.PositionCode, cq); err != nil {
			return err
		}
	}

	// 現金余力の更新
	if err := s.strategyStore.AddStrategyCash(order.StrategyCode, contract.Price*contract.Quantity); err != nil {
		return err
	}

	// 戦略の最終約定情報より新しい約定情報であれば更新
	if strategy.LastContractDateTime.Before(contract.ContractDateTime) {
		if err := s.strategyStore.SetContractPrice(order.StrategyCode, contract.Price, contract.ContractDateTime); err != nil {
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
