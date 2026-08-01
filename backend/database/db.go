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

// AutoMigrate 自动迁移数据表
func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},           // 用户模型
		&models.SystemCheck{},    // Linux 加固检查模型
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
