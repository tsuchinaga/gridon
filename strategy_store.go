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
	mtx   sync.Mutex
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

	_, ok := s.store[strategyCode]
	if !ok {
		return nil
	}
	s.store[strategyCode].Cash += cashDiff

	return nil
}

// SetContract - 最終約定情報をセットする
func (s *strategyStore) SetContract(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[strategyCode]; ok {
		s.store[strategyCode].LastContractPrice = contractPrice
		s.store[strategyCode].LastContractDateTime = contractDateTime
	}

	return nil
}
