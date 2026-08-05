package models

import (
	"time"

	"gorm.io/gorm"
)

// ReportSchedule 报告计划表
type ReportSchedule struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:100;not null" json:"name"`                                                      // 报告名称
	ScheduleType   string         `gorm:"size:20;not null" json:"schedule_type"`                                               // daily/every_n_days/weekly/every_n_weeks/monthly/every_n_months
	SendTime       string         `gorm:"size:8;not null" json:"send_time"`                                                   // HH:mm
	IntervalDays   int            `gorm:"default:1" json:"interval_days"`                                                     // 间隔天数（every_n_days）
	Weekday        int            `gorm:"default:1" json:"weekday"`                                                           // 星期几（1-7，weekly）
	DayOfMonth     int            `gorm:"default:1" json:"day_of_month"`                                                      // 日期（1-31，monthly）
	IntervalWeeks  int            `gorm:"default:1" json:"interval_weeks"`                                                    // 间隔周数（every_n_weeks）
	IntervalMonths int            `gorm:"default:1" json:"interval_months"`                                                   // 间隔月数（every_n_months）
	Recipients     string         `gorm:"type:text" json:"recipients"`                                                        // 收件人列表（逗号分隔）
	Subject        string         `gorm:"size:200" json:"subject"`                                                            // 邮件主题
	IsEnabled      bool           `gorm:"default:true" json:"is_enabled"`                                                     // 启用状态
	LastRunAt      *time.Time     `json:"last_run_at"`                                                                         // 上次发送时间
	LastStatus     string         `gorm:"size:20" json:"last_status"`                                                         // success/failed
	LastError      string         `gorm:"size:2000" json:"last_error"`                                                        // 错误信息
	CreatedBy      string         `gorm:"size:50" json:"created_by"`                                                          // 创建人
	LastUpdatedBy  string         `gorm:"size:50" json:"last_updated_by"`                                                     // 最后更新人
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (ReportSchedule) TableName() string {
	return "report_schedules"
}
