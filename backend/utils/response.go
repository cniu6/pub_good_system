package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code:    200,
		Message: "OK",
		Data:    data,
	})
}

func SuccessMsg(c *gin.Context, message string, data any) {
	c.JSON(200, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(200, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
