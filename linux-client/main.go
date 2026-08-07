package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var version = "dev" // 版本号，在编译时通过 -ldflags 注入
var config Config
var tokenManager *TokenManager

func main() {
	log.Printf("=== Linux Hardening Client v%s ===", version)
	
	// 加载配置
	configPath := "/opt/linux-hardening-client/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	
	var err error
	config, err = LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	log.Printf("Server URL: %s", config.ServerURL)
	log.Printf("Device: %s (%s)", config.DeviceName, config.IPAddress)
	
	// 初始化 Token Manager
	tokenManager = NewTokenManager(config.LocalDBPath)
	
	// 尝试加载现有 tokens
	if err := tokenManager.Load(); err != nil {
		log.Printf("没有现有 tokens，正在尝试注册...")
		// 没有 tokens，需要注册
		tempResp, err := RequestTempToken(config.DeviceName, config.IPAddress)
		if err != nil {
			log.Fatalf("请求临时 token 失败：%v", err)
		}
		
		log.Printf("获取到临时 token: %s...", tempResp.TempToken[:20])
		log.Println("正在注册客户端...")
		
		// 获取真实操作系统信息
		osVersion := GetOSInfo()
		log.Printf("检测到操作系统：%s", osVersion)
		
		regResp, err := RegisterWithTempToken(tempResp.TempToken, config.DeviceName, config.IPAddress, osVersion, version)
		if err != nil {
			log.Fatalf("注册失败：%v", err)
		}
		
		// 解析 expiration 时间
		expiresAt, err := time.Parse(time.RFC3339, regResp.ExpiresAt)
		if err != nil {
			log.Fatalf("解析过期时间失败：%v", err)
		}
		
		// 保存新 tokens
		if err := tokenManager.Save(regResp.ShortToken, regResp.RefreshToken, expiresAt); err != nil {
			log.Fatalf("保存 tokens 失败：%v", err)
		}
		
		log.Printf("✅ 客户端注册成功! UUID: %s", regResp.ClientUUID)
	} else {
		log.Println("从数据库加载了现有 tokens")
	}
	
	// 启动定时任务
	go dailyTaskScheduler()
	
	// 启动心跳循环（每 2 分钟发送一次）
	go heartbeatLoop()
	
	log.Println("Client started and waiting for tasks...")
	
	// 启动版本检查协程（延迟 30 秒后开始首次检查）
	go func() {
		time.Sleep(30 * time.Second)
	}()
	go checkUpdateLoop()
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	
	log.Println("Shutting down client...")
	os.Exit(0)
}

func dailyTaskScheduler() {
	log.Println("Starting daily task scheduler...")
	
	// 立即执行第一次检查（首次安装或更新后）
	log.Println("🚀 Performing initial security check...")
	runDailyCheck()
	
	// 之后每天执行一次
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		runDailyCheck()
	}
}

func runDailyCheck() {
	log.Println("[CHECK] Starting daily security check...")
	
	// 1. 执行加固脚本
	logString, err := executeScript()
	if err != nil {
		log.Printf("[ERROR] Script execution failed: %v", err)
		return
	}

	// 2. 解析输出
	checkData := parseOutput(logString)
	if checkData == nil {
		log.Println("[ERROR] Failed to parse script output")
		return
	}
	
	// 设置客户端版本号 (从 main.go 的 var version)
	checkData.ClientVersion = version

	// 3. 检查 Token 是否有效
	if tokenManager.IsExpired() {
		log.Println("[TOKEN] Token expired or expiring, attempting refresh...")
		if err := tokenManager.Refresh(); err != nil {
			log.Printf("[ERROR] Token refresh failed: %v", err)
			log.Println("[ERROR] Please reinstall the client with install_client_interactive.sh")
			return
		}
		log.Println("[TOKEN] Token refreshed successfully")
	}

	// 4. 上传数据
	log.Printf("📊 UploadData - DnfConfGpgcheck: %s, RedhatRepoGpgcheck: %s", checkData.DnfConfGpgcheck, checkData.RedhatRepoGpgcheck)
	if err := uploadData(checkData); err != nil {
		log.Printf("[ERROR] Upload failed: %v", err)
		return
	}

	log.Println("[CHECK] ✅ Daily check completed successfully")
}

// heartbeatLoop 每隔 2 分钟发送一次心跳
func heartbeatLoop() {
	log.Println("Starting heartbeat loop (every 2 minutes)...")

	// 立即发送第一次心跳
	log.Println("💓 Sending initial heartbeat...")
	if err := sendHeartbeat(); err != nil {
		if isAuthError(err) {
			reRegister()
		} else {
			log.Printf("[HEARTBEAT] Error: %v", err)
		}
	}

	// 之后每 2 分钟发送一次
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := sendHeartbeat(); err != nil {
			if isAuthError(err) {
				reRegister()
			} else {
				log.Printf("[HEARTBEAT] Error: %v", err)
			}
		}
	}
}

// sendHeartbeat 向服务器发送心跳
func sendHeartbeat() error {
	token := tokenManager.GetShortToken()
	if token == "" {
		log.Printf("[HEARTBEAT] No token available, skipping heartbeat")
		return nil
	}

	resp, err := SendHeartbeat(token)
	if err != nil {
		return fmt.Errorf("heartbeat failed: %v", err)
	}
	log.Printf("💓 Heartbeat sent successfully: %s", resp.Status)
	return nil
}

// isAuthError 判断错误是否为认证失败（HTTP 401/403）
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "HTTP 403")
}

// reRegister 清除本地 Token 并重新注册
func reRegister() {
	log.Println("[AUTH] 认证失败，清除本地 Token 并重新注册...")
	tokenManager.Clear()

	tempResp, err := RequestTempToken(config.DeviceName, config.IPAddress)
	if err != nil {
		log.Printf("[AUTH] 请求临时 token 失败: %v", err)
		return
	}

	osVersion := GetOSInfo()
	regResp, err := RegisterWithTempToken(tempResp.TempToken, config.DeviceName, config.IPAddress, osVersion, version)
	if err != nil {
		log.Printf("[AUTH] 重新注册失败: %v", err)
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, regResp.ExpiresAt)
	if err != nil {
		log.Printf("[AUTH] 解析过期时间失败: %v", err)
		return
	}

	if err := tokenManager.Save(regResp.ShortToken, regResp.RefreshToken, expiresAt); err != nil {
		log.Printf("[AUTH] 保存 tokens 失败: %v", err)
		return
	}

	log.Printf("[AUTH] ✅ 重新注册成功! UUID: %s", regResp.ClientUUID)
}
