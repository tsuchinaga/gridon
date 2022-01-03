package gridon

import (
	"errors"
	"reflect"
	"testing"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
)

type testDB struct {
	IDB
	SaveStrategy1               error
	SaveStrategyCount           int
	SaveStrategyHistory         []interface{}
	DeleteStrategyByCode1       error
	DeleteStrategyByCodeCount   int
	DeleteStrategyByCodeHistory []interface{}
	SaveOrder1                  error
	SaveOrderCount              int
	SaveOrderHistory            []interface{}
	SavePosition1               error
	SavePositionCount           int
	SavePositionHistory         []interface{}
	GetStrategies1              []*Strategy
	GetStrategies2              error
	GetActiveOrders1            []*Order
	GetActiveOrders2            error
	GetActivePositions1         []*Position
	GetActivePositions2         error
	CleanupOrders1              error
	CleanupPositions1           error
}

func (t *testDB) GetStrategies() ([]*Strategy, error) {
	return t.GetStrategies1, t.GetStrategies2
}
func (t *testDB) GetActiveOrders() ([]*Order, error) {
	return t.GetActiveOrders1, t.GetActiveOrders2
}
func (t *testDB) GetActivePositions() ([]*Position, error) {
	return t.GetActivePositions1, t.GetActivePositions2
}
func (t *testDB) SaveStrategy(strategy *Strategy) error {
	t.SaveStrategyHistory = append(t.SaveStrategyHistory, strategy)
	t.SaveStrategyCount++
	return t.SaveStrategy1
}
func (t *testDB) DeleteStrategyByCode(code string) error {
	t.DeleteStrategyByCodeHistory = append(t.DeleteStrategyByCodeHistory, code)
	t.DeleteStrategyByCodeCount++
	return t.DeleteStrategyByCode1
}
func (t *testDB) SaveOrder(order *Order) error {
	t.SaveOrderHistory = append(t.SaveOrderHistory, order)
	t.SaveOrderCount++
	return t.SaveOrder1
}
func (t *testDB) SavePosition(position *Position) error {
	t.SavePositionHistory = append(t.SavePositionHistory, position)
	t.SavePositionCount++
	return t.SavePosition1
}
func (t *testDB) CleanupOrders() error    { return t.CleanupOrders1 }
func (t *testDB) CleanupPositions() error { return t.CleanupPositions1 }

func Test_db_SaveStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		logger         *testLogger
		dataset        []*Strategy
		arg            *Strategy
		want           error
		wantStrategies []*Strategy
	}{
		{name: "同じコードのデータがなければinsertされる",
			logger: &testLogger{},
			dataset: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			arg:  &Strategy{Code: "strategy-code-004"},
			want: nil,
			wantStrategies: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
				{Code: "strategy-code-004"},
			}},
		{name: "同じコードのデータがあったら上書きされる",
			logger: &testLogger{},
			dataset: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			arg:  &Strategy{Code: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
			want: nil,
			wantStrategies: []*Strategy{
				{Code: "strategy-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into strategies values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.SaveStrategy(test.arg)

			strategies := make([]*Strategy, 0)
			res, _ := d.Query("select * from strategies order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var strategy Strategy
				_ = document.StructScan(d, &strategy)
				strategies = append(strategies, &strategy)
				return nil
			})

			if !reflect.DeepEqual(test.wantStrategies, strategies) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantStrategies, got, strategies)
			}
		})
	}
}

func Test_db_SaveOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		logger     *testLogger
		dataset    []*Order
		arg        *Order
		want       error
		wantOrders []*Order
	}{
		{name: "同じコードのデータがなければinsertされる",
			logger: &testLogger{},
			dataset: []*Order{
				{Code: "order-code-001"},
				{Code: "order-code-002"},
				{Code: "order-code-003"},
			},
			arg:  &Order{Code: "order-code-004"},
			want: nil,
			wantOrders: []*Order{
				{Code: "order-code-001"},
				{Code: "order-code-002"},
				{Code: "order-code-003"},
				{Code: "order-code-004"},
			}},
		{name: "同じコードのデータがあったら上書きされる",
			logger: &testLogger{},
			dataset: []*Order{
				{Code: "order-code-001"},
				{Code: "order-code-002"},
				{Code: "order-code-003"},
			},
			arg:  &Order{Code: "order-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
			want: nil,
			wantOrders: []*Order{
				{Code: "order-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
				{Code: "order-code-002"},
				{Code: "order-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into orders values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.SaveOrder(test.arg)

			orders := make([]*Order, 0)
			res, _ := d.Query("select * from orders order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var order Order
				_ = document.StructScan(d, &order)
				orders = append(orders, &order)
				return nil
			})

			if !reflect.DeepEqual(test.wantOrders, orders) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantOrders, got, orders)
			}
		})
	}
}

