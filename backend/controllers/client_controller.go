package controllers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/configs"
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
	TempToken  string    `json:"temp_token"`
	ExpiresAt  time.Time `json:"expires_at"`
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	TempToken     string `json:"temp_token" binding:"required"`
	DeviceName    string `json:"device_name" binding:"required"`
	IPAddress     string `json:"ip_address" binding:"required"`
	OSVersion     string `json:"os_version"`
	ClientVersion string `json:"client_version"` // 新增：客户端版本
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ClientUUID   string    `json:"client_uuid"`
	ShortToken   string    `json:"short_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	DeviceName   string    `json:"device_name"`
	IPAddress    string    `json:"ip_address"`
}

// RefreshTokenRequest Token 刷新请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ClientUUID   string `json:"client_uuid"`
}

// RefreshTokenResponse Token 刷新响应
type RefreshTokenResponse struct {
	ShortToken string `json:"short_token"`
	ExpiresAt  string `json:"expires_at"` // Changed from time.Time to string
}

// UploadDataRequest 上传数据请求
type UploadDataRequest struct {
	ClientUUID string             `json:"client_uuid"`
	Data       models.SystemCheck `json:"data"`
}

// HeartbeatResponse 心跳响应 (含版本检测信息)
type HeartbeatResponse struct {
	Status        string `json:"status"`
	ClientUUID    string `json:"client_uuid"`
	HasUpdate     bool   `json:"has_update,omitempty"`         // 是否有新版本
	NewVersion    string `json:"new_version,omitempty"`        // 新版本号
	DownloadURL   string `json:"download_url,omitempty"`       // 下载路径 URL
}

// ClientItem 客户端列表项类型
type ClientItem struct {
	ID            uint       `json:"id"`
	DeviceName    string     `json:"device_name"`
	IPAddress     string     `json:"ip_address"`
	OSVersion     string     `json:"os_version"`
	ClientVersion string     `json:"client_version"` // 客户端版本
	Status        string     `json:"status"` // online/offline
	LastCheckTime *time.Time `json:"last_check_time"`
	CreatedAt     time.Time  `json:"created_at"`
}

var (
	// 全局配置引用（将在初始化时调用）
	globalConfig *configs.Config
)

// NewClientController 创建客户端控制器
func NewClientController(db *gorm.DB) *ClientController {
	return &ClientController{db: db}
}

// InitPackageCache 启动时从数据库加载缓存
func InitPackageCache(db *gorm.DB) {
	var packages []models.PackageMeta
	if err := db.Find(&packages).Error; err != nil {
		log.Printf(" 加载包元数据缓存失败：%v", err)
		return
	}

	log.Printf("✅ 成功加载 %d 个安装包元数据", len(packages))
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
		DeviceName: req.DeviceName,
		IPAddress:  req.IPAddress,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
		Used:       false,
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
		// 检查是否存在同 device_name + ip_address 的软删除记录（删除后重装场景）
		var deletedClient models.Client
		deletedResult := cc.db.Unscoped().Where("device_name = ? AND ip_address = ? AND deleted_at IS NOT NULL", req.DeviceName, req.IPAddress).First(&deletedClient)

		if deletedResult.Error == nil && deletedClient.ID > 0 {
			// 复活软删除的客户端：分配新 UUID，清除删除标记和时间戳
			newUUID := generateUUID()
			if err := cc.db.Unscoped().Model(&deletedClient).Updates(map[string]interface{}{
				"client_uuid":      newUUID,
				"deleted_at":       nil,
				"status":           "active",
				"os_version":       req.OSVersion,
				"client_version":   req.ClientVersion, // 新增：更新客户端版本
				"last_check_time":  nil,
				"last_upload_time": nil,
			}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to resurrect client: %v", err)})
				c.Abort()
				return
			}
			log.Printf("✅ Resurrected soft-deleted client: id=%d, old_uuid=%s, new_uuid=%s", deletedClient.ID, deletedClient.ClientUUID, newUUID)
			deletedClient.ClientUUID = newUUID
			client = deletedClient
			refreshToken = generateRefreshToken()
		} else {
			// 全新客户端，创建记录
			client = models.Client{
				ClientUUID:    generateUUID(),
				DeviceName:    req.DeviceName,
				IPAddress:     req.IPAddress,
				OSVersion:     req.OSVersion,
				ClientVersion: req.ClientVersion, // 新增：保存客户端版本
				Status:        "active",
			}

			if err := cc.db.Create(&client).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client record: %v", err)})
				c.Abort()
				return
			}

			// 生成新的 Refresh Token (90 天)
			refreshToken = generateRefreshToken()
		}
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
	var reqBody struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	
	// 尝试提取 client_version
	var tempData map[string]interface{}
	var clientVersion string
	if json.Unmarshal(reqBody.Data, &tempData) == nil {
		if cv, ok := tempData["client_version"].(string); ok && cv != "" {
			clientVersion = cv
			log.Printf("✅ Extracted client_version: %s", cv)
		}
	}

	// 关联 ClientUUID
	var reqData models.SystemCheck
	if err := json.Unmarshal(reqBody.Data, &reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		c.Abort()
		return
	}
	reqData.ClientUUID = token.ClientUUID

	// 检查该客户端是否已有记录
	var existingRecord models.SystemCheck
	result = cc.db.Where("client_uuid = ?", token.ClientUUID).Order("id DESC").First(&existingRecord)

	if result.Error == nil {
		// 记录存在，执行 UPDATE 操作
		reqData.ID = existingRecord.ID // 保留原 ID
		if err := cc.db.Model(&models.SystemCheck{}).Where("id = ?", existingRecord.ID).Updates(reqData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Updated existing systemcheck record for client: %s (ID=%d)", token.ClientUUID, existingRecord.ID)
	} else {
		// 记录不存在，执行 CREATE 操作
		if err := cc.db.Create(&reqData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save check data"})
			c.Abort()
			return
		}
		log.Printf("✅ Created new systemcheck record for client: %s", token.ClientUUID)
	}

	// 更新时间戳和客户端版本
	now := time.Now()
	updates := map[string]interface{}{
		"last_check_time":  now,
		"last_upload_time": &now,
		"status":           "active",
	}
	
	// 如果数据中包含客户端版本号，也进行更新
	if clientVersion != "" {
		updates["client_version"] = clientVersion
		log.Printf("✅ Updated client version to: %s", clientVersion)
	}
	
	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(updates).Error; err != nil {
		log.Printf("Warning: Failed to update client last activity: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"record_id": reqData.ID,
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
	var reqBody struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	
	// 先尝试解析出 client_version
	var tempData map[string]interface{}
	if err := json.Unmarshal(reqBody.Data, &tempData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		c.Abort()
		return
	}
	
	var clientVersion string
	if cv, ok := tempData["client_version"].(string); ok && cv != "" {
		clientVersion = cv
		log.Printf("✅ Extracted client_version from Windows data: %s", cv)
	}

	// 关联 ClientUUID
	var reqData models.WindowsSystemCheck
	if err := json.Unmarshal(reqBody.Data, &reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		c.Abort()
		return
	}
	reqData.ClientUUID = token.ClientUUID
	
	req := struct {
		Data models.WindowsSystemCheck
	}{Data: reqData}

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

	// 更新时间戳和客户端版本
	now := time.Now()
	updates := map[string]interface{}{
		"last_check_time":  now,
		"last_upload_time": &now,
		"status":           "active",
	}
	
	// 如果数据中包含客户端版本号，也进行更新
	if clientVersion != "" {
		updates["client_version"] = clientVersion
		log.Printf("✅ Updated client version to: %s", clientVersion)
	}
	
	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(updates).Error; err != nil {
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
	updates := map[string]interface{}{
		"last_check_time": now,
		"status":          "active",
	}

	// 【关键】心跳可选携带当前运行版本，更新重启后及时同步 client_version，避免重复更新死循环
	var heartbeatBody struct {
		ClientVersion string `json:"client_version"`
	}
	if err := c.ShouldBindJSON(&heartbeatBody); err == nil && heartbeatBody.ClientVersion != "" {
		updates["client_version"] = heartbeatBody.ClientVersion
		log.Printf("✅ Heartbeat updated client version to: %s", heartbeatBody.ClientVersion)
	}

	if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
		Updates(updates).Error; err != nil {
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

// CheckUpdate 检查是否有新版本 (独立接口)
func (cc *ClientController) CheckUpdate(c *gin.Context) {
	// 验证短期 Token
	tokenStr := c.GetHeader("X-Client-Token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token in header"})
		c.Abort()
		return
	}

	var token models.ClientToken
	result := cc.db.Where("short_token = ? AND expires_at > ?", tokenStr, time.Now()).First(&token)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		c.Abort()
		return
	}

	// 获取当前客户端信息
	var client models.Client
	if err := cc.db.Where("client_uuid = ?", token.ClientUUID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		c.Abort()
		return
	}

	// 【关键】优先使用客户端上报的本地实际版本（更新重启后 DB 记录可能滞后），并同步回数据库
	if reportedVersion := c.GetHeader("X-Client-Version"); reportedVersion != "" && reportedVersion != client.ClientVersion {
		log.Printf("✅ Syncing client version from check-update: %s -> %s", client.ClientVersion, reportedVersion)
		if err := cc.db.Model(&models.Client{}).Where("client_uuid = ?", token.ClientUUID).
			Update("client_version", reportedVersion).Error; err != nil {
			log.Printf("Warning: Failed to sync client version: %v", err)
		}
		client.ClientVersion = reportedVersion
	}

	// 根据客户端类型获取最新安装包版本
	var latestPackage models.PackageMeta
	pkgType := "linux"
	if strings.Contains(client.OSVersion, "Windows") {
		pkgType = "windows"
	}

	if err := cc.db.Where("type = ?", pkgType).Order("created_at DESC").First(&latestPackage).Error; err != nil {
		log.Printf("⚠️ No package found for type: %s", pkgType)
		c.JSON(http.StatusOK, gin.H{
			"has_update":    false,
			"current_version": client.ClientVersion,
			"new_version":   "",
			"message":       "No package available",
		})
		return
	}

	// 对比版本号 (简单字符串比较，实际可使用 semver 库)
	hasUpdate := false
	if latestPackage.Version != "" && client.ClientVersion != latestPackage.Version {
		hasUpdate = true
	}

	// 构建下载 URL
	downloadURL := fmt.Sprintf("%s/api/packages/%s/download", globalConfig.Packages.ServerURL, pkgType)

	c.JSON(http.StatusOK, gin.H{
		"has_update":    hasUpdate,
		"current_version": client.ClientVersion,
		"new_version":   latestPackage.Version,
		"download_url":  downloadURL,
		"hash":          latestPackage.Hash,
		"size":          latestPackage.Size,
		"filename":      latestPackage.Filename,
	})
}

// ListClients 获取所有客户端列表（需认证），支持搜索、筛选和分页
func (cc *ClientController) ListClients(c *gin.Context) {
	// 解析查询参数
	search := c.Query("search")
	status := c.Query("status")     // online / offline
	osType := c.Query("os_type")    // windows / linux
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建数据库查询（搜索和系统类型过滤在 DB 层完成）
	query := cc.db.Model(&models.Client{})
	if search != "" {
		keyword := "%" + search + "%"
		query = query.Where("device_name LIKE ? OR ip_address LIKE ?", keyword, keyword)
	}
	if osType != "" {
		query = query.Where("LOWER(os_version) LIKE ?", "%"+strings.ToLower(osType)+"%")
	}

	var clients []models.Client
	if err := query.Order("created_at DESC").Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	now := time.Now()
	allItems := make([]ClientItem, 0, len(clients))

	for _, client := range clients {
		// 在线状态判定：5 分钟内有心跳为 online
		itemStatus := "offline"
		if client.LastCheckTime != nil && now.Sub(*client.LastCheckTime) <= 5*time.Minute {
			itemStatus = "online"
		}

		// 状态过滤（在线状态是计算值，在内存中过滤）
		if status != "" && itemStatus != status {
			continue
		}

		allItems = append(allItems, ClientItem{
			ID:            client.ID,
			DeviceName:    client.DeviceName,
			IPAddress:     client.IPAddress,
			OSVersion:     client.OSVersion,
			ClientVersion: client.ClientVersion, // 添加客户端版本
			Status:        itemStatus,
			LastCheckTime: client.LastCheckTime,
			CreatedAt:     client.CreatedAt,
		})
	}

	// 计算总数并分页
	total := len(allItems)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  allItems[start:end],
		"total": total,
	})
}

// DeleteClient 删除客户端及其关联数据（软删除客户端，硬删除关联数据）（需认证）
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

	// 1. 删除 Linux systemcheck 记录
	if err := tx.Where("client_uuid = ?", client.ClientUUID).Delete(&models.SystemCheck{}).Error; err != nil {
		log.Printf("❌ Error deleting systemcheck records: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete systemcheck records"})
		c.Abort()
		return
	}

	// 2. 删除 Windows systemcheck 记录
	if err := tx.Where("client_uuid = ?", client.ClientUUID).Delete(&models.WindowsSystemCheck{}).Error; err != nil {
		log.Printf("❌ Error deleting windows systemcheck records: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete windows systemcheck records"})
		c.Abort()
		return
	}

	// 3. 删除 client_tokens 记录
	if err := tx.Where("client_uuid = ?", client.ClientUUID).Delete(&models.ClientToken{}).Error; err != nil {
		log.Printf("❌ Error deleting client tokens: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client tokens"})
		c.Abort()
		return
	}

	// 4. 清理 region_clients 多对多关联
	if err := tx.Exec("DELETE FROM region_clients WHERE client_id = ?", client.ID).Error; err != nil {
		log.Printf("❌ Error clearing region associations: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear region associations"})
		c.Abort()
		return
	}

	// 5. 软删除 client 记录（保留记录以便重装时复活）
	if err := tx.Delete(&client).Error; err != nil {
		log.Printf("❌ Error deleting client: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		c.Abort()
		return
	}

	log.Printf("✅ Client deleted successfully: id=%d, uuid=%s", client.ID, client.ClientUUID)
	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}

// SetGlobalConfig 设置全局配置（供初始化时调用）
func SetGlobalConfig(cfg *configs.Config) {
	globalConfig = cfg
}
func (cc *ClientController) getPackagePath(pkgType string) string {
	var dir string
	if globalConfig != nil {
		if pkgType == "linux" {
			dir = globalConfig.Packages.LinuxPackageDir
		} else if pkgType == "windows" {
			dir = globalConfig.Packages.WindowsPackageDir
		}
	}
	if dir == "" {
		return ""
	}
	// 根据类型确定文件扩展名
	ext := ".zip"
	if pkgType == "windows" {
		ext = ".exe"
	}
	// 查找目录下的第一个匹配扩展名的文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ext) {
			return dir + "/" + file.Name()
		}
	}
	return ""
}

// CalculateFileHash 计算文件内容的 MD5 哈希值
func (cc *ClientController) CalculateFileHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// extractVersionFromFilename 从文件名中提取版本号
// 支持格式:
//   - linux-hardening-client_v1.1.0.zip
//   - WindowsHardeningClient_Setup_1.1.0.exe
func extractVersionFromFilename(filename string) string {
	// 提取不带扩展名的文件名
	ext := strings.ToLower(filepath.Ext(filename))
	baseName := strings.TrimSuffix(filename, ext)
	
	if ext == ".zip" {
		// Linux: linux-hardening-client_v1.1.0.zip
		// 查找 _v 后面的版本号
		if idx := strings.Index(baseName, "_v"); idx != -1 {
			versionStr := baseName[idx+2:] // 跳过 "_v"
			// 验证版本号格式 (语义化版本)
			if isValidVersion(versionStr) {
				return versionStr
			}
		}
	} else if ext == ".exe" {
		// Windows: WindowsHardeningClient_Setup_1.1.0.exe
		// 查找 Setup_ 或 setup_ 后面的版本号
		if idx := strings.Index(baseName, "Setup_"); idx != -1 {
			versionStr := baseName[idx+6:] // 跳过 "Setup_"
			if isValidVersion(versionStr) {
				return versionStr
			}
		}
		// 或者查找 _setup_ (小写)
		if idx := strings.Index(baseName, "_setup_"); idx != -1 {
			versionStr := baseName[idx+7:] // 跳过 "_setup_"
			if isValidVersion(versionStr) {
				return versionStr
			}
		}
	}
	
	return ""
}

// isValidVersion 验证版本号格式是否为语义化版本 (X.Y.Z)
func isValidVersion(version string) bool {
	// 版本号应该只包含数字和点，至少有两个点 (X.Y.Z)
	if len(version) < 5 {
		return false
	}
	
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return false
	}
	
	for _, part := range parts {
		if part == "" {
			return false
		}
		// 检查是否全为数字
		isDigit := true
		for _, c := range part {
			if c < '0' || c > '9' {
				isDigit = false
				break
			}
		}
		if !isDigit {
			return false
		}
	}
	
	return true
}

// GetPackageInfo 获取包信息（大小和哈希值）
func (cc *ClientController) GetPackageInfo(c *gin.Context) {
	pkgType := c.Param("type")
	if pkgType != "linux" && pkgType != "windows" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的包类型，必须是 linux 或 windows"})
		c.Abort()
		return
	}

	// === 优先从数据库读取包元数据 ===
	result := make(map[string]interface{})
	var pkgMeta models.PackageMeta
	err := cc.db.Where("type = ?", pkgType).First(&pkgMeta).Error

	if err == nil && pkgMeta.Filepath != "" {
		// 数据库有记录且文件存在，直接返回
		if _, statErr := os.Stat(pkgMeta.Filepath); statErr == nil {
			result["exists"] = true
			result["size"] = pkgMeta.Size
			result["hash"] = pkgMeta.Hash
			result["version"] = pkgMeta.Version // 新增：返回版本号
			c.JSON(http.StatusOK, result)
			return
		}
		log.Printf("⚠️ 数据库记录的文件不存在：%s，回退到文件扫描", pkgMeta.Filepath)
	}

	c.JSON(http.StatusOK, result)
}

// UploadPackage 上传安装包
func (cc *ClientController) UploadPackage(c *gin.Context) {
	pkgType := c.PostForm("type")
	if pkgType != "linux" && pkgType != "windows" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的包类型，必须是 linux 或 windows"})
		c.Abort()
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("package")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到上传文件"})
		c.Abort()
		return
	}

	// === 安全验证：文件名检查 ===
	originalFilename := file.Filename

	// 提取文件名和扩展名
	ext := strings.ToLower(filepath.Ext(originalFilename))
	baseName := strings.TrimSuffix(originalFilename, ext)

	// 验证扩展名
	if ext != ".zip" && ext != ".exe" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持.zip 或.exe 格式的安装包"})
		c.Abort()
		return
	}

	// 防止路径遍历：检查文件名是否包含特殊字符
	if baseName == "" ||
		strings.ContainsAny(baseName, `\/:*?"<>|`) ||
		strings.Contains(baseName, "..") ||
		strings.HasPrefix(baseName, ".") ||
		!regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(baseName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名非法，只能包含字母、数字、下划线、点和连字符"})
		c.Abort()
		return
	}

	// 限制文件大小（200MB）
	const maxFileSize int64 = 200 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("文件大小不能超过%d MB", 200)})
		c.Abort()
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"})
		c.Abort()
		return
	}
	defer src.Close()

	// 读取文件内容以计算哈希
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		c.Abort()
		return
	}
	
	// === 先提取版本号 (在文件改名之前) ===
	version := extractVersionFromFilename(originalFilename)
	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法从文件名提取版本号，请使用格式：linux-hardening-client_v1.1.0.zip 或 WindowsHardeningClient_Setup_1.1.0.exe"})
		c.Abort()
		return
	}

	// === 安全验证：文件内容类型检查 ===
	switch ext {
	case ".zip":
		// ZIP 文件魔数：50 4B 03 04
		if len(data) < 4 || !bytes.Equal(data[:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ZIP 文件"})
			c.Abort()
			return
		}
	case ".exe":
		// DOS/EXE 文件魔数：4D 5A ('MZ')
		if len(data) < 2 || !bytes.Equal(data[:2], []byte{'M', 'Z'}) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 EXE 文件"})
			c.Abort()
			return
		}
	}

	// 计算新文件的哈希值
	hash := cc.CalculateFileHash(data)

	// 先从数据库中查询是否已存在该类型的安装包
	var existingPkgMeta models.PackageMeta
	err = cc.db.Where("type = ?", pkgType).First(&existingPkgMeta).Error

	if err == nil && existingPkgMeta.ID > 0 {
		// 数据库中存在该类型的安装包记录
		if strings.TrimSpace(existingPkgMeta.Hash) == hash {
			// 哈希值一致，说明是相同的安装包，拒绝上传
			log.Printf("❌ %s 安装包与数据库记录一致，拒绝重复上传：hash=%s", pkgType, hash)
			c.JSON(http.StatusBadRequest, gin.H{"error": "已经存在该安装包，无需重复上传"})
			c.Abort()
			return
		}

		// 哈希值不一致，需要替换
		log.Printf("⚠️ %s 安装包哈希不一致，准备替换：old_hash=%s, new_hash=%s", 
			pkgType, existingPkgMeta.Hash, hash)

		// 检查数据库记录的文件是否存在
		_, fileStatErr := os.Stat(existingPkgMeta.Filepath)
		if fileStatErr != nil {
			// 文件不存在，记录日志但继续执行
			log.Printf("⚠️ 数据库记录的文件已丢失：%s", existingPkgMeta.Filepath)
		} else {
			// 文件存在，删除旧文件
			if err := os.Remove(existingPkgMeta.Filepath); err != nil {
				log.Printf("❌ 删除旧安装包失败：%v", err)
				// 继续执行，不阻止上传流程
			}
		}
	}

	// 确定保存目录
	var saveDir string
	if globalConfig != nil {
		if pkgType == "linux" {
			saveDir = globalConfig.Packages.LinuxPackageDir
		} else if pkgType == "windows" {
			saveDir = globalConfig.Packages.WindowsPackageDir
		}
	}

	// 如果配置中未设置，使用默认目录
	if saveDir == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "安装包目录配置未设置"})
		c.Abort()
		return
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建目录失败：%v", err)})
		c.Abort()
		return
	}

	// 生成目标文件名（使用哈希值作为文件名前缀）
	filename := fmt.Sprintf("%s_%s%s", pkgType, hash[:8], ext)
	savePath := filepath.Join(saveDir, filename)

	// 保存新文件
	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("文件保存失败：%v", err)})
		c.Abort()
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("文件写入失败：%v", err)})
		c.Abort()
		return
	}

	// === 保存到数据库（存在则更新，不存在则创建）===
	pkgMeta := &models.PackageMeta{
		Hash:     hash,
		Size:     file.Size,
		Filename: file.Filename,
		Filepath: savePath,
		Version:  version, // 使用之前提取的版本号
	}
	
	// 检查是否已存在该类型的记录
	var existingMeta models.PackageMeta
	queryResult := cc.db.Where("type = ?", pkgType).First(&existingMeta)
	
	log.Printf("🔍 [UploadPackage] 查询结果 - queryResult.Error=%v, existingMeta.ID=%d, err=%v", queryResult.Error, existingMeta.ID, queryResult.Error)
	
	if queryResult.Error == nil && existingMeta.ID > 0 {
		log.Printf("⚙️ [UploadPackage] 执行 UPDATE，ID=%d", existingMeta.ID)
		// 记录存在，执行 UPDATE（使用 map 避免 updated_at 类型错误）
		updateResult := cc.db.Model(&existingMeta).Updates(map[string]interface{}{
			"hash":     pkgMeta.Hash,
			"size":     pkgMeta.Size,
			"filename": pkgMeta.Filename,
			"filepath": pkgMeta.Filepath,
			"version":  pkgMeta.Version, // 新增：更新版本号
		})
		if updateErr := updateResult.Error; updateErr != nil {
			log.Printf("❌ [UploadPackage] 更新包元数据失败：%v", updateErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("更新数据库记录失败：%v", updateErr)})
			c.Abort()
			return
		}
		if rowsAffected := updateResult.RowsAffected; rowsAffected > 0 {
			log.Printf("✅ 已更新 %s 安装包：ID=%d, path=%s, size=%d, hash=%s", pkgType, existingMeta.ID, savePath, file.Size, hash)
		} else {
			log.Printf("ℹ️  已更新 %s 安装包：ID=%d, path=%s (无变化，大小=%d, hash=%s)", pkgType, existingMeta.ID, savePath, file.Size, hash)
		}
	} else {
		log.Printf("ℹ️ [UploadPackage] 不存在 %s 记录，将创建新记录", pkgType)
		// 记录不存在，执行 CREATE
		pkgMeta.Type = pkgType
		createErr := cc.db.Create(pkgMeta).Error
		if createErr != nil {
			log.Printf("❌ 创建包元数据失败：%v", createErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建数据库记录失败：%v", createErr)})
			c.Abort()
			return
		}
		log.Printf("✅ 已创建 %s 安装包：ID=%d, path=%s, size=%d, hash=%s", pkgType, pkgMeta.ID, savePath, file.Size, hash)
	}

	isReplaced := (existingMeta.ID > 0)
	log.Printf("✅ %s 安装包上传成功，路径：%s，大小：%d bytes，哈希：%s，是否替换：%v",
		pkgType, savePath, file.Size, hash, isReplaced)

	c.JSON(http.StatusOK, gin.H{
		"message":  "上传成功",
		"replaced": isReplaced,
		"size":     file.Size,
		"hash":     hash,
		"path":     savePath,
	})
}

