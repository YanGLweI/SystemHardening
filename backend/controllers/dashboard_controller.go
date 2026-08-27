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
	LinuxHostCount         int64 `json:"linux_host_count"`
	LinuxCompliantCount    int64 `json:"linux_compliant_count"`
	LinuxNonCompliantCount int64 `json:"linux_non_compliant_count"`

	// Windows 加固统计
	WindowsHostCount         int64 `json:"windows_host_count"`
	WindowsCompliantCount    int64 `json:"windows_compliant_count"`
	WindowsNonCompliantCount int64 `json:"windows_non_compliant_count"`

	// 区域统计
	RegionCount int64 `json:"region_count"`

	// 各区域合规统计
	RegionCompliance []RegionComplianceItem `json:"region_compliance"`

	// 最近新增客户端（最新 5 个）
	RecentClients []RecentClientItem `json:"recent_clients"`
}

// RecentClientItem 最近新增客户端摘要
type RecentClientItem struct {
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
	OSVersion  string    `json:"os_version"`
	CreatedAt  time.Time `json:"created_at"`
}

// RegionComplianceItem 单个区域的合规数量统计
type RegionComplianceItem struct {
	RegionName        string `json:"region_name"`
	CompliantCount    int64  `json:"compliant_count"`
	NonCompliantCount int64  `json:"non_compliant_count"`
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

	// 加载区域及其客户端，构建 clientUUID -> 区域索引映射，用于按区域聚合合规数
	var regions []models.Region
	db.Order("id ASC").Preload("Clients").Find(&regions)
	stats.RegionCompliance = make([]RegionComplianceItem, 0, len(regions))
	clientRegionIdx := make(map[string]int)
	for i, r := range regions {
		stats.RegionCompliance = append(stats.RegionCompliance, RegionComplianceItem{RegionName: r.Name})
		for _, cli := range r.Clients {
			clientRegionIdx[cli.ClientUUID] = i
		}
	}

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

	// 加载 Linux 字段例外配置（clientUUID -> 豁免字段集合）
	linuxExemptionMap := models.LoadExemptionMap(db, "linux")

	// 计算 Linux 合规情况
	for i := range linuxChecks {
		result := models.CompareCompliance(&linuxChecks[i], linuxStandardMap, linuxExemptionMap[linuxChecks[i].ClientUUID])
		idx, inRegion := clientRegionIdx[linuxChecks[i].ClientUUID]
		if result.Status == "compliant" {
			stats.LinuxCompliantCount++
			if inRegion {
				stats.RegionCompliance[idx].CompliantCount++
			}
		} else {
			stats.LinuxNonCompliantCount++
			if inRegion {
				stats.RegionCompliance[idx].NonCompliantCount++
			}
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

	// 加载 Windows 字段例外配置（clientUUID -> 豁免字段集合）
	windowsExemptionMap := models.LoadExemptionMap(db, "windows")

	// 计算 Windows 合规情况
	for i := range windowsChecks {
		result := models.CompareWindowsCompliance(&windowsChecks[i], windowsStandardMap, windowsExemptionMap[windowsChecks[i].ClientUUID])
		idx, inRegion := clientRegionIdx[windowsChecks[i].ClientUUID]
		if result.Status == "compliant" {
			stats.WindowsCompliantCount++
			if inRegion {
				stats.RegionCompliance[idx].CompliantCount++
			}
		} else {
			stats.WindowsNonCompliantCount++
			if inRegion {
				stats.RegionCompliance[idx].NonCompliantCount++
			}
		}
	}

	// 4. 区域统计
	db.Model(&models.Region{}).Where("deleted_at IS NULL").Count(&stats.RegionCount)

	// 5. 最近新增客户端（按创建时间倒序取 5 个）
	var recentClients []models.Client
	db.Model(&models.Client{}).Where("deleted_at IS NULL").Order("created_at DESC, id DESC").Limit(5).Find(&recentClients)
	stats.RecentClients = make([]RecentClientItem, 0, len(recentClients))
	for _, cli := range recentClients {
		stats.RecentClients = append(stats.RecentClients, RecentClientItem{
			DeviceName: cli.DeviceName,
			IPAddress:  cli.IPAddress,
			OSVersion:  cli.OSVersion,
			CreatedAt:  cli.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}
