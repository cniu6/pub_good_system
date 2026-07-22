package utils

import "testing"

func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		page, pageSize         int
		wantPage, wantPageSize int
	}{
		{0, 0, 1, 20},
		{-1, -5, 1, 20},
		{2, 50, 2, 50},
		{3, 999, 3, 100}, // 超过上限裁剪
		{1, 100, 1, 100}, // 正好等于上限
	}
	for _, c := range cases {
		gotPage, gotPageSize := NormalizePagination(c.page, c.pageSize)
		if gotPage != c.wantPage || gotPageSize != c.wantPageSize {
			t.Errorf("NormalizePagination(%d,%d) = (%d,%d), want (%d,%d)",
				c.page, c.pageSize, gotPage, gotPageSize, c.wantPage, c.wantPageSize)
		}
	}
}

func TestNormalizePaginationWithLimits(t *testing.T) {
	page, pageSize := NormalizePaginationWithLimits(0, 0, 50, 200)
	if page != 1 || pageSize != 50 {
		t.Errorf("got (%d,%d), want (1,50)", page, pageSize)
	}
	// maxSize <= 0 表示不限制上限
	page, pageSize = NormalizePaginationWithLimits(1, 99999, 50, 0)
	if pageSize != 99999 {
		t.Errorf("maxSize<=0 应不限制上限，got %d", pageSize)
	}
}

func TestNormalizeTimeRange(t *testing.T) {
	// 全部默认：结束=当前，开始=结束-30天
	start, end, err := NormalizeTimeRange(0, 0, 30, 365)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if end-start != 30*24*60*60 {
		t.Errorf("默认时间窗应为 30 天，got %d 秒", end-start)
	}

	// defaultDays 超过 maxDays 应被裁剪
	start2, end2, err := NormalizeTimeRange(0, 0, 400, 365)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if end2-start2 != 365*24*60*60 {
		t.Errorf("defaultDays 应被裁剪到 365 天，got %d 秒", end2-start2)
	}

	// start > end 应返回错误
	_, _, err = NormalizeTimeRange(200, 100, 30, 365)
	if err != ErrInvalidTimeRange {
		t.Errorf("start>end 应返回 ErrInvalidTimeRange, got %v", err)
	}

	// 显式传入的 start/end 应原样保留（在合法范围内）
	start3, end3, err := NormalizeTimeRange(1000, 5000, 30, 365)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start3 != 1000 || end3 != 5000 {
		t.Errorf("显式传入的时间应保留，got (%d,%d)", start3, end3)
	}
}
