package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testFourPriceStore struct {
	IFourPriceStore
	GetBySymbolCodeAndExchange1           []*FourPrice
	GetBySymbolCodeAndExchange2           error
	GetBySymbolCodeAndExchangeCount       int
	GetBySymbolCodeAndExchangeHistory     []interface{}
	Save1                                 error
	SaveCount                             int
	SaveHistory                           []interface{}
	GetLastBySymbolCodeAndExchange1       *FourPrice
	GetLastBySymbolCodeAndExchange2       error
	GetLastBySymbolCodeAndExchangeCount   int
	GetLastBySymbolCodeAndExchangeHistory []interface{}
}

func (t *testFourPriceStore) GetBySymbolCodeAndExchange(symbolCode string, exchange Exchange, num int) ([]*FourPrice, error) {
	t.GetBySymbolCodeAndExchangeCount++
	t.GetBySymbolCodeAndExchangeHistory = append(t.GetBySymbolCodeAndExchangeHistory, symbolCode)
	t.GetBySymbolCodeAndExchangeHistory = append(t.GetBySymbolCodeAndExchangeHistory, exchange)
	t.GetBySymbolCodeAndExchangeHistory = append(t.GetBySymbolCodeAndExchangeHistory, num)
	return t.GetBySymbolCodeAndExchange1, t.GetBySymbolCodeAndExchange2
}
func (t *testFourPriceStore) Save(fourPrice *FourPrice) error {
	t.SaveCount++
	t.SaveHistory = append(t.SaveHistory, fourPrice)
	return t.Save1
}
func (t *testFourPriceStore) GetLastBySymbolCodeAndExchange(symbolCode string, exchange Exchange) (*FourPrice, error) {
	t.GetLastBySymbolCodeAndExchangeCount++
	t.GetLastBySymbolCodeAndExchangeHistory = append(t.GetLastBySymbolCodeAndExchangeHistory, symbolCode)
	t.GetLastBySymbolCodeAndExchangeHistory = append(t.GetLastBySymbolCodeAndExchangeHistory, exchange)
	return t.GetLastBySymbolCodeAndExchange1, t.GetLastBySymbolCodeAndExchange2
}

func Test_getFourPriceStore(t *testing.T) {
	t.Parallel()
	db := &testDB{}
	want1 := &fourPriceStore{db: db, store: map[SymbolKey]*FourPrice{}}
	got1 := getFourPriceStore(db)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_fourPriceStore_GetBySymbolCodeAndExchange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                           string
		db                                             *testDB
		arg1                                           string
		arg2                                           Exchange
		arg3                                           int
		want1                                          []*FourPrice
		want2                                          error
		wantGetFourPriceBySymbolCodeAndExchangeHistory []interface{}
	}{
		{name: "dbがエラーを返したらエラー",
			db:    &testDB{GetFourPriceBySymbolCodeAndExchange2: ErrUnknown},
			arg1:  "1475",
			arg2:  ExchangeToushou,
			arg3:  3,
			want1: nil,
			want2: ErrUnknown,
			wantGetFourPriceBySymbolCodeAndExchangeHistory: []interface{}{"1475", ExchangeToushou, 3}},
		{name: "dbが値を返したらその値を返す",
			db: &testDB{GetFourPriceBySymbolCodeAndExchange1: []*FourPrice{
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 20, 15, 0, 0, 0, time.Local), Open: 1969, High: 2000, Low: 1960, Close: 1993},
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 19, 15, 0, 0, 0, time.Local), Open: 1999, High: 2010, Low: 1966, Close: 1973},
			}},
			arg1: "1475",
			arg2: ExchangeToushou,
			arg3: 3,
			want1: []*FourPrice{
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 20, 15, 0, 0, 0, time.Local), Open: 1969, High: 2000, Low: 1960, Close: 1993},
				{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 19, 15, 0, 0, 0, time.Local), Open: 1999, High: 2010, Low: 1966, Close: 1973},
			},
			want2: nil,
			wantGetFourPriceBySymbolCodeAndExchangeHistory: []interface{}{"1475", ExchangeToushou, 3}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fourPriceStore{db: test.db}
			got1, got2 := store.GetBySymbolCodeAndExchange(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantGetFourPriceBySymbolCodeAndExchangeHistory, test.db.GetFourPriceBySymbolCodeAndExchangeHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantGetFourPriceBySymbolCodeAndExchangeHistory,
					got1, got2, test.db.GetFourPriceBySymbolCodeAndExchangeHistory)
			}
		})
	}
}

