package middleware

import (
	"strings"

	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth JWT 鉴权中间件
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Err(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Err(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Err(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Err(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		// 注入用户信息到上下文
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("campus_id", uint(claims["campus_id"].(float64)))
		c.Set("role", claims["role"].(string))

		c.Next()
	}
}
