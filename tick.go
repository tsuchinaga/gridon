package gridon

import "math"

// newTick - 新しいtickの取得
func newTick() ITick {
	return &tick{}
}

// ITick - ティック計算のインターフェース
type ITick interface {
	GetTick(tickGroup TickGroup, price float64) float64
	TickAddedPrice(tickGroup TickGroup, price float64, tick int) float64
}

// tick - ティック計算
type tick struct{}

// GetTick - 1ティックの幅
func (t *tick) GetTick(tickGroup TickGroup, price float64) float64 {
	switch tickGroup {
	case TickGroupTOPIX100:
		return t.getTopix100Tick(price)
	default:
		return t.getOtherTick(price)
	}
}

// getTopix100Tick - TOPIX100テーブル対象の銘柄の呼値単位
func (t *tick) getTopix100Tick(price float64) float64 {
	switch {
	case price <= 1_000:
		return 0.1
	case price <= 3_000:
		return 0.5
	case price <= 10_000:
		return 1
	case price <= 30_000:
		return 5
	case price <= 100_000:
		return 10
	case price <= 300_000:
		return 50
	case price <= 1_000_000:
		return 100
	case price <= 3_000_000:
		return 500
	case price <= 10_000_000:
		return 1_000
	case price <= 30_000_000:
		return 5_000
	default:
		return 10_000
	}
}

// getOtherTick - TOPIX100テーブル対象以外の銘柄の呼値単位
func (t *tick) getOtherTick(price float64) float64 {
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

// TickAddedPrice - 呼値単位を加味した価格
func (t *tick) TickAddedPrice(tickGroup TickGroup, price float64, tick int) float64 {
	for tick != 0 {
		if tick < 0 {
			price -= t.GetTick(tickGroup, price-0.01)
			tick++
		} else {
			price += t.GetTick(tickGroup, price+0.01)
			tick--
		}
	}
	return math.Round(price*10) / 10 // 小数点以下第一で四捨五入
}
