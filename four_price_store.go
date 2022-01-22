package gridon

import "sync"

var (
	fourPriceStoreSingleton    IFourPriceStore
	fourPriceStoreSingletonMtx sync.Mutex
)

// getFourPriceStore - 四本値ストアの取得
func getFourPriceStore(db IDB) IFourPriceStore {
	fourPriceStoreSingletonMtx.Lock()
	defer fourPriceStoreSingletonMtx.Unlock()

	if fourPriceStoreSingleton == nil {
		fourPriceStoreSingleton = &fourPriceStore{
			db: db,
		}
	}

	return fourPriceStoreSingleton
}

// IFourPriceStore - 四本値ストアのインターフェース
type IFourPriceStore interface {
	GetBySymbolCodeAndExchange(symbolCode string, exchange Exchange, num int) ([]*FourPrice, error)
	Save(fourPrice *FourPrice) error
}

// fourPriceStore - 四本値ストア
// 利用頻度が高くないストアになるのでメモリ上に保持せずすべてファイルDBへの操作にする
type fourPriceStore struct {
	db  IDB
	mtx sync.Mutex
}

// GetBySymbolCodeAndExchange - 四本値データの取得
func (s *fourPriceStore) GetBySymbolCodeAndExchange(symbolCode string, exchange Exchange, num int) ([]*FourPrice, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.db.GetFourPriceBySymbolCodeAndExchange(symbolCode, exchange, num)
}

// Save - 四本値の保存
func (s *fourPriceStore) Save(fourPrice *FourPrice) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	go s.db.SaveFourPrice(fourPrice)

	return nil
}
