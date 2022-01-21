package gridon

import (
	"reflect"
	"testing"
	"time"
)

func Test_service_runnableContractTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		service             *service
		want1               bool
		wantContractRunning bool
	}{
		{name: "実行中ならfalse",
			service:             &service{contractRunning: true},
			want1:               false,
			wantContractRunning: true},
		{name: "実行中でないなら実行中に変更してtrue",
			service:             &service{contractRunning: false},
			want1:               true,
			wantContractRunning: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.service.runnableContractTask()
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantContractRunning, test.service.contractRunning) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantContractRunning, got1, test.service.contractRunning)
			}
		})
	}
}

func Test_service_finishContractTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		service             *service
		wantContractRunning bool
	}{
		{name: "実行中ならfalseにできる",
			service:             &service{contractRunning: true},
			wantContractRunning: false},
		{name: "実行中でなくてもfalseにできる",
			service:             &service{contractRunning: false},
			wantContractRunning: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.finishContractTask()
			if !reflect.DeepEqual(test.wantContractRunning, test.service.contractRunning) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantContractRunning, test.service.contractRunning)
			}
		})
	}
}

func Test_service_runnableOrderTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		service          *service
		want1            bool
		wantOrderRunning bool
	}{
		{name: "実行中ならfalse",
			service:          &service{orderRunning: true},
			want1:            false,
			wantOrderRunning: true},
		{name: "実行中でないなら実行中に変更してtrue",
			service:          &service{orderRunning: false},
			want1:            true,
			wantOrderRunning: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.service.runnableOrderTask()
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantOrderRunning, test.service.orderRunning) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantOrderRunning, got1, test.service.orderRunning)
			}
		})
	}
}

func Test_service_finishOrderTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		service          *service
		wantOrderRunning bool
	}{
		{name: "実行中ならfalseにできる",
			service:          &service{orderRunning: true},
			wantOrderRunning: false},
		{name: "実行中でなくてもfalseにできる",
			service:          &service{orderRunning: false},
			wantOrderRunning: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.finishOrderTask()
			if !reflect.DeepEqual(test.wantOrderRunning, test.service.orderRunning) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrderRunning, test.service.orderRunning)
			}
		})
	}
}

