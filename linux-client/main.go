package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var version = "dev" // 版本号，在编译时通过 -ldflags 注入
var config Config
var tokenManager *TokenManager
var clientUUID string

func main() {
	// 【关键】日志双路输出：同时写入 logs/client.log 和 stderr（systemd journal）
	setupLogging()

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
		// 实时采集设备名/主 IP（对齐 Windows：改名/换 IP 后重注册不上报旧值，config 仅作兜底）
		deviceName, ipAddress := currentDeviceInfo()
		tempResp, err := RequestTempToken(deviceName, ipAddress)
		if err != nil {
			log.Fatalf("请求临时 token 失败：%v", err)
		}

		log.Printf("获取到临时 token: %s...", tempResp.TempToken[:20])
		log.Println("正在注册客户端...")

		// 获取真实操作系统信息
		osVersion := GetOSInfo()
		log.Printf("检测到操作系统：%s", osVersion)

		// 采集硬件 UUID（SMBIOS，无 DMI 环境返回空串降级）
		hwUUID := CollectHardwareUUID()
		if hwUUID != "" {
			log.Printf("采集到硬件 UUID：%s", hwUUID)
		}

		regResp, err := RegisterWithTempToken(tempResp.TempToken, deviceName, ipAddress, osVersion, version, hwUUID)
		if err != nil {
			log.Fatalf("注册失败：%v", err)
		}

		// 解析 expiration 时间
		expiresAt, err := time.Parse(time.RFC3339, regResp.ExpiresAt)
		if err != nil {
			log.Fatalf("解析过期时间失败：%v", err)
		}

		// 设置客户端 UUID（必须先设置再保存）
		tokenManager.SetClientUUID(regResp.ClientUUID)

		// 设置硬件 UUID（与 UUID 一并持久化）
		tokenManager.SetHardwareUUID(hwUUID)

		// 保存新 tokens（包含 UUID）
		if err := tokenManager.Save(regResp.ShortToken, regResp.RefreshToken, expiresAt); err != nil {
			log.Fatalf("保存 tokens 失败：%v", err)
		}

		clientUUID = regResp.ClientUUID
		log.Printf("✅ 客户端注册成功！UUID: %s", regResp.ClientUUID)
	} else {
		log.Println("从数据库加载了现有 tokens")
		// 恢复客户端 UUID
		clientUUID = tokenManager.GetClientUUID()

		// 启动重校验：缓存非标准 UUID（空/遗留值）一律重采并覆盖（对齐 Windows 2.3.3 语义）
		cachedHW := tokenManager.GetHardwareUUID()
		if !IsValidHardwareUUID(cachedHW) {
			log.Printf("缓存 hardware_uuid 为空/遗留值（%s），重新采集...", cachedHW)
			newHW := CollectHardwareUUID()
			if newHW != cachedHW {
				if err := tokenManager.PersistHardwareUUID(newHW); err != nil {
					log.Printf("⚠️ 保存 hardware_uuid 失败：%v", err)
				} else {
					log.Printf("✅ hardware_uuid 已更新: %s -> %s", cachedHW, newHW)
				}
			}
		}
	}

	// 启动检查计划拉取协程（每 5 分钟从服务端获取计划）
	go scheduleLoop()

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

	// 启动任务轮询协程（每 5 分钟检查一次 pending tasks）
	go taskScheduler()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("Shutting down client...")
	os.Exit(0)
}

func dailyTaskScheduler() {
	log.Println("Starting daily task scheduler...")

	// 立即执行第一次检查（首次安装或更新后的兜底补检）
	log.Println("🚀 Performing initial security check...")
	runDailyCheck()

	// 之后按服务端下发的检查计划执行（每 30 秒判断一次是否到达计划时刻）
	for {
		time.Sleep(30 * time.Second)

		next := getNextCheckTime()
		if next.IsZero() || time.Now().Before(next) {
			continue // 尚未应用计划或未到达计划时刻
		}

		log.Printf("[CHECK] 到达计划检查时刻（%s），开始执行检查...", next.Format("2006-01-02 15:04:05"))
		runDailyCheck()
		recomputeNextCheck()
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
		// 【关键】检测 tokens.json 文件是否被删除，若丢失则自动重新注册（无需重启服务）
		if !tokenManager.FileExists() {
			log.Println("[TOKEN] tokens.json 文件丢失，自动重新注册...")
			reRegister()
			continue
		}
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

	// 实时采集设备信息（与 Windows 客户端心跳一致：服务端据此同步 device_name/ip_address）
	resp, err := SendHeartbeat(token, tokenManager.GetHardwareUUID(), GetHostname(), GetLocalIP())
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

// currentDeviceInfo 实时采集设备名与主 IP，采集失败时回退 config.yaml 静态值（仅作引导兜底）
func currentDeviceInfo() (deviceName, ipAddress string) {
	deviceName = GetHostname()
	if deviceName == "" {
		deviceName = config.DeviceName
	}
	ipAddress = GetLocalIP()
	if ipAddress == "" {
		ipAddress = config.IPAddress
	}
	return deviceName, ipAddress
}

// reRegister 清除本地 Token 并重新注册
func reRegister() {
	log.Println("[AUTH] 认证失败，清除本地 Token 并重新注册...")
	tokenManager.Clear()

	// 实时采集设备名/主 IP（改名/换 IP 后重注册不上报旧值，config 仅作兜底）
	deviceName, ipAddress := currentDeviceInfo()
	tempResp, err := RequestTempToken(deviceName, ipAddress)
	if err != nil {
		log.Printf("[AUTH] 请求临时 token 失败: %v", err)
		return
	}

	osVersion := GetOSInfo()
	hwUUID := CollectHardwareUUID()
	regResp, err := RegisterWithTempToken(tempResp.TempToken, deviceName, ipAddress, osVersion, version, hwUUID)
	if err != nil {
		log.Printf("[AUTH] 重新注册失败: %v", err)
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, regResp.ExpiresAt)
	if err != nil {
		log.Printf("[AUTH] 解析过期时间失败: %v", err)
		return
	}

	// 设置客户端 UUID（必须先设置再保存，确保 save 时一并持久化）
	tokenManager.SetClientUUID(regResp.ClientUUID)

	// 硬件 UUID 同步更新（重注册场景下重新采集）
	tokenManager.SetHardwareUUID(hwUUID)

	if err := tokenManager.Save(regResp.ShortToken, regResp.RefreshToken, expiresAt); err != nil {
		log.Printf("[AUTH] 保存 tokens 失败：%v", err)
		return
	}

	clientUUID = regResp.ClientUUID
	log.Printf("[AUTH] ✅ 重新注册成功！UUID: %s", regResp.ClientUUID)
}

// setupLogging 日志同时写入 logs/client.log 和 stderr（journal）
// 超过 10MB 时轮转为 client.log.1，避免日志无限增长
func setupLogging() {
	logPath := filepath.Join(logDir, "client.log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("⚠️ 创建日志目录失败（%v），仅输出到控制台", err)
		return
	}

	// 简单轮转：超过 10MB 时重命名为 client.log.1
	if info, err := os.Stat(logPath); err == nil && info.Size() > 10*1024*1024 {
		os.Remove(logPath + ".1")
		os.Rename(logPath, logPath+".1")
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("⚠️ 打开日志文件失败（%v），仅输出到控制台", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	log.Printf("日志输出到: %s 和 systemd journal", logPath)
}

// taskScheduler 定时轮询待执行任务（每 5 分钟检查一次）
func taskScheduler() {
	log.Println("Starting task scheduler loop (every 5 minutes)...")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ProcessPendingTasks()
	}
}
