package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testStrategyStore struct {
	IStrategyStore
	AddStrategyCash1           error
	AddStrategyCashHistory     []interface{}
	AddStrategyCashCount       int
	SetBasePrice1              error
	SetBasePriceHistory        []interface{}
	SetBasePriceCount          int
	GetByCode1                 *Strategy
	GetByCode2                 error
	GetByCodeHistory           []interface{}
	GetByCodeCount             int
	GetStrategies1             []*Strategy
	GetStrategies2             error
	GetStrategiesCount         int
	DeployFromDB1              error
	DeployFromDBCount          int
	SetSymbolInfo1             error
	SetSymbolInfoCount         int
	SetSymbolInfoHistory       []interface{}
	SetContractPrice1          error
	SetContractPriceHistory    []interface{}
	SetContractPriceCount      int
	SetMaxContractPrice1       error
	SetMaxContractPriceHistory []interface{}
	SetMaxContractPriceCount   int
	SetMinContractPrice1       error
	SetMinContractPriceHistory []interface{}
	SetMinContractPriceCount   int
	Save1                      error
	SaveHistory                []interface{}
	SaveCount                  int
	DeleteByCode1              error
	DeleteByCodeHistory        []interface{}
	DeleteByCodeCount          int
}

func (t *testStrategyStore) GetByCode(code string) (*Strategy, error) {
	t.GetByCodeHistory = append(t.GetByCodeHistory, code)
	t.GetByCodeCount++
	return t.GetByCode1, t.GetByCode2
}
func (t *testStrategyStore) AddStrategyCash(strategyCode string, cashDiff float64) error {
	t.AddStrategyCashHistory = append(t.AddStrategyCashHistory, strategyCode)
	t.AddStrategyCashHistory = append(t.AddStrategyCashHistory, cashDiff)
	t.AddStrategyCashCount++
	return t.AddStrategyCash1
}
func (t *testStrategyStore) SetBasePrice(strategyCode string, basePrice float64, basePriceDateTime time.Time) error {
	t.SetBasePriceHistory = append(t.SetBasePriceHistory, strategyCode)
	t.SetBasePriceHistory = append(t.SetBasePriceHistory, basePrice)
	t.SetBasePriceHistory = append(t.SetBasePriceHistory, basePriceDateTime)
	t.SetBasePriceCount++
	return t.SetBasePrice1
}
func (t *testStrategyStore) SetContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	t.SetContractPriceHistory = append(t.SetContractPriceHistory, strategyCode)
	t.SetContractPriceHistory = append(t.SetContractPriceHistory, contractPrice)
	t.SetContractPriceHistory = append(t.SetContractPriceHistory, contractDateTime)
	t.SetContractPriceCount++
	return t.SetContractPrice1
}
func (t *testStrategyStore) SetMaxContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	t.SetMaxContractPriceHistory = append(t.SetMaxContractPriceHistory, strategyCode)
	t.SetMaxContractPriceHistory = append(t.SetMaxContractPriceHistory, contractPrice)
	t.SetMaxContractPriceHistory = append(t.SetMaxContractPriceHistory, contractDateTime)
	t.SetMaxContractPriceCount++
	return t.SetMaxContractPrice1
}
func (t *testStrategyStore) SetMinContractPrice(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	t.SetMinContractPriceHistory = append(t.SetMinContractPriceHistory, strategyCode)
	t.SetMinContractPriceHistory = append(t.SetMinContractPriceHistory, contractPrice)
	t.SetMinContractPriceHistory = append(t.SetMinContractPriceHistory, contractDateTime)
	t.SetMinContractPriceCount++
	return t.SetMinContractPrice1
}
func (t *testStrategyStore) GetStrategies() ([]*Strategy, error) {
	t.GetStrategiesCount++
	return t.GetStrategies1, t.GetStrategies2
}
func (t *testStrategyStore) DeployFromDB() error {
	t.DeployFromDBCount++
	return t.DeployFromDB1
}
func (t *testStrategyStore) SetSymbolInfo(strategyCode string, tickGroup TickGroup, tradingUnit float64) error {
	t.SetSymbolInfoHistory = append(t.SetSymbolInfoHistory, strategyCode)
	t.SetSymbolInfoHistory = append(t.SetSymbolInfoHistory, tickGroup)
	t.SetSymbolInfoHistory = append(t.SetSymbolInfoHistory, tradingUnit)
	t.SetSymbolInfoCount++
	return t.SetSymbolInfo1
}
func (t *testStrategyStore) Save(strategy *Strategy) error {
	t.SaveHistory = append(t.SaveHistory, strategy)
	t.SaveCount++
	return t.Save1
}
func (t *testStrategyStore) DeleteByCode(code string) error {
	t.DeleteByCodeHistory = append(t.DeleteByCodeHistory, code)
	t.DeleteByCodeCount++
	return t.DeleteByCode1
}

