package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/gorm"
)

type ClientController struct {
	db *gorm.DB
}

type clientRequest struct {
	db *gorm.DB
}

// RequestTempTokenResponse 临时 Token 响应
type RequestTempTokenResponse struct {
	TempToken   string    `json:"temp_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	DeviceName  string    `json:"device_name"`
	IPAddress   string    `json:"ip_address"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	TempToken  string `json:"temp_token" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
	IPAddress  string `json:"ip_address" binding:"required"`
	OSVersion  string `json:"os_version"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ClientUUID     string    `json:"client_uuid"`
	ShortToken     string    `json:"short_token"`
	RefreshToken   string    `json:"refresh_token"`
	ExpiresAt      time.Time `json:"expires_at"`
	DeviceName     string    `json:"device_name"`
	IPAddress      string    `json:"ip_address"`
}

// RefreshTokenRequest Token 刷新请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ClientUUID   string `json:"client_uuid"`
}

// RefreshTokenResponse Token 刷新响应
type RefreshTokenResponse struct {
	ShortToken string    `json:"short_token"`
	ExpiresAt  string    `json:"expires_at"` // Changed from time.Time to string
}

// UploadDataRequest 上传数据请求
type UploadDataRequest struct {
	ClientUUID string               `json:"client_uuid"`
	Data       models.SystemCheck   `json:"data"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Status      string `json:"status"`
	ClientUUID  string `json:"client_uuid"`
}

// ClientItem 客户端列表项类型
type ClientItem struct {
	ID            uint       `json:"id"`
	DeviceName    string     `json:"device_name"`
	IPAddress     string     `json:"ip_address"`
	OSVersion     string     `json:"os_version"`
	Status        string     `json:"status"` // online/offline
	LastCheckTime *time.Time `json:"last_check_time"`
	CreatedAt     time.Time  `json:"created_at"`
}

// NewClientController 创建客户端控制器
func NewClientController(db *gorm.DB) *ClientController {
	return &ClientController{db: db}
}

// RequestTempToken 请求临时安装 Token
func (cc *ClientController) RequestTempToken(c *gin.Context) {
	var req struct {
		DeviceName string `json:"device_name" binding:"required"`
		IPAddress  string `json:"ip_address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		c.Abort()
		return
	}

	// 生成临时 Token（5 分钟有效期）
	tempToken := generateTempToken(req.DeviceName, req.IPAddress)
	expiresAt := time.Now().Add(5 * time.Minute)

	// TODO: 将临时 Token 存储到缓存（Redis/Memory），设置 5 分钟过期
	// 这里使用内存存储，生产环境建议使用 Redis
	tempTokenStore[tempToken] = TempTokenInfo{
		DeviceName:  req.DeviceName,
		IPAddress:   req.IPAddress,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Used:        false,
	}

	c.JSON(http.StatusOK, gin.H{
		"temp_token":  tempToken,
		"expires_in":  300, // 秒
		"expires_at":  expiresAt.Format(time.RFC3339),
		"device_name": req.DeviceName,
	})
}

