package gridon

import (
	"sort"
	"sync"
)

var (
	positionStoreSingleton    IPositionStore
	positionStoreSingletonMtx sync.Mutex
)

// getPositionStore - ポジションストアの取得
func getPositionStore(db IDB) IPositionStore {
	positionStoreSingletonMtx.Lock()
	defer positionStoreSingletonMtx.Unlock()

	if positionStoreSingleton == nil {
		positionStoreSingleton = &positionStore{
			db:    db,
			store: map[string]*Position{},
		}
	}

	return positionStoreSingleton
}

// IPositionStore - ポジションストアのインターフェース
type IPositionStore interface {
	DeployFromDB() error
	Save(position *Position) error
	ExitContract(positionCode string, quantity float64) error
	Release(positionCode string, quantity float64) error
	GetActivePositionsByStrategyCode(strategyCode string) ([]*Position, error)
	Hold(positionCode string, quantity float64) error
}

// positionStore - ポジションストア
type positionStore struct {
	db    IDB
	store map[string]*Position
	mtx   sync.Mutex
}

// DeployFromDB - DBからmapに展開する
func (s *positionStore) DeployFromDB() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions, err := s.db.GetActivePositions()
	if err != nil {
		return err
	}

	store := make(map[string]*Position)
	for _, position := range positions {
		store[position.Code] = position
	}
	s.store = store

	return nil
}

// Save - ポジションの保存
func (s *positionStore) Save(position *Position) error {
	if position == nil {
		return ErrNilArgument
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[position.Code] = position

	go s.db.SavePosition(s.store[position.Code])

	return nil
}

// ExitContract - エグジット約定による数量の変動を反映
func (s *positionStore) ExitContract(positionCode string, quantity float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[positionCode]; ok {
		s.store[positionCode].OwnedQuantity -= quantity
		s.store[positionCode].HoldQuantity -= quantity

		go s.db.SavePosition(s.store[positionCode])
	}

	return nil
}

// Release - 拘束しているポジションの解放
func (s *positionStore) Release(positionCode string, quantity float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[positionCode]; ok {
		s.store[positionCode].HoldQuantity -= quantity

		go s.db.SavePosition(s.store[positionCode])
	}

	return nil
}

// GetActivePositionsByStrategyCode - 戦略を指定してポジションを取り出す
func (s *positionStore) GetActivePositionsByStrategyCode(strategyCode string) ([]*Position, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*Position, 0)
	for _, p := range s.store {
		// 戦略コードが違うか、無効なポジションだったらスキップ
		if p.StrategyCode != strategyCode || !p.IsActive() {
			continue
		}
		positions = append(positions, p)
	}

	// 約定日時で並び替え
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].ContractDateTime.Before(positions[j].ContractDateTime)
	})

	return positions, nil
}

// Hold - 指定したポジションを拘束する
func (s *positionStore) Hold(positionCode string, quantity float64) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.store[positionCode]; ok {
		s.store[positionCode].HoldQuantity += quantity

		go s.db.SavePosition(s.store[positionCode])
	}

	return nil
}
