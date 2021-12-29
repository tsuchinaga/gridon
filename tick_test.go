package gridon

import (
	"reflect"
	"testing"
)

func Test_tick_getOtherTick(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  float64
		want float64
	}{
		{name: "3000円未満なら1", arg: 2999, want: 1},
		{name: "3000円なら1", arg: 3000, want: 1},
		{name: "5000円未満なら5", arg: 4999, want: 5},
		{name: "5000円なら5", arg: 5000, want: 5},
		{name: "30000円未満なら10", arg: 29999, want: 10},
		{name: "30000円なら10", arg: 30000, want: 10},
		{name: "50000円未満なら50", arg: 49999, want: 50},
		{name: "50000円なら50", arg: 50000, want: 50},
		{name: "300000円未満なら100", arg: 299999, want: 100},
		{name: "300000円なら100", arg: 300000, want: 100},
		{name: "500000円未満なら500", arg: 499999, want: 500},
		{name: "500000円なら500", arg: 500000, want: 500},
		{name: "3000000円未満なら1000", arg: 2999999, want: 1000},
		{name: "3000000円なら1000", arg: 3000000, want: 1000},
		{name: "5000000円未満なら5000", arg: 4999999, want: 5000},
		{name: "5000000円なら5000", arg: 5000000, want: 5000},
		{name: "30000000円未満なら10000", arg: 29999999, want: 10000},
		{name: "30000000円なら10000", arg: 30000000, want: 10000},
		{name: "50000000円未満なら50000", arg: 49999999, want: 50000},
		{name: "50000000円なら50000", arg: 50000000, want: 50000},
		{name: "50000000円超なら100000", arg: 50000001, want: 100000},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got := tick.getOtherTick(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_tick_getTopix100Tick(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  float64
		want float64
	}{
		{name: "1,000円未満なら0.1", arg: 999, want: 0.1},
		{name: "1,000円なら0.1", arg: 1_000, want: 0.1},
		{name: "3,000円未満なら0.5", arg: 2_999, want: 0.5},
		{name: "3,000円なら0.5", arg: 3_000, want: 0.5},
		{name: "10,000円未満なら1", arg: 9_999, want: 1},
		{name: "10,000円なら1", arg: 10_000, want: 1},
		{name: "30,000円未満なら5", arg: 29_999, want: 5},
		{name: "30,000円なら5", arg: 30_000, want: 5},
		{name: "100,0000円未満なら10", arg: 99_999, want: 10},
		{name: "100,000円なら10", arg: 100_000, want: 10},
		{name: "300,000円未満なら50", arg: 299_999, want: 50},
		{name: "300,000円なら50", arg: 300_000, want: 50},
		{name: "1,000,000円未満なら100", arg: 999_999, want: 100},
		{name: "1,000,000円なら100", arg: 1_000_000, want: 100},
		{name: "3,000,000円未満なら500", arg: 2_999_999, want: 500},
		{name: "3,000,000円なら500", arg: 3_000_000, want: 500},
		{name: "10,000,000円未満なら1,000", arg: 9_999_999, want: 1_000},
		{name: "10,000,000円なら1,000", arg: 10_000_000, want: 1_000},
		{name: "30,000,000円未満なら5,000", arg: 29_999_999, want: 5_000},
		{name: "30,000,000円なら5,000", arg: 30_000_000, want: 5_000},
		{name: "30,000,000円超なら10,000", arg: 30_000_001, want: 10_000},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got := tick.getTopix100Tick(test.arg)
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
		arg1 TickGroup
		arg2 float64
		arg3 int
		want float64
	}{
		{name: "0ならそのまま", arg1: TickGroupOther, arg2: 3000, arg3: 0, want: 3000},
		{name: "5なら5ティック上", arg1: TickGroupOther, arg2: 3000, arg3: 5, want: 3025},
		{name: "1なら1ティック上", arg1: TickGroupOther, arg2: 3000, arg3: 1, want: 3005},
		{name: "-1なら1ティック下", arg1: TickGroupOther, arg2: 3000, arg3: -1, want: 2999},
		{name: "-5なら5ティック下", arg1: TickGroupOther, arg2: 3000, arg3: -5, want: 2995},
		{name: "範囲を越えて5ティック上", arg1: TickGroupOther, arg2: 2998, arg3: 5, want: 3015},
		{name: "範囲を超えて5ティック下", arg1: TickGroupOther, arg2: 3010, arg3: -5, want: 2997},
		{name: "TOPIX100テーブルで1ティック上", arg1: TickGroupTopix100, arg2: 250.5, arg3: 1, want: 250.6},
		{name: "TOPIX100テーブルで1ティック下", arg1: TickGroupTopix100, arg2: 250.5, arg3: -1, want: 250.4},
		{name: "TOPIX100テーブルで範囲を超えて5ティック上", arg1: TickGroupTopix100, arg2: 999.8, arg3: 5, want: 1001.5},
		{name: "TOPIX100テーブルで範囲を超えて5ティック下", arg1: TickGroupTopix100, arg2: 1001.0, arg3: -5, want: 999.7},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got := tick.TickAddedPrice(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_newTick(t *testing.T) {
	t.Parallel()
	want1 := &tick{}
	got1 := newTick()
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}

func Test_tick_Ticks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		arg1  TickGroup
		arg2  float64
		arg3  float64
		want1 int
	}{
		{name: "aとbが一致しているなら0",
			arg1:  TickGroupOther,
			arg2:  200,
			arg3:  200,
			want1: 0},
		{name: "aとbがa<bで、1ティック未満の差なら1",
			arg1:  TickGroupOther,
			arg2:  200,
			arg3:  200.1,
			want1: 1},
		{name: "aとbがa>bで、1ティック未満の差なら1",
			arg1:  TickGroupOther,
			arg2:  200.1,
			arg3:  200,
			want1: 1},
		{name: "aとbがともに最初の水準でなければ、最初の水準での計算はスキップされる",
			arg1:  TickGroupOther,
			arg2:  3100,
			arg3:  3200,
			want1: 20},
		{name: "指定してtickGroupが存在しなければ、その他のとして計算する",
			arg1:  TickGroup("hoge"),
			arg2:  3100,
			arg3:  3200,
			want1: 20},
		{name: "TOPIX100グループで、minとmaxが同じ値段水準にあるときにTICK数が得られる",
			arg1:  TickGroupTopix100,
			arg2:  200,
			arg3:  201.5,
			want1: 15},
		{name: "TOPIX100グループで、minとmaxが水準が1段階離れているときにTICK数が得られる",
			arg1:  TickGroupTopix100,
			arg2:  900,
			arg3:  1_100,
			want1: 1_200},
		{name: "TOPIX100グループで、minとmaxが水準が2段階離れているときにTICK数が得られる",
			arg1:  TickGroupTopix100,
			arg2:  900,
			arg3:  3_100,
			want1: 5_100},
		{name: "その他グループで、minとmaxが同じ値段水準にあるときにTICK数が得られる",
			arg1:  TickGroupOther,
			arg2:  2000,
			arg3:  2050,
			want1: 50},
		{name: "その他グループで、minとmaxが水準が1段階離れているときにTICK数が得られる",
			arg1:  TickGroupOther,
			arg2:  2500,
			arg3:  3500,
			want1: 600},
		{name: "その他グループで、minとmaxが水準が2段階離れているときにTICK数が得られる",
			arg1:  TickGroupOther,
			arg2:  2500,
			arg3:  5_500,
			want1: 950},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tick := &tick{}
			got1 := tick.Ticks(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
