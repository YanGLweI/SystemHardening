package database

import (
	"fmt"
	"log"

	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(config configs.DatabaseConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully")
}

// CleanupIncompatibleTables 清理与新的数据模型不兼容的旧表
func CleanupIncompatibleTables() {
	// linux_standards 表由于旧结构（field_name 为 bigint）与新模型（varchar）冲突，必须重建
	// 使用 Raw SQL 检查表是否存在
	err := DB.Exec("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'linux_standards'").Error
	if err != nil {
		log.Printf("Warning: Failed to check linux_standards table existence: %v\n", err)
		return
	}
	log.Println("Checking linux_standards table - will drop if exists to recreate with correct schema")
	
	// 强制删除旧表（如果存在），让 GORM 重新创建正确结构
	err = DB.Exec("DROP TABLE IF EXISTS `linux_standards`").Error
	if err != nil {
		log.Printf("Warning: Failed to drop linux_standards: %v\n", err)
	} else {
		log.Println("Successfully dropped linux_standards table for recreation by GORM AutoMigrate")
	}
}

// AutoMigrate 自动迁移数据表
func AutoMigrate() {
	// 首先清理可能不兼容的旧表
	CleanupIncompatibleTables()

	err := DB.AutoMigrate(
		&models.User{},             // 用户模型
		&models.SystemCheck{},      // Linux 加固检查模型
		&models.Client{},           // 客户端管理模型
		&models.ClientToken{},      // Token 管理模型
		&models.LinuxStandard{},    // Linux 标准配置模型
		&models.LinuxField{},       // Linux 字段定义模型
	)

	if err != nil {
		log.Fatalf("AutoMigration failed: %v", err)
	}

	log.Println("Database tables migrated successfully")
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