func Test_strategyStore_AddStrategyCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		logger                *testLogger
		store                 map[string]*Strategy
		arg1                  string
		arg2                  float64
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
		wantCashFlowCount     int
	}{
		{name: "該当する戦略がなければ変更しない",
			db:     &testDB{},
			logger: &testLogger{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  10_000,
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0,
			wantCashFlowCount:     0},
		{name: "該当する戦略の現金余力に加算できる",
			db:     &testDB{},
			logger: &testLogger{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", Cash: 100_000},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  10_000,
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", Cash: 110_000},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1,
			wantCashFlowCount:     1},
		{name: "該当する戦略の現金余力に減算できる",
			db:     &testDB{},
			logger: &testLogger{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", Cash: 100_000},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  -10_000,
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", Cash: 90_000},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1,
			wantCashFlowCount:     1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db, logger: test.logger}
			got1 := store.AddStrategyCash(test.arg1, test.arg2)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantStore, store.store) ||
				!reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) ||
				!reflect.DeepEqual(test.wantCashFlowCount, test.logger.CashFlowCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount, test.wantCashFlowCount,
					got1, store.store, test.db.SaveStrategyCount, test.logger.CashFlowCount)
			}
		})
	}
}

func Test_strategyStore_SetBasePrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		store                 map[string]*Strategy
		arg1                  string
		arg2                  float64
		arg3                  time.Time
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
	}{
		{name: "該当する戦略がなければ変更なし",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0},
		{name: "該当する戦略があれば更新する",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", BasePrice: 10_000, BasePriceDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local)},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.SetBasePrice(test.arg1, test.arg2, test.arg3)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_GetByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store map[string]*Strategy
		arg1  string
		want1 *Strategy
		want2 error
	}{
		{name: "指定したコードがなければerror",
			store: map[string]*Strategy{},
			arg1:  "strategy-code-001",
			want1: nil,
			want2: ErrNoData},
		{name: "mapがnilならerror",
			store: nil,
			arg1:  "strategy-code-001",
			want1: nil,
			want2: ErrNoData},
		{name: "指定したコードが存在すればStrategyを返す",
			store: map[string]*Strategy{"strategy-code-001": {Code: "strategy-code-001"}},
			arg1:  "strategy-code-001",
			want1: &Strategy{Code: "strategy-code-001"},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := strategyStore{store: test.store}
			got1, got2 := store.GetByCode(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_strategyStore_DeployFromDB(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		db        *testDB
		want1     error
		wantStore map[string]*Strategy
	}{
		{name: "dbがエラーを返したらエラーを返す",
			db:        &testDB{GetStrategies2: ErrUnknown},
			want1:     ErrUnknown,
			wantStore: nil},
		{name: "dbが空を返したらstoreを空にする",
			db:        &testDB{GetStrategies1: []*Strategy{}},
			want1:     nil,
			wantStore: map[string]*Strategy{}},
		{name: "dbが要素のある配列を返したらstoreに展開される",
			db: &testDB{GetStrategies1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}}},
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{db: test.db}
			got1 := store.DeployFromDB()
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantStore, got1, store.store)
			}
		})
	}
}

