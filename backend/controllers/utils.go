package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// normalizeHardwareUUID 硬件 UUID 统一归一化：去首尾空白 + 转大写 + 去内部空格。
// 注册查询、创建、更新、心跳回填全部读写点必须复用此函数，避免格式差异导致去重碎片化。
func normalizeHardwareUUID(s string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(s)), " ", "")
}

// TempTokenInfo 临时 Token 信息
type TempTokenInfo struct {
	DeviceName string
	IPAddress  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Used       bool
}

// tempTokenStore 临时 Token 存储（生产环境使用 Redis）
var tempTokenStore = make(map[string]TempTokenInfo)
var tokenMutex sync.RWMutex

// generateTempToken 生成临时安装 Token（20 字符随机字符串）
func generateTempToken(deviceName, ipAddress string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	// 添加时间戳前缀增加安全性
	prefix := fmt.Sprintf("%d_", time.Now().Unix())
	return prefix + token
}

// generateRefreshToken 生成刷新 Token（90 天有效期）
func generateRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateShortToken 生成短期访问 Token（14 天有效期）
func generateShortToken() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateUUID 生成 UUID v4
func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[8] = (uuid[8] & 0xff &^ 0x80) | 0x80
	uuid[6] = (uuid[6] & 0xff & 0x4f) | 0x40

	return fmt.Sprintf("%X-%X-%X-%X-%X",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:])
}
