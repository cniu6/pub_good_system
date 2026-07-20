package utils

import "github.com/gin-gonic/gin"

// CtxBizOK 业务是否成功标记（供幂等中间件在 c.Next 后判断：成功锁定 / 失败释放）
const CtxBizOK = "bizOK"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.Set(CtxBizOK, true)
	c.JSON(200, Response{
		Code:    200,
		Message: "OK",
		Data:    data,
	})
}

func SuccessMsg(c *gin.Context, message string, data any) {
	c.Set(CtxBizOK, true)
	c.JSON(200, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// httpStatus 把业务 code 映射为真实 HTTP 状态码。
// 项目里 Fail 传入的 code 约定就是标准 HTTP 错误码（400/401/403/404/409/429/500 等），
// 落在合法错误状态区间 [400,599] 时直接原样作为 HTTP 状态，
// 让网关/中间件按 c.Writer.Status() 统计的 4xx/5xx 才准确；
// 否则（如历史遗留的非标准业务码）兜底成 500，避免 c.JSON 传入非法状态码。
func httpStatus(code int) int {
	if code >= 400 && code <= 599 {
		return code
	}
	return 500
}

func Fail(c *gin.Context, code int, message string) {
	c.Set(CtxBizOK, false)
	c.JSON(httpStatus(code), Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
