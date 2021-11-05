package gridon

import (
	"reflect"
	"testing"
)

func Test_tick_GetTick(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  float64
		want float64
	}{
		{name: "3000円以下なら1", arg: 2999, want: 1},
		{name: "3000円なら1", arg: 3000, want: 1},
		{name: "5000円以下なら5", arg: 4999, want: 5},
		{name: "5000円なら5", arg: 5000, want: 5},
		{name: "30000円以下なら10", arg: 29999, want: 10},
		{name: "30000円なら10", arg: 30000, want: 10},
		{name: "50000円以下なら50", arg: 49999, want: 50},
		{name: "50000円なら50", arg: 50000, want: 50},
		{name: "300000円以下なら100", arg: 299999, want: 100},
		{name: "300000円なら100", arg: 300000, want: 100},
		{name: "500000円以下なら500", arg: 499999, want: 500},
		{name: "500000円なら500", arg: 500000, want: 500},
		{name: "3000000円以下なら1000", arg: 2999999, want: 1000},
		{name: "3000000円なら1000", arg: 3000000, want: 1000},
		{name: "5000000円以下なら5000", arg: 4999999, want: 5000},
		{name: "5000000円なら5000", arg: 5000000, want: 5000},
		{name: "30000000円以下なら10000", arg: 29999999, want: 10000},
		{name: "30000000円なら10000", arg: 30000000, want: 10000},
		{name: "50000000円以下なら50000", arg: 49999999, want: 50000},
		{name: "50000000円なら50000", arg: 50000000, want: 50000},
		{name: "50000000円超なら100000", arg: 50000001, want: 100000},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got := tick.GetTick(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_tick_TickAddedPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 float64
		arg2 int
		want float64
	}{
		{name: "0ならそのまま", arg1: 3000, arg2: 0, want: 3000},
		{name: "5なら5ティック上", arg1: 3000, arg2: 5, want: 3025},
		{name: "1なら1ティック上", arg1: 3000, arg2: 1, want: 3005},
		{name: "-1なら1ティック下", arg1: 3000, arg2: -1, want: 2999},
		{name: "-5なら5ティック下", arg1: 3000, arg2: -5, want: 2995},
		{name: "範囲を越えて5ティック上", arg1: 2998, arg2: 5, want: 3015},
		{name: "範囲を超えて5ティック下", arg1: 3010, arg2: -5, want: 2997},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got := tick.TickAddedPrice(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
