package services

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/gorm"
)

// Scheduler 定时调度器
type Scheduler struct {
	db          *gorm.DB
	mu          sync.Mutex // 互斥锁
	activePlans map[uint]bool // 监控的计划 ID 集
}

// NewScheduler 创建调度器实例
func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		db:          db,
		activePlans: make(map[uint]bool),
	}
}

// Run 启动调度器主循环（每分钟执行一次）
func (s *Scheduler) Run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("✅ Email scheduler started")

	for range ticker.C {
		s.mu.Lock()

		// 1. 从数据库获取所有启用的计划
		plans := s.fetchEnabledSchedules()

		// 2. 比对当前时间与上次发送时间，判断是否需要发送
		for _, plan := range plans {
			if s.shouldSend(&plan) {
				err := s.sendReport(&plan)
				s.updateLastRun(&plan, err)

				if err != nil {
					log.Printf("❌ Failed to send report for plan %d: %v", plan.ID, err)
				} else {
					log.Printf("✅ Report sent successfully for plan %d at %s", plan.ID, time.Now().Format("2006-01-02 15:04:05"))
				}
			}
		}

		s.mu.Unlock()
	}
}

// fetchEnabledSchedules 从数据库获取所有启用的计划
func (s *Scheduler) fetchEnabledSchedules() []models.ReportSchedule {
	var plans []models.ReportSchedule
	err := s.db.Where("is_enabled = ? AND deleted_at IS NULL", true).Find(&plans).Error
	if err != nil {
		log.Printf("Warning: Failed to fetch enabled schedules: %v", err)
		return []models.ReportSchedule{}
	}
	return plans
}

// shouldSend 判断计划是否需要发送报告
func (s *Scheduler) shouldSend(plan *models.ReportSchedule) bool {
	now := time.Now()
	sendTime := s.parseSendTime(plan.SendTime)

	// 检查分钟是否匹配（允许 ±1 分钟的误差）
	if now.Hour() != sendTime.Hour() || now.Minute() != sendTime.Minute() {
		return false
	}

	// 如果是首次发送（LastRunAt 为 null），且当前时间已到发送时间
	if plan.LastRunAt == nil {
		return true
	}

	lastRun := *plan.LastRunAt

	switch plan.ScheduleType {
	case "daily":
		// 每日：只要昨天没发过就可以发
		return lastRun.Before(now.AddDate(0, 0, -1))

	case "every_n_days":
		// 每 N 日：距离上次发送 >= N 天
		daysDiff := int(now.Sub(lastRun).Hours() / 24)
		return daysDiff >= plan.IntervalDays

	case "weekly":
		// 每周：检查星期几且本周未发送（Go 的 Weekday 周日为 0，转换为 7）
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		if weekday != plan.Weekday {
			return false
		}
		// 检查是否是本周首次发送
		startOfWeek := now.AddDate(0, 0, -int(now.Weekday())+1)
		return lastRun.Before(startOfWeek)

	case "every_n_weeks":
		// 每 N 周：距离上次发送 >= N 周
		weeksDiff := int(now.Sub(lastRun).Hours() / (24 * 7))
		return weeksDiff >= plan.IntervalWeeks

	case "monthly":
		// 每月：检查日期且上月未发送
		if now.Day() != plan.DayOfMonth {
			return false
		}
		// 检查是否是本月首次发送
		lastMonth := lastRun.Month()
		currentMonth := now.Month()
		return lastMonth != currentMonth

	case "every_n_months":
		// 每 N 月：检查日期且间隔月份达标
		if now.Day() != plan.DayOfMonth {
			return false
		}
		monthsDiff := (now.Year() - lastRun.Year())*12 + int(now.Month()-lastRun.Month())
		return monthsDiff >= plan.IntervalMonths

	default:
		return false
	}
}

// parseSendTime 解析 HH:mm 格式的时间字符串
func (s *Scheduler) parseSendTime(timeStr string) time.Time {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		// 默认使用 09:00
		t, _ = time.Parse("15:04", "09:00")
	}
	// 使用当前日期
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
}

// sendReport 发送单个计划的报告
func (s *Scheduler) sendReport(plan *models.ReportSchedule) error {
	mailService := NewMailService(s.db)
	html := mailService.GenerateReportHTML(*plan)
	to := strings.Split(plan.Recipients, ",")
	to = SanitizeEmailList(to)
	
	subject := plan.Subject
	if subject == "" {
		subject = "[系统加固平台] 安全合规报告 - " + plan.Name
	}

	return mailService.SendEmail(to, subject, html)
}

// updateLastRun 更新计划的最后运行状态
func (s *Scheduler) updateLastRun(plan *models.ReportSchedule, err error) {
	now := time.Now()
	status := "success"
	lastError := ""

	if err != nil {
		status = "failed"
		lastError = err.Error()
	}

	updates := map[string]interface{}{
		"last_run_at":   now,
		"last_status":   status,
		"last_error":    lastError,
		"updated_at":    now,
	}

	// 如果是非首次发送，更新时间字段可能需要特殊处理
	if plan.LastRunAt != nil {
		updates["created_by"] = plan.CreatedBy // 保留原有创建人
	}

	err = s.db.Model(plan).Updates(updates).Error
	if err != nil {
		log.Printf("Warning: Failed to update schedule %d last run time: %v", plan.ID, err)
	}
}

// SanitizeEmailList 清理邮箱列表，过滤无效邮箱
func SanitizeEmailList(emails []string) []string {
	result := make([]string, 0)
	validSuffixes := []string{"@gmail.com", "@qq.com", "@163.com", "@outlook.com", "@hotmail.com"}

	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		hasValidSuffix := false
		for _, suffix := range validSuffixes {
			if strings.HasSuffix(strings.ToLower(email), suffix) {
				hasValidSuffix = true
				break
			}
		}

		// 如果没有匹配的标准后缀，但包含 @ 符号，也认为有效（可能是企业邮箱）
		if strings.Contains(email, "@") && !hasValidSuffix {
			hasValidSuffix = true
		}

		if hasValidSuffix {
			result = append(result, email)
		}
	}

	return result
}
