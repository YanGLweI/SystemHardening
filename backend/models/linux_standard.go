package models

import (
	"time"
	"gorm.io/gorm"
)

// LinuxStandard Linux 标准配置记录
type LinuxStandard struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FieldName     string         `gorm:"size:50;not null;index:idx_field_name" json:"field_name"` // 字段名：pass_max_days
	FieldLabel    string         `gorm:"size:50;not null" json:"field_label"`                     // 字段标签：PASS_MAX_DAYS
	StandardValue string         `gorm:"size:200;not null" json:"standard_value"`                // 标准值
	GroupName     string         `gorm:"size:50;not null;index:idx_group_name" json:"group_name"` // 分组：系统更新/用户账户策略等
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (LinuxStandard) TableName() string {
	return "linux_standards"
}
