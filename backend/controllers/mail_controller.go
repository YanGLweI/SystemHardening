package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/models"
	"github.com/yeung/system-hardening/backend/services"
	"gorm.io/gorm"
)

// MailController 邮件控制器
type MailController struct {
	db          *gorm.DB
	mailService *services.MailService
}

// NewMailController 创建邮件控制器
func NewMailController(db *gorm.DB, mailService *services.MailService) *MailController {
	return &MailController{
		db:          db,
		mailService: mailService,
	}
}

// GetMailConfigRequest 获取邮件配置请求
type GetMailConfigRequest struct {
}

// GetMailConfigResponse 获取邮件配置响应
type GetMailConfigResponse struct {
	ID        uint   `json:"id"`
	SMTPHost  string `json:"smtp_host"`
	SMTPPort  int    `json:"smtp_port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	IsEnabled bool   `json:"is_enabled"`
}

// SaveMailConfigRequest 保存邮件配置请求
type SaveMailConfigRequest struct {
	SMTPHost  string `json:"smtp_host" binding:"required"`
	SMTPPort  int    `json:"smtp_port" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FromEmail string `json:"from_email"`
	IsEnabled bool   `json:"is_enabled"`
}

// TestEmailRequest 测试邮件请求
type TestEmailRequest struct {
	Recipient string `json:"recipient" binding:"required"`
}

// ReportScheduleRequest 报告计划创建/更新请求
type ReportScheduleRequest struct {
	Name             string `json:"name" binding:"required"`
	ScheduleType     string `json:"schedule_type" binding:"required"` // daily/every_n_days/weekly/every_n_weeks/monthly/every_n_months
	SendTime         string `json:"send_time" binding:"required"`    // HH:mm
	IntervalDays     int    `json:"interval_days"`
	Weekday          int    `json:"weekday"`
	DayOfMonth       int    `json:"day_of_month"`
	IntervalWeeks    int    `json:"interval_weeks"`
	IntervalMonths   int    `json:"interval_months"`
	Recipients       string `json:"recipients" binding:"required"`
	Subject          string `json:"subject"`
	IsEnabled        bool   `json:"is_enabled"`
	CreatedBy        string `json:"created_by"`
	LastUpdatedBy    string `json:"last_updated_by"`
}

// ListSchedulesResponse 列表响应
type ListSchedulesResponse struct {
	List []models.ReportSchedule `json:"list"`
	Total int                       `json:"total"`
	Page  int                       `json:"page"`
	PageSize int                     `json:"page_size"`
}

// 1. 获取邮件配置
func (mc *MailController) GetMailConfig(c *gin.Context) {
	var config models.MailConfig
	
	// 查询最新的配置
	err := mc.db.Last(&config).Error
	if err != nil {
		// 如果没有配置，返回默认值
		config = models.MailConfig{
			SMTPPort: 587,
			IsEnabled: true,
		}
		
		c.JSON(http.StatusOK, gin.H{
			"id":           0,
			"smtp_host":    "",
			"smtp_port":    config.SMTPPort,
			"username":     "",
			"password":     "",
			"from_email":   "",
			"is_enabled":   config.IsEnabled,
			"test_result":  "",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"id":          config.ID,
		"smtp_host":   config.SMTPHost,
		"smtp_port":   config.SMTPPort,
		"username":    config.Username,
		"password":    "", // 不返回密码
		"from_email":  config.FromEmail,
		"is_enabled":  config.IsEnabled,
		"test_result": config.TestResult,
	})
}

// 2. 保存邮件配置
func (mc *MailController) SaveMailConfig(c *gin.Context) {
	var req SaveMailConfigRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 查询现有配置
	var existingConfig models.MailConfig
	result := mc.db.Last(&existingConfig)
	
	nowTime := time.Now()
	
	if result.Error == gorm.ErrRecordNotFound || existingConfig.ID == 0 {
		// 创建新配置
		config := models.MailConfig{
			SMTPHost:  req.SMTPHost,
			SMTPPort:  req.SMTPPort,
			Username:  req.Username,
			Password:  req.Password,
			FromEmail: req.FromEmail,
			IsEnabled: req.IsEnabled,
		}
		
		if err := mc.db.Create(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save mail config"})
			return
		}
		
		c.JSON(http.StatusCreated, gin.H{
			"message": "Mail config created successfully",
			"id":      config.ID,
		})
	} else {
		// 更新现有配置（密码为空时保留原密码，避免覆盖为空白）
		password := req.Password
		if password == "" {
			password = existingConfig.Password
		}
		
		config := models.MailConfig{
			ID:        existingConfig.ID,
			SMTPHost:  req.SMTPHost,
			SMTPPort:  req.SMTPPort,
			Username:  req.Username,
			Password:  password,
			FromEmail: req.FromEmail,
			IsEnabled: req.IsEnabled,
			UpdatedAt: nowTime,
		}
		
		if err := mc.db.Save(&config).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mail config"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Mail config updated successfully",
		})
	}
}

