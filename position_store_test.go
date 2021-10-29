package gridon

import (
	"errors"
	"reflect"
	"testing"
)

type testPositionStore struct {
	IPositionStore
	Save1               error
	SaveHistory         []interface{}
	SaveCount           int
	ExitContract1       error
	ExitContractHistory []interface{}
	ExitContractCount   int
	Release1            error
	ReleaseHistory      []interface{}
	ReleaseCount        int
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

func Test_positionStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     map[string]*Position
		arg1      *Position
		want1     error
		wantStore map[string]*Position
	}{
		{name: "引数がnilならerr",
			store:     map[string]*Position{},
			arg1:      nil,
			want1:     ErrNilArgument,
			wantStore: map[string]*Position{}},
		{name: "codeが一致するpositionがなければ追加される",
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
				"position-code-004": {Code: "position-code-004"},
			}},
		{name: "codeが一致するpositionがあれば更新される",
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
				"position-code-003": {Code: "position-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store}
			got1 := store.Save(test.arg1)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantStore, got1, store.store)
			}
		})
	}
}

func Test_positionStore_ExitContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     map[string]*Position
		arg1      string
		arg2      float64
		want1     error
		wantStore map[string]*Position
	}{
		{name: "該当するポジションがなければ何もしない",
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
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			}},
		{name: "該当するポジションの保有数と拘束数を減算する",
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
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store}
			got1 := store.ExitContract(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantStore, got1, store.store)
			}
		})
	}
}

func Test_positionStore_Release(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     map[string]*Position
		arg1      string
		arg2      float64
		want1     error
		wantStore map[string]*Position
	}{
		{name: "該当するポジションがなければ何もしない",
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
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			}},
		{name: "該当するポジションの拘束数を減算する",
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
				"position-code-003": {Code: "position-code-003", OwnedQuantity: 300, HoldQuantity: 200},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &positionStore{store: test.store}
			got1 := store.Release(test.arg1, test.arg2)
			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantStore, got1, store.store)
			}
		})
	}
}
