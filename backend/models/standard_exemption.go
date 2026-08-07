package models

import (
	"time"

	"gorm.io/gorm"
)

// StandardExemption 标准字段例外配置（字段 × 客户端维度）
// 被例外的客户端在合规比对时跳过该字段的标准值校验
type StandardExemption struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FieldType  string         `gorm:"size:10;uniqueIndex:uk_exemption,priority:1" json:"field_type"`    // linux / windows
	FieldName  string         `gorm:"size:100;uniqueIndex:uk_exemption,priority:2" json:"field_name"`   // 标准字段名
	ClientUUID string         `gorm:"size:64;uniqueIndex:uk_exemption,priority:3;index" json:"client_uuid"` // 客户端 UUID
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (StandardExemption) TableName() string {
	return "standard_exemptions"
}