func Test_service_contractTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		logger                  *testLogger
		strategyStore           *testStrategyStore
		contractService         *testContractService
		gridService             *testGridService
		contractRunning         bool
		wantWarningCount        int
		wantConfirmCount        int
		wantConfirmGridEndCount int
		wantLevelingCount       int
	}{
		{name: "実行中なら何もせず終了",
			logger:            &testLogger{},
			strategyStore:     &testStrategyStore{},
			contractService:   &testContractService{},
			gridService:       &testGridService{},
			contractRunning:   true,
			wantWarningCount:  0,
			wantConfirmCount:  0,
			wantLevelingCount: 0},
		{name: "戦略一覧取得に失敗したらエラーを吐いて終了",
			logger:            &testLogger{},
			strategyStore:     &testStrategyStore{GetStrategies2: ErrUnknown},
			contractService:   &testContractService{},
			gridService:       &testGridService{},
			contractRunning:   false,
			wantWarningCount:  1,
			wantConfirmCount:  0,
			wantLevelingCount: 0},
		{name: "戦略がなければ何もせずに終了",
			logger:            &testLogger{},
			strategyStore:     &testStrategyStore{GetStrategies1: []*Strategy{}},
			contractService:   &testContractService{},
			gridService:       &testGridService{},
			contractRunning:   false,
			wantWarningCount:  0,
			wantConfirmCount:  0,
			wantLevelingCount: 0},
		{name: "約定確認でエラーが発生したらエラーを吐いて終了",
			logger:            &testLogger{},
			strategyStore:     &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			contractService:   &testContractService{Confirm1: ErrUnknown},
			gridService:       &testGridService{},
			contractRunning:   false,
			wantWarningCount:  1,
			wantConfirmCount:  1,
			wantLevelingCount: 0},
		{name: "グリッド終了時約定確認でエラーが発生したらエラーを吐いて終了",
			logger:                  &testLogger{},
			strategyStore:           &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			contractService:         &testContractService{ConfirmGridEnd1: ErrUnknown},
			gridService:             &testGridService{},
			contractRunning:         false,
			wantWarningCount:        1,
			wantConfirmCount:        1,
			wantConfirmGridEndCount: 1,
			wantLevelingCount:       0},
		{name: "グリッドの整地でエラーが発生したらエラーを吐いて終了",
			logger:                  &testLogger{},
			strategyStore:           &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			contractService:         &testContractService{},
			gridService:             &testGridService{Leveling1: ErrUnknown},
			contractRunning:         false,
			wantWarningCount:        1,
			wantConfirmCount:        1,
			wantConfirmGridEndCount: 1,
			wantLevelingCount:       1},
		{name: "戦略の数だけ約定確認とグリッドの整地をする",
			logger: &testLogger{},
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}}},
			contractService:         &testContractService{},
			gridService:             &testGridService{},
			contractRunning:         false,
			wantWarningCount:        0,
			wantConfirmCount:        3,
			wantConfirmGridEndCount: 3,
			wantLevelingCount:       3},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &service{
				logger:          test.logger,
				strategyStore:   test.strategyStore,
				contractService: test.contractService,
				gridService:     test.gridService,
				contractRunning: test.contractRunning,
			}
			service.contractTask()

			// 非同期処理を少し待つ
			time.Sleep(100 * time.Millisecond)

			if !reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) ||
				!reflect.DeepEqual(test.wantConfirmCount, test.contractService.ConfirmCount) ||
				!reflect.DeepEqual(test.wantConfirmGridEndCount, test.contractService.ConfirmGridEndCount) ||
				!reflect.DeepEqual(test.wantLevelingCount, test.gridService.LevelingCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.wantWarningCount, test.wantConfirmCount, test.wantConfirmGridEndCount, test.wantLevelingCount,
					test.logger.WarningCount, test.contractService.ConfirmCount, test.contractService.ConfirmGridEndCount, test.gridService.LevelingCount)
			}
		})
	}
}

