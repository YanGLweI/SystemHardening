package models

import (
	"time"

	"gorm.io/gorm"
)

// Region 区域模型
type Region struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"size:50;not null;index:idx_name" json:"name"` // 区域名（重名由应用层检查，避免软删除后无法重建）
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Clients   []Client   `gorm:"many2many:region_clients;" json:"clients,omitempty"`
}

// TableName specify table name
func (Region) TableName() string {
	return "regions"
}

// RegionClientItem 区域客户端信息（用于 API 响应）
type RegionClientItem struct {
	ID        uint   `json:"id"`
	DeviceName string `json:"device_name"`
	IPAddress   string `json:"ip_address"`
}
