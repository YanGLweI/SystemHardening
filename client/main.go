package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "1.0.0"
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
		
		regResp, err := RegisterWithTempToken(tempResp.TempToken, config.DeviceName, config.IPAddress, "Red Hat Enterprise Linux release 9.7 (Plow)")
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
	
	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Client started and waiting for tasks...")
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
	
	// 3. 检查 Token 是否有效
	if tokenManager.IsExpired() {
		log.Println("[TOKEN] Token expired or expiring, attempting refresh...")
		if err := tokenManager.Refresh(); err != nil {
			log.Printf("[ERROR] Token refresh failed: %v", err)
			log.Println("[ERROR] Please reinstall the client with install.sh")
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
