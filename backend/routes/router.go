package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/handlers"
	"github.com/yeung/system-hardening/backend/middleware"
	"github.com/yeung/system-hardening/backend/services"
)

// SetupRouter 设置路由
func SetupRouter(config *configs.Config, ldapService *services.LDAPService) *gin.Engine {
	router := gin.Default()

	// CORS 配置
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}
	router.Use(cors.New(corsConfig))

	// 中间件
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// 初始化认证处理器
	authHandler := handlers.NewAuthHandler(ldapService, config.JWT)

	// 公开路由（无需认证）- 登录接口
	router.POST("/api/auth/login", authHandler.LoginHandler)

	// API 路由组（需要认证）
	api := router.Group("/api")
	api.Use(middleware.JWTAuth(config.JWT))
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})

		// 示例路由
		api.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "API is working",
			})
		})

		// 用户相关接口
		api.GET("/profile", authHandler.GetProfileHandler)
	}

	return router
}
