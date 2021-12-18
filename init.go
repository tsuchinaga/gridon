package gridon

import (
	"time"
)

func init() {
	// タイムゾーンがズレないよう固定
	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
}
