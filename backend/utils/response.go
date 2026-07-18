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

func Fail(c *gin.Context, code int, message string) {
	c.Set(CtxBizOK, false)
	c.JSON(200, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
