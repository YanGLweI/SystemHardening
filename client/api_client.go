package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// RequestTempTokenResponse 临时 Token 响应
type RequestTempTokenResponse struct {
	TempToken   string `json:"temp_token"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   string `json:"expires_at"`
	DeviceName  string `json:"device_name"`
	IPAddress   string `json:"ip_address"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	TempToken  string `json:"temp_token"`
	DeviceName string `json:"device_name"`
	IPAddress  string `json:"ip_address"`
	OSVersion  string `json:"os_version,omitempty"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ClientUUID   string `json:"client_uuid"`
	ShortToken   string `json:"short_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	DeviceName   string `json:"device_name"`
	IPAddress    string `json:"ip_address"`
}

// RequestTempToken 请求临时安装 Token
func RequestTempToken(deviceName, ipAddress string) (*RequestTempTokenResponse, error) {
	payload := map[string]string{
		"device_name": deviceName,
		"ip_address":  ipAddress,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %v", err)
	}
	
	resp, err := http.Post(
		config.ServerURL+"/api/client/request-temp-token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request temp token failed: %s", string(body))
	}
	
	var result RequestTempTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}
	
	return &result, nil
}

// RegisterWithTempToken 使用临时 Token 注册客户端
func RegisterWithTempToken(tempToken, deviceName, ipAddress, osVersion string) (*RegisterResponse, error) {
	reqBody := RegisterRequest{
		TempToken:  tempToken,
		DeviceName: deviceName,
		IPAddress:  ipAddress,
		OSVersion:  osVersion,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}
	
	resp, err := http.Post(
		config.ServerURL+"/api/client/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("register failed: %s", string(body))
	}
	
	var result RegisterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}
	
	return &result, nil
}

// UploadData 上传加固检查结果
func uploadData(checkData *SystemCheckData) error {
	log.Printf("[API] Uploading data for device: %s", config.DeviceName)
	
	payload := map[string]interface{}{
		"client_uuid": "", // TODO: 可以从配置或数据库中获取
		"data":        checkData,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal data failed: %v", err)
	}
	
	req, err := http.NewRequest(
		"POST",
		config.ServerURL+"/api/client/upload-data",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	
	// 设置 Token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Token", tokenManager.GetShortToken())
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: HTTP %d, body: %s", resp.StatusCode, string(body))
	}
	
	log.Printf("[API] ✅ Data uploaded successfully, status code: %d", resp.StatusCode)
	return nil
}
