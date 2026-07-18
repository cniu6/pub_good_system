package utils

import (
	"testing"
)

func TestYuanFenRoundTrip(t *testing.T) {
	cases := []float64{0, 0.01, 0.1, 0.2, 1.23, 10, 99.99, -0.01, -1.5}
	for _, yuan := range cases {
		fen, err := YuanToFen(yuan)
		if err != nil {
			t.Fatalf("YuanToFen(%v) err=%v", yuan, err)
		}
		back := FenToYuan(fen)
		backFen, err := YuanToFen(back)
		if err != nil || backFen != fen {
			t.Fatalf("round-trip fail yuan=%v fen=%d back=%v backFen=%d", yuan, fen, back, backFen)
		}
	}
}

func TestYuanToFenClassicFloatTrap(t *testing.T) {
	// 0.1+0.2 在 float 上不等于 0.3，但转分后应对齐
	sum := 0.1 + 0.2
	fen, err := YuanToFen(sum)
	if err != nil {
		t.Fatal(err)
	}
	if fen != 30 {
		t.Fatalf("期望 30 分，得到 %d（原始 float=%v）", fen, sum)
	}
	if FenToYuan(fen) != 0.3 {
		t.Fatalf("FenToYuan(30)=%v，期望 0.3", FenToYuan(fen))
	}
}
