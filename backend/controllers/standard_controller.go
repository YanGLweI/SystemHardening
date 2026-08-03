package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

type StandardController struct{}

// NewStandardController create new StandardController instance
func NewStandardController() *StandardController {
	return &StandardController{}
}

// CreateStandards 批量创建标准配置
func (sc *StandardController) CreateStandards(c *gin.Context) {
	var standards []models.LinuxStandard
	if err := c.ShouldBindJSON(&standards); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := database.DB

	for i := range standards {
		// 检查是否已存在相同 field_name 的记录
		var count int64
		db.Model(&models.LinuxStandard{}).Where("field_name = ? AND deleted_at IS NULL", standards[i].FieldName).Count(&count)
		if count > 0 {
			c.JSON(409, gin.H{
				"error": "字段「" + standards[i].FieldLabel + "」已被配置",
			})
			return
		}

		if err := db.Create(&standards[i]).Error; err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(201, gin.H{
		"message": "添加成功",
		"count":   len(standards),
	})
}

// ListStandards 获取所有标准配置列表
func (sc *StandardController) ListStandards(c *gin.Context) {
	var standards []models.LinuxStandard

	db := database.DB
	query := db.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL")

	// 关键词搜索（字段名或字段标签）
	keyword := c.Query("keyword")
	if keyword != "" {
		query = query.Where("field_name LIKE ? OR field_label LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型分组过滤
	groupBy := c.Query("group_by")
	if groupBy != "" {
		// 构建 field_name 的条件
		var fieldNames []string
		for fieldName, fieldGroup := range FieldGroups {
			if fieldGroup.Group == groupBy {
				fieldNames = append(fieldNames, fieldName)
			}
		}
		if len(fieldNames) > 0 {
			query = query.Where("field_name IN (?)", fieldNames)
		} else {
			// 如果没有匹配的类型，返回空数组
			query = query.Where("1 = 0")
		}
	}

	query.Order("group_name, field_name").Find(&standards)

	// 确保每条记录都有正确的 group_name
	for i := range standards {
		if fg, ok := FieldGroups[standards[i].FieldName]; ok {
			standards[i].GroupName = fg.Group
		}
	}

	c.JSON(200, standards)
}

// UpdateStandard 更新单个标准配置
func (sc *StandardController) UpdateStandard(c *gin.Context) {
	id := c.Param("id")
	data := make(map[string]interface{})

	db := database.DB
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 检查是否与其他记录重复
	var count int64
	db.Model(&models.LinuxStandard{}).Where("field_name = ? AND id != ? AND deleted_at IS NULL", data["field_name"], id).Count(&count)
	if count > 0 {
		c.JSON(409, gin.H{
			"error": "字段已被其他记录使用",
		})
		return
	}

	if err := db.Model(&models.LinuxStandard{}).Where("id = ?", id).Updates(data).Error; err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "更新成功",
	})
}

// DeleteStandard 删除标准配置
func (sc *StandardController) DeleteStandard(c *gin.Context) {
	id := c.Param("id")

	db := database.DB
	if err := db.Delete(&models.LinuxStandard{}, id).Error; err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "删除成功",
	})
}

// GetAvailableFields 获取所有可用的 Linux 加固字段列表（未配置的）
func (sc *StandardController) GetAvailableFields(c *gin.Context) {
	// 查询已存在的记录
	db := database.DB
	var existingRecords []models.LinuxStandard
	db.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL").Find(&existingRecords)

	// 提取已使用的 field_name
	existingFieldNames := make(map[string]bool)
	for _, record := range existingRecords {
		existingFieldNames[record.FieldName] = true
	}

	// 从数据库查询所有可用字段
	var allFields []models.LinuxField
	db.Model(&models.LinuxField{}).Where("deleted_at IS NULL").Order("sort_order, field_name").Find(&allFields)

	// 构建返回的字段列表（排除已配置的字段）
	fields := []gin.H{}
	for _, field := range allFields {
		if !existingFieldNames[field.FieldName] {
			fields = append(fields, gin.H{
				"field_name":  field.FieldName,
				"field_label": field.FieldLabel,
				"group_name":  field.FieldGroup,
			})
		}
	}

	c.JSON(200, fields)
}
