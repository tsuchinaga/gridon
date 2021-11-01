package gridon

import (
	"reflect"
	"testing"
	"time"
)

func Test_Order_IsActive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *Order
		want1 bool
	}{
		{name: "注文中の注文が有効な注文",
			order: &Order{Status: OrderStatusInOrder},
			want1: true},
		{name: "約定済みの注文は有効ではない",
			order: &Order{Status: OrderStatusDone},
			want1: false},
		{name: "取消済みの注文は有効ではない",
			order: &Order{Status: OrderStatusCanceled},
			want1: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.order.IsActive()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Order_IsEqualSecurityOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *Order
		arg1  SecurityOrder
		want1 bool
	}{
		{name: "各項目が一致していたらtrue",
			order: &Order{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			arg1:  SecurityOrder{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			want1: true},
		{name: "ステータスが一致していなければfalse",
			order: &Order{Status: OrderStatusInOrder, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			arg1:  SecurityOrder{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			want1: false},
		{name: "約定数量が一致していなければfalse",
			order: &Order{Status: OrderStatusDone, ContractQuantity: 100, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			arg1:  SecurityOrder{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			want1: false},
		{name: "約定日時が一致していなければfalse",
			order: &Order{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 10, 0, 0, time.Local), CancelDateTime: time.Time{}},
			arg1:  SecurityOrder{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			want1: false},
		{name: "取消日時が一致していなければfalse",
			order: &Order{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Time{}},
			arg1:  SecurityOrder{Status: OrderStatusDone, ContractQuantity: 200, ContractDateTime: time.Date(2021, 10, 28, 9, 45, 0, 0, time.Local), CancelDateTime: time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local)},
			want1: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.order.IsEqualSecurityOrder(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Order_ReflectSecurityOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		order     *Order
		arg1      SecurityOrder
		wantOrder *Order
	}{
		{name: "securityOrderの内容を反映する",
			order: &Order{},
			arg1: SecurityOrder{
				Code:             "order-code-001",
				Status:           OrderStatusInOrder,
				SymbolCode:       "1475",
				Exchange:         ExchangeToushou,
				Product:          ProductMargin,
				MarginTradeType:  MarginTradeTypeDay,
				TradeType:        TradeTypeEntry,
				Side:             SideBuy,
				Price:            2050,
				OrderQuantity:    30,
				ContractQuantity: 10,
				AccountType:      AccountTypeSpecific,
				ExpireDay:        time.Date(2021, 10, 28, 0, 0, 0, 0, time.Local),
				OrderDateTime:    time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Date(2021, 10, 28, 10, 1, 0, 0, time.Local),
				CancelDateTime:   time.Time{},
				Contracts: []Contract{
					{
						OrderCode: "order-code-001", PositionCode: "position-code-001",
						Price:            2050,
						Quantity:         10,
						ContractDateTime: time.Date(2021, 10, 28, 10, 1, 0, 0, time.Local),
					},
				}},
			wantOrder: &Order{
				Code:             "",
				StrategyCode:     "",
				SymbolCode:       "",
				Exchange:         "",
				Status:           OrderStatusInOrder,
				Product:          "",
				MarginTradeType:  "",
				TradeType:        "",
				Side:             "",
				Price:            0,
				OrderQuantity:    0,
				ContractQuantity: 10,
				AccountType:      "",
				OrderDateTime:    time.Date(2021, 10, 28, 10, 0, 0, 0, time.Local),
				ContractDateTime: time.Date(2021, 10, 28, 10, 1, 0, 0, time.Local),
				CancelDateTime:   time.Time{},
				Contracts: []Contract{
					{
						OrderCode: "order-code-001", PositionCode: "position-code-001",
						Price:            2050,
						Quantity:         10,
						ContractDateTime: time.Date(2021, 10, 28, 10, 1, 0, 0, time.Local),
					}},
				HoldPositions: nil,
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.order.ReflectSecurityOrder(test.arg1)
			if !reflect.DeepEqual(test.wantOrder, test.order) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrder, test.order)
			}
		})
	}
}

func Test_Order_HasContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *Order
		arg1  Contract
		want1 bool
	}{
		{name: "contractsがnilならfalse",
			order: &Order{},
			arg1:  Contract{PositionCode: "position-code-001"},
			want1: false},
		{name: "contractsが空配列ならfalse",
			order: &Order{Contracts: []Contract{}},
			arg1:  Contract{PositionCode: "position-code-001"},
			want1: false},
		{name: "contractsにPositionCodeが一致するContractがなかったらfalse",
			order: &Order{Contracts: []Contract{{PositionCode: "position-code-001"}, {PositionCode: "position-code-002"}}},
			arg1:  Contract{PositionCode: "position-code-003"},
			want1: false},
		{name: "contractsにPositionCodeが一致するContractがあったらtrue",
			order: &Order{Contracts: []Contract{{PositionCode: "position-code-001"}, {PositionCode: "position-code-002"}}},
			arg1:  Contract{PositionCode: "position-code-002"},
			want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.order.HasContract(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Order_ContractDiff(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *Order
		arg1  SecurityOrder
		want1 []Contract
	}{
		{name: "SecurityOrderの約定がnilなら空配列",
			order: &Order{Contracts: []Contract{}},
			arg1:  SecurityOrder{Contracts: nil},
			want1: []Contract{}},
		{name: "SecurityOrderに約定がなければ空配列",
			order: &Order{Contracts: []Contract{}},
			arg1:  SecurityOrder{Contracts: []Contract{}},
			want1: []Contract{}},
		{name: "SecurityOrderにある約定がOrderになければ配列に入れて返される",
			order: &Order{Contracts: []Contract{}},
			arg1: SecurityOrder{Contracts: []Contract{
				{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2000, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 9, 0, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-002", Price: 2000, Quantity: 3, ContractDateTime: time.Date(2021, 10, 28, 9, 1, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-003", Price: 2000, Quantity: 2, ContractDateTime: time.Date(2021, 10, 28, 9, 2, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-004", Price: 2000, Quantity: 1, ContractDateTime: time.Date(2021, 10, 28, 9, 3, 0, 0, time.Local)},
			}},
			want1: []Contract{
				{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2000, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 9, 0, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-002", Price: 2000, Quantity: 3, ContractDateTime: time.Date(2021, 10, 28, 9, 1, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-003", Price: 2000, Quantity: 2, ContractDateTime: time.Date(2021, 10, 28, 9, 2, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-004", Price: 2000, Quantity: 1, ContractDateTime: time.Date(2021, 10, 28, 9, 3, 0, 0, time.Local)},
			}},
		{name: "SecurityOrderにある約定がすべてOrderにあれば空配列",
			order: &Order{Contracts: []Contract{
				{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2000, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 9, 0, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-002", Price: 2000, Quantity: 3, ContractDateTime: time.Date(2021, 10, 28, 9, 1, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-003", Price: 2000, Quantity: 2, ContractDateTime: time.Date(2021, 10, 28, 9, 2, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-004", Price: 2000, Quantity: 1, ContractDateTime: time.Date(2021, 10, 28, 9, 3, 0, 0, time.Local)},
			}},
			arg1: SecurityOrder{Contracts: []Contract{
				{OrderCode: "order-code-001", PositionCode: "position-code-001", Price: 2000, Quantity: 4, ContractDateTime: time.Date(2021, 10, 28, 9, 0, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-002", Price: 2000, Quantity: 3, ContractDateTime: time.Date(2021, 10, 28, 9, 1, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-003", Price: 2000, Quantity: 2, ContractDateTime: time.Date(2021, 10, 28, 9, 2, 0, 0, time.Local)},
				{OrderCode: "order-code-001", PositionCode: "position-code-004", Price: 2000, Quantity: 1, ContractDateTime: time.Date(2021, 10, 28, 9, 3, 0, 0, time.Local)},
			}},
			want1: []Contract{}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.order.ContractDiff(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Position_IsActive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *Position
		want1    bool
	}{
		{name: "保有数量が0ならfalse", position: &Position{OwnedQuantity: 0}, want1: false},
		{name: "保有数量が0より大きければtrue", position: &Position{OwnedQuantity: 0.1}, want1: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.position.IsActive()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Position_LeaveQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *Position
		want1    float64
	}{
		{name: "保有数量、拘束数量が0なら0", position: &Position{OwnedQuantity: 0, HoldQuantity: 0}, want1: 0},
		{name: "保有数量 - 拘束数量が0なら0", position: &Position{OwnedQuantity: 100, HoldQuantity: 100}, want1: 0},
		{name: "保有数量 - 拘束数量が100なら100", position: &Position{OwnedQuantity: 100, HoldQuantity: 0}, want1: 100},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.position.LeaveQuantity()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
