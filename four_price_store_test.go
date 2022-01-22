package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testFourPriceStore struct {
	IFourPriceStore
	GetBySymbolCodeAndExchange1       []*FourPrice
	GetBySymbolCodeAndExchange2       error
	GetBySymbolCodeAndExchangeCount   int
	GetBySymbolCodeAndExchangeHistory []interface{}
	Save1                             error
	SaveCount                         int
	SaveHistory                       []interface{}
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

func Test_getFourPriceStore(t *testing.T) {
	t.Parallel()
	db := &testDB{}
	want1 := &fourPriceStore{db: db}
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
		arg1                     *FourPrice
		want1                    error
		wantSaveFourPriceHistory []interface{}
	}{
		{name: "非同期でDBの保存をたたくだけ",
			db:                       &testDB{},
			arg1:                     &FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981},
			want1:                    nil,
			wantSaveFourPriceHistory: []interface{}{&FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fourPriceStore{db: test.db}
			got1 := store.Save(test.arg1)

			// 非同期なので少し待機
			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantSaveFourPriceHistory, test.db.SaveFourPriceHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantSaveFourPriceHistory, got1, test.db.SaveFourPriceHistory)
			}
		})
	}
}
