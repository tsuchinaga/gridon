package gridon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type testWebService struct {
	IWebService
	StartWebServer1    error
	StartWebServerLock bool
}

func (t *testWebService) StartWebServer() error {
	if t.StartWebServerLock {
		select {}
	}
	return t.StartWebServer1
}

func Test_NewWebService(t *testing.T) {
	t.Parallel()
	strategyStore := &testStrategyStore{}
	want1 := &webService{
		port:          ":18083",
		strategyStore: strategyStore,
		routes:        map[string]map[string]http.Handler{},
	}
	got1 := NewWebService(":18083", strategyStore)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_webService_ServeHTTP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		routes         map[string]map[string]http.Handler
		path           string
		method         string
		wantStatusCode int
		wantBody       string
	}{
		{name: "指定したパスがなければ404",
			path:           "/error",
			method:         "GET",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "404 Not Found"},
		{name: "指定したルーティングがあっても、メソッドがなければ405",
			path:           "/",
			method:         "POST",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "405 Method Not Allowed"},
		{name: "指定したルーティングがあって、メソッドもあれば、HandlerFuncが実行される",
			path:           "/",
			method:         "GET",
			wantStatusCode: http.StatusOK,
			wantBody:       "200 OK"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &webService{}
			ts := httptest.NewServer(http.HandlerFunc(service.ServeHTTP))
			defer ts.Close()
			service.routes = map[string]map[string]http.Handler{
				"/": {"GET": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("200 OK")) })},
			}

			var res *http.Response
			var err error

			switch test.method {
			case "GET":
				res, err = http.Get(ts.URL + test.path)
			case "POST":
				res, err = http.Post(ts.URL+test.path, "", nil)
			}
			if err != nil {
				t.Errorf("%s request error\nerr: %+v\n", t.Name(), err)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("%s read body error\nerr: %+v\n", t.Name(), err)
			}
			strBody := strings.Trim(string(body), "\n")

			if !reflect.DeepEqual(test.wantStatusCode, res.StatusCode) ||
				!reflect.DeepEqual(test.wantBody, strBody) {
				t.Errorf("%s error\nresult: %v, %v\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					!reflect.DeepEqual(test.wantStatusCode, res.StatusCode), !reflect.DeepEqual(test.wantBody, strBody),
					test.wantStatusCode, test.wantBody,
					res.StatusCode, strBody)
			}
		})
	}
}

func Test_webService_getStrategies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		strategyStore  *testStrategyStore
		wantStatusCode int
		wantBody       string
	}{
		{name: "storeがエラーを返したらエラー",
			strategyStore:  &testStrategyStore{GetStrategies2: ErrUnknown},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       ErrUnknown.Error()},
		{name: "storeがnilを返したらnullを返す",
			strategyStore:  &testStrategyStore{GetStrategies1: nil},
			wantStatusCode: 200,
			wantBody:       `null`},
		{name: "storeが空配列を返したら空配列を返す",
			strategyStore:  &testStrategyStore{GetStrategies1: []*Strategy{}},
			wantStatusCode: 200,
			wantBody:       `[]`},
		{name: "storeが戦略の一覧を返したら戦略の一覧を返す",
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{
					Code:                 "1458-buy",
					SymbolCode:           "1458",
					Exchange:             ExchangeToushou,
					Product:              ProductMargin,
					MarginTradeType:      MarginTradeTypeDay,
					EntrySide:            SideBuy,
					Cash:                 858_010,
					BasePrice:            17_995,
					BasePriceDateTime:    time.Date(2021, 12, 17, 15, 0, 0, 0, time.Local),
					LastContractPrice:    17_995,
					LastContractDateTime: time.Date(2021, 12, 17, 15, 0, 0, 0, time.Local),
					TickGroup:            TickGroupTopix100,
					RebalanceStrategy: RebalanceStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
							time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
						},
					},
					GridStrategy: GridStrategy{
						Runnable:      true,
						Width:         12,
						Quantity:      1,
						NumberOfGrids: 3,
						TimeRanges: []TimeRange{
							{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
							{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
						},
					},
					CancelStrategy: CancelStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 11, 28, 0, 0, time.Local),
							time.Date(0, 1, 1, 14, 58, 0, 0, time.Local),
						},
					},
					ExitStrategy: ExitStrategy{
						Runnable: true,
						Conditions: []ExitCondition{
							{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 29, 0, 0, time.Local)},
							{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
						},
					},
					Account: Account{Password: "Password1234", AccountType: AccountTypeSpecific},
				},
				{
					Code:                 "1458-sell",
					SymbolCode:           "1458",
					Exchange:             ExchangeToushou,
					Product:              ProductMargin,
					MarginTradeType:      MarginTradeTypeDay,
					EntrySide:            SideSell,
					Cash:                 885_680,
					BasePrice:            17_995,
					BasePriceDateTime:    time.Date(2021, 12, 17, 15, 0, 0, 0, time.Local),
					LastContractPrice:    17_995,
					LastContractDateTime: time.Date(2021, 12, 17, 15, 0, 0, 0, time.Local),
					TickGroup:            TickGroupTopix100,
					RebalanceStrategy: RebalanceStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
							time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
						},
					},
					GridStrategy: GridStrategy{
						Runnable:      true,
						Width:         12,
						Quantity:      1,
						NumberOfGrids: 3,
						TimeRanges: []TimeRange{
							{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
							{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
						},
					},
					CancelStrategy: CancelStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 11, 28, 0, 0, time.Local),
							time.Date(0, 1, 1, 14, 58, 0, 0, time.Local),
						},
					},
					ExitStrategy: ExitStrategy{
						Runnable: true,
						Conditions: []ExitCondition{
							{ExecutionType: ExecutionTypeMarketMorningClose, Timing: time.Date(0, 1, 1, 11, 29, 0, 0, time.Local)},
							{ExecutionType: ExecutionTypeMarketAfternoonClose, Timing: time.Date(0, 1, 1, 14, 59, 0, 0, time.Local)},
						},
					},
					Account: Account{Password: "Password1234", AccountType: AccountTypeSpecific},
				},
			}},
			wantStatusCode: 200,
			wantBody:       `[{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","TickGroup":"topix100","RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Width":12,"Quantity":1,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}]},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}},{"Code":"1458-sell","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"sell","Cash":885680,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","TickGroup":"topix100","RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Width":12,"Quantity":1,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}]},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}]`},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &webService{strategyStore: test.strategyStore}
			ts := httptest.NewServer(http.HandlerFunc(service.getStrategies))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Errorf("%s request error\nerr: %+v\n", t.Name(), err)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("%s read body error\nerr: %+v\n", t.Name(), err)
			}
			strBody := strings.Trim(string(body), "\n")

			if !reflect.DeepEqual(test.wantStatusCode, res.StatusCode) ||
				!reflect.DeepEqual(test.wantBody, strBody) {
				t.Errorf("%s error\nresult: %v, %v\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					!reflect.DeepEqual(test.wantStatusCode, res.StatusCode), !reflect.DeepEqual(test.wantBody, strBody),
					test.wantStatusCode, test.wantBody,
					res.StatusCode, strBody)
			}
		})
	}
}
