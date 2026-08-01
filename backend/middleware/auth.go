package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/utils"
)

// JWTAuth 创建 JWT 认证中间件
func JWTAuth(jwtConfig configs.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证 token",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的 authorization header 格式",
			})
			c.Abort()
			return
		}

		// 验证 token
		claims, err := utils.ValidateToken(tokenString, jwtConfig)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token 无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存入 context
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		if claims.UserID != "" {
			c.Set("user_id", claims.UserID)
		}

		c.Next()
	}
}
