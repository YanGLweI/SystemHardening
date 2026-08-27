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

	// 先查出 field_name，用于同步清理该字段的例外配置
	var std models.LinuxStandard
	if err := db.First(&std, id).Error; err == nil {
		db.Where("field_type = ? AND field_name = ?", "linux", std.FieldName).Delete(&models.StandardExemption{})
	}

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

// ======================== 标准字段例外管理 ========================

// listExemptions 通用实现：获取指定类型的全部字段例外列表
func (sc *StandardController) listExemptions(c *gin.Context, fieldType string) {
	db := database.DB
	var exemptions []models.StandardExemption
	db.Model(&models.StandardExemption{}).Where("field_type = ?", fieldType).Find(&exemptions)

	// 收集客户端 UUID，批量查询显示信息
	uuidSet := make(map[string]bool)
	for _, e := range exemptions {
		uuidSet[e.ClientUUID] = true
	}
	clientMap := make(map[string]models.Client)
	if len(uuidSet) > 0 {
		uuids := make([]string, 0, len(uuidSet))
		for u := range uuidSet {
			uuids = append(uuids, u)
		}
		var clients []models.Client
		db.Where("client_uuid IN ?", uuids).Find(&clients)
		for _, cl := range clients {
			clientMap[cl.ClientUUID] = cl
		}
	}

	items := make([]gin.H, 0, len(exemptions))
	for _, e := range exemptions {
		item := gin.H{
			"id":          e.ID,
			"field_name":  e.FieldName,
			"client_uuid": e.ClientUUID,
			"device_name": "",
			"ip_address":  "",
		}
		if cl, ok := clientMap[e.ClientUUID]; ok {
			item["device_name"] = cl.DeviceName
			item["ip_address"] = cl.IPAddress
		}
		items = append(items, item)
	}

	c.JSON(200, items)
}

// updateExemptions 通用实现：全量替换指定标准的例外客户端
func (sc *StandardController) updateExemptions(c *gin.Context, fieldType string) {
	id := c.Param("id")

	var body struct {
		ClientUUIDs []string `json:"client_uuids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	db := database.DB

	// 根据类型查出标准记录的 field_name
	var fieldName string
	if fieldType == "linux" {
		var std models.LinuxStandard
		if err := db.First(&std, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "标准配置不存在"})
			return
		}
		fieldName = std.FieldName
	} else {
		var std models.WindowsStandard
		if err := db.First(&std, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "标准配置不存在"})
			return
		}
		fieldName = std.FieldName
	}

	// 事务内先删后建，实现全量替换
	tx := db.Begin()
	if err := tx.Where("field_type = ? AND field_name = ?", fieldType, fieldName).Delete(&models.StandardExemption{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	for _, uuid := range body.ClientUUIDs {
		if uuid == "" {
			continue
		}
		exemption := models.StandardExemption{
			FieldType:  fieldType,
			FieldName:  fieldName,
			ClientUUID: uuid,
		}
		if err := tx.Create(&exemption).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	c.JSON(200, gin.H{
		"message": "保存成功",
	})
}

// ListLinuxExemptions 获取 Linux 标准字段例外列表
func (sc *StandardController) ListLinuxExemptions(c *gin.Context) {
	sc.listExemptions(c, "linux")
}

// UpdateLinuxExemptions 更新 Linux 标准字段例外客户端
func (sc *StandardController) UpdateLinuxExemptions(c *gin.Context) {
	sc.updateExemptions(c, "linux")
}

// ListWindowsExemptions 获取 Windows 标准字段例外列表
func (sc *StandardController) ListWindowsExemptions(c *gin.Context) {
	sc.listExemptions(c, "windows")
}

// UpdateWindowsExemptions 更新 Windows 标准字段例外客户端
func (sc *StandardController) UpdateWindowsExemptions(c *gin.Context) {
	sc.updateExemptions(c, "windows")
}

// ======================== Windows 标准配置方法 ========================

// CreateWindowsStandards 批量创建 Windows 标准配置
func (sc *StandardController) CreateWindowsStandards(c *gin.Context) {
	var standards []models.WindowsStandard
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
		db.Model(&models.WindowsStandard{}).Where("field_name = ? AND deleted_at IS NULL", standards[i].FieldName).Count(&count)
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

// ListWindowsStandards 获取所有 Windows 标准配置列表
func (sc *StandardController) ListWindowsStandards(c *gin.Context) {
	var standards []models.WindowsStandard

	db := database.DB
	query := db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL")

	// 关键词搜索（字段名或字段标签）
	keyword := c.Query("keyword")
	if keyword != "" {
		query = query.Where("field_name LIKE ? OR field_label LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型分组过滤
	groupBy := c.Query("group_by")
	if groupBy != "" {
		var fieldNames []string
		for fieldName, fieldGroup := range WindowsFieldGroups {
			if fieldGroup == groupBy {
				fieldNames = append(fieldNames, fieldName)
			}
		}
		if len(fieldNames) > 0 {
			query = query.Where("field_name IN (?)", fieldNames)
		} else {
			query = query.Where("1 = 0")
		}
	}

	query.Order("group_name, field_name").Find(&standards)

	// 确保每条记录都有正确的 group_name
	for i := range standards {
		if fg, ok := WindowsFieldGroups[standards[i].FieldName]; ok {
			standards[i].GroupName = fg
		}
	}

	c.JSON(200, standards)
}

// UpdateWindowsStandard 更新单个 Windows 标准配置
func (sc *StandardController) UpdateWindowsStandard(c *gin.Context) {
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
	db.Model(&models.WindowsStandard{}).Where("field_name = ? AND id != ? AND deleted_at IS NULL", data["field_name"], id).Count(&count)
	if count > 0 {
		c.JSON(409, gin.H{
			"error": "字段已被其他记录使用",
		})
		return
	}

	if err := db.Model(&models.WindowsStandard{}).Where("id = ?", id).Updates(data).Error; err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "更新成功",
	})
}

// DeleteWindowsStandard 删除 Windows 标准配置
func (sc *StandardController) DeleteWindowsStandard(c *gin.Context) {
	id := c.Param("id")

	db := database.DB

	// 先查出 field_name，用于同步清理该字段的例外配置
	var std models.WindowsStandard
	if err := db.First(&std, id).Error; err == nil {
		db.Where("field_type = ? AND field_name = ?", "windows", std.FieldName).Delete(&models.StandardExemption{})
	}

	if err := db.Delete(&models.WindowsStandard{}, id).Error; err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "删除成功",
	})
}

// GetAvailableWindowsFields 获取所有可用的 Windows 加固字段列表（未配置的）
func (sc *StandardController) GetAvailableWindowsFields(c *gin.Context) {
	// 查询已存在的记录
	db := database.DB
	var existingRecords []models.WindowsStandard
	db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&existingRecords)

	// 提取已使用的 field_name
	existingFieldNames := make(map[string]bool)
	for _, record := range existingRecords {
		existingFieldNames[record.FieldName] = true
	}

	// 从数据库查询所有可用字段
	var allFields []models.WindowsField
	db.Model(&models.WindowsField{}).Where("deleted_at IS NULL").Order("sort_order, field_name").Find(&allFields)

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
