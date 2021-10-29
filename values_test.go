package gridon

import (
	"reflect"
	"testing"
)

func Test_HoldPosition_LeaveQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		holdPosition HoldPosition
		want1        float64
	}{
		{name: "Hold - ExitContract - Releaseが返される 1", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 0}, want1: 100},
		{name: "Hold - ExitContract - Releaseが返される 2", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 20, ReleaseQuantity: 0}, want1: 80},
		{name: "Hold - ExitContract - Releaseが返される 3", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 100, ReleaseQuantity: 0}, want1: 0},
		{name: "Hold - ExitContract - Releaseが返される 4", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 20}, want1: 80},
		{name: "Hold - ExitContract - Releaseが返される 5", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 0, ReleaseQuantity: 100}, want1: 0},
		{name: "Hold - ExitContract - Releaseが返される 6", holdPosition: HoldPosition{HoldQuantity: 100, ContractQuantity: 30, ReleaseQuantity: 70}, want1: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.holdPosition.LeaveQuantity()
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
