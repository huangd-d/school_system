package middleware

import (
	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 异常恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic 已恢复", zap.Any("错误信息", err))
				response.Err(c, apperror.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}
