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
// 注意：只针对确有需要迁移的旧表，不再删除 linux_standards
func CleanupIncompatibleTables() {
	// linux_standards 表由于旧结构（field_name 为 bigint）与新模型（varchar）冲突
	// 现在已经通过 migration sql 文件修复，不需要每次都重建
	// 只有在首次安装或明确需要重置时才调用 DROP TABLE
	// 当前没有需要清理的表，此函数保留供未来扩展使用
}

// downgradeHardwareUUIDIndex 将存量库中 clients.hardware_uuid 的唯一索引降级为普通索引。
// v2.2.5~v2.2.8 曾使用 uniqueIndex，但旧版客户端/采集失败会写入空串，
// MySQL 唯一索引不豁免重复空串，会导致注册冲突；GORM AutoMigrate 不会修改已存在的索引，
// 因此需在迁移前显式删除旧索引，再由 AutoMigrate 重建为普通索引。
func downgradeHardwareUUIDIndex() {
	var count int64
	if err := DB.Raw("SELECT COUNT(*) FROM information_schema.STATISTICS " +
		"WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'clients' " +
		"AND INDEX_NAME = 'idx_hardware_uuid' AND NON_UNIQUE = 0").Scan(&count).Error; err != nil {
		log.Printf("Warning: failed to check idx_hardware_uuid: %v", err)
		return
	}
	if count > 0 {
		log.Println("Dropping legacy unique index idx_hardware_uuid (will recreate as normal index)...")
		if err := DB.Exec("ALTER TABLE clients DROP INDEX idx_hardware_uuid").Error; err != nil {
			log.Printf("Warning: failed to drop legacy unique index idx_hardware_uuid: %v", err)
		}
	}
}

// AutoMigrate 自动迁移数据表
func AutoMigrate() {
	// 首先清理可能不兼容的旧表
	CleanupIncompatibleTables()

	// 存量库索引兼容迁移（新装库无此索引时自动跳过）
	downgradeHardwareUUIDIndex()

	err := DB.AutoMigrate(
		&models.SystemCheck{},        // Linux 加固检查模型
		&models.Client{},             // 客户端管理模型
		&models.ClientToken{},        // Token 管理模型
		&models.LinuxStandard{},      // Linux 标准配置模型
		&models.LinuxField{},         // Linux 字段定义模型
		&models.Region{},             // 区域管理模型
		&models.WindowsSystemCheck{}, // Windows 加固检查模型
		&models.WindowsStandard{},    // Windows 标准配置模型
		&models.WindowsField{},       // Windows 字段定义模型
		&models.MailConfig{},         // 邮件配置模型
		&models.ReportSchedule{},     // 报告计划模型
		&models.CheckSchedule{},      // 加固检查计划模型
		&models.PackageMeta{},        // 安装包元数据模型
		&models.StandardExemption{},  // 标准字段例外模型
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
