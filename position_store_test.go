package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testPositionStore struct {
	IPositionStore
	Save1                                   error
	SaveHistory                             []interface{}
	SaveCount                               int
	ExitContract1                           error
	ExitContractHistory                     []interface{}
	ExitContractCount                       int
	Release1                                error
	ReleaseHistory                          []interface{}
	ReleaseCount                            int
	GetActivePositionsByStrategyCode1       []*Position
	GetActivePositionsByStrategyCode2       error
	GetActivePositionsByStrategyCodeCount   int
	GetActivePositionsByStrategyCodeHistory []interface{}
	Hold1                                   error
	HoldCount                               int
	HoldHistory                             []interface{}
}

func (t *testPositionStore) Save(position *Position) error {
	t.SaveHistory = append(t.SaveHistory, position)
	t.SaveCount++
	return t.Save1
}
func (t *testPositionStore) ExitContract(positionCode string, quantity float64) error {
	t.ExitContractHistory = append(t.ExitContractHistory, positionCode)
	t.ExitContractHistory = append(t.ExitContractHistory, quantity)
	t.ExitContractCount++
	return t.ExitContract1
}
func (t *testPositionStore) Release(positionCode string, quantity float64) error {
	t.ReleaseHistory = append(t.ReleaseHistory, positionCode)
	t.ReleaseHistory = append(t.ReleaseHistory, quantity)
	t.ReleaseCount++
	return t.Release1
}
func (t *testPositionStore) GetActivePositionsByStrategyCode(strategyCode string) ([]*Position, error) {
	t.GetActivePositionsByStrategyCodeHistory = append(t.GetActivePositionsByStrategyCodeHistory, strategyCode)
	t.GetActivePositionsByStrategyCodeCount++
	return t.GetActivePositionsByStrategyCode1, t.GetActivePositionsByStrategyCode2
}
func (t *testPositionStore) Hold(positionCode string, quantity float64) error {
	t.HoldHistory = append(t.HoldHistory, positionCode)
	t.HoldHistory = append(t.HoldHistory, quantity)
	t.HoldCount++
	return t.Hold1
}

func Test_positionStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                    *testDB
		name                  string
		store                 map[string]*Position
		arg1                  *Position
		want1                 error
		wantStore             map[string]*Position
		wantSavePositionCount int
	}{
		{name: "引数がnilならerr",
			db:        &testDB{},
			store:     map[string]*Position{},
			arg1:      nil,
			want1:     ErrNilArgument,
			wantStore: map[string]*Position{}},
		{name: "codeが一致するpositionがなければ追加される",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001"},
				"position-code-002": {Code: "position-code-002"},
				"position-code-003": {Code: "position-code-003"},
			},
			arg1:  &Position{Code: "position-code-004"},
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001"},
				"position-code-002": {Code: "position-code-002"},
				"position-code-003": {Code: "position-code-003"},
				"position-code-004": {Code: "position-code-004"}},
			wantSavePositionCount: 1},
		{name: "codeが一致するpositionがあれば更新される",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001"},
				"position-code-002": {Code: "position-code-002"},
				"position-code-003": {Code: "position-code-003"},
			},
			arg1:  &Position{Code: "position-code-002", HoldQuantity: 100},
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001"},
				"position-code-002": {Code: "position-code-002", HoldQuantity: 100},
				"position-code-003": {Code: "position-code-003"}},
			wantSavePositionCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store, db: test.db}
			got1 := store.Save(test.arg1)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSavePositionCount, test.db.SavePositionCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSavePositionCount,
					got1, store.store, test.db.SavePositionCount)
			}
		})
	}
}

func Test_positionStore_ExitContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                    *testDB
		name                  string
		store                 map[string]*Position
		arg1                  string
		arg2                  float64
		want1                 error
		wantStore             map[string]*Position
		wantSavePositionCount int
	}{
		{name: "該当するポジションがなければ何もしない",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			},
			arg1:  "position-code-000",
			arg2:  100,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200}}},
		{name: "該当するポジションの保有数と拘束数を減算する",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			},
			arg1:  "position-code-002",
			arg2:  200,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 100, HoldQuantity: 100},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200}},
			wantSavePositionCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store, db: test.db}
			got1 := store.ExitContract(test.arg1, test.arg2)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSavePositionCount, test.db.SavePositionCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSavePositionCount,
					got1, store.store, test.db.SavePositionCount)
			}
		})
	}
}

