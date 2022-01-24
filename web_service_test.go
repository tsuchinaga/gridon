package gridon

import (
	"fmt"
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
	kabusAPI := &testKabusAPI{}
	want1 := &webService{
		port:          ":18083",
		strategyStore: strategyStore,
		kabusAPI:      kabusAPI,
		routes:        map[string]map[string]http.Handler{},
	}
	got1 := NewWebService(":18083", strategyStore, kabusAPI)
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
					TradingUnit:          1,
					RebalanceStrategy: RebalanceStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
							time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
						},
					},
					GridStrategy: GridStrategy{
						Runnable:      true,
						BaseWidth:     12,
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
					TradingUnit:          1,
					RebalanceStrategy: RebalanceStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
							time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
						},
					},
					GridStrategy: GridStrategy{
						Runnable:      true,
						BaseWidth:     12,
						Quantity:      1,
						NumberOfGrids: 3,
						TimeRanges: []TimeRange{
							{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
							{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
						},
						DynamicGridPrevDay: DynamicGridPrevDay{
							Valid:         true,
							Rate:          0.8,
							NumberOfGrids: 6,
							Rounding:      RoundingRound,
						},
						DynamicGridMinMax: DynamicGridMinMax{
							Valid:     true,
							Divide:    5,
							Rounding:  RoundingCeil,
							Operation: OperationPlus,
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
			wantBody:       `[{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"topix100","TradingUnit":1,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Quantity":1,"BaseWidth":12,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"DynamicGridPrevDay":{"Valid":false,"Rate":0,"NumberOfGrids":0,"Rounding":"","Operation":""},"DynamicGridMinMax":{"Valid":false,"Divide":0,"Rounding":"","Operation":""}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}},{"Code":"1458-sell","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"sell","Cash":885680,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"topix100","TradingUnit":1,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Quantity":1,"BaseWidth":12,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"DynamicGridPrevDay":{"Valid":true,"Rate":0.8,"NumberOfGrids":6,"Rounding":"round","Operation":""},"DynamicGridMinMax":{"Valid":true,"Divide":5,"Rounding":"ceil","Operation":"+"}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}]`},
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

func Test_webService_postSaveStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		strategyStore           *testStrategyStore
		kabusAPI                *testKabusAPI
		body                    string
		wantStatusCode          int
		wantBody                string
		wantGetSymbolHistory    []interface{}
		wantSaveStrategyHistory []interface{}
	}{
		{name: "bodyがjson形式でなければエラー",
			strategyStore:  &testStrategyStore{},
			kabusAPI:       &testKabusAPI{},
			body:           `a`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `invalid character 'a' looking for beginning of value`},
		{name: "bodyにcodeがなければエラー",
			strategyStore:  &testStrategyStore{},
			kabusAPI:       &testKabusAPI{},
			body:           `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `code is required`},
		{name: "bodyにcodeがなければエラー",
			strategyStore:  &testStrategyStore{},
			kabusAPI:       &testKabusAPI{},
			body:           `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `code is required`},
		{name: "銘柄情報取得に失敗したらエラー",
			strategyStore:        &testStrategyStore{},
			kabusAPI:             &testKabusAPI{GetSymbol2: ErrUnknown},
			body:                 `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","TickGroup":"topix100","RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"BaseWidth":12,"Quantity":1,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}]},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantStatusCode:       http.StatusInternalServerError,
			wantBody:             `unknown`,
			wantGetSymbolHistory: []interface{}{"1458", ExchangeToushou}},
		{name: "saveに失敗したらエラー",
			strategyStore:        &testStrategyStore{Save1: ErrUnknown},
			kabusAPI:             &testKabusAPI{GetSymbol1: &Symbol{Code: "1458", Exchange: ExchangeToushou, TradingUnit: 1, TickGroup: TickGroupTopix100}},
			body:                 `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"BaseWidth":12,"Quantity":1,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}]},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantStatusCode:       http.StatusInternalServerError,
			wantBody:             `unknown`,
			wantGetSymbolHistory: []interface{}{"1458", ExchangeToushou},
			wantSaveStrategyHistory: []interface{}{&Strategy{
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
				TradingUnit:          1,
				RebalanceStrategy: RebalanceStrategy{
					Runnable: true,
					Timings: []time.Time{
						time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
						time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
					},
				},
				GridStrategy: GridStrategy{
					Runnable:      true,
					BaseWidth:     12,
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
			}}},
		{name: "saveに成功したら保存したstrategyを返す",
			strategyStore:        &testStrategyStore{},
			kabusAPI:             &testKabusAPI{GetSymbol1: &Symbol{Code: "1458", Exchange: ExchangeToushou, TradingUnit: 1, TickGroup: TickGroupTopix100}},
			body:                 `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"BaseWidth":12,"Quantity":1,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"GridType":"min_max","DynamicGridMinMax":{"Divide":5,"Rounding":"ceil","Operation":"+"}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantStatusCode:       http.StatusOK,
			wantBody:             `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"topix100","TradingUnit":1,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Quantity":1,"BaseWidth":12,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"DynamicGridPrevDay":{"Valid":false,"Rate":0,"NumberOfGrids":0,"Rounding":"","Operation":""},"DynamicGridMinMax":{"Valid":false,"Divide":5,"Rounding":"ceil","Operation":"+"}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantGetSymbolHistory: []interface{}{"1458", ExchangeToushou},
			wantSaveStrategyHistory: []interface{}{&Strategy{
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
				TradingUnit:          1,
				RebalanceStrategy: RebalanceStrategy{
					Runnable: true,
					Timings: []time.Time{
						time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
						time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
					},
				},
				GridStrategy: GridStrategy{
					Runnable:      true,
					BaseWidth:     12,
					Quantity:      1,
					NumberOfGrids: 3,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					},
					DynamicGridMinMax: DynamicGridMinMax{
						Valid:     false,
						Divide:    5,
						Rounding:  RoundingCeil,
						Operation: OperationPlus,
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
			}}},
		{name: "rebalance戦略のsaveに成功したら保存したstrategyを返す",
			strategyStore:        &testStrategyStore{},
			kabusAPI:             &testKabusAPI{GetSymbol1: &Symbol{Code: "1458", Exchange: ExchangeToushou, TradingUnit: 1, TickGroup: TickGroupOther}},
			body:                 `{"Code":"1475-rebalance","SymbolCode":"1475","Exchange":"toushou","Product":"stock","EntrySide":"buy","Cash":75056,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantStatusCode:       http.StatusOK,
			wantBody:             `{"Code":"1475-rebalance","SymbolCode":"1475","Exchange":"toushou","Product":"stock","MarginTradeType":"","EntrySide":"buy","Cash":75056,"BasePrice":0,"BasePriceDateTime":"0001-01-01T00:00:00Z","LastContractPrice":0,"LastContractDateTime":"0001-01-01T00:00:00Z","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"other","TradingUnit":1,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":false,"Quantity":0,"BaseWidth":0,"NumberOfGrids":0,"TimeRanges":null,"DynamicGridPrevDay":{"Valid":false,"Rate":0,"NumberOfGrids":0,"Rounding":"","Operation":""},"DynamicGridMinMax":{"Valid":false,"Divide":0,"Rounding":"","Operation":""}},"CancelStrategy":{"Runnable":false,"Timings":null},"ExitStrategy":{"Runnable":false,"Conditions":null},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantGetSymbolHistory: []interface{}{"1475", ExchangeToushou},
			wantSaveStrategyHistory: []interface{}{&Strategy{
				Code:        "1475-rebalance",
				SymbolCode:  "1475",
				Exchange:    ExchangeToushou,
				Product:     ProductStock,
				EntrySide:   SideBuy,
				Cash:        75_056,
				TickGroup:   TickGroupOther,
				TradingUnit: 1,
				RebalanceStrategy: RebalanceStrategy{
					Runnable: true,
					Timings: []time.Time{
						time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
						time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
					},
				},
				Account: Account{Password: "Password1234", AccountType: AccountTypeSpecific},
			}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &webService{strategyStore: test.strategyStore, kabusAPI: test.kabusAPI}
			ts := httptest.NewServer(http.HandlerFunc(service.postSaveStrategy))
			defer ts.Close()

			res, err := http.Post(ts.URL, "application/json; charset=utf-8", strings.NewReader(test.body))
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
				!reflect.DeepEqual(test.wantBody, strBody) ||
				!reflect.DeepEqual(test.wantGetSymbolHistory, test.kabusAPI.GetSymbolHistory) ||
				!reflect.DeepEqual(test.wantSaveStrategyHistory, test.strategyStore.SaveHistory) {
				t.Errorf("%s error\nresult: %v, %v, %v, %v\nwant: %v, %+v, %+v, %v\ngot: %v, %+v, %+v, %v\n", t.Name(),
					!reflect.DeepEqual(test.wantStatusCode, res.StatusCode),
					!reflect.DeepEqual(test.wantBody, strBody),
					!reflect.DeepEqual(test.wantGetSymbolHistory, test.kabusAPI.GetSymbolHistory),
					!reflect.DeepEqual(test.wantSaveStrategyHistory, test.strategyStore.SaveHistory),
					test.wantStatusCode, test.wantBody, test.wantGetSymbolHistory, test.wantSaveStrategyHistory,
					res.StatusCode, strBody, test.kabusAPI.GetSymbolHistory, test.strategyStore.SaveHistory)
			}
		})
	}
}

