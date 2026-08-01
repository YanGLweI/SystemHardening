package main

import (
	"log"

	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/routes"
	"github.com/yeung/system-hardening/backend/services"
)

func main() {
	// 加载配置文件
	config := configs.LoadConfig()

	// 初始化数据库连接
	database.ConnectDB(config.Database)
	
	// 自动迁移数据表（创建或更新表结构）
	database.AutoMigrate()

	// 初始化 LDAP 服务
	ldapService, err := services.NewLDAPService(config.LDAP)
	if err != nil {
		log.Printf("Warning: Failed to initialize LDAP service: %v", err)
		log.Printf("Note: Starting without LDAP support. Login will fail until LDAP is fixed.")
	}
	defer func() {
		if ldapService != nil {
			ldapService.Close()
		}
	}()

	// 初始化路由（传入配置和 LDAP 服务）
	router := routes.SetupRouter(config, ldapService)

	log.Printf("Server starting on port %s", config.Server.Port)
	if err := router.Run(":" + config.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
