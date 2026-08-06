package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// TokenData token 数据结构
type TokenData struct {
	ShortToken   string    `json:"short_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TokenManager Token 管理器（使用 JSON 文件）
type TokenManager struct {
	dbPath       string
	shortToken   string
	refreshToken string
	expiresAt    time.Time
}

// NewTokenManager 创建 Token Manager
func NewTokenManager(dbPath string) *TokenManager {
	return &TokenManager{
		dbPath: dbPath,
	}
}

// Load 从 JSON 文件加载 Token
func (tm *TokenManager) Load() error {
	data, err := os.ReadFile(tm.dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no token file found")
		}
		return fmt.Errorf("read failed: %v", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return fmt.Errorf("parse failed: %v", err)
	}

	tm.shortToken = tokenData.ShortToken
	tm.refreshToken = tokenData.RefreshToken
	tm.expiresAt = tokenData.ExpiresAt

	return nil
}

// Save 保存 Token 到 JSON 文件
func (tm *TokenManager) Save(shortToken, refreshToken string, expiresAt time.Time) error {
	tokenData := TokenData{
		ShortToken:   shortToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}

	if err := os.WriteFile(tm.dbPath, data, 0600); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	tm.shortToken = shortToken
	tm.refreshToken = refreshToken
	tm.expiresAt = expiresAt

	log.Printf("✅ Tokens saved to %s", tm.dbPath)
	return nil
}

// GetShortToken 获取短期 Token
func (tm *TokenManager) GetShortToken() string {
	return tm.shortToken
}

// GetRefreshToken 获取刷新 Token
func (tm *TokenManager) GetRefreshToken() string {
	return tm.refreshToken
}

// IsExpired 检查 Token 是否过期（提前 24 小时预警）
func (tm *TokenManager) IsExpired() bool {
	if tm.expiresAt.IsZero() {
		return true
	}

	// 如果剩余时间少于 24 小时，视为即将过期
	return time.Until(tm.expiresAt) < 24*time.Hour
}

// AutoRefresh 自动刷新 Token（如果即将过期）
func (tm *TokenManager) AutoRefresh() error {
	if !tm.IsExpired() {
		return nil // Token 有效，无需刷新
	}

	return tm.Refresh()
}

// Refresh 刷新 Token
func (tm *TokenManager) Refresh() error {
	// 调用服务器刷新接口
	payload := map[string]string{
		"refresh_token": tm.refreshToken,
	}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(
		config.ServerURL+"/api/client/refresh-token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed: %s", string(body))
	}

	// 解析响应
	var result struct {
		ShortToken string `json:"short_token"`
		ExpiresAt  string `json:"expires_at"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response failed: %v", err)
	}

	// 保存新 Token
	expiresAt, _ := time.Parse(time.RFC3339, result.ExpiresAt)
	if err := tm.Save(result.ShortToken, tm.refreshToken, expiresAt); err != nil {
		return fmt.Errorf("save token failed: %v", err)
	}

	log.Printf("Token refreshed successfully, expires at: %s", result.ExpiresAt)
	return nil
}

// Clear 清除本地 Token（删除文件并重置内存状态）
func (tm *TokenManager) Clear() {
	tm.shortToken = ""
	tm.refreshToken = ""
	tm.expiresAt = time.Time{}

	if _, err := os.Stat(tm.dbPath); err == nil {
		if err := os.Remove(tm.dbPath); err != nil {
			log.Printf("[TOKEN] 删除 Token 文件失败: %v", err)
		} else {
			log.Printf("[TOKEN] 已清除本地 Token 文件: %s", tm.dbPath)
		}
	}
}