// Register 注册客户端
func (cc *ClientController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		c.Abort()
		return
	}

	// 验证临时 Token
	tempTokenInfo, exists := tempTokenStore[req.TempToken]
	if !exists || tempTokenInfo.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "临时 Token 已过期或无效"})
		c.Abort()
		return
	}

	// 检查是否已经使用过
	if tempTokenInfo.Used {
		c.JSON(http.StatusForbidden, gin.H{"error": "临时 Token 已被使用"})
		c.Abort()
		return
	}

	// 验证设备信息匹配
	if tempTokenInfo.DeviceName != req.DeviceName || tempTokenInfo.IPAddress != req.IPAddress {
		c.JSON(http.StatusForbidden, gin.H{"error": "设备信息与临时 Token 不匹配"})
		c.Abort()
		return
	}

	// 标记临时 Token 为已使用
	tempTokenInfo.Used = true
	tempTokenStore[req.TempToken] = tempTokenInfo

	// 检查是否已注册
	var client models.Client
	result := cc.db.Where("device_name = ? AND ip_address = ?", req.DeviceName, req.IPAddress).First(&client)

	var refreshToken string
	if result.Error == nil && client.ID > 0 {
		// 已注册，只刷新 Token
		refreshToken = generateRefreshToken()

		// 查找并更新现有 Token
		var token models.ClientToken
		tokenResult := cc.db.Where("client_uuid = ?", client.ClientUUID).First(&token)
		if tokenResult.Error == nil {
			token.RefreshToken = refreshToken
			token.ShortToken = "" // 强制重新生成
			token.ExpiresAt = time.Now().Add(14 * 24 * time.Hour)
			cc.db.Save(&token)
			
			// 使用新生成的 token
			shortToken := generateShortToken()
			token.ShortToken = shortToken
			cc.db.Save(&token)
			
			// 直接返回，不再创建新 token
			c.JSON(http.StatusOK, RegisterResponse{
				ClientUUID:   client.ClientUUID,
				ShortToken:   shortToken,
				RefreshToken: refreshToken,
				ExpiresAt:    time.Now().Add(14 * 24 * time.Hour),
				DeviceName:   client.DeviceName,
				IPAddress:    client.IPAddress,
			})
			return
		}
	} else {
		// 新客户端，创建记录
		client = models.Client{
			ClientUUID:  generateUUID(),
			DeviceName:  req.DeviceName,
			IPAddress:   req.IPAddress,
			OSVersion:   req.OSVersion,
			Status:      "active",
		}

		if err := cc.db.Create(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client record: %v", err)})
			c.Abort()
			return
		}

		// 生成新的 Refresh Token (90 天)
		refreshToken = generateRefreshToken()
	}

	// 生成 Short Token (14 天)
	shortToken := generateShortToken()
	expiresAt := time.Now().Add(14 * 24 * time.Hour)

	// 保存 Token
	token := models.ClientToken{
		ClientUUID:   client.ClientUUID,
		RefreshToken: refreshToken,
		ShortToken:   shortToken,
		ExpiresAt:    expiresAt,
	}

	if err := cc.db.Create(&token).Error; err != nil {
		log.Printf("❌ Error creating token: %v, ClientUUID: %s", err, client.ClientUUID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save token: %v", err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		ClientUUID:   client.ClientUUID,
		ShortToken:   shortToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		DeviceName:   client.DeviceName,
		IPAddress:    client.IPAddress,
	})
}

// RefreshToken 刷新 Token
func (cc *ClientController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		c.Abort()
		return
	}

	// 查询有效 Token
	var token models.ClientToken
	result := cc.db.Where("refresh_token = ? AND expires_at > ?", 
		req.RefreshToken, time.Now()).First(&token)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		c.Abort()
		return
	}

	// 生成新的短期 Token
	newShortToken := generateShortToken()
	newExpiresAt := time.Now().Add(14 * 24 * time.Hour)

	// 更新 Token
	token.ShortToken = newShortToken
	token.ExpiresAt = newExpiresAt
	if err := cc.db.Save(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		ShortToken: newShortToken,
		ExpiresAt:  newExpiresAt.Format(time.RFC3339), // Format as string
	})
}

