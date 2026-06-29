package middleware

import "github.com/gin-gonic/gin"

// DataScope 数据隔离中间件
// 总部管理员不限制，校区操作员限制本校区，活动联系人限制关联活动
// 具体过滤逻辑在 service 层根据 user_id、campus_id、role 实现
func DataScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
