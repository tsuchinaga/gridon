package gridon

// newStrategyService - 戦略サービスの取得
func newStrategyService(kabusAPI IKabusAPI, strategyStore IStrategyStore) IStrategyService {
	return &strategyService{
		kabusAPI:      kabusAPI,
		strategyStore: strategyStore,
	}
}

// IStrategyService - 戦略サービスのインターフェース
type IStrategyService interface {
	UpdateStrategyTickGroup(strategyCode string) error
}

// strategyService - 戦略サービス
type strategyService struct {
	kabusAPI      IKabusAPI
	strategyStore IStrategyStore
}

// UpdateStrategyTickGroup - 指定した戦略の呼値グループを更新する
func (s *strategyService) UpdateStrategyTickGroup(strategyCode string) error {
	strategy, err := s.strategyStore.GetByCode(strategyCode)
	if err != nil {
		return err
	}

	symbol, err := s.kabusAPI.GetSymbol(strategy.SymbolCode, strategy.Exchange)
	if err != nil {
		return err
	}

	return s.strategyStore.SetSymbolInfo(strategyCode, symbol.TickGroup, symbol.TradingUnit)
}
