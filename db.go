package gridon

import (
	"errors"
	"fmt"
	"sync"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"

	"github.com/genjidb/genji"
	gerrors "github.com/genjidb/genji/errors"
)

var (
	dbSingleton    IDB
	dbSingletonMtx sync.Mutex
)

// getDB - dbの取得
func getDB(path string, logger ILogger) (IDB, error) {
	dbSingletonMtx.Lock()
	defer dbSingletonMtx.Unlock()

	gdb, err := openDB(path)
	if err != nil {
		return nil, err
	}
	dbSingleton = &db{
		db:     gdb,
		logger: logger,
	}

	return dbSingleton, nil
}

// openDB - genjiDBを開き、初期セットアップをする
func openDB(path string) (*genji.DB, error) {
	db, err := genji.Open(path)
	if err != nil {
		return nil, err
	}

	// 必須テーブルの作成
	sqlList := []string{
		// strategies
		`create table if not exists strategies`,
		`create unique index if not exists strategies_code on strategies (code)`,
		// orders
		`create table if not exists orders`,
		`create unique index if not exists orders_code on orders (code)`,
		`create index if not exists orders_strategy_code on orders (strategycode)`,
		`create index if not exists orders_status on orders (status)`,
		// positions
		`create table if not exists positions`,
		`create unique index if not exists positions_code on positions (code)`,
	}

	for _, sql := range sqlList {
		if err := db.Exec(sql); err != nil {
			return nil, fmt.Errorf("error sql: `%s`: %w", sql, err)
		}
	}

	return db, nil
}

// IDB - データベースのインターフェース
type IDB interface {
	GetStrategies() ([]*Strategy, error)
	SaveStrategy(strategy *Strategy) error
	DeleteStrategyByCode(code string) error
	GetActiveOrders() ([]*Order, error)
	SaveOrder(order *Order) error
	GetActivePositions() ([]*Position, error)
	SavePosition(position *Position) error
	CleanupOrders() error
	CleanupPositions() error
}

// db - データベース
type db struct {
	db     *genji.DB
	logger ILogger
}

func (d *db) wrapErr(err error) error {
	switch {
	case errors.Is(gerrors.ErrDocumentNotFound, err):
		return fmt.Errorf("genji error: %s: %w", err, ErrNoData)
	case errors.Is(gerrors.ErrDuplicateDocument, err):
		return fmt.Errorf("genji error: %s: %w", err, ErrDuplicateData)
	case errors.Is(gerrors.AlreadyExistsError{}, err):
		return fmt.Errorf("genji error: %s: %w", err, ErrAlreadyExists)
	case errors.Is(gerrors.NotFoundError{}, err):
		return fmt.Errorf("genji error: %s: %w", err, ErrNotFound)
	default:
		return err
	}
}

// GetStrategies - 戦略一覧の取得
func (d *db) GetStrategies() ([]*Strategy, error) {
	res, err := d.db.Query(`select * from strategies`)
	if err != nil {
		return nil, d.wrapErr(err)
	}
	defer res.Close()

	result := make([]*Strategy, 0)
	err = res.Iterate(func(d types.Document) error {
		var strategy Strategy
		if err := document.StructScan(d, &strategy); err != nil {
			return err
		}
		result = append(result, &strategy)
		return nil
	})
	if err != nil {
		return nil, d.wrapErr(err)
	}
	return result, nil
}

// SaveStrategy - 戦略の保存
func (d *db) SaveStrategy(strategy *Strategy) error {
	d.logger.Notice(fmt.Sprintf("save strategy: %+v", strategy))

	tx, err := d.db.Begin(true)
	if err != nil {
		return d.wrapErr(err)
	}

	if err := tx.Exec(`delete from strategies where code = ?`, strategy.Code); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	if err := tx.Exec(`insert into strategies values ?`, strategy); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	_ = tx.Commit()
	return nil
}

// DeleteStrategyByCode - 戦略の削除
func (d *db) DeleteStrategyByCode(code string) error {
	d.logger.Notice(fmt.Sprintf("delete strategy: %+v", code))

	if err := d.db.Exec(`delete from strategies where code = ?`, code); err != nil {
		d.logger.Warning(err)
		return d.wrapErr(err)
	}
	return nil
}

// GetActiveOrders - 有効な注文一覧の取得
func (d *db) GetActiveOrders() ([]*Order, error) {
	res, err := d.db.Query(`select * from orders where status = 'in_order'`)
	if err != nil {
		return nil, d.wrapErr(err)
	}
	defer res.Close()

	result := make([]*Order, 0)
	err = res.Iterate(func(d types.Document) error {
		var order Order
		if err := document.StructScan(d, &order); err != nil {
			return err
		}
		result = append(result, &order)
		return nil
	})
	if err != nil {
		return nil, d.wrapErr(err)
	}
	return result, nil
}

// SaveOrder - 注文の保存
func (d *db) SaveOrder(order *Order) error {
	d.logger.Notice(fmt.Sprintf("save order: %+v", order))

	tx, err := d.db.Begin(true)
	if err != nil {
		return d.wrapErr(err)
	}

	if err := tx.Exec(`delete from orders where code = ?`, order.Code); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	if err := tx.Exec(`insert into orders values ?`, order); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	_ = tx.Commit()
	return nil
}

// GetActivePositions - 有効なポジション一覧の取得
func (d *db) GetActivePositions() ([]*Position, error) {
	res, err := d.db.Query(`select * from positions where ownedquantity > 0`)
	if err != nil {
		return nil, d.wrapErr(err)
	}
	defer res.Close()

	result := make([]*Position, 0)
	err = res.Iterate(func(d types.Document) error {
		var position Position
		if err := document.StructScan(d, &position); err != nil {
			return err
		}
		result = append(result, &position)
		return nil
	})
	if err != nil {
		return nil, d.wrapErr(err)
	}
	return result, nil
}

// SavePosition - ポジションの保存
func (d *db) SavePosition(position *Position) error {
	d.logger.Notice(fmt.Sprintf("save position: %+v", position))

	tx, err := d.db.Begin(true)
	if err != nil {
		return d.wrapErr(err)
	}

	if err := tx.Exec(`delete from positions where code = ?`, position.Code); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	if err := tx.Exec(`insert into positions values ?`, position); err != nil {
		_ = tx.Rollback()
		d.logger.Warning(err)
		return d.wrapErr(err)
	}

	_ = tx.Commit()
	return nil
}

// CleanupOrders - 不要な注文データの削除
func (d *db) CleanupOrders() error {
	if err := d.db.Exec(`delete from orders where status != 'in_order'`); err != nil {
		return err
	}

	return nil
}

// CleanupPositions - 不要なポジションデータの削除
func (d *db) CleanupPositions() error {
	if err := d.db.Exec(`delete from positions where ownedquantity <= 0`); err != nil {
		return err
	}

	return nil
}
