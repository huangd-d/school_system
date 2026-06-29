package response

import (
	"errors"
	"net/http"

	"school-system/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
// 所有接口统一返回 HTTP 200，用 code 区分：0=成功，非0=错误
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Err 错误响应
// 传入 *AppError 时提取其 Code 和 Message
// 传入普通 error 时统一按 50000（服务器内部错误）处理
func Err(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(http.StatusOK, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
			Data:    nil,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    apperror.ErrInternal.Code,
		Message: err.Error(),
		Data:    nil,
	})
}
