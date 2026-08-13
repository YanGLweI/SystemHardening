package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

type WindowsController struct{}

// NewWindowsController create new WindowsController instance
func NewWindowsController() *WindowsController {
	return &WindowsController{}
}

// List 获取 Windows 加固检查列表（分页）
func (wc *WindowsController) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	keyword := c.Query("keyword")
	complianceStatus := c.Query("compliance_status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var checks []models.WindowsSystemCheck
	var total int64

	db := database.DB

	// 获取所有 Windows 标准配置
	var standards []models.WindowsStandard
	db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&standards)

	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}

	// 加载字段例外配置（clientUUID -> 豁免字段集合）
	exemptionMap := models.LoadExemptionMap(db, "windows")

	// 如果需要按合规状态过滤，需要先获取全部数据再内存过滤
	if complianceStatus != "" {
		var allChecks []models.WindowsSystemCheck
		query := db.Model(&models.WindowsSystemCheck{})
		if keyword != "" {
			query = query.Where("hostname LIKE ? OR ip LIKE ? OR domainname LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
		if err := query.Order("id DESC").Find(&allChecks).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch data"})
			return
		}
		// 计算合规状态并过滤
		for i := range allChecks {
			result := models.CompareWindowsCompliance(&allChecks[i], standardMap, exemptionMap[allChecks[i].ClientUUID])
			allChecks[i].ComplianceStatus = result.Status
		}
		for _, check := range allChecks {
			if check.ComplianceStatus == complianceStatus {
				checks = append(checks, check)
			}
		}
		total = int64(len(checks))
		// 手动分页
		offset := (page - 1) * pageSize
		end := offset + pageSize
		if offset > len(checks) {
			checks = []models.WindowsSystemCheck{}
		} else {
			if end > len(checks) {
				end = len(checks)
			}
			checks = checks[offset:end]
		}
		// nil 切片会被序列化为 null，前端无法识别，统一转为空数组
		if checks == nil {
			checks = []models.WindowsSystemCheck{}
		}
	} else {
		// 无需合规过滤，正常 SQL 分页
		query := db.Model(&models.WindowsSystemCheck{})
		if keyword != "" {
			query = query.Where("hostname LIKE ? OR ip LIKE ? OR domainname LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
		query.Count(&total)
		if err := query.Order("id DESC").Limit(pageSize).Offset((page-1)*pageSize).Find(&checks).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch data"})
			return
		}
		for i := range checks {
			result := models.CompareWindowsCompliance(&checks[i], standardMap, exemptionMap[checks[i].ClientUUID])
			checks[i].ComplianceStatus = result.Status
		}
	}

	c.JSON(200, gin.H{
		"list":     checks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Detail 获取单个 Windows 主机加固检查详情
func (wc *WindowsController) Detail(c *gin.Context) {
	id := c.Param("id")

	var check models.WindowsSystemCheck
	if err := database.DB.First(&check, id).Error; err != nil {
		c.JSON(404, gin.H{
			"error": "Record not found",
		})
		return
	}

	// 获取所有标准配置
	var standards []models.WindowsStandard
	database.DB.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&standards)

	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}

	// 加载字段例外配置（clientUUID -> 豁免字段集合）
	exemptionMap := models.LoadExemptionMap(database.DB, "windows")

	// 计算合规状态
	result := models.CompareWindowsCompliance(&check, standardMap, exemptionMap[check.ClientUUID])
	check.ComplianceStatus = result.Status

	c.JSON(200, gin.H{
		"check":      check,
		"compliance": result,
	})
}

// 合规比对逻辑已迁移至 models 包 (models/compliance.go)