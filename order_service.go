package gridon

// IOrderService - 注文サービスのインターフェース
type IOrderService interface {
	CancelAll(strategy *Strategy) error
}

// orderService - 注文サービス
type orderService struct {
	kabusAPI      IKabusAPI
	orderStore    IOrderStore
	positionStore IPositionStore
}

// CancelAll - 戦略に関連する全ての注文を取り消す
func (s *orderService) CancelAll(strategy *Strategy) error {
	if strategy == nil {
		return ErrNilArgument
	}

	// 有効な注文を取り出す
	orders, err := s.orderStore.GetActiveOrdersByStrategyCode(strategy.Code)
	if err != nil {
		return err
	}

	// キャンセルに流す
	for _, o := range orders {
		_, err := s.kabusAPI.CancelOrder(strategy.Account.Password, o.Code)
		if err != nil {
			return err
		}
		// TODO 取消注文に失敗したログを残す？
	}
	return nil
}