func Test_webService_getStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		strategyStore        *testStrategyStore
		params               string
		wantStatusCode       int
		wantBody             string
		wantGetByCodeHistory []interface{}
	}{
		{name: "codeが指定されていなければエラー",
			strategyStore:        &testStrategyStore{GetByCode2: ErrNoData},
			params:               "",
			wantStatusCode:       http.StatusBadRequest,
			wantBody:             `no data`,
			wantGetByCodeHistory: []interface{}{""}},
		{name: "指定したcodeがなければエラー",
			strategyStore:        &testStrategyStore{GetByCode2: ErrNoData},
			params:               "?code=strategy-code-001",
			wantStatusCode:       http.StatusBadRequest,
			wantBody:             `no data`,
			wantGetByCodeHistory: []interface{}{"strategy-code-001"}},
		{name: "指定したcodeの戦略があれば、戦略の中身を返す",
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{
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
				TradingUnit:          1,
				RebalanceStrategy: RebalanceStrategy{
					Runnable: true,
					Timings: []time.Time{
						time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
						time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
					},
				},
				GridStrategy: GridStrategy{
					Runnable:      true,
					BaseWidth:     12,
					Quantity:      1,
					NumberOfGrids: 3,
					TimeRanges: []TimeRange{
						{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
						{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
					},
					DynamicGridMinMax: DynamicGridMinMax{
						Valid:     true,
						Divide:    5,
						Rounding:  RoundingCeil,
						Operation: OperationPlus,
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
			}},
			params:               "?code=1458-buy",
			wantStatusCode:       http.StatusOK,
			wantBody:             `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"topix100","TradingUnit":1,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Quantity":1,"BaseWidth":12,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"DynamicGridPrevDay":{"Valid":false,"Rate":0,"NumberOfGrids":0,"Rounding":"","Operation":""},"DynamicGridMinMax":{"Valid":true,"Divide":5,"Rounding":"ceil","Operation":"+"}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantGetByCodeHistory: []interface{}{"1458-buy"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &webService{strategyStore: test.strategyStore}
			ts := httptest.NewServer(http.HandlerFunc(service.getStrategy))
			defer ts.Close()

			res, err := http.Get(fmt.Sprintf("%s%s", ts.URL, test.params))
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
				!reflect.DeepEqual(test.wantBody, strBody) ||
				!reflect.DeepEqual(test.wantGetByCodeHistory, test.strategyStore.GetByCodeHistory) {
				t.Errorf("%s error\nresult: %v, %v, %v\nwant: %+v, %+v, %v\ngot: %+v, %+v, %v\n", t.Name(),
					!reflect.DeepEqual(test.wantStatusCode, res.StatusCode),
					!reflect.DeepEqual(test.wantBody, strBody),
					!reflect.DeepEqual(test.wantGetByCodeHistory, test.strategyStore.GetByCodeHistory),
					test.wantStatusCode, test.wantBody, test.wantGetByCodeHistory,
					res.StatusCode, strBody, test.strategyStore.GetByCodeHistory)
			}
		})
	}
}

func Test_webService_deleteStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		strategyStore           *testStrategyStore
		params                  string
		wantStatusCode          int
		wantBody                string
		wantGetByCodeHistory    []interface{}
		wantDeleteByCodeHistory []interface{}
	}{
		{name: "codeが指定されていなければエラー",
			strategyStore:        &testStrategyStore{GetByCode2: ErrNoData},
			params:               "",
			wantStatusCode:       http.StatusBadRequest,
			wantBody:             `no data`,
			wantGetByCodeHistory: []interface{}{""}},
		{name: "指定したcodeがなければエラー",
			strategyStore:        &testStrategyStore{GetByCode2: ErrNoData},
			params:               "?code=strategy-code-001",
			wantStatusCode:       http.StatusBadRequest,
			wantBody:             `no data`,
			wantGetByCodeHistory: []interface{}{"strategy-code-001"}},
		{name: "指定したcodeの戦略があっても、削除に失敗すればエラー",
			strategyStore: &testStrategyStore{
				GetByCode1: &Strategy{
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
					TradingUnit:          1,
					RebalanceStrategy: RebalanceStrategy{
						Runnable: true,
						Timings: []time.Time{
							time.Date(0, 1, 1, 8, 59, 0, 0, time.Local),
							time.Date(0, 1, 1, 12, 29, 0, 0, time.Local),
						},
					},
					GridStrategy: GridStrategy{
						Runnable:      true,
						BaseWidth:     12,
						Quantity:      1,
						NumberOfGrids: 3,
						TimeRanges: []TimeRange{
							{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
							{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
						},
						DynamicGridMinMax: DynamicGridMinMax{
							Divide:    5,
							Rounding:  "ceil",
							Operation: "+",
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
				DeleteByCode1: ErrUnknown},
			params:                  "?code=1458-buy",
			wantStatusCode:          http.StatusInternalServerError,
			wantBody:                `unknown`,
			wantGetByCodeHistory:    []interface{}{"1458-buy"},
			wantDeleteByCodeHistory: []interface{}{"1458-buy"}},
		{name: "指定したcodeの戦略があり、削除に成功すれば、戦略の中身を返す",
			strategyStore: &testStrategyStore{
				GetByCode1: &Strategy{
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
						BaseWidth:     12,
						Quantity:      1,
						NumberOfGrids: 3,
						TimeRanges: []TimeRange{
							{Start: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), End: time.Date(0, 1, 1, 11, 28, 0, 0, time.Local)},
							{Start: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), End: time.Date(0, 1, 1, 14, 58, 0, 0, time.Local)},
						},
						DynamicGridMinMax: DynamicGridMinMax{
							Divide:    5,
							Rounding:  "ceil",
							Operation: "+",
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
				DeleteByCode1: nil},
			params:                  "?code=1458-buy",
			wantStatusCode:          http.StatusOK,
			wantBody:                `{"Code":"1458-buy","SymbolCode":"1458","Exchange":"toushou","Product":"margin","MarginTradeType":"day","EntrySide":"buy","Cash":858010,"BasePrice":17995,"BasePriceDateTime":"2021-12-17T15:00:00+09:00","LastContractPrice":17995,"LastContractDateTime":"2021-12-17T15:00:00+09:00","MaxContractPrice":0,"MaxContractDateTime":"0001-01-01T00:00:00Z","MinContractPrice":0,"MinContractDateTime":"0001-01-01T00:00:00Z","TickGroup":"topix100","TradingUnit":0,"RebalanceStrategy":{"Runnable":true,"Timings":["0000-01-01T08:59:00+09:00","0000-01-01T12:29:00+09:00"]},"GridStrategy":{"Runnable":true,"Quantity":1,"BaseWidth":12,"NumberOfGrids":3,"TimeRanges":[{"Start":"0000-01-01T09:00:00+09:00","End":"0000-01-01T11:28:00+09:00"},{"Start":"0000-01-01T12:30:00+09:00","End":"0000-01-01T14:58:00+09:00"}],"DynamicGridPrevDay":{"Valid":false,"Rate":0,"NumberOfGrids":0,"Rounding":"","Operation":""},"DynamicGridMinMax":{"Valid":false,"Divide":5,"Rounding":"ceil","Operation":"+"}},"CancelStrategy":{"Runnable":true,"Timings":["0000-01-01T11:28:00+09:00","0000-01-01T14:58:00+09:00"]},"ExitStrategy":{"Runnable":true,"Conditions":[{"ExecutionType":"market_morning_close","Timing":"0000-01-01T11:29:00+09:00"},{"ExecutionType":"market_afternoon_close","Timing":"0000-01-01T14:59:00+09:00"}]},"Account":{"Password":"Password1234","AccountType":"specific"}}`,
			wantGetByCodeHistory:    []interface{}{"1458-buy"},
			wantDeleteByCodeHistory: []interface{}{"1458-buy"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &webService{strategyStore: test.strategyStore}
			ts := httptest.NewServer(http.HandlerFunc(service.deleteStrategy))
			defer ts.Close()

			client := &http.Client{}
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s%s", ts.URL, test.params), nil)
			res, err := client.Do(req)
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
				!reflect.DeepEqual(test.wantBody, strBody) ||
				!reflect.DeepEqual(test.wantGetByCodeHistory, test.strategyStore.GetByCodeHistory) ||
				!reflect.DeepEqual(test.wantDeleteByCodeHistory, test.strategyStore.DeleteByCodeHistory) {
				t.Errorf("%s error\nresult: %v, %v, %v, %v\nwant: %v, %+v, %+v, %v\ngot: %v, %+v, %+v, %v\n", t.Name(),
					!reflect.DeepEqual(test.wantStatusCode, res.StatusCode),
					!reflect.DeepEqual(test.wantBody, strBody),
					!reflect.DeepEqual(test.wantGetByCodeHistory, test.strategyStore.GetByCodeHistory),
					!reflect.DeepEqual(test.wantDeleteByCodeHistory, test.strategyStore.DeleteByCodeHistory),
					test.wantStatusCode, test.wantBody, test.wantGetByCodeHistory, test.wantDeleteByCodeHistory,
					res.StatusCode, strBody, test.strategyStore.GetByCodeHistory, test.strategyStore.DeleteByCodeHistory)
			}
		})
	}
}
