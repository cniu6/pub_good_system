package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParamID 解析路径参数为 uint64；失败时已写好 400 响应，调用方直接 return。
// 收敛各 controller 里重复的 ParseUint(c.Param(...)) + Fail 样板。
func ParamID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		Fail(c, 400, "Invalid ID")
		return 0, false
	}
	return id, true
}

// QueryPage 读取 page/page_size 并套用统一默认值与上限（空值走 NormalizePagination 的默认值）。
func QueryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	return NormalizePagination(page, pageSize)
}
