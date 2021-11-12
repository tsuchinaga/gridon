package gridon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/tsuchinaga/kabus-grpc-server/kabuspb"

	"google.golang.org/grpc"
)

func NewService() (IService, error) {
	logger, err := getLogger()
	if err != nil {
		return nil, err
	}

	db, err := getDB("gridon.db", logger)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(context.Background(), "localhost:18082", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	kabucom := kabuspb.NewKabusServiceClient(conn)

	return &service{
		logger:        logger,
		clock:         newClock(),
		strategyStore: getStrategyStore(db),
		orderStore:    getOrderStore(db),
		positionStore: getPositionStore(db),
		contractService: newContractService(
			newKabusAPI(kabucom),
			getStrategyStore(db),
			getOrderStore(db),
			getPositionStore(db)),
		rebalanceService: newRebalanceService(
			newClock(),
			newKabusAPI(kabucom),
			getPositionStore(db),
			newOrderService(
				newClock(),
				newKabusAPI(kabucom),
				getStrategyStore(db),
				getOrderStore(db),
				getPositionStore(db))),
		gridService: newGridService(
			newClock(),
			newTick(),
			newKabusAPI(kabucom),
			newOrderService(
				newClock(),
				newKabusAPI(kabucom),
				getStrategyStore(db),
				getOrderStore(db),
				getPositionStore(db))),
		orderService: newOrderService(
			newClock(),
			newKabusAPI(kabucom),
			getStrategyStore(db),
			getOrderStore(db),
			getPositionStore(db)),
	}, nil
}

// IService - gridonサービスのインターフェース
type IService interface {
	Start() error
	SaveStrategy(strategy *Strategy) error
}

// service - gridonサービス
type service struct {
	logger             ILogger
	clock              IClock
	strategyStore      IStrategyStore
	orderStore         IOrderStore
	positionStore      IPositionStore
	contractService    IContractService
	rebalanceService   IRebalanceService
	gridService        IGridService
	orderService       IOrderService
	contractRunning    bool
	contractRunningMtx sync.Mutex
	orderRunning       bool
	orderRunningMtx    sync.Mutex
}

func (s *service) Start() error {
	// DBからデータの読み込み
	if err := s.strategyStore.DeployFromDB(); err != nil {
		return err
	}
	if err := s.orderStore.DeployFromDB(); err != nil {
		return err
	}
	if err := s.positionStore.DeployFromDB(); err != nil {
		return err
	}

	// 約定確認スケジューラの起動
	go s.contractScheduler()

	// 注文に関するスケジューラの起動 (リバランス、グリッド、全エグジット)
	go s.orderScheduler()
	select {}
}

// contractScheduler - 約定確認スケジューラ
func (s *service) contractScheduler() {
	s.logger.Notice("約定確認スケジューラ起動")

	// 4秒に1回非同期で処理を実行する
	ticker := time.NewTicker(4 * time.Second)
	for {
		go s.contractTask()
		<-ticker.C
	}
}

// runnableContractTask - 約定確認タスクが実行可能かどうか、可能なら実行中にする
func (s *service) runnableContractTask() bool {
	s.contractRunningMtx.Lock()
	defer s.contractRunningMtx.Unlock()

	if s.contractRunning {
		return false
	}
	s.contractRunning = true
	return true
}

// finishContractTask - 約定確認タスクの実行状態を終了に更新する
func (s *service) finishContractTask() {
	s.contractRunningMtx.Lock()
	defer s.contractRunningMtx.Unlock()
	s.contractRunning = false
}

// contractTask - 約定確認のタスク
func (s *service) contractTask() {
	if !s.runnableContractTask() {
		return
	}
	defer s.finishContractTask()

	// 戦略一覧の取得
	strategies, err := s.strategyStore.GetStrategies()
	if err != nil {
		s.logger.Warning(fmt.Errorf("約定確認処理の戦略一覧取得でエラーが発生しました: %w", err))
	}

	// 約定確認の実行
	var wg sync.WaitGroup
	for _, strategy := range strategies {
		strategy := strategy
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := s.contractService.Confirm(strategy); err != nil {
				s.logger.Warning(fmt.Errorf("%s の約定確認処理でエラーが発生しました: %w", strategy.Code, err))
				return
			}

			if err := s.gridService.Leveling(strategy); err != nil {
				s.logger.Warning(fmt.Errorf("%s のグリッド処理でエラーが発生しました: %w", strategy.Code, err))
			}
		}()
	}
	wg.Wait()
}

// orderScheduler - 注文スケジューラ
func (s *service) orderScheduler() {
	s.logger.Notice("注文スケジューラ起動")

	// 次の0秒のタイミングまで待機
	now := s.clock.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	time.Sleep(nextRun.Sub(now))

	// 1分に1回非同期で処理を実行する
	ticker := time.NewTicker(1 * time.Minute)
	for {
		go s.orderTask()
		<-ticker.C
	}
}

// runnableOrderTask - 注文タスクが実行可能かどうか、可能なら実行中にする
func (s *service) runnableOrderTask() bool {
	s.orderRunningMtx.Lock()
	defer s.orderRunningMtx.Unlock()

	if s.orderRunning {
		return false
	}
	s.orderRunning = true
	return true
}

// finishOrderTask - 注文タスクの実行状態を終了に更新する
func (s *service) finishOrderTask() {
	s.orderRunningMtx.Lock()
	defer s.orderRunningMtx.Unlock()
	s.orderRunning = false
}

// orderTask - 注文のタスク
func (s *service) orderTask() {
	if !s.runnableOrderTask() {
		return
	}
	defer s.finishOrderTask()

	// 戦略一覧の取得
	strategies, err := s.strategyStore.GetStrategies()
	if err != nil {
		s.logger.Warning(fmt.Errorf("注文処理の戦略一覧取得でエラーが発生しました: %w", err))
	}

	// rebalanceの実行
	var wg sync.WaitGroup
	for _, strategy := range strategies {
		strategy := strategy
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := s.rebalanceService.Rebalance(strategy); err != nil {
				s.logger.Warning(fmt.Errorf("%s のリバランス処理でエラーが発生しました: %w", strategy.Code, err))
			}

			if err := s.orderService.CancelAll(strategy); err != nil {
				s.logger.Warning(fmt.Errorf("%s の全取消処理でエラーが発生しました: %w", strategy.Code, err))
			}

			if err := s.orderService.ExitAll(strategy); err != nil {
				s.logger.Warning(fmt.Errorf("%s の全エグジット処理でエラーが発生しました: %w", strategy.Code, err))
			}
		}()
	}
	wg.Wait()
}

// SaveStrategy - 戦略の保存
func (s *service) SaveStrategy(strategy *Strategy) error {
	return s.strategyStore.Save(strategy)
}