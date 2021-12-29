package gridon

import "math"

var tickTables = map[TickGroup][]struct {
	Lower float64
	Upper float64
	Tick  float64
}{
	TickGroupTopix100: {
		{Lower: 0, Upper: 1_000, Tick: 0.1},
		{Lower: 1_000, Upper: 3_000, Tick: 0.5},
		{Lower: 3_000, Upper: 10_000, Tick: 1},
		{Lower: 10_000, Upper: 30_000, Tick: 5},
		{Lower: 30_000, Upper: 100_000, Tick: 10},
		{Lower: 100_000, Upper: 300_000, Tick: 50},
		{Lower: 300_000, Upper: 1_000_000, Tick: 100},
		{Lower: 1_000_000, Upper: 3_000_000, Tick: 500},
		{Lower: 3_000_000, Upper: 10_000_000, Tick: 1_000},
		{Lower: 10_000_000, Upper: 30_000_000, Tick: 5_000},
		{Lower: 30_000_000, Upper: math.Inf(1), Tick: 10_000},
	},
	TickGroupOther: {
		{Lower: 0, Upper: 3_000, Tick: 1},
		{Lower: 3_000, Upper: 5_000, Tick: 5},
		{Lower: 5_000, Upper: 10_000, Tick: 10},
		{Lower: 10_000, Upper: 50_000, Tick: 50},
		{Lower: 50_000, Upper: 100_000, Tick: 100},
		{Lower: 100_000, Upper: 500_000, Tick: 500},
		{Lower: 500_000, Upper: 1_000_000, Tick: 1_000},
		{Lower: 1_000_000, Upper: 5_000_000, Tick: 5_000},
		{Lower: 5_000_000, Upper: 10_000_000, Tick: 10_000},
		{Lower: 10_000_000, Upper: 50_000_000, Tick: 50_000},
		{Lower: 50_000_000, Upper: math.Inf(1), Tick: 100_000},
	},
}

// newTick - 新しいtickの取得
func newTick() ITick {
	return &tick{}
}

// ITick - ティック計算のインターフェース
type ITick interface {
	GetTick(tickGroup TickGroup, price float64) float64
	TickAddedPrice(tickGroup TickGroup, price float64, tick int) float64
	Ticks(tickGroup TickGroup, a float64, b float64) int
}

// tick - ティック計算
type tick struct{}

// GetTick - 1ティックの幅
func (t *tick) GetTick(tickGroup TickGroup, price float64) float64 {
	switch tickGroup {
	case TickGroupTopix100:
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

// Ticks - minからmaxになるのに何tickあるか
func (t *tick) Ticks(tickGroup TickGroup, a float64, b float64) int {
	max, min := a, b
	if max < min {
		max, min = min, max
	}

	tickTable, ok := tickTables[tickGroup]
	if !ok {
		tickTable = tickTables[TickGroupOther]
	}

	// minの含まれているborderを特定
	// maxが同じborderに含まれていればmaxまでのTICK数を、含まれていなければupperまでのTICK数を計算し、TICK数に加算する
	// maxまでのTICK数が計算されたならreturn
	// 次のborderに移って、繰り返す
	var tick int
	for _, t := range tickTable {
		if min < t.Lower || t.Upper <= min {
			continue
		}

		// maxが含まれている値段水準に達したら、minとmaxのTICK数を加味して結果を返す
		if t.Lower <= max && max < t.Upper {
			tick += int(math.Ceil((max - min) / t.Tick))
			break
		}

		// maxが含まれている値段水準でなかったので、TICK数にその値段水準の最大値までのTICK数を加算し、minを次の水準に移して次のループにいく
		tick += int(math.Ceil((t.Upper - min) / t.Tick))
		min = t.Upper
	}

	return tick
}
