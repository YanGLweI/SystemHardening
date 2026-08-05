package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

// DashboardController 看板数据控制器
type DashboardController struct{}

// NewDashboardController 创建看板控制器实例
func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// DashboardStats 看板统计数据
type DashboardStats struct {
	// 客户端统计
	TotalClients   int64 `json:"total_clients"`
	OnlineClients  int64 `json:"online_clients"`
	OfflineClients int64 `json:"offline_clients"`

	// Linux 加固统计
	LinuxHostCount          int64 `json:"linux_host_count"`
	LinuxCompliantCount     int64 `json:"linux_compliant_count"`
	LinuxNonCompliantCount  int64 `json:"linux_non_compliant_count"`

	// Windows 加固统计
	WindowsHostCount          int64 `json:"windows_host_count"`
	WindowsCompliantCount     int64 `json:"windows_compliant_count"`
	WindowsNonCompliantCount  int64 `json:"windows_non_compliant_count"`

	// 区域统计
	RegionCount int64 `json:"region_count"`
}

// GetStats 获取看板统计数据
func (dc *DashboardController) GetStats(c *gin.Context) {
	db := database.DB
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	var stats DashboardStats

	// 1. 客户端在线/离线统计
	db.Model(&models.Client{}).Where("deleted_at IS NULL").Count(&stats.TotalClients)
	db.Model(&models.Client{}).Where("deleted_at IS NULL AND last_check_time >= ?", fiveMinutesAgo).Count(&stats.OnlineClients)
	stats.OfflineClients = stats.TotalClients - stats.OnlineClients

	// 2. Linux 加固检查统计
	var linuxChecks []models.SystemCheck
	db.Model(&models.SystemCheck{}).Where("deleted_at IS NULL").Find(&linuxChecks)
	stats.LinuxHostCount = int64(len(linuxChecks))

	// 获取 Linux 标准配置
	var linuxStandards []models.LinuxStandard
	db.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL").Find(&linuxStandards)
	linuxStandardMap := make(map[string]string)
	for _, std := range linuxStandards {
		linuxStandardMap[std.FieldName] = std.StandardValue
	}

	// 计算 Linux 合规情况
	for i := range linuxChecks {
		result := models.CompareCompliance(&linuxChecks[i], linuxStandardMap)
		if result.Status == "compliant" {
			stats.LinuxCompliantCount++
		} else {
			stats.LinuxNonCompliantCount++
		}
	}

	// 3. Windows 加固检查统计
	var windowsChecks []models.WindowsSystemCheck
	db.Model(&models.WindowsSystemCheck{}).Where("deleted_at IS NULL").Find(&windowsChecks)
	stats.WindowsHostCount = int64(len(windowsChecks))

	// 获取 Windows 标准配置
	var windowsStandards []models.WindowsStandard
	db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&windowsStandards)
	windowsStandardMap := make(map[string]string)
	for _, std := range windowsStandards {
		windowsStandardMap[std.FieldName] = std.StandardValue
	}

	// 计算 Windows 合规情况
	for i := range windowsChecks {
		result := models.CompareWindowsCompliance(&windowsChecks[i], windowsStandardMap)
		if result.Status == "compliant" {
			stats.WindowsCompliantCount++
		} else {
			stats.WindowsNonCompliantCount++
		}
	}

	// 4. 区域统计
	db.Model(&models.Region{}).Where("deleted_at IS NULL").Count(&stats.RegionCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}
