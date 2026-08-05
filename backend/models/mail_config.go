package models

import (
	"time"

	"gorm.io/gorm"
)

// MailConfig 邮件配置表（单行记录）
type MailConfig struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SMTPHost   string         `gorm:"size:100;not null" json:"smtp_host"`  // SMTP 服务器地址
	SMTPPort   int            `gorm:"not null" json:"smtp_port"`           // SMTP 端口 (25/465/587)
	Username   string         `gorm:"size:100;not null" json:"username"`   // 账号
	Password   string         `gorm:"size:255;not null" json:"-"`          // 密码（不序列化到 JSON）
	FromEmail  string         `gorm:"size:100" json:"from_email"`          // 发件人邮箱（为空则用账号）
	IsEnabled  bool           `gorm:"default:true" json:"is_enabled"`      // 是否启用
	TestResult string         `gorm:"size:500" json:"test_result"`         // 测试结果
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (MailConfig) TableName() string {
	return "mail_configs"
}
