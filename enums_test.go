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

func Test_Rounding_Calc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		rounding Rounding
		arg1     float64
		want1    float64
	}{
		{name: "-1の未指定は-1", rounding: RoundingUnspecified, arg1: -1, want1: -1},
		{name: "-1の切り捨ては-1", rounding: RoundingFloor, arg1: -1, want1: -1},
		{name: "-1の四捨五入は-1", rounding: RoundingRound, arg1: -1, want1: -1},
		{name: "-1の切り上げは-1", rounding: RoundingCeil, arg1: -1, want1: -1},
		{name: "-0.9の未指定は-0.9", rounding: RoundingUnspecified, arg1: -0.9, want1: -0.9},
		{name: "-0.9の切り捨ては-1", rounding: RoundingFloor, arg1: -0.9, want1: -1},
		{name: "-0.9の四捨五入は-1", rounding: RoundingRound, arg1: -0.9, want1: -1},
		{name: "-0.9の切り上げは0", rounding: RoundingCeil, arg1: -0.9, want1: 0},
		{name: "-0.5の未指定は-0.5", rounding: RoundingUnspecified, arg1: -0.5, want1: -0.5},
		{name: "-0.5の切り捨ては-1", rounding: RoundingFloor, arg1: -0.5, want1: -1},
		{name: "-0.5の四捨五入は-1", rounding: RoundingRound, arg1: -0.5, want1: -1},
		{name: "-0.5の切り上げは0", rounding: RoundingCeil, arg1: -0.5, want1: 0},
		{name: "-0.4の未指定は-0.4", rounding: RoundingUnspecified, arg1: -0.4, want1: -0.4},
		{name: "-0.4の切り捨ては-1", rounding: RoundingFloor, arg1: -0.4, want1: -1},
		{name: "-0.4の四捨五入は0", rounding: RoundingRound, arg1: -0.4, want1: 0},
		{name: "-0.4の切り上げは0", rounding: RoundingCeil, arg1: -0.4, want1: 0},
		{name: "-0.1の未指定は-0.1", rounding: RoundingUnspecified, arg1: -0.1, want1: -0.1},
		{name: "-0.1の切り捨ては-1", rounding: RoundingFloor, arg1: -0.1, want1: -1},
		{name: "-0.1の四捨五入は0", rounding: RoundingRound, arg1: -0.1, want1: 0},
		{name: "-0.1の切り上げは0", rounding: RoundingCeil, arg1: -0.1, want1: 0},
		{name: "0の未指定は0.0", rounding: RoundingUnspecified, arg1: 0, want1: 0},
		{name: "0の切り捨ては0", rounding: RoundingFloor, arg1: 0, want1: 0},
		{name: "0の四捨五入は0", rounding: RoundingRound, arg1: 0, want1: 0},
		{name: "0の切り上げは0", rounding: RoundingCeil, arg1: 0, want1: 0},
		{name: "0.1の未指定は0.1", rounding: RoundingUnspecified, arg1: 0.1, want1: 0.1},
		{name: "0.1の切り捨ては0", rounding: RoundingFloor, arg1: 0.1, want1: 0},
		{name: "0.1の四捨五入は0", rounding: RoundingRound, arg1: 0.1, want1: 0},
		{name: "0.1の切り上げは1", rounding: RoundingCeil, arg1: 0.1, want1: 1},
		{name: "0.4の未指定は0.4", rounding: RoundingUnspecified, arg1: 0.4, want1: 0.4},
		{name: "0.4の切り捨ては0", rounding: RoundingFloor, arg1: 0.4, want1: 0},
		{name: "0.4の四捨五入は0", rounding: RoundingRound, arg1: 0.4, want1: 0},
		{name: "0.4の切り上げは1", rounding: RoundingCeil, arg1: 0.4, want1: 1},
		{name: "0.5の未指定は0.5", rounding: RoundingUnspecified, arg1: 0.5, want1: 0.5},
		{name: "0.5の切り捨ては0", rounding: RoundingFloor, arg1: 0.5, want1: 0},
		{name: "0.5の四捨五入は1", rounding: RoundingRound, arg1: 0.5, want1: 1},
		{name: "0.5の切り上げは1", rounding: RoundingCeil, arg1: 0.5, want1: 1},
		{name: "0.9の未指定は0.9", rounding: RoundingUnspecified, arg1: 0.9, want1: 0.9},
		{name: "0.9の切り捨ては0", rounding: RoundingFloor, arg1: 0.9, want1: 0},
		{name: "0.9の四捨五入は1", rounding: RoundingRound, arg1: 0.9, want1: 1},
		{name: "0.9の切り上げは1", rounding: RoundingCeil, arg1: 0.9, want1: 1},
		{name: "1の未指定は1", rounding: RoundingUnspecified, arg1: 1, want1: 1},
		{name: "1の切り捨ては1", rounding: RoundingFloor, arg1: 1, want1: 1},
		{name: "1の四捨五入は1", rounding: RoundingRound, arg1: 1, want1: 1},
		{name: "1の切り上げは1", rounding: RoundingCeil, arg1: 1, want1: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.rounding.Calc(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_Operation_Calc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		operation Operation
		arg1      float64
		arg2      float64
		want1     float64
	}{
		{name: "未指定なら加算", operation: OperationUnspecified, arg1: 100, arg2: 0.1, want1: 100.1},
		{name: "+なら加算", operation: OperationPlus, arg1: 100, arg2: 0.1, want1: 100.1},
		{name: "-なら減算", operation: OperationMinus, arg1: 100, arg2: 0.1, want1: 99.9},
		{name: "*なら積算", operation: OperationMultiple, arg1: 100, arg2: 0.1, want1: 10},
		{name: "/なら除算", operation: OperationDived, arg1: 100, arg2: 0.1, want1: 1000},
		{name: "minなら小さいほうを返す", operation: OperationMin, arg1: 100, arg2: 50, want1: 50},
		{name: "maxなら大きいほうを返す", operation: OperationMax, arg1: 100, arg2: 50, want1: 100},
		{name: "overwriteなら後の方を返す", operation: OperationOverwrite, arg1: 100, arg2: 50, want1: 50},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.operation.Calc(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}