func Test_db_SavePosition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		logger        *testLogger
		dataset       []*Position
		arg           *Position
		want          error
		wantPositions []*Position
	}{
		{name: "同じコードのデータがなければinsertされる",
			logger: &testLogger{},
			dataset: []*Position{
				{Code: "position-code-001"},
				{Code: "position-code-002"},
				{Code: "position-code-003"},
			},
			arg:  &Position{Code: "position-code-004"},
			want: nil,
			wantPositions: []*Position{
				{Code: "position-code-001"},
				{Code: "position-code-002"},
				{Code: "position-code-003"},
				{Code: "position-code-004"},
			}},
		{name: "同じコードのデータがあったら上書きされる",
			logger: &testLogger{},
			dataset: []*Position{
				{Code: "position-code-001"},
				{Code: "position-code-002"},
				{Code: "position-code-003"},
			},
			arg:  &Position{Code: "position-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
			want: nil,
			wantPositions: []*Position{
				{Code: "position-code-001", SymbolCode: "1475", Exchange: ExchangeToushou},
				{Code: "position-code-002"},
				{Code: "position-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into positions values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.SavePosition(test.arg)

			positions := make([]*Position, 0)
			res, _ := d.Query("select * from positions order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var position Position
				_ = document.StructScan(d, &position)
				positions = append(positions, &position)
				return nil
			})

			if !reflect.DeepEqual(test.wantPositions, positions) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantPositions, got, positions)
			}
		})
	}
}

func Test_db_GetStrategies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		dataset []*Strategy
		want1   []*Strategy
		want2   error
	}{
		{name: "戦略がなければ空スライスを返す",
			dataset: []*Strategy{},
			want1:   []*Strategy{},
			want2:   nil},
		{name: "戦略があればstrategyに詰めてスライスに入れて返す",
			dataset: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			want1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into strategies values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			store := &db{db: d}
			got1, got2 := store.GetStrategies()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_db_GetActiveOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		dataset []*Order
		want1   []*Order
		want2   error
	}{
		{name: "注文がなければ空スライスを返す",
			dataset: []*Order{},
			want1:   []*Order{},
			want2:   nil},
		{name: "有効な注文があればOrderに詰めてスライスに入れて返す",
			dataset: []*Order{
				{Code: "order-code-001", Status: OrderStatusUnspecified},
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-003", Status: OrderStatusDone},
				{Code: "order-code-004", Status: OrderStatusCanceled},
				{Code: "order-code-005", Status: OrderStatusInOrder},
			},
			want1: []*Order{
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-005", Status: OrderStatusInOrder},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into orders values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			store := &db{db: d}
			got1, got2 := store.GetActiveOrders()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_db_GetActivePositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		dataset []*Position
		want1   []*Position
		want2   error
	}{
		{name: "注文がなければ空スライスを返す",
			dataset: []*Position{},
			want1:   []*Position{},
			want2:   nil},
		{name: "有効なポジションがあればPositionに詰めてスライスに入れて返す",
			dataset: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 4},
				{Code: "position-code-002", OwnedQuantity: 0, HoldQuantity: 0},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 2},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 0},
			},
			want1: []*Position{
				{Code: "position-code-001", OwnedQuantity: 4, HoldQuantity: 4},
				{Code: "position-code-003", OwnedQuantity: 4, HoldQuantity: 2},
				{Code: "position-code-004", OwnedQuantity: 4, HoldQuantity: 0},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into positions values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			store := &db{db: d}
			got1, got2 := store.GetActivePositions()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_getDB(t *testing.T) {
	t.Parallel()

	gdb, err := openDB(":memory:")
	if err != nil {
		t.Errorf("%s error\nerror: %s\n", t.Name(), err)
	}
	logger := &logger{}
	want1 := &db{db: gdb, logger: logger}
	got1, err := getDB(":memory:", logger)
	if err != nil {
		t.Errorf("%s error\nerror: %s\n", t.Name(), err)
	}

	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_db_CleanupOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		logger     *testLogger
		dataset    []*Order
		want       error
		wantOrders []*Order
	}{
		{name: "データがなければ何もしない",
			logger:     &testLogger{},
			dataset:    []*Order{},
			want:       nil,
			wantOrders: []*Order{}},
		{name: "削除対象のデータがなければ何もしない",
			logger: &testLogger{},
			dataset: []*Order{
				{Code: "order-code-001", Status: OrderStatusInOrder},
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-003", Status: OrderStatusInOrder},
			},
			want: nil,
			wantOrders: []*Order{
				{Code: "order-code-001", Status: OrderStatusInOrder},
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-003", Status: OrderStatusInOrder},
			}},
		{name: "削除対象のデータがあれば削除する",
			logger: &testLogger{},
			dataset: []*Order{
				{Code: "order-code-001", Status: OrderStatusDone},
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-003", Status: OrderStatusCanceled},
				{Code: "order-code-004", Status: OrderStatusUnspecified},
				{Code: "order-code-005", Status: OrderStatusInOrder},
			},
			want: nil,
			wantOrders: []*Order{
				{Code: "order-code-002", Status: OrderStatusInOrder},
				{Code: "order-code-005", Status: OrderStatusInOrder},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into orders values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.CleanupOrders()

			orders := make([]*Order, 0)
			res, _ := d.Query("select * from orders order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var order Order
				_ = document.StructScan(d, &order)
				orders = append(orders, &order)
				return nil
			})

			if !reflect.DeepEqual(test.wantOrders, orders) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantOrders, got, orders)
			}
		})
	}
}

