package models

import "time"

// CheckTask 检查任务模型
type CheckTask struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TaskID         string    `gorm:"size:64;uniqueIndex;not null" json:"task_id"`
	ClientUUID     string    `gorm:"size:64;not null;index" json:"client_uuid"`
	ClientName     string    `gorm:"size:128" json:"client_name"`
	IPAddress      string    `gorm:"size:45" json:"ip_address"`
	TriggeredBy    string    `gorm:"size:64" json:"triggered_by"`
	Status         string    `gorm:"size:20;not null;default:'pending'" json:"status"` // pending/sent/executing/completed/failed/timeout
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ResultSummary  string    `gorm:"type:text" json:"result_summary,omitempty"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	RetryCount     int       `gorm:"default:0" json:"retry_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName specify table name
func (CheckTask) TableName() string {
	return "check_tasks"
}

// CheckTaskStatus 任务状态类型
type CheckTaskStatus string

const (
	StatusPending    CheckTaskStatus = "pending"    // 等待客户端拉取
	StatusSent       CheckTaskStatus = "sent"       // 已下发给客户端
	StatusExecuting  CheckTaskStatus = "executing"  // 正在执行
	StatusCompleted  CheckTaskStatus = "completed"  // 已完成
	StatusFailed     CheckTaskStatus = "failed"     // 执行失败
	StatusTimeout    CheckTaskStatus = "timeout"    // 超时
)
