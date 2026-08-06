package models

import (
	"time"

	"gorm.io/gorm"
)

// Client Linux 加固客户端信息
type Client struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientUUID    string         `gorm:"uniqueIndex:idx_client_uuid;size:64" json:"client_uuid"`                // 客户端唯一标识
	DeviceName    string         `gorm:"column:device_name;size:100" json:"device_name"`                         // 设备名称 (Hostname)
	IPAddress     string         `gorm:"column:ip_address;size:50" json:"ip_address"`                            // IP 地址
	OSVersion     string         `gorm:"column:os_version;size:100" json:"os_version"`                           // 操作系统版本
	ClientVersion string         `gorm:"column:client_version;size:50" json:"client_version"`                    // 客户端实际运行版本
	Status        string         `gorm:"column:status;size:20;default:'active'" json:"status"`                   // active|inactive|expired
	LastCheckTime *time.Time     `gorm:"column:last_check_time" json:"last_check_time"`                          // 最后检查时间
	LastUploadTime *time.Time     `gorm:"column:last_upload_time" json:"last_upload_time"`                        // 最后上传时间
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (Client) TableName() string {
	return "clients"
}

// ClientToken 客户端 Token 管理表
type ClientToken struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientUUID   string    `gorm:"size:64" json:"client_uuid"` // 外键字段
	RefreshToken string    `gorm:"uniqueIndex:idx_refresh_token;size:255" json:"refresh_token"` // 长期刷新令牌 (90 天)
	ShortToken   string    `gorm:"uniqueIndex:idx_short_token;size:255" json:"short_token"`     // 短期访问令牌 (7 天)
	ExpiresAt    time.Time `gorm:"index:idx_expires_at" json:"expires_at"`                     // 短期令牌过期时间
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specify table name
func (ClientToken) TableName() string {
	return "client_tokens"
}

// PackageMeta 安装包元数据表
type PackageMeta struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string    `gorm:"type:varchar(20);not null;uniqueIndex:uk_type" json:"type"` // linux 或 windows
	Version   string    `gorm:"type:varchar(50);not null" json:"version"`                 // 版本号 (语义化版本)
	Hash      string    `gorm:"type:varchar(64);not null" json:"hash"`                    // MD5 哈希值
	Size      int64     `gorm:"type:bigint;not null" json:"size"`                        // 文件大小（字节）
	Filename  string    `gorm:"type:varchar(255);not null" json:"filename"`              // 文件名
	Filepath  string    `gorm:"type:varchar(500);not null" json:"filepath"`              // 文件路径
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specify table name
func (PackageMeta) TableName() string {
	return "package_meta"
}
