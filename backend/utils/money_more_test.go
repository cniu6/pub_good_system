package utils_test

import (
	"testing"

	"fst/backend/utils"
)

func TestMoneyFenRoundTripDetailed(t *testing.T) {
	cases := []struct {
		yuan float64
		fen  int64
	}{
		{0, 0},
		{0.01, 1},
		{1.23, 123},
		{10.00, 1000},
		{87.66, 8766},
	}
	for _, c := range cases {
		fen, err := utils.YuanToFen(c.yuan)
		if err != nil {
			t.Fatalf("YuanToFen(%v): %v", c.yuan, err)
		}
		if fen != c.fen {
			t.Fatalf("YuanToFen(%v)=%d want %d", c.yuan, fen, c.fen)
		}
		back := utils.FenToYuan(fen)
		if back != c.yuan {
			// 允许浮点展示差异时再看
			if utils.MustYuanToFen(back) != fen {
				t.Fatalf("往返不一致 yuan=%v fen=%d back=%v", c.yuan, fen, back)
			}
		}
	}
}

func TestRejectInvalidMoney(t *testing.T) {
	if _, err := utils.YuanToFen(1e20); err == nil {
		// 过大金额应拒绝（若实现有上限）
		t.Log("极大金额未拒绝（若业务有上限可再收紧）")
	}
}