func Test_strategyStore_GetStrategies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store map[string]*Strategy
		want1 []*Strategy
		want2 error
	}{
		{name: "storeがnilなら空配列が返される",
			store: nil,
			want1: []*Strategy{},
			want2: nil},
		{name: "storeが空配列なら空配列が返される",
			store: map[string]*Strategy{},
			want1: []*Strategy{},
			want2: nil},
		{name: "storeにある要素がコードの昇順で返される",
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			want1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store}
			got1, got2 := store.GetStrategies()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_getStrategyStore(t *testing.T) {
	t.Parallel()

	db := &testDB{}
	logger := &testLogger{}
	want1 := &strategyStore{store: map[string]*Strategy{}, db: db, logger: logger}
	got1 := getStrategyStore(db, logger)

	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_strategyStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                    *testDB
		name                  string
		store                 map[string]*Strategy
		arg1                  *Strategy
		want1                 error
		wantStore             map[string]*Strategy
		wantSaveStrategyCount int
	}{
		{name: "引数がnilならエラー",
			db:        &testDB{},
			store:     map[string]*Strategy{},
			arg1:      nil,
			want1:     ErrNilArgument,
			wantStore: map[string]*Strategy{}},
		{name: "同一コードの注文がなければ追加",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  &Strategy{Code: "strategy-code-004"},
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
				"strategy-code-004": {Code: "strategy-code-004"}},
			wantSaveStrategyCount: 1},
		{name: "同一コードの注文があれば上書き",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  &Strategy{Code: "strategy-code-002", Cash: 100},
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", Cash: 100},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantSaveStrategyCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.Save(test.arg1)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSaveStrategyCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSaveStrategyCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_SetSymbolInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		store                 map[string]*Strategy
		arg1                  string
		arg2                  TickGroup
		arg3                  float64
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
	}{
		{name: "該当する戦略がなければ変更なし",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  TickGroupTopix100,
			arg3:  100,
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0},
		{name: "該当する戦略があれば更新する",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  TickGroupTopix100,
			arg3:  100,
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", TickGroup: TickGroupTopix100, TradingUnit: 100},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.SetSymbolInfo(test.arg1, test.arg2, test.arg3)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_SetContractPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		store                 map[string]*Strategy
		arg1                  string
		arg2                  float64
		arg3                  time.Time
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
	}{
		{name: "該当する戦略がなければ変更なし",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0},
		{name: "該当する戦略があれば更新する",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", BasePrice: 10_000, BasePriceDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local), LastContractPrice: 10_000, LastContractDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local)},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.SetContractPrice(test.arg1, test.arg2, test.arg3)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_SetMaxContractPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		store                 map[string]*Strategy
		arg1                  string
		arg2                  float64
		arg3                  time.Time
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
	}{
		{name: "該当する戦略がなければ変更なし",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0},
		{name: "該当する戦略があれば更新する",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", MaxContractPrice: 10_000, MaxContractDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local)},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.SetMaxContractPrice(test.arg1, test.arg2, test.arg3)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_SetMinContractPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		db                    *testDB
		store                 map[string]*Strategy
		arg1                  string
		arg2                  float64
		arg3                  time.Time
		want1                 error
		wantStore             map[string]*Strategy
		wantStrategySaveCount int
	}{
		{name: "該当する戦略がなければ変更なし",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 0},
		{name: "該当する戦略があれば更新する",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			arg2:  10_000,
			arg3:  time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local),
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002", MinContractPrice: 10_000, MinContractDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local)},
				"strategy-code-003": {Code: "strategy-code-003"}},
			wantStrategySaveCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.SetMinContractPrice(test.arg1, test.arg2, test.arg3)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantStrategySaveCount, test.db.SaveStrategyCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantStrategySaveCount,
					got1, store.store, test.db.SaveStrategyCount)
			}
		})
	}
}

func Test_strategyStore_DeleteByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                          string
		db                            *testDB
		store                         map[string]*Strategy
		arg1                          string
		want1                         error
		wantStore                     map[string]*Strategy
		wantDeleteStrategyByCodeCount int
	}{
		{name: "指定した戦略がstoreになければ何もしない",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-004",
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			wantDeleteStrategyByCodeCount: 0},
		{name: "指定した戦略がstoreにあれば、storeから消し、DBからも消す",
			db: &testDB{},
			store: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-002": {Code: "strategy-code-002"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			arg1:  "strategy-code-002",
			want1: nil,
			wantStore: map[string]*Strategy{
				"strategy-code-001": {Code: "strategy-code-001"},
				"strategy-code-003": {Code: "strategy-code-003"},
			},
			wantDeleteStrategyByCodeCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store, db: test.db}
			got1 := store.DeleteByCode(test.arg1)

			time.Sleep(100 * time.Millisecond) // 非同期処理が実行されることの確認のため少し待機

			if !errors.Is(got1, test.want1) ||
				!reflect.DeepEqual(test.wantStore, store.store) ||
				!reflect.DeepEqual(test.wantDeleteStrategyByCodeCount, test.db.DeleteStrategyByCodeCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantDeleteStrategyByCodeCount,
					got1, store.store, test.db.DeleteStrategyByCodeCount)
			}
		})
	}
}
