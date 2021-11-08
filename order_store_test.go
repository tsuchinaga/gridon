package gridon

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testOrderStore struct {
	IOrderStore
	GetActiveOrdersByStrategyCode1       []*Order
	GetActiveOrdersByStrategyCode2       error
	GetActiveOrdersByStrategyCodeCount   int
	GetActiveOrdersByStrategyCodeHistory []interface{}
	Save1                                error
	SaveCount                            int
	SaveHistory                          []interface{}
}

func (t *testOrderStore) GetActiveOrdersByStrategyCode(strategyCode string) ([]*Order, error) {
	t.GetActiveOrdersByStrategyCodeHistory = append(t.GetActiveOrdersByStrategyCodeHistory, strategyCode)
	t.GetActiveOrdersByStrategyCodeCount++
	return t.GetActiveOrdersByStrategyCode1, t.GetActiveOrdersByStrategyCode2
}
func (t *testOrderStore) Save(order *Order) error {
	t.SaveHistory = append(t.SaveHistory, order)
	t.SaveCount++
	return t.Save1
}

func Test_orderStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		db                 *testDB
		name               string
		store              map[string]*Order
		arg1               *Order
		want1              error
		wantStore          map[string]*Order
		wantSaveOrderCount int
	}{
		{name: "引数がnilならエラー",
			db:        &testDB{},
			store:     map[string]*Order{},
			arg1:      nil,
			want1:     ErrNilArgument,
			wantStore: map[string]*Order{}},
		{name: "同一コードの注文がなければ追加",
			db: &testDB{},
			store: map[string]*Order{
				"order-code-001": {Code: "order-code-001"},
				"order-code-002": {Code: "order-code-002"},
				"order-code-003": {Code: "order-code-003"},
			},
			arg1:  &Order{Code: "order-code-004"},
			want1: nil,
			wantStore: map[string]*Order{
				"order-code-001": {Code: "order-code-001"},
				"order-code-002": {Code: "order-code-002"},
				"order-code-003": {Code: "order-code-003"},
				"order-code-004": {Code: "order-code-004"}},
			wantSaveOrderCount: 1},
		{name: "同一コードの注文があれば上書き",
			db: &testDB{},
			store: map[string]*Order{
				"order-code-001": {Code: "order-code-001"},
				"order-code-002": {Code: "order-code-002"},
				"order-code-003": {Code: "order-code-003"},
			},
			arg1:  &Order{Code: "order-code-002", ContractQuantity: 100},
			want1: nil,
			wantStore: map[string]*Order{
				"order-code-001": {Code: "order-code-001"},
				"order-code-002": {Code: "order-code-002", ContractQuantity: 100},
				"order-code-003": {Code: "order-code-003"}},
			wantSaveOrderCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &orderStore{store: test.store, db: test.db}
			got1 := store.Save(test.arg1)

			time.Sleep(100 * time.Millisecond)

			if !errors.Is(got1, test.want1) || !reflect.DeepEqual(test.wantStore, store.store) || !reflect.DeepEqual(test.wantSaveOrderCount, test.db.SaveOrderCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.wantStore, test.wantSaveOrderCount,
					got1, store.store, test.db.SaveOrderCount)
			}
		})
	}
}

func Test_orderStore_GetActiveOrdersByStrategyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store map[string]*Order
		arg1  string
		want1 []*Order
		want2 error
	}{
		{name: "対象の戦略コードの注文がなければnil",
			store: map[string]*Order{},
			arg1:  "strategy-code-001",
			want1: []*Order{},
			want2: nil},
		{name: "対象の戦略コードの注文があっても生きていない注文ならスキップ",
			store: map[string]*Order{
				"order-code-001": {Code: "order-code-001", StrategyCode: "strategy-code-001", Status: OrderStatusDone},
				"order-code-002": {Code: "order-code-002", StrategyCode: "strategy-code-001", Status: OrderStatusCanceled},
				"order-code-003": {Code: "order-code-003", StrategyCode: "strategy-code-001", Status: OrderStatusUnspecified},
			},
			arg1:  "strategy-code-001",
			want1: []*Order{},
			want2: nil},
		{name: "注文は注文日時昇順で返される",
			store: map[string]*Order{
				"order-code-001": {Code: "order-code-001", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 9, 0, 0, 0, time.Local)},
				"order-code-002": {Code: "order-code-002", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 10, 0, 0, 0, time.Local)},
				"order-code-003": {Code: "order-code-003", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 11, 0, 0, 0, time.Local)},
			},
			arg1: "strategy-code-001",
			want1: []*Order{
				{Code: "order-code-001", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 9, 0, 0, 0, time.Local)},
				{Code: "order-code-002", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 10, 0, 0, 0, time.Local)},
				{Code: "order-code-003", StrategyCode: "strategy-code-001", Status: OrderStatusInOrder, OrderDateTime: time.Date(2021, 10, 25, 11, 0, 0, 0, time.Local)},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &orderStore{store: test.store}
			got1, got2 := store.GetActiveOrdersByStrategyCode(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}