// 3. 发送测试邮件
func (mc *MailController) TestEmail(c *gin.Context) {
	var req TestEmailRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := mc.mailService.TestEmail(req.Recipient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send test email: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
	})
}

// 4. 报告计划列表
func (mc *MailController) ListSchedules(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	var total int64
	var plans []models.ReportSchedule
	
	db := mc.db.Model(&models.ReportSchedule{})
	db.Where("deleted_at IS NULL").Count(&total)
	
	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"list":     plans,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

// 5. 创建报告计划
func (mc *MailController) CreateSchedule(c *gin.Context) {
	var req ReportScheduleRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	now := time.Now()
	plan := models.ReportSchedule{
		Name:            req.Name,
		ScheduleType:    req.ScheduleType,
		SendTime:        req.SendTime,
		IntervalDays:    req.IntervalDays,
		Weekday:         req.Weekday,
		DayOfMonth:      req.DayOfMonth,
		IntervalWeeks:   req.IntervalWeeks,
		IntervalMonths:  req.IntervalMonths,
		Recipients:      req.Recipients,
		Subject:         req.Subject,
		IsEnabled:       req.IsEnabled,
		CreatedBy:       req.CreatedBy,
		LastUpdatedBy:   req.LastUpdatedBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	
	if err := mc.db.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Schedule created successfully",
		"id":      plan.ID,
	})
}

// 6. 更新报告计划
func (mc *MailController) UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	var req ReportScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var plan models.ReportSchedule
	if err := mc.db.First(&plan, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}
	
	now := time.Now()
	plan.Name = req.Name
	plan.ScheduleType = req.ScheduleType
	plan.SendTime = req.SendTime
	plan.IntervalDays = req.IntervalDays
	plan.Weekday = req.Weekday
	plan.DayOfMonth = req.DayOfMonth
	plan.IntervalWeeks = req.IntervalWeeks
	plan.IntervalMonths = req.IntervalMonths
	plan.Recipients = req.Recipients
	plan.Subject = req.Subject
	plan.IsEnabled = req.IsEnabled
	plan.LastUpdatedBy = req.LastUpdatedBy
	plan.UpdatedAt = now
	
	if err := mc.db.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Schedule updated successfully"})
}

// 7. 删除报告计划
func (mc *MailController) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	if err := mc.db.Delete(&models.ReportSchedule{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
}

// 8. 立即发送报告
func (mc *MailController) ImmediateSend(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	var plan models.ReportSchedule
	if err := mc.db.First(&plan, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}
	
	html := mc.mailService.GenerateReportHTML(plan)
	to := strings.Split(plan.Recipients, ",")
	to = services.SanitizeEmailList(to)
	
	subject := plan.Subject
	if subject == "" {
		subject = "[系统加固平台] 安全合规报告（手动触发）- " + plan.Name
	}
	
	err = mc.mailService.SendEmail(to, subject, html)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send report: " + err.Error()})
		return
	}
	
	// 更新状态
	now := time.Now()
	mc.db.Model(&plan).Updates(models.ReportSchedule{
		LastRunAt:   &now,
		LastStatus:  "success",
		LastError:   "",
		UpdatedAt:   now,
	})
	
	c.JSON(http.StatusOK, gin.H{"message": "Report sent successfully"})
}
