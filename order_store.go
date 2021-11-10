package gridon

import (
	"sort"
	"sync"
)

var (
	orderStoreSingleton    IOrderStore
	orderStoreSingletonMtx sync.Mutex
)

// getOrderStore - 注文ストアの取得
func getOrderStore(db IDB) IOrderStore {
	orderStoreSingletonMtx.Lock()
	defer orderStoreSingletonMtx.Unlock()

	if orderStoreSingleton == nil {
		orderStoreSingleton = &orderStore{
			db:    db,
			store: map[string]*Order{},
		}
	}

	return orderStoreSingleton
}

// IOrderStore - 注文ストアのインターフェース
type IOrderStore interface {
	DeployFromDB() error
	GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error)
	Save(order *Order) error
}

// orderStore - 注文ストア
type orderStore struct {
	db    IDB
	store map[string]*Order
	mtx   sync.Mutex
}

// DeployFromDB - DBからmapに展開する
func (s *orderStore) DeployFromDB() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orders, err := s.db.GetActiveOrders()
	if err != nil {
		return err
	}

	store := make(map[string]*Order)
	for _, order := range orders {
		store[order.Code] = order
	}
	s.store = store
	return nil
}

// GetActiveOrdersByStrategyCode - 戦略を指定して有効な注文を取り出す
func (s *orderStore) GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orders := make([]*Order, 0)
	for _, o := range s.store {
		// 戦略コードが違うか、無効な注文だったらスキップ
		if o.StrategyCode != strategyCode || !o.IsActive() {
			continue
		}
		orders = append(orders, o)
	}

	// 注文日時で並び替え
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].OrderDateTime.Before(orders[j].OrderDateTime)
	})

	return orders, nil
}

// Save - 注文の保存
func (s *orderStore) Save(order *Order) error {
	if order == nil {
		return ErrNilArgument
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[order.Code] = order

	go s.db.SaveOrder(s.store[order.Code])

	return nil
}
