package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testPriceService struct {
	IPriceService
	SaveFourPrice1       error
	SaveFourPriceCount   int
	SaveFourPriceHistory []interface{}
}

func (t *testPriceService) SaveFourPrice(symbolCode string, exchange Exchange) error {
	t.SaveFourPriceCount++
	t.SaveFourPriceHistory = append(t.SaveFourPriceHistory, symbolCode)
	t.SaveFourPriceHistory = append(t.SaveFourPriceHistory, exchange)
	return t.SaveFourPrice1
}

func Test_priceService_SaveFourPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		kabusAPI                *testKabusAPI
		fourPriceStore          *testFourPriceStore
		arg1                    string
		arg2                    Exchange
		want1                   error
		wantGetFourPriceHistory []interface{}
		wantSaveHistory         []interface{}
	}{
		{name: "四本値取得でエラーがあればエラーを返す",
			kabusAPI:                &testKabusAPI{GetFourPrice2: ErrUnknown},
			fourPriceStore:          &testFourPriceStore{},
			arg1:                    "1475",
			arg2:                    ExchangeToushou,
			want1:                   ErrUnknown,
			wantGetFourPriceHistory: []interface{}{"1475", ExchangeToushou}},
		{name: "四本値を取得できればDBに保存する",
			kabusAPI:                &testKabusAPI{GetFourPrice1: &FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981}},
			fourPriceStore:          &testFourPriceStore{},
			arg1:                    "1475",
			arg2:                    ExchangeToushou,
			want1:                   nil,
			wantGetFourPriceHistory: []interface{}{"1475", ExchangeToushou},
			wantSaveHistory:         []interface{}{&FourPrice{SymbolCode: "1475", Exchange: ExchangeToushou, DateTime: time.Date(2022, 1, 21, 15, 0, 0, 0, time.Local), Open: 1966, High: 1985, Low: 1952, Close: 1981}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &priceService{kabusAPI: test.kabusAPI, fourPriceStore: test.fourPriceStore}
			got1 := service.SaveFourPrice(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantGetFourPriceHistory, test.kabusAPI.GetFourPriceHistory) ||
				!reflect.DeepEqual(test.wantSaveHistory, test.fourPriceStore.SaveHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantGetFourPriceHistory, test.wantSaveHistory,
					got1, test.kabusAPI.GetFourPriceHistory, test.fourPriceStore.SaveHistory)
			}
		})
	}
}

func Test_newPriceService(t *testing.T) {
	t.Parallel()
	kabusAPI := &testKabusAPI{}
	fourPriceStore := &testFourPriceStore{}
	want1 := &priceService{
		kabusAPI:       kabusAPI,
		fourPriceStore: fourPriceStore,
	}
	got1 := newPriceService(kabusAPI, fourPriceStore)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}