// UploadData 接收加固检查结果
func (cc *ClientController) UploadData(c *gin.Context) {
	// 验证短期 Token
	tokenStr := c.GetHeader("X-Client-Token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token in header"})
		c.Abort()
		return
	}

	var token models.ClientToken
	result := cc.db.Where("short_token = ? AND expires_at > ?", 
		tokenStr, time.Now()).First(&token)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		c.Abort()
		return
	}

	// 解析请求体
	var req struct {
		Data models.SystemCheck `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// 关联 ClientUUID
	req.Data.ClientUUID = token.ClientUUID

	// 检查该客户端是否已有记录
	var existingRecord models.SystemCheck
	result = cc.db.Where("client_uuid = ?", token.ClientUUID).Order("id DESC").First(&existingRecord)

	if result.Error == nil {
		// 记录存在，执行 UPDATE 操作
		req.Data.ID = existingRecord.ID // 保留原 ID
		if err := cc.db.Model(&models.SystemCheck{}).Where("id = ?", existingRecord.ID).Updates(req.Data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Updated existing systemcheck record for client: %s (ID=%d)", token.ClientUUID, existingRecord.ID)
	} else {
		// 记录不存在，执行 CREATE 操作
		if err := cc.db.Create(&req.Data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Created new systemcheck record for client: %s", token.ClientUUID)
	}

	// 更新时间戳
	now := time.Now()
	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(map[string]interface{}{
			"last_check_time":  now,
			"last_upload_time": &now,
			"status":           "active",
		}).Error; err != nil {
		log.Printf("Warning: Failed to update client last activity: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"record_id": req.Data.ID,
		"message":   "Data uploaded successfully",
	})
}

// UploadWindowsData 接收 Windows 加固检查结果
func (cc *ClientController) UploadWindowsData(c *gin.Context) {
	// 验证短期 Token
	tokenStr := c.GetHeader("X-Client-Token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token in header"})
		c.Abort()
		return
	}

	var token models.ClientToken
	result := cc.db.Where("short_token = ? AND expires_at > ?",
		tokenStr, time.Now()).First(&token)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		c.Abort()
		return
	}

	// 解析请求体
	var req struct {
		Data models.WindowsSystemCheck `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// 关联 ClientUUID
	req.Data.ClientUUID = token.ClientUUID

	// 检查该客户端是否已有记录
	var existingRecord models.WindowsSystemCheck
	result = cc.db.Where("client_uuid = ?", token.ClientUUID).Order("id DESC").First(&existingRecord)

	if result.Error == nil {
		// 记录存在，执行 UPDATE 操作
		req.Data.ID = existingRecord.ID // 保留原 ID
		if err := cc.db.Model(&models.WindowsSystemCheck{}).Where("id = ?", existingRecord.ID).Updates(req.Data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update windows check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Updated existing windows check record for client: %s (ID=%d)", token.ClientUUID, existingRecord.ID)
	} else {
		// 记录不存在，执行 CREATE 操作
		if err := cc.db.Create(&req.Data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save windows check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Created new windows check record for client: %s", token.ClientUUID)
	}

	// 更新时间戳
	now := time.Now()
	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(map[string]interface{}{
			"last_check_time":  now,
			"last_upload_time": &now,
			"status":           "active",
		}).Error; err != nil {
		log.Printf("Warning: Failed to update client last activity: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"record_id": req.Data.ID,
		"message":   "Windows data uploaded successfully",
	})
}

// Heartbeat 接收客户端心跳
func (cc *ClientController) Heartbeat(c *gin.Context) {
	// 验证短期 Token
	tokenStr := c.GetHeader("X-Client-Token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token in header"})
		c.Abort()
		return
	}

	var token models.ClientToken
	result := cc.db.Where("short_token = ? AND expires_at > ?", 
		tokenStr, time.Now()).First(&token)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		c.Abort()
		return
	}

	// 更新客户端心跳时间
	now := time.Now()
	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(map[string]interface{}{
			"last_check_time": now,
			"status":          "active",
		}).Error; err != nil {
		log.Printf("Warning: Failed to update client heartbeat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HeartbeatResponse{
		Status:     "ok",
		ClientUUID: token.ClientUUID,
	})
}

// ListClients 获取所有客户端列表（需认证）
func (cc *ClientController) ListClients(c *gin.Context) {
	var clients []models.Client
	
	if err := cc.db.Order("created_at DESC").Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	now := time.Now()
	result := make([]ClientItem, 0, len(clients))

	for _, client := range clients {
		// 在线状态判定：5 分钟内有心跳为 online
		status := "offline"
		if client.LastCheckTime != nil && now.Sub(*client.LastCheckTime) <= 5*time.Minute {
			status = "online"
		}

		result = append(result, ClientItem{
			ID:            client.ID,
			DeviceName:    client.DeviceName,
			IPAddress:     client.IPAddress,
			OSVersion:     client.OSVersion,
			Status:        status,
			LastCheckTime: client.LastCheckTime,
			CreatedAt:     client.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  result,
		"total": len(result),
	})
}

// DeleteClient 删除客户端及其关联数据（硬删除）（需认证）
func (cc *ClientController) DeleteClient(c *gin.Context) {
	id, parseErr := strconv.ParseUint(c.Param("id"), 10, 32)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		c.Abort()
		return
	}

	var client models.Client
	if err := cc.db.First(&client, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		c.Abort()
		return
	}

	// 启动事务
	tx := cc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 1. 删除 systemcheck 记录
	if err := tx.Where("client_uuid = ?", client.ClientUUID).Delete(&models.SystemCheck{}).Error; err != nil {
		log.Printf("❌ Error deleting systemcheck records: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete systemcheck records"})
		c.Abort()
		return
	}

	// 2. 删除 client_tokens 记录
	if err := tx.Where("client_uuid = ?", client.ClientUUID).Delete(&models.ClientToken{}).Error; err != nil {
		log.Printf("❌ Error deleting client tokens: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client tokens"})
		c.Abort()
		return
	}

	// 3. 删除 client 记录（硬删除）
	if err := tx.Delete(&models.Client{}, client.ID).Error; err != nil {
		log.Printf("❌ Error deleting client: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		c.Abort()
		return
	}

	log.Printf("✅ Client deleted successfully: id=%d, uuid=%s", client.ID, client.ClientUUID)
	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}