// DownloadPackage 下载安装包
func (cc *ClientController) DownloadPackage(c *gin.Context) {
	pkgType := c.Param("type")
	if pkgType != "linux" && pkgType != "windows" {
		c.JSON(http.StatusNotFound, gin.H{"error": "无效的包类型"})
		c.Abort()
		return
	}

	// 先从数据库读取包元数据中的文件路径
	var pkgMeta models.PackageMeta
	err := cc.db.Where("type = ?", pkgType).First(&pkgMeta).Error

	if err == nil && pkgMeta.Filepath != "" {
		// 如果数据库中记录了文件路径，优先使用
		// 检查文件是否存在
		_, err := os.Stat(pkgMeta.Filepath)
		if err == nil {
			// 文件存在，直接下载
			// === 添加安全响应头 ===
			filename := filepath.Base(pkgMeta.Filepath)
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Download-Options", "noopen")
			c.Header("Cache-Control", "no-cache")
			c.Header("X-File-Size", fmt.Sprintf("%d", pkgMeta.Size))

			c.File(pkgMeta.Filepath)
			c.Abort()
			return
		}
		log.Printf("⚠️ 数据库记录的文件不存在：%s", pkgMeta.Filepath)
	}

	// 如果数据库中不存在记录或文件不存在，回退到自动查找
	path := cc.getPackagePath(pkgType)
	if path == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s 安装包不存在", pkgType)})
		c.Abort()
		return
	}

	// 文件存在，直接下载
	filename := filepath.Base(path)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Download-Options", "noopen")
	c.Header("Cache-Control", "no-cache")

	// 提供文件下载
	c.File(path)
	c.Abort()
}