func Test_positionStore_Release(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                    *testDB
		name                  string
		store                 map[string]*Position
		arg1                  string
		arg2                  float64
		want1                 error
		wantStore             map[string]*Position
		wantSavePositionCount int
	}{
		{name: "該当するポジションがなければ何もしない",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			},
			arg1:  "position-code-000",
			arg2:  100,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200}}},
		{name: "該当するポジションの拘束数を減算する",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 300},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			},
			arg1:  "position-code-002",
			arg2:  200,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 300, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 300, HoldQuantity: 100},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200}},
			wantSavePositionCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store, db: test.db}
			got1 := store.Release(test.arg1, test.arg2)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSavePositionCount, test.db.SavePositionCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSavePositionCount,
					got1, store.store, test.db.SavePositionCount)
			}
		})
	}
}

func Test_positionStore_GetActivePositionsByStrategyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store map[string]*Position
		arg1  string
		want1 []*Position
		want2 error
	}{
		{name: "対象の戦略コードのポジションがなければ空配列",
			store: map[string]*Position{},
			arg1:  "strategy-code-001",
			want1: []*Position{},
			want2: nil},
		{name: "対象の戦略コードのポジションがあっても生きていないポジションならスキップ",
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", StrategyCode: "strategy-code-001", ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)},
				"position-code-002": {Code: "position-code-002", StrategyCode: "strategy-code-001", ContractDateTime: time.Date(2021, 10, 29, 10, 1, 0, 0, time.Local)},
				"position-code-003": {Code: "position-code-003", StrategyCode: "strategy-code-001", ContractDateTime: time.Date(2021, 10, 29, 10, 2, 0, 0, time.Local)},
			},
			arg1:  "strategy-code-001",
			want1: []*Position{},
			want2: nil},
		{name: "ポジションは約定日時昇順で返される",
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)},
				"position-code-002": {Code: "position-code-002", StrategyCode: "strategy-code-001", OwnedQuantity: 0, ContractDateTime: time.Date(2021, 10, 29, 10, 1, 0, 0, time.Local)},
				"position-code-003": {Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 30, ContractDateTime: time.Date(2021, 10, 29, 10, 2, 0, 0, time.Local)},
			},
			arg1: "strategy-code-001",
			want1: []*Position{
				{Code: "position-code-001", StrategyCode: "strategy-code-001", OwnedQuantity: 100, ContractDateTime: time.Date(2021, 10, 29, 10, 0, 0, 0, time.Local)},
				{Code: "position-code-003", StrategyCode: "strategy-code-001", OwnedQuantity: 30, ContractDateTime: time.Date(2021, 10, 29, 10, 2, 0, 0, time.Local)},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store}
			got1, got2 := store.GetActivePositionsByStrategyCode(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_positionStore_Hold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                    *testDB
		name                  string
		store                 map[string]*Position
		arg1                  string
		arg2                  float64
		want1                 error
		wantStore             map[string]*Position
		wantSavePositionCount int
	}{
		{name: "指定したpositionCodeがなければ何もしない",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 200, HoldQuantity: 0},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 0},
			},
			arg1:  "position-code-000",
			arg2:  100,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 200, HoldQuantity: 0},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 0}}},
		{name: "指定したpositionCodeがあれば、拘束数に加算する",
			db: &testDB{},
			store: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 0},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 200, HoldQuantity: 0},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 0}},
			arg1:  "position-code-001",
			arg2:  100,
			want1: nil,
			wantStore: map[string]*Position{
				"position-code-001": {Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 100},
				"position-code-002": {Code: "position-code-002", OwnedQuantity: 200, HoldQuantity: 0},
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 0}},
			wantSavePositionCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store, db: test.db}
			got1 := store.Hold(test.arg1, test.arg2)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSavePositionCount, test.db.SavePositionCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSavePositionCount,
					got1, store.store, test.db.SavePositionCount)
			}
		})
	}
}
