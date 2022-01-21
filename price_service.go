package gridon

// newPriceService - 新しい価格サービスの取得
func newPriceService(kabusAPI IKabusAPI, fourPriceStore IFourPriceStore) IPriceService {
	return &priceService{
		kabusAPI:       kabusAPI,
		fourPriceStore: fourPriceStore,
	}
}

// IPriceService - 価格サービスのインターフェース
type IPriceService interface {
	SaveFourPrice(symbolCode string, exchange Exchange) error
}

// priceService - 価格サービス
type priceService struct {
	kabusAPI       IKabusAPI
	fourPriceStore IFourPriceStore
}

// SaveFourPrice - 現時点での最終の四本値の保存
func (s *priceService) SaveFourPrice(symbolCode string, exchange Exchange) error {
	fourPrice, err := s.kabusAPI.GetFourPrice(symbolCode, exchange)
	if err != nil {
		return err
	}

	return s.fourPriceStore.Save(fourPrice)
}
