package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/gorm"
)

// RegionController 区域控制器
type RegionController struct {
	db *gorm.DB
}

// NewRegionController create new RegionController instance
func NewRegionController(db *gorm.DB) *RegionController {
	return &RegionController{db: db}
}

// CreateRegionRequest 创建区域请求
type CreateRegionRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateRegionClientsRequest 更新区域客户端关联请求
type UpdateRegionClientsRequest struct {
	ClientIDs []uint `json:"client_ids" binding:"required"`
}

// CreateRegion 创建区域
func (rc *RegionController) CreateRegion(c *gin.Context) {
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查区域名是否已存在
	var count int64
	if err := rc.db.Model(&models.Region{}).Where("name = ? AND deleted_at IS NULL", req.Name).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check region name"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "区域名称「" + req.Name + "」已被使用"})
		return
	}

	// 创建区域
	region := models.Region{Name: req.Name}
	if err := rc.db.Create(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"id":      region.ID,
	})
}

// ListRegions 获取所有区域及其关联的客户端（自动过滤软删除记录）
func (rc *RegionController) ListRegions(c *gin.Context) {
	var regions []models.Region

	if err := rc.db.Preload("Clients").Order("id ASC").Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 转换为 API 响应格式
	result := make([]gin.H, 0, len(regions))
	for _, region := range regions {
		clientItems := make([]models.RegionClientItem, 0, len(region.Clients))
		for _, client := range region.Clients {
			clientItems = append(clientItems, models.RegionClientItem{
				ID:         client.ID,
				DeviceName: client.DeviceName,
				IPAddress:  client.IPAddress,
			})
		}

		result = append(result, gin.H{
			"id":         region.ID,
			"name":       region.Name,
			"created_at": region.CreatedAt,
			"clients":    clientItems,
		})
	}

	c.JSON(http.StatusOK, result)
}

// UpdateRegionClients 更新区域关联的客户端列表（全量替换）
func (rc *RegionController) UpdateRegionClients(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	var req UpdateRegionClientsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查区域是否存在
	var region models.Region
	if err := rc.db.First(&region, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "区域不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query region"})
		}
		return
	}

	clientIDs := req.ClientIDs
	if len(clientIDs) == 0 {
		// 清空全部关联
		if err := rc.db.Model(&region).Association("Clients").Clear(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing associations"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "关联成功"})
		return
	}

	// 校验客户端 ID 是否全部存在
	var foundCount int64
	if err := rc.db.Model(&models.Client{}).Where("id IN ?", clientIDs).Count(&foundCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query clients"})
		return
	}
	if foundCount != int64(len(clientIDs)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分客户端不存在，请刷新后重试"})
		return
	}

	// 根据客户端 ID 查询客户端记录
	var clients []models.Client
	if err := rc.db.Where("id IN ?", clientIDs).Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query clients"})
		return
	}

	// 事务内全量替换关联关系（先删除旧关联，再插入新关联）
	err = rc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&region).Association("Clients").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&region).Association("Clients").Append(clients); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update associations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "关联成功"})
}

// DeleteRegion 删除区域
func (rc *RegionController) DeleteRegion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region ID"})
		return
	}

	// 检查区域是否存在
	var region models.Region
	if err := rc.db.First(&region, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "区域不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query region"})
		}
		return
	}

	// 删除区域（触发软删除），同时清理 region_clients 关联关系
	err = rc.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&region).Association("Clients").Clear(); err != nil {
			return err
		}
		return tx.Delete(&region).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
