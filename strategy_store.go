package gridon

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	strategyStoreSingleton    IStrategyStore
	strategyStoreSingletonMtx sync.Mutex
)

// getStrategyStore - 戦略ストアの取得
func getStrategyStore(db IDB, logger ILogger) IStrategyStore {
	strategyStoreSingletonMtx.Lock()
	defer strategyStoreSingletonMtx.Unlock()

	if strategyStoreSingleton == nil {
		strategyStoreSingleton = &strategyStore{
			store:  map[string]*Strategy{},
			db:     db,
			logger: logger,
		}
	}

	return strategyStoreSingleton
}

// IStrategyStore - 戦略ストアのインターフェース
type IStrategyStore interface {
	DeployFromDB() error
	GetByCode(code string) (*Strategy, error)
	GetStrategies() ([]*Strategy, error)
	AddStrategyCash(strategyCode string, cashDiff float64) error
	SetBasePrice(strategyCode string, basePrice float64, basePriceDateTime time.Time) error
	SetContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error
	SetMaxContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error
	SetMinContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error
	SetTickGroup(strategyCode string, tickGroup TickGroup) error
	Save(strategy *Strategy) error
	DeleteByCode(code string) error
}

// strategyStore - 戦略ストア
type strategyStore struct {
	store  map[string]*Strategy
	db     IDB
	logger ILogger
	mtx    sync.Mutex
}

// DeployFromDB - DBからmapに展開する
func (s *strategyStore) DeployFromDB() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	strategies, err := s.db.GetStrategies()
	if err != nil {
		return err
	}

	store := make(map[string]*Strategy)
	for _, strategy := range strategies {
		store[strategy.Code] = strategy
	}
	s.store = store
	return nil
}

// GetByCode - コードを指定して取り出す
func (s *strategyStore) GetByCode(code string) (*Strategy, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	strategy, ok := s.store[code]
	if !ok {
		return nil, ErrNoData
	}

	return strategy, nil
}

// GetStrategies - 戦略一覧の取得
func (s *strategyStore) GetStrategies() ([]*Strategy, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	strategies := make([]*Strategy, 0)
	for _, strategy := range s.store {
		strategies = append(strategies, strategy)
	}
	sort.Slice(strategies, func(i, j int) bool {
		return strategies[i].Code < strategies[j].Code
	})

	return strategies, nil
}

// AddStrategyCash - 現金余力に加算する
// 負の値を与えると減算になる
func (s *strategyStore) AddStrategyCash(strategyCode string, cashDiff float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if strategy, ok := s.store[strategyCode]; ok {
		calc := s.store[strategyCode].Cash + cashDiff
		s.logger.CashFlow(fmt.Sprintf("strategyCode: %s, symbolCode: %s, cash: %.2f, diff: %.2f, calc: %.2f", strategy.Code, strategy.SymbolCode, strategy.Cash, cashDiff, calc))
		s.store[strategyCode].Cash = calc

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// SetBasePrice - 基準情報をセットする
func (s *strategyStore) SetBasePrice(strategyCode string, basePrice float64, basePriceDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].BasePrice = basePrice
		s.store[strategyCode].BasePriceDateTime = basePriceDateTime

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// SetContractPrice - 最終約定情報をセットする
// 約定値は基準価格にもなるので、更新時に基準価格も更新する
func (s *strategyStore) SetContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].LastContractPrice = contractPrice
		s.store[strategyCode].LastContractDateTime = contractDateTime
		s.store[strategyCode].BasePrice = contractPrice
		s.store[strategyCode].BasePriceDateTime = contractDateTime

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// SetMaxContractPrice - 最大約定情報をセットする
func (s *strategyStore) SetMaxContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].MaxContractPrice = contractPrice
		s.store[strategyCode].MaxContractDateTime = contractDateTime

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// SetMinContractPrice - 最小約定情報をセットする
func (s *strategyStore) SetMinContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].MinContractPrice = contractPrice
		s.store[strategyCode].MinContractDateTime = contractDateTime

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// SetTickGroup - 呼値グループを設定する
func (s *strategyStore) SetTickGroup(strategyCode string, tickGroup TickGroup) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].TickGroup = tickGroup

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}

// Save - 戦略の保存
func (s *strategyStore) Save(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[strategy.Code] = strategy

	go s.db.SaveStrategy(s.store[strategy.Code])

	return nil
}

// DeleteByCode - 戦略の削除
func (s *strategyStore) DeleteByCode(code string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[code]; ok {
		delete(s.store, code)
		go s.db.DeleteStrategyByCode(code)
	}

	return nil
}