func Test_db_CleanupPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		logger        *testLogger
		dataset       []*Position
		want          error
		wantPositions []*Position
	}{
		{name: "データがなければ何もしない",
			logger:        &testLogger{},
			dataset:       []*Position{},
			want:          nil,
			wantPositions: []*Position{}},
		{name: "削除対象のデータがなければ何もしない",
			logger: &testLogger{},
			dataset: []*Position{
				{Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 100},
				{Code: "position-code-002", OwnedQuantity: 100, HoldQuantity: 0},
				{Code: "position-code-003", OwnedQuantity: 100, HoldQuantity: 50},
			},
			want: nil,
			wantPositions: []*Position{
				{Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 100},
				{Code: "position-code-002", OwnedQuantity: 100, HoldQuantity: 0},
				{Code: "position-code-003", OwnedQuantity: 100, HoldQuantity: 50},
			}},
		{name: "削除対象のデータがあれば削除する",
			logger: &testLogger{},
			dataset: []*Position{
				{Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 100},
				{Code: "position-code-002", OwnedQuantity: 0, HoldQuantity: 0},
				{Code: "position-code-003", OwnedQuantity: 0, HoldQuantity: 0},
			},
			want: nil,
			wantPositions: []*Position{
				{Code: "position-code-001", OwnedQuantity: 100, HoldQuantity: 100},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into positions values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.CleanupPositions()

			positions := make([]*Position, 0)
			res, _ := d.Query("select * from positions order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var position Position
				_ = document.StructScan(d, &position)
				positions = append(positions, &position)
				return nil
			})

			if !reflect.DeepEqual(test.wantPositions, positions) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantPositions, got, positions)
			}
		})
	}
}

func Test_db_DeleteStrategyByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		logger         *testLogger
		dataset        []*Strategy
		arg            string
		want           error
		wantStrategies []*Strategy
	}{
		{name: "同じコードのデータがなければなにもしない",
			logger: &testLogger{},
			dataset: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			arg:  "strategy-code-004",
			want: nil,
			wantStrategies: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			}},
		{name: "同じコードのデータがあったら削除される",
			logger: &testLogger{},
			dataset: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"},
			},
			arg:  "strategy-code-002",
			want: nil,
			wantStrategies: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-003"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			d, _ := openDB(":memory:")
			defer d.Close()
			for _, data := range test.dataset {
				if err := d.Exec(`insert into strategies values ?`, data); err != nil {
					t.Errorf("%s insert error\n%+v\n", t.Name(), err)
				}
			}

			db := &db{db: d, logger: test.logger}
			got := db.DeleteStrategyByCode(test.arg)

			strategies := make([]*Strategy, 0)
			res, _ := d.Query("select * from strategies order by code")
			defer res.Close()
			_ = res.Iterate(func(d types.Document) error {
				var strategy Strategy
				_ = document.StructScan(d, &strategy)
				strategies = append(strategies, &strategy)
				return nil
			})

			if !reflect.DeepEqual(test.wantStrategies, strategies) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantStrategies, got, strategies)
			}
		})
	}
}
