package gridon

import "sync"

// IPositionStore - ポジションストアのインターフェース
type IPositionStore interface {
	Save(position *Position) error
	ExitContract(positionCode string, quantity float64) error
	Release(positionCode string, quantity float64) error
}

// positionStore - ポジションストア
type positionStore struct {
	store map[string]*Position
	mtx   sync.Mutex
}

// Save - ポジションの保存
func (s *positionStore) Save(position *Position) error {
	if position == nil {
		return ErrNilArgument
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[position.Code] = position

	return nil
}

// ExitContract - エグジット約定による数量の変動を反映
func (s *positionStore) ExitContract(positionCode string, quantity float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[positionCode]; ok {
		s.store[positionCode].OwnedQuantity -= quantity
		s.store[positionCode].HoldQuantity -= quantity
	}

	return nil
}

// Release - 拘束しているポジションの解放
func (s *positionStore) Release(positionCode string, quantity float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[positionCode]; ok {
		s.store[positionCode].HoldQuantity -= quantity
	}

	return nil
}
