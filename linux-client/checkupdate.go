package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// CheckUpdateResponse 检查更新响应
type CheckUpdateResponse struct {
	HasUpdate     bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	NewVersion    string `json:"new_version"`
	DownloadURL   string `json:"download_url"`
	Hash          string `json:"hash"`
	Size          int64  `json:"size"`
	Filename      string `json:"filename"`
}

// CheckForUpdate 检查是否有新版本
func CheckForUpdate() (*CheckUpdateResponse, error) {
	token := tokenManager.GetShortToken()
	if token == "" {
		return nil, fmt.Errorf("no token available")
	}

	// 构建请求 URL
	url := fmt.Sprintf("%s/api/client/check-update", config.ServerURL)

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置 Token
	req.Header.Set("X-Client-Token", token)

	// 【关键】携带本地实际运行版本，供后端比对并同步更新记录，避免更新后版本不同步导致重复更新
	req.Header.Set("X-Client-Version", version)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check update failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result CheckUpdateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	log.Printf("✅ 版本检查结果：当前=%s, 最新=%s, has_update=%v", 
		result.CurrentVersion, result.NewVersion, result.HasUpdate)

	return &result, nil
}

// checkUpdateLoop 版本检查循环
func checkUpdateLoop() {
	log.Println("Starting version check loop (every 5 minutes)...")

	// 立即执行第一次检查
	checkAndDownloadUpdate()

	// 之后每 5 分钟检查一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		checkAndDownloadUpdate()
	}
}

// checkAndDownloadUpdate 检查并下载更新
func checkAndDownloadUpdate() {
	log.Println("[UPDATE] Checking for new version...")

	response, err := CheckForUpdate()
	if err != nil {
		log.Printf("[UPDATE] Check failed: %v", err)
		return
	}

	if !response.HasUpdate {
		log.Println("[UPDATE] No new version available")
		return
	}

	log.Printf("[UPDATE] New version found: %s -> %s", response.CurrentVersion, response.NewVersion)

	// 【关键】本地防护：后端返回的目标版本与本地实际版本一致时跳过，避免重复更新死循环
	if response.NewVersion == version {
		log.Printf("[UPDATE] Target version equals local version %s, skip installation", version)
		return
	}
	
	// 调用下载器
	tempPath, err := DownloadUpdate(response.DownloadURL, response.Filename, response.Hash)
	if err != nil {
		log.Printf("[UPDATE] Download failed: %v", err)
		return
	}

	log.Println("[UPDATE] Update package downloaded successfully")

	// 调用安装器进行安装
	if err := InstallUpdate(tempPath); err != nil {
		log.Printf("[UPDATE] Installation failed: %v", err)
		return
	}

	log.Println("[UPDATE] ✅ Update installed successfully!")
}
