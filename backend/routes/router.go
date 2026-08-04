package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/controllers"
	"github.com/yeung/system-hardening/backend/handlers"
	"github.com/yeung/system-hardening/backend/middleware"
	"github.com/yeung/system-hardening/backend/services"
	"gorm.io/gorm"
)

// SetupRouter 设置路由
func SetupRouter(config *configs.Config, ldapService *services.LDAPService, db *gorm.DB) *gin.Engine {
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
	
	// 初始化 Linux 控制器
	linuxController := controllers.NewLinuxController()
	
	// 初始化标准配置控制器
	standardController := controllers.NewStandardController()
	
	// 初始化客户端控制器
	clientController := controllers.NewClientController(db)
	clientHandler := handlers.NewClientHandler(clientController)

	// 初始化区域控制器
	regionController := controllers.NewRegionController(db)

	// 公开路由（无需认证）- 登录接口
	router.POST("/api/auth/login", authHandler.LoginHandler)

	// 客户端 API（无需认证）- 在 API 组之前定义
	clientRouter := router.Group("/api/client")
	{
		// 公共接口：请求临时 Token、注册、刷新 Token、上传数据、心跳
		clientRouter.POST("/request-temp-token", clientHandler.RequestTempToken)
		clientRouter.POST("/register", clientHandler.Register)
		clientRouter.POST("/refresh-token", clientHandler.RefreshToken)
		clientRouter.POST("/upload-data", clientHandler.UploadData)
		clientRouter.POST("/heartbeat", clientHandler.Heartbeat) // 新增心跳接口
	}

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
		
		// Linux 加固检查接口
		api.GET("/linux-checks", linuxController.List)
		api.GET("/linux-checks/:id", linuxController.Detail)
		
		// Linux 标准配置接口
		api.POST("/linux-standards", standardController.CreateStandards)
		api.GET("/linux-standards", standardController.ListStandards)
		api.PUT("/linux-standards/:id", standardController.UpdateStandard)
		api.DELETE("/linux-standards/:id", standardController.DeleteStandard)
		api.GET("/linux-standards/fields", standardController.GetAvailableFields)
		
		// 区域管理接口
		api.POST("/regions", regionController.CreateRegion)
		api.GET("/regions", regionController.ListRegions)
		api.PUT("/regions/:id/clients", regionController.UpdateRegionClients)
		api.DELETE("/regions/:id", regionController.DeleteRegion)
		
		// 客户端管理接口（需认证）
		api.GET("/clients", clientController.ListClients)
		api.DELETE("/clients/:id", clientController.DeleteClient)
	}

	return router
}