func Test_fourPrice_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		db                       *testDB
		store                    map[SymbolKey]*FourPrice
		arg1                     *FourPrice
		want1                    error
		wantStore                map[SymbolKey]*FourPrice
		wantSaveFourPriceHistory []interface{}
	}{
		{name: "storeに同一銘柄のデータがなければ追加",
			db: &testDB{},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
			},
			arg1:  &FourPrice{SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			want1: nil,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			wantSaveFourPriceHistory: []interface{}{&FourPrice{SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940}}},
		{name: "storeに同一銘柄のデータがあれば上書き",
			db: &testDB{},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
			},
			arg1:  &FourPrice{SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 0, Close: 17020},
			want1: nil,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 0, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
			},
			wantSaveFourPriceHistory: []interface{}{&FourPrice{SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 0, Close: 17020}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fourPriceStore{db: test.db, store: test.store}
			got1 := store.Save(test.arg1)

			// 非同期なので少し待機
			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantStore, store.store) ||
				!reflect.DeepEqual(test.wantSaveFourPriceHistory, test.db.SaveFourPriceHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSaveFourPriceHistory,
					got1, store.store, test.db.SaveFourPriceHistory)
			}
		})
	}
}

func Test_fourPriceStore_GetLastBySymbolCodeAndExchange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                                           string
		db                                             *testDB
		store                                          map[SymbolKey]*FourPrice
		arg1                                           string
		arg2                                           Exchange
		want1                                          *FourPrice
		want2                                          error
		wantStore                                      map[SymbolKey]*FourPrice
		wantGetFourPriceBySymbolCodeAndExchangeHistory []interface{}
	}{
		{name: "storeに指定したデータがあればそれを返す",
			db: &testDB{},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			arg1:  "1475",
			arg2:  ExchangeToushou,
			want1: &FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
			want2: nil,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			wantGetFourPriceBySymbolCodeAndExchangeHistory: nil},
		{name: "storeに指定したデータがなければDBから取得して返す",
			db: &testDB{GetFourPriceBySymbolCodeAndExchange1: []*FourPrice{{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981}}},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			arg1:  "1475",
			arg2:  ExchangeToushou,
			want1: &FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
			want2: nil,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1475", Exchange: ExchangeToushou}: {SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			wantGetFourPriceBySymbolCodeAndExchangeHistory: []interface{}{"1475", ExchangeToushou, 1}},
		{name: "storeに指定したデータがなくDBから取得する際にエラーが出たらエラーを返す",
			db: &testDB{GetFourPriceBySymbolCodeAndExchange2: ErrUnknown},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			arg1:  "1475",
			arg2:  ExchangeToushou,
			want1: nil,
			want2: ErrUnknown,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			wantGetFourPriceBySymbolCodeAndExchangeHistory: []interface{}{"1475", ExchangeToushou, 1}},
		{name: "storeに指定したデータがなくDBから取得する際にエラーがなくても、DBからデータがとれなかったらエラーを返す",
			db: &testDB{GetFourPriceBySymbolCodeAndExchange1: []*FourPrice{}},
			store: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			arg1:  "1475",
			arg2:  ExchangeToushou,
			want1: nil,
			want2: ErrNoData,
			wantStore: map[SymbolKey]*FourPrice{
				SymbolKey{SymbolCode: "1458", Exchange: ExchangeToushou}: {SymbolCode: "1458", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 16355, High: 16785, Low: 16220, Close: 17020},
				SymbolKey{SymbolCode: "1699", Exchange: ExchangeToushou}: {SymbolCode: "1699", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 241.9, High: 242.4, Low: 238.0, Close: 247},
				SymbolKey{SymbolCode: "1476", Exchange: ExchangeToushou}: {SymbolCode: "1476", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1883, High: 1948, Low: 1846, Close: 1940},
			},
			wantGetFourPriceBySymbolCodeAndExchangeHistory: []interface{}{"1475", ExchangeToushou, 1}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fourPriceStore{db: test.db, store: test.store}
			got1, got2 := store.GetLastBySymbolCodeAndExchange(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantStore, store.store) ||
				!reflect.DeepEqual(test.wantGetFourPriceBySymbolCodeAndExchangeHistory, test.db.GetFourPriceBySymbolCodeAndExchangeHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantStore, test.wantGetFourPriceBySymbolCodeAndExchangeHistory,
					got1, got2, store.store, test.db.GetFourPriceBySymbolCodeAndExchangeHistory)
			}
		})
	}
}
