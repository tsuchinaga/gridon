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
			db:    db,
			store: map[SymbolKey]*FourPrice{},
		}
	}

	return fourPriceStoreSingleton
}

// IFourPriceStore - 四本値ストアのインターフェース
type IFourPriceStore interface {
	GetBySymbolCodeAndExchange(symbolCode string, exchange Exchange, num int) ([]*FourPrice, error)
	GetLastBySymbolCodeAndExchange(symbolCode string, exchange Exchange) (*FourPrice, error)
	Save(fourPrice *FourPrice) error
}

// fourPriceStore - 四本値ストア
type fourPriceStore struct {
	db    IDB
	store map[SymbolKey]*FourPrice
	mtx   sync.Mutex
}

// GetBySymbolCodeAndExchange - 四本値データの取得
func (s *fourPriceStore) GetBySymbolCodeAndExchange(symbolCode string, exchange Exchange, num int) ([]*FourPrice, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.db.GetFourPriceBySymbolCodeAndExchange(symbolCode, exchange, num)
}

// GetLastBySymbolCodeAndExchange - 指定した銘柄の最新の四本値データの取得
func (s *fourPriceStore) GetLastBySymbolCodeAndExchange(symbolCode string, exchange Exchange) (*FourPrice, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if f, ok := s.store[SymbolKey{SymbolCode: symbolCode, Exchange: exchange}]; ok {
		return f, nil
	}

	fs, err := s.db.GetFourPriceBySymbolCodeAndExchange(symbolCode, exchange, 1)
	if err != nil {
		return nil, err
	}
	if len(fs) == 0 {
		return nil, ErrNoData
	}
	s.store[SymbolKey{SymbolCode: symbolCode, Exchange: exchange}] = fs[0]
	return s.store[SymbolKey{SymbolCode: symbolCode, Exchange: exchange}], nil
}

// Save - 四本値の保存
func (s *fourPriceStore) Save(fourPrice *FourPrice) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[SymbolKey{SymbolCode: fourPrice.SymbolCode, Exchange: fourPrice.Exchange}] = fourPrice
	go s.db.SaveFourPrice(fourPrice)

	return nil
}
