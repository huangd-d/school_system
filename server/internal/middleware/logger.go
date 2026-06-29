package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 请求日志中间件
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		logger.Info("请求日志",
			zap.Int("状态码", c.Writer.Status()),
			zap.String("方法", c.Request.Method),
			zap.String("路径", path),
			zap.Duration("耗时", latency),
		)
	}
}