func Test_service_orderTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		logger             *testLogger
		strategyStore      *testStrategyStore
		rebalanceService   *testRebalanceService
		orderService       *testOrderService
		orderRunning       bool
		wantWarningCount   int
		wantRebalanceCount int
		wantCancelAllCount int
		wantExitAllCount   int
	}{
		{name: "実行中なら何もせず終了",
			logger:           &testLogger{},
			strategyStore:    &testStrategyStore{},
			rebalanceService: &testRebalanceService{},
			orderService:     &testOrderService{},
			orderRunning:     true},
		{name: "戦略一覧の取得に失敗したらログを吐いて終了",
			logger:           &testLogger{},
			strategyStore:    &testStrategyStore{GetStrategies2: ErrUnknown},
			rebalanceService: &testRebalanceService{},
			orderService:     &testOrderService{},
			orderRunning:     false,
			wantWarningCount: 1},
		{name: "戦略がなければ何もせず終了",
			logger:           &testLogger{},
			strategyStore:    &testStrategyStore{GetStrategies1: []*Strategy{}},
			rebalanceService: &testRebalanceService{},
			orderService:     &testOrderService{},
			orderRunning:     false,
			wantWarningCount: 0},
		{name: "リバランスでエラーがあればログを吐き、後続の処理を実行",
			logger:             &testLogger{},
			strategyStore:      &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			rebalanceService:   &testRebalanceService{Rebalance1: ErrUnknown},
			orderService:       &testOrderService{},
			orderRunning:       false,
			wantWarningCount:   1,
			wantRebalanceCount: 1,
			wantCancelAllCount: 1,
			wantExitAllCount:   1},
		{name: "全取消でエラーがあればログを吐き、後続の処理を実行",
			logger:             &testLogger{},
			strategyStore:      &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			rebalanceService:   &testRebalanceService{},
			orderService:       &testOrderService{CancelAll1: ErrUnknown},
			orderRunning:       false,
			wantWarningCount:   1,
			wantRebalanceCount: 1,
			wantCancelAllCount: 1,
			wantExitAllCount:   1},
		{name: "全エグジットでエラーがあればログを吐き、後続の処理を実行",
			logger:             &testLogger{},
			strategyStore:      &testStrategyStore{GetStrategies1: []*Strategy{{Code: "strategy-code-001"}}},
			rebalanceService:   &testRebalanceService{},
			orderService:       &testOrderService{ExitAll1: ErrUnknown},
			orderRunning:       false,
			wantWarningCount:   1,
			wantRebalanceCount: 1,
			wantCancelAllCount: 1,
			wantExitAllCount:   1},
		{name: "戦略の数だけ各処理を実行する",
			logger: &testLogger{},
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}}},
			rebalanceService:   &testRebalanceService{},
			orderService:       &testOrderService{},
			orderRunning:       false,
			wantWarningCount:   0,
			wantRebalanceCount: 3,
			wantCancelAllCount: 3,
			wantExitAllCount:   3},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &service{
				logger:           test.logger,
				strategyStore:    test.strategyStore,
				rebalanceService: test.rebalanceService,
				orderService:     test.orderService,
				orderRunning:     test.orderRunning,
			}
			service.orderTask()

			// 非同期処理を少し待つ
			time.Sleep(100 * time.Millisecond)

			if !reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) ||
				!reflect.DeepEqual(test.wantRebalanceCount, test.rebalanceService.RebalanceCount) ||
				!reflect.DeepEqual(test.wantCancelAllCount, test.orderService.CancelAllCount) ||
				!reflect.DeepEqual(test.wantExitAllCount, test.orderService.ExitAllCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.wantWarningCount, test.wantRebalanceCount, test.wantCancelAllCount, test.wantExitAllCount,
					test.logger.WarningCount, test.rebalanceService.RebalanceCount, test.orderService.CancelAllCount, test.orderService.ExitAllCount)
			}
		})
	}
}

func Test_service_Start(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		strategyStore *testStrategyStore
		orderStore    *testOrderStore
		positionStore *testPositionStore
		want1         error
	}{
		{name: "戦略ストアのデプロイに失敗したらエラー",
			strategyStore: &testStrategyStore{DeployFromDB1: ErrUnknown},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			want1:         ErrUnknown},
		{name: "注文ストアのデプロイに失敗したらエラー",
			strategyStore: &testStrategyStore{},
			orderStore:    &testOrderStore{DeployFromDB1: ErrUnknown},
			positionStore: &testPositionStore{},
			want1:         ErrUnknown},
		{name: "ポジションストアのデプロイに失敗したらエラー",
			strategyStore: &testStrategyStore{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{DeployFromDB1: ErrUnknown},
			want1:         ErrUnknown},
		{name: "デプロイに成功すればタスクが起動され、エラーなし",
			strategyStore: &testStrategyStore{},
			orderStore:    &testOrderStore{},
			positionStore: &testPositionStore{},
			want1:         nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var got1 error
			service := &service{
				logger:        &testLogger{},
				clock:         &testClock{},
				strategyStore: test.strategyStore,
				orderStore:    test.orderStore,
				positionStore: test.positionStore,
				webService:    &testWebService{},
			}
			go func() {
				got1 = service.Start()
			}()
			<-time.After(1 * time.Second)

			if !reflect.DeepEqual(test.want1, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want1, got1)
			}
		})
	}
}

