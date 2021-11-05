package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testStrategyStore struct {
	IStrategyStore
	AddStrategyCash1       error
	AddStrategyCashHistory []interface{}
	AddStrategyCashCount   int
	SetContract1           error
	SetContractHistory     []interface{}
	SetContractCount       int
	GetByCode1             *Strategy
	GetByCode2             error
}

func (t *testStrategyStore) GetByCode(string) (*Strategy, error) {
	return t.GetByCode1, t.GetByCode2
}
func (t *testStrategyStore) AddStrategyCash(strategyCode string, cashDiff float64) error {
	t.AddStrategyCashHistory = append(t.AddStrategyCashHistory, strategyCode)
	t.AddStrategyCashHistory = append(t.AddStrategyCashHistory, cashDiff)
	t.AddStrategyCashCount++
	return t.AddStrategyCash1
}
func (t *testStrategyStore) SetContract(strategyCode string, contractPrice float64, contractDateTime time.Time) error {
	t.SetContractHistory = append(t.SetContractHistory, strategyCode)
	t.SetContractHistory = append(t.SetContractHistory, contractPrice)
	t.SetContractHistory = append(t.SetContractHistory, contractDateTime)
	t.SetContractCount++
	return t.SetContract1
}

func Test_strategyStore_AddStrategyCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     map[string]*Strategy
		arg1      string
		arg2      float64
		want1     error
		wantStore map[string]*Strategy
	}{
		{name: "該当する戦略がなければ変更しない",
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
				"strategy-code-003": {Code: "strategy-code-003"},
			}},
		{name: "該当する戦略の現金余力に加算できる",
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
				"strategy-code-003": {Code: "strategy-code-003"},
			}},
		{name: "該当する戦略の現金余力に減算できる",
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
				"strategy-code-003": {Code: "strategy-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store}
			got1 := store.AddStrategyCash(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_strategyStore_SetContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     map[string]*Strategy
		arg1      string
		arg2      float64
		arg3      time.Time
		want1     error
		wantStore map[string]*Strategy
	}{
		{name: "該当する戦略がなければ変更なし",
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
				"strategy-code-003": {Code: "strategy-code-003"},
			}},
		{name: "該当する戦略があれば更新する",
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
				"strategy-code-002": {Code: "strategy-code-002", LastContractPrice: 10_000, LastContractDateTime: time.Date(2021, 10, 26, 10, 0, 0, 0, time.Local)},
				"strategy-code-003": {Code: "strategy-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &strategyStore{store: test.store}
			got1 := store.SetContract(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
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
