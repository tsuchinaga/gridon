package gridon

import (
	"reflect"
	"testing"
)

type testStrategyService struct {
	IStrategyService
	UpdateStrategyTickGroup1       error
	UpdateStrategyTickGroupCount   int
	UpdateStrategyTickGroupHistory []interface{}
}

func (t *testStrategyService) UpdateStrategyTickGroup(strategyCode string) error {
	t.UpdateStrategyTickGroupHistory = append(t.UpdateStrategyTickGroupHistory, strategyCode)
	t.UpdateStrategyTickGroupCount++
	return t.UpdateStrategyTickGroup1
}

func Test_strategyService_UpdateStrategyTickGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		strategyStore           *testStrategyStore
		kabusAPI                *testKabusAPI
		arg1                    string
		want1                   error
		wantSetTickGroupHistory []interface{}
	}{
		{name: "戦略が見つからなければエラー",
			strategyStore: &testStrategyStore{GetByCode2: ErrNoData},
			kabusAPI:      &testKabusAPI{},
			arg1:          "strategy-code-001",
			want1:         ErrNoData},
		{name: "kabusapiで銘柄情報が取れなければエラー",
			strategyStore: &testStrategyStore{GetByCode1: &Strategy{}},
			kabusAPI:      &testKabusAPI{GetSymbol2: ErrUnknown},
			arg1:          "strategy-code-001",
			want1:         ErrUnknown},
		{name: "呼値グループの更新に失敗したらエラー",
			strategyStore:           &testStrategyStore{GetByCode1: &Strategy{}, SetTickGroup1: ErrUnknown},
			kabusAPI:                &testKabusAPI{GetSymbol1: &Symbol{TickGroup: TickGroupTopix100}},
			arg1:                    "strategy-code-001",
			want1:                   ErrUnknown,
			wantSetTickGroupHistory: []interface{}{"strategy-code-001", TickGroupTopix100}},
		{name: "途中でエラーがなければnilを返す",
			strategyStore:           &testStrategyStore{GetByCode1: &Strategy{}},
			kabusAPI:                &testKabusAPI{GetSymbol1: &Symbol{TickGroup: TickGroupTopix100}},
			arg1:                    "strategy-code-001",
			want1:                   nil,
			wantSetTickGroupHistory: []interface{}{"strategy-code-001", TickGroupTopix100}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &strategyService{strategyStore: test.strategyStore, kabusAPI: test.kabusAPI}
			got1 := service.UpdateStrategyTickGroup(test.arg1)
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantSetTickGroupHistory, test.strategyStore.SetTickGroupHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantSetTickGroupHistory, got1, test.strategyStore.SetTickGroupHistory)
			}
		})
	}
}

func Test_newStrategyService(t *testing.T) {
	t.Parallel()
	strategyStore := &testStrategyStore{}
	kabusAPI := &testKabusAPI{}
	want1 := &strategyService{strategyStore: strategyStore, kabusAPI: kabusAPI}
	got1 := newStrategyService(kabusAPI, strategyStore)
	if !reflect.DeepEqual(want1, got1) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want1, got1)
	}
}
