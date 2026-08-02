package models

import (
	"time"
	"gorm.io/gorm"
)

// LinuxField 加固字段定义表 - 定义每个可用的加固字段
type LinuxField struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FieldName      string    `gorm:"size:50;not null;uniqueIndex" json:"field_name"`     // 字段名：pass_max_days
	FieldLabel     string    `gorm:"size:50;not null" json:"field_label"`               // 字段标签：PASS_MAX_DAYS
	FieldGroup     string    `gorm:"size:50" json:"field_group"`                        // 所属分组
	Category       string    `gorm:"size:50" json:"category"`                           // 业务分类（用于前端 Tab 分组）
	SortOrder      int       `gorm:"default:0" json:"sort_order"`                       // 排序顺序
	Description    string    `gorm:"size:500" json:"description"`                       // 字段描述
	IsRequired     bool      `gorm:"default:false" json:"is_required"`                  // 是否必填
	DataType       string    `gorm:"size:20;default:'string'" json:"data_type"`         // 数据类型：string/number/int/bool
	DefaultValue   string    `gorm:"size:100" json:"default_value"`                     // 默认值
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (LinuxField) TableName() string {
	return "linux_fields"
}
