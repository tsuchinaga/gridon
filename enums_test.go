package gridon

import (
	"reflect"
	"testing"
)

func Test_Side_Turn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		side  Side
		want1 Side
	}{
		{name: "未指定 の反対は未指定", side: SideUnspecified, want1: SideUnspecified},
		{name: "買い の反対は売り", side: SideBuy, want1: SideSell},
		{name: "売り の反対は買い", side: SideSell, want1: SideBuy},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.side.Turn()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
