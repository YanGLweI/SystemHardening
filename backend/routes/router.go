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
	
	// 初始化 Windows 控制器
	windowsController := controllers.NewWindowsController()
	
	// 初始化标准配置控制器
	standardController := controllers.NewStandardController()
	
	// 初始化客户端控制器
	clientController := controllers.NewClientController(db)
	clientHandler := handlers.NewClientHandler(clientController)
	// 设置全局配置供包管理器使用
	controllers.SetGlobalConfig(config)

	// 初始化区域控制器
	regionController := controllers.NewRegionController(db)
	
	// 初始化看板控制器
	dashboardController := controllers.NewDashboardController()
	
	// 初始化邮件服务
	mailService := services.NewMailService(db)

	// 初始化任务控制器
	taskController := controllers.NewTaskController()

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
		clientRouter.POST("/upload-data-windows", clientHandler.UploadWindowsData)
		clientRouter.POST("/heartbeat", clientHandler.Heartbeat) // 新增心跳接口
		clientRouter.GET("/check-update", clientHandler.CheckUpdate) // 检查更新接口
		clientRouter.GET("/check-schedule", clientHandler.GetCheckSchedule) // 获取加固检查计划接口
		// 任务管理接口
		clientRouter.GET("/tasks/pending", taskController.GetPendingTasksForClient) // 客户端拉取待执行任务
		clientRouter.PUT("/tasks/:id/result", taskController.SubmitTaskResult)       // 客户端上报执行结果
	}
	
	// 安装包下载接口（公开，无需认证）
	downloads := router.Group("/api/packages")
	{
		downloads.GET("/:type/download", clientController.DownloadPackage)
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
		
		// Windows 加固检查接口
		api.GET("/windows-checks", windowsController.List)
		api.GET("/windows-checks/:id", windowsController.Detail)
		
		// Linux 标准配置接口
		api.POST("/linux-standards", standardController.CreateStandards)
		api.GET("/linux-standards", standardController.ListStandards)
		api.PUT("/linux-standards/:id", standardController.UpdateStandard)
		api.DELETE("/linux-standards/:id", standardController.DeleteStandard)
		api.GET("/linux-standards/fields", standardController.GetAvailableFields)
		api.GET("/linux-standards/exemptions", standardController.ListLinuxExemptions)
		api.PUT("/linux-standards/:id/exemptions", standardController.UpdateLinuxExemptions)
		
		// Windows 标准配置接口
		api.POST("/windows-standards", standardController.CreateWindowsStandards)
		api.GET("/windows-standards", standardController.ListWindowsStandards)
		api.PUT("/windows-standards/:id", standardController.UpdateWindowsStandard)
		api.DELETE("/windows-standards/:id", standardController.DeleteWindowsStandard)
		api.GET("/windows-standards/fields", standardController.GetAvailableWindowsFields)
		api.GET("/windows-standards/exemptions", standardController.ListWindowsExemptions)
		api.PUT("/windows-standards/:id/exemptions", standardController.UpdateWindowsExemptions)
		
		// 区域管理接口
		api.POST("/regions", regionController.CreateRegion)
		api.GET("/regions", regionController.ListRegions)
		api.PUT("/regions/:id", regionController.UpdateRegion)
		api.PUT("/regions/:id/clients", regionController.UpdateRegionClients)
		api.DELETE("/regions/:id", regionController.DeleteRegion)
		
		// 客户端管理接口（需认证）
		api.GET("/clients", clientController.ListClients)
		api.DELETE("/clients/:id", clientController.DeleteClient)

		// 加固检查计划接口（需认证）
		api.GET("/check-schedule", clientController.GetCheckSchedule)
		api.PUT("/check-schedule", clientController.SaveCheckSchedule)
		
		// 安装包管理接口（需认证）
		api.POST("/packages/upload", clientController.UploadPackage)
		api.GET("/packages/:type/info", clientController.GetPackageInfo)
		
		// 任务管理接口（需认证）- 必须在 /packages/:type/info 之前定义
		api.POST("/tasks/trigger", taskController.TriggerCheckTask) // 触发立即检查任务
		api.GET("/tasks/:id", taskController.GetTaskStatus)         // 查询任务状态
		api.DELETE("/tasks/:id", taskController.DeleteTask)         // 删除任务（卡死任务重试）
		api.PUT("/tasks/:id/result", taskController.SubmitTaskResult) // 客户端上报结果
		api.GET("/tasks/client/:client_uuid", taskController.GetClientLatestTask) // 获取客户端最新任务
			
		// 看板统计接口
		api.GET("/dashboard/stats", dashboardController.GetStats)
		
		// 邮件通知配置接口（需认证）
		mailController := controllers.NewMailController(db, mailService)
		api.GET("/mail-config", mailController.GetMailConfig)
		api.PUT("/mail-config", mailController.SaveMailConfig)
		api.POST("/mail/test", mailController.TestEmail)
		
		// 报告计划接口（需认证）
		api.GET("/report-schedules", mailController.ListSchedules)
		api.POST("/report-schedules", mailController.CreateSchedule)
		api.PUT("/report-schedules/:id", mailController.UpdateSchedule)
		api.DELETE("/report-schedules/:id", mailController.DeleteSchedule)
		api.POST("/report-schedules/:id/send", mailController.ImmediateSend)
	}

	return router
}
