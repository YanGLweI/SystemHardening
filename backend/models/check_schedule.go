package models

import (
	"time"
)

// CheckSchedule 加固检查计划表（全局单一计划，所有客户端共用）
type CheckSchedule struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ScheduleType string    `gorm:"size:20;not null;default:'daily'" json:"schedule_type"` // daily/weekly/monthly
	CheckTime    string    `gorm:"size:5;not null;default:'02:00'" json:"check_time"`     // HH:mm，半小时粒度
	Weekday      int       `gorm:"default:1" json:"weekday"`                              // 星期几（1-7，weekly）
	DayOfMonth   int       `gorm:"default:1" json:"day_of_month"`                         // 日期（1-31，monthly）
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specify table name
func (CheckSchedule) TableName() string {
	return "check_schedules"
}
