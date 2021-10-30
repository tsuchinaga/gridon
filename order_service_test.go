package gridon

import (
	"reflect"
	"testing"
)

func Test_orderService_CancelAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		orderStore             *testOrderStore
		kabusAPI               *testKabusAPI
		arg1                   *Strategy
		want1                  error
		wantCancelOrderHistory []interface{}
	}{
		{name: "引数がnilならエラー",
			orderStore:             &testOrderStore{},
			kabusAPI:               &testKabusAPI{},
			arg1:                   nil,
			want1:                  ErrNilArgument,
			wantCancelOrderHistory: nil},
		{name: "注文一覧取得に失敗したらエラー",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode2: ErrUnknown},
			kabusAPI:               &testKabusAPI{},
			arg1:                   &Strategy{Code: "strategy-code-001"},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: nil},
		{name: "注文一覧が空なら何もしない",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{}},
			kabusAPI:               &testKabusAPI{},
			arg1:                   &Strategy{Code: "strategy-code-001"},
			want1:                  nil,
			wantCancelOrderHistory: nil},
		{name: "取消注文の実行に失敗したらエラー",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder2: ErrUnknown},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}},
			want1:                  ErrUnknown,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001"}},
		{name: "取消注文の実行結果が失敗でもエラーなし",
			orderStore:             &testOrderStore{GetActiveOrdersByStrategyCode1: []*Order{{Code: "order-code-001"}, {Code: "order-code-002"}, {Code: "order-code-003"}}},
			kabusAPI:               &testKabusAPI{CancelOrder1: OrderResult{Result: false, ResultCode: -1}},
			arg1:                   &Strategy{Code: "strategy-code-001", Account: Account{Password: "Password1234"}},
			want1:                  nil,
			wantCancelOrderHistory: []interface{}{"Password1234", "order-code-001", "Password1234", "order-code-002", "Password1234", "order-code-003"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &orderService{kabusAPI: test.kabusAPI, orderStore: test.orderStore}
			got1 := service.CancelAll(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantCancelOrderHistory, test.kabusAPI.CancelOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantCancelOrderHistory, got1, test.kabusAPI.CancelOrderHistory)
			}
		})
	}
}
