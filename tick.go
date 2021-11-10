package gridon

// newTick - 新しいtickの取得
func newTick() ITick {
	return &tick{}
}

// ITick - ティック計算のインターフェース
type ITick interface {
	GetTick(price float64) float64
	TickAddedPrice(price float64, tick int) float64
}

// tick - ティック計算
type tick struct{}

// GetTick - 1ティックの幅
func (t *tick) GetTick(price float64) float64 {
	switch {
	case price <= 3_000:
		return 1
	case price <= 5_000:
		return 5
	case price <= 30_000:
		return 10
	case price <= 50_000:
		return 50
	case price <= 300_000:
		return 100
	case price <= 500_000:
		return 500
	case price <= 3_000_000:
		return 1_000
	case price <= 5_000_000:
		return 5_000
	case price <= 30_000_000:
		return 10_000
	case price <= 50_000_000:
		return 50_000
	default:
		return 100_000
	}
}

// TickAddedPrice - TICKを加味した価格
func (t *tick) TickAddedPrice(price float64, tick int) float64 {
	for tick != 0 {
		if tick < 0 {
			price -= t.GetTick(price - 0.1)
			tick++
		} else {
			price += t.GetTick(price + 0.1)
			tick--
		}
	}
	return price
}
