package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// DownloadUpdate 下载更新包并验证 hash，返回临时文件路径
func DownloadUpdate(url, filename, expectedHash string) (string, error) {
	log.Printf("[DOWNLOADER] Starting download: %s", url)

	// 1. 生成临时文件路径
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	tempPath := fmt.Sprintf("/tmp/systemhardening-update-%s.zip", timestamp)

	// 2. 发起 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// 3. 创建临时文件
	file, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %v", err)
	}
	defer file.Close()

	// 4. 计算 MD5 hash
	hasher := md5.New()

	// 5. 写入文件并计算 hash
	written, err := io.Copy(io.MultiWriter(file, hasher), resp.Body)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("download failed: %v", err)
	}

	log.Printf("[DOWNLOADER] Downloaded %d bytes to %s", written, tempPath)

	// 6. 关闭文件
	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("close file failed: %v", err)
	}

	// 7. 验证 hash
	actualHash := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("[DOWNLOADER] MD5 verification: expected=%s, actual=%s", expectedHash, actualHash)

	if actualHash != expectedHash {
		os.Remove(tempPath)
		return "", fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	log.Printf("[DOWNLOADER] ✅ MD5 verified successfully")

	return tempPath, nil
}
