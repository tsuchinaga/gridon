package gridon

import (
	"sort"
	"sync"
	"time"
)

var (
	strategyStoreSingleton    IStrategyStore
	strategyStoreSingletonMtx sync.Mutex
)

// getStrategyStore - 戦略ストアの取得
func getStrategyStore(db IDB) IStrategyStore {
	strategyStoreSingletonMtx.Lock()
	defer strategyStoreSingletonMtx.Unlock()

	if strategyStoreSingleton == nil {
		strategyStoreSingleton = &strategyStore{
			store: map[string]*Strategy{},
			db:    db,
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
	SetTickGroup(strategyCode string, tickGroup TickGroup) error
	Save(strategy *Strategy) error
}

// strategyStore - 戦略ストア
type strategyStore struct {
	store map[string]*Strategy
	db    IDB
	mtx   sync.Mutex
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

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].Cash += cashDiff

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
