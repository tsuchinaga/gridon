package gridon

import (
	"sync"
	"time"
)

// IStrategyStore - 戦略ストアのインターフェース
type IStrategyStore interface {
	GetByCode(code string) (*Strategy, error)
	AddStrategyCash(strategyCode string, cashDiff float64) error
	SetContract(strategyCode string, contractPrice float64, contractDateTime time.Time) error
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

// SetContract - 最終約定情報をセットする
func (s *strategyStore) SetContract(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].LastContractPrice = contractPrice
		s.store[strategyCode].LastContractDateTime = contractDateTime

		go s.db.SaveStrategy(s.store[strategyCode])
	}

	return nil
}
