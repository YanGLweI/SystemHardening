package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// normalizeHardwareUUID 硬件 UUID 统一归一化：去首尾空白 + 转大写 + 去内部空格。
// 注册查询、创建、更新、心跳回填全部读写点必须复用此函数，避免格式差异导致去重碎片化。
func normalizeHardwareUUID(s string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(s)), " ", "")
}

var hardwareUUIDRegex = regexp.MustCompile(`(?i)^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$`)

// looksLikeHardwareUUID 是否为标准 SMBIOS UUID 格式（用于识别 BIOS 序列号等遗留无效存储值）
func looksLikeHardwareUUID(s string) bool {
	return hardwareUUIDRegex.MatchString(strings.TrimSpace(s))
}

// shouldAdoptHardwareUUID 判定服务端是否应采纳客户端上报的硬件 UUID：
// 1. 存储为空：采纳（调用方另行查重）
// 2. 上报为标准 UUID 格式且存储为非标准（遗留 BIOS 序列号/占位值）：采纳作为纠正
// 3. 其余情况保留存储值（防冒认/克隆机覆盖）
func shouldAdoptHardwareUUID(stored, reported string) bool {
	if strings.TrimSpace(stored) == "" {
		return true
	}
	return looksLikeHardwareUUID(reported) && !looksLikeHardwareUUID(stored)
}

// isPlaceholderHardwareUUID 判定归一化后的硬件 ID 是否为 BIOS 占位值（与客户端 is_placeholder_serial 同源）。
// 入参须先经 normalizeHardwareUUID 处理（大写、无空格）。
func isPlaceholderHardwareUUID(normalized string) bool {
	switch normalized {
	case "DEFAULTSTRING", "TOBEFILLEDBYO.E.M.", "TOBEFILLEDBYOEM", "NOTSPECIFIED",
		"NONE", "NULL", "N/A", "UNKNOWN", "O.E.M.",
		"SYSTEMSERIALNUMBER", "CHASSISSERIALNUMBER", "BASEBOARDSERIALNUMBER",
		"SERIAL", "0", "0000000000", "123456789", "0123456789":
		return true
	}
	return false
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