func Test_service_updateStrategyTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                               string
		strategyStore                      *testStrategyStore
		strategyService                    *testStrategyService
		logger                             *testLogger
		wantWarningCount                   int
		wantUpdateStrategyTickGroupHistory []interface{}
	}{
		{name: "戦略一覧の取得に失敗したらエラーログを吐いて終了",
			strategyStore:    &testStrategyStore{GetStrategies2: ErrUnknown},
			strategyService:  &testStrategyService{},
			logger:           &testLogger{},
			wantWarningCount: 1},
		{name: "戦略一覧が空なら何もせずに終了",
			strategyStore:    &testStrategyStore{GetStrategies1: []*Strategy{}},
			strategyService:  &testStrategyService{},
			logger:           &testLogger{},
			wantWarningCount: 0},
		{name: "戦略の更新に失敗したらエラーログを吐き、後続の処理を実行",
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}}},
			strategyService:                    &testStrategyService{UpdateStrategyTickGroup1: ErrUnknown},
			logger:                             &testLogger{},
			wantWarningCount:                   3,
			wantUpdateStrategyTickGroupHistory: []interface{}{"strategy-code-001", "strategy-code-002", "strategy-code-003"}},
		{name: "戦略の更新に成功したら終了",
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{Code: "strategy-code-001"},
				{Code: "strategy-code-002"},
				{Code: "strategy-code-003"}}},
			strategyService:                    &testStrategyService{UpdateStrategyTickGroup1: nil},
			logger:                             &testLogger{},
			wantWarningCount:                   0,
			wantUpdateStrategyTickGroupHistory: []interface{}{"strategy-code-001", "strategy-code-002", "strategy-code-003"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &service{logger: test.logger, strategyStore: test.strategyStore, strategyService: test.strategyService}
			service.updateStrategyTask()
			if !reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) ||
				!reflect.DeepEqual(test.wantUpdateStrategyTickGroupHistory, test.strategyService.UpdateStrategyTickGroupHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					test.wantWarningCount, test.wantUpdateStrategyTickGroupHistory,
					test.logger.WarningCount, test.strategyService.UpdateStrategyTickGroupHistory)
			}
		})
	}
}

func Test_service_startWebServerTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		webService       *testWebService
		logger           *testLogger
		wantWarningCount int
	}{
		{name: "webServiceがエラーを返したらログを吐いて終了",
			webService:       &testWebService{StartWebServer1: ErrUnknown},
			logger:           &testLogger{},
			wantWarningCount: 1},
		{name: "webServiceがエラーを返さなければ何もなく終了",
			webService:       &testWebService{},
			logger:           &testLogger{},
			wantWarningCount: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &service{webService: test.webService, logger: test.logger}
			service.startWebServerTask()
			if !reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantWarningCount, test.logger.WarningCount)
			}
		})
	}
}

func Test_service_dailyTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		strategyStore          *testStrategyStore
		priceService           *testPriceService
		logger                 *testLogger
		wantSaveFourPriceCount int
		wantWarningCount       int
	}{
		{name: "戦略一覧の取得に失敗したらログを吐いて終了",
			strategyStore:    &testStrategyStore{GetStrategies2: ErrUnknown},
			priceService:     &testPriceService{},
			logger:           &testLogger{},
			wantWarningCount: 1},
		{name: "四本値の保存に失敗したらログを吐いて終了",
			strategyStore:          &testStrategyStore{GetStrategies1: []*Strategy{{SymbolCode: "1475", Exchange: ExchangeToushou}}},
			priceService:           &testPriceService{SaveFourPrice1: ErrUnknown},
			logger:                 &testLogger{},
			wantSaveFourPriceCount: 1,
			wantWarningCount:       1},
		{name: "四本値の保存でエラーがなければそのまま終了",
			strategyStore: &testStrategyStore{GetStrategies1: []*Strategy{
				{SymbolCode: "1475", Exchange: ExchangeToushou},
				{SymbolCode: "1476", Exchange: ExchangeToushou},
			}},
			priceService:           &testPriceService{SaveFourPrice1: nil},
			logger:                 &testLogger{},
			wantSaveFourPriceCount: 2,
			wantWarningCount:       0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &service{
				logger:        test.logger,
				strategyStore: test.strategyStore,
				priceService:  test.priceService}
			service.dailyTask()
			if !reflect.DeepEqual(test.wantSaveFourPriceCount, test.priceService.SaveFourPriceCount) ||
				!reflect.DeepEqual(test.wantWarningCount, test.logger.WarningCount) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					test.wantSaveFourPriceCount, test.wantWarningCount,
					test.wantWarningCount, test.logger.WarningCount)
			}
		})
	}
}
