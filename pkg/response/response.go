package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// R 统一响应结构
type R struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, R{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, R{
		Code:    -1,
		Message: message,
	})
}

// ErrorWithCode 带业务错误码的错误响应
func ErrorWithCode(c *gin.Context, httpCode int, bizCode int, message string) {
	c.JSON(httpCode, R{
		Code:    bizCode,
		Message: message,
	})
}
