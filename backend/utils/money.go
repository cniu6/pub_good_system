package utils

import (
	"errors"
	"math"
)

// 金额约定（兼容现有 DECIMAL(10,2)「元」库表）：
//   - 库表 / API 边界仍用「元」（两位小数）
//   - 内部加减一律先转成「分」int64 再算，写回前再转成元
//   - 这样 0.1+0.2 不会漂成 0.3000000004，对账按分对齐

const (
	// MoneyFenPerYuan 1 元 = 100 分
	MoneyFenPerYuan int64 = 100
	// MoneyMaxFen 余额上限（与原先 999999999999 元同量级，按分计）
	MoneyMaxFen int64 = 999999999999 * MoneyFenPerYuan
)

// YuanToFen 元 → 分（四舍五入到分）。非法浮点返回错误。
func YuanToFen(yuan float64) (int64, error) {
	if math.IsNaN(yuan) || math.IsInf(yuan, 0) {
		return 0, errors.New("金额非法")
	}
	// 先限制在合理范围，避免 *100 溢出 int64
	if yuan > float64(MoneyMaxFen)/float64(MoneyFenPerYuan) || yuan < -float64(MoneyMaxFen)/float64(MoneyFenPerYuan) {
		return 0, errors.New("金额超出上限")
	}
	return int64(math.Round(yuan * float64(MoneyFenPerYuan))), nil
}

// FenToYuan 分 → 元（精确两位小数对应的 float64）
func FenToYuan(fen int64) float64 {
	return float64(fen) / float64(MoneyFenPerYuan)
}

// NormalizeYuan 把任意 float 元规范成「分」精度后再变回元（展示/落库前用）
func NormalizeYuan(yuan float64) (float64, error) {
	fen, err := YuanToFen(yuan)
	if err != nil {
		return 0, err
	}
	return FenToYuan(fen), nil
}

// MustYuanToFen 元→分，非法时返回 0（仅用于确定已校验过的路径；优先用 YuanToFen）
func MustYuanToFen(yuan float64) int64 {
	fen, err := YuanToFen(yuan)
	if err != nil {
		return 0
	}
	return fen
}
