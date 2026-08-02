package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

type LinuxController struct {
}

// NewLinuxController create new LinuxController instance
func NewLinuxController() *LinuxController {
	return &LinuxController{}
}

// List 获取 Linux 加固检查列表（分页）
func (lc *LinuxController) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var checks []models.SystemCheck
	var total int64

	db := database.DB

	// 获取总数
	db.Model(&models.SystemCheck{}).Count(&total)

	// 计算偏移量
	offset := (page - 1) * pageSize
	
	// 分页查询
	if err := db.Order("id DESC").Limit(pageSize).Offset(offset).Find(&checks).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to fetch data",
		})
		return
	}

	// 获取所有标准配置
	var standards []models.LinuxStandard
	db.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL").Find(&standards)
	
	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}
	
	// 为每条记录计算合规状态
	for i := range checks {
		result := CompareCompliance(&checks[i], standardMap)
		checks[i].ComplianceStatus = result.Status
	}

	c.JSON(200, gin.H{
		"list":     checks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Detail 获取单个主机加固检查详情
func (lc *LinuxController) Detail(c *gin.Context) {
	id := c.Param("id")

	var check models.SystemCheck
	if err := database.DB.First(&check, id).Error; err != nil {
		c.JSON(404, gin.H{
			"error": "Record not found",
		})
		return
	}

	// 获取所有标准配置
	var standards []models.LinuxStandard
	database.DB.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL").Find(&standards)
	
	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}
	
	// 计算合规状态
	result := CompareCompliance(&check, standardMap)
	check.ComplianceStatus = result.Status

	c.JSON(200, gin.H{
		"check": check,
		"compliance": result,
	})
}
