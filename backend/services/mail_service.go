package services

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/models"
	"gorm.io/gorm"
)

// MailService 邮件服务
type MailService struct {
	db *gorm.DB
}

// NewMailService 初始化邮件服务
func NewMailService(db *gorm.DB) *MailService {
	return &MailService{db: db}
}

// ClientCheckInfo 客户端检查信息（用于 HTML 报告渲染）
type ClientCheckInfo struct {
	Client     models.Client
	Check      interface{}
	Compliance *models.ComplianceResult
	SystemType string // Linux / Windows
}

// SendEmail SMTP 发送邮件
func (s *MailService) SendEmail(to []string, subject string, body string) error {
	// 获取邮件配置
	var config models.MailConfig
	if err := s.db.Last(&config).Error; err != nil {
		return fmt.Errorf("failed to get mail config: %v", err)
	}

	if !config.IsEnabled {
		return fmt.Errorf("mail service is disabled")
	}

	log.Printf("📧 SendEmail 开始: 收件人=%v 主题=%s 服务器=%s 端口=%d 账号=%s",
		to, subject, config.SMTPHost, config.SMTPPort, config.Username)

	// 构建发件人地址（用于邮件头部的视觉 From）
	fromEmail := config.Username
	if config.FromEmail != "" {
		fromEmail = config.FromEmail
	}

	// SMTP 信封 MAIL FROM 必须使用认证用户，否则服务器会静默丢弃
	mailFrom := config.Username

	// 检查密码是否为空
	if config.Password == "" {
		log.Printf("❌ SMTP 密码为空，请在前端重新配置密码")
		return fmt.Errorf("SMTP password is empty, please configure password in the frontend")
	}

	// 构建邮件头
	headers := make(map[string]string)
	// 对 From 显示名进行 RFC 2047 编码，避免中文等非 ASCII 字符被邮件服务器拒绝
	headers["From"] = fmt.Sprintf("=?UTF-8?B?%s?= <%s>", encodeBase64("系统加固平台"), fromEmail)
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", encodeBase64(subject))
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Content-Transfer-Encoding"] = "8bit"
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	headers["Message-ID"] = fmt.Sprintf("<%s.%x@%s>", time.Now().Format("20060102150405"), rand.Intn(0xffffffff), extractDomain(config.Username))
	headers["X-Mailer"] = "System Hardening Platform"

	// 构造完整邮件内容
	msg := ""
	for key, value := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	msg += "\r\n" + body

	auth := smtp.PlainAuth(
		"",
		config.Username,
		config.Password,
		config.SMTPHost,
	)

	var err error
	log.Printf("📧 准备连接 SMTP 服务器: %s:%d", config.SMTPHost, config.SMTPPort)
	switch config.SMTPPort {
	case 465:
		// 端口 465：隐式 TLS
		err = s.sendWithTLS(config.SMTPHost, config.SMTPPort, mailFrom, fromEmail, config.Username, config.Password, to, msg)
	case 587:
		// 端口 587：STARTTLS
		serverName := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
		err = s.sendWithSTARTTLS(serverName, auth, mailFrom, fromEmail, to, msg)
	default:
		// 端口 25：明文
		serverName := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
		err = s.sendPlain(serverName, auth, mailFrom, fromEmail, to, msg)
	}
	if err != nil {
		log.Printf("❌ SendEmail 发送失败: %v", err)
		return err
	}

	log.Printf("✅ Email sent successfully to %v: %s", to, subject)
	return nil
}

// dialSMTP 建立带超时的 SMTP 连接
func dialSMTP(addr string) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	// 设置整体超时，避免邮件服务器不可达时请求长时间挂起
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return client, nil
}

// sendWithTLS 使用 TLS 发送邮件（适用于端口 465）
func (s *MailService) sendWithTLS(host string, port int, mailFrom string, fromEmail, username, password string, to []string, msg string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("🔌 sendWithTLS: 连接 %s", addr)

	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		log.Printf("❌ TCP 连接失败: %v", err)
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	log.Printf("✅ TCP 连接成功")

	// 隐式 TLS 握手
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	})
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("❌ TLS 握手失败: %v", err)
		return err
	}
	log.Printf("✅ TLS 握手成功")

	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		log.Printf("❌ SMTP 客户端创建失败: %v", err)
		return err
	}
	defer client.Close()
	log.Printf("✅ SMTP 客户端创建成功")

	auth := smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		log.Printf("❌ SMTP 认证失败: %v", err)
		return err
	}
	log.Printf("✅ SMTP 认证成功")

	// 使用认证用户作 SMTP 信封 MAIL FROM，确保邮件不因发件人不匹配被静默丢弃
	if err := client.Mail(mailFrom); err != nil {
		log.Printf("❌ MAIL FROM 失败 (%s): %v", mailFrom, err)
		return err
	}
	log.Printf("✅ MAIL FROM 成功: %s", mailFrom)

	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			log.Printf("❌ RCPT TO 失败 (%s): %v", rcpt, err)
			return err
		}
		log.Printf("✅ RCPT TO 成功: %s", rcpt)
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("❌ DATA 命令失败: %v", err)
		return err
	}
	log.Printf("✅ DATA 命令成功，开始写入邮件内容 (%d bytes)", len(msg))

	if _, err := w.Write([]byte(msg)); err != nil {
		log.Printf("❌ 写入邮件内容失败: %v", err)
		return err
	}

	if err := w.Close(); err != nil {
		log.Printf("❌ DATA 关闭失败: %v", err)
		return err
	}
	log.Printf("✅ DATA 写入完成")

	if err := client.Quit(); err != nil {
		log.Printf("❌ QUIT 失败: %v", err)
		return err
	}
	log.Printf("✅ QUIT 成功，邮件已发送至 SMTP 服务器")
	return nil
}

// sendWithSTARTTLS 使用 STARTTLS 发送邮件（适用于端口 587）
func (s *MailService) sendWithSTARTTLS(addr string, auth smtp.Auth, mailFrom string, fromEmail string, to []string, msg string) error {
	conn, err := dialSMTP(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 启用 STARTTLS
	if ok, _ := conn.Extension("STARTTLS"); ok {
		if err = conn.StartTLS(nil); err != nil {
			return err
		}
	}

	// 认证
	if auth != nil {
		if err = conn.Auth(auth); err != nil {
			return err
		}
	}

	// 设置发件人（使用认证用户，避免被静默丢弃）
	if err = conn.Mail(mailFrom); err != nil {
		return err
	}

	// 添加收件人
	for _, addr := range to {
		if err = conn.Rcpt(addr); err != nil {
			return err
		}
	}

	// 写入邮件内容
	writer, err := conn.Data()
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	return conn.Quit()
}

// sendPlain 使用普通 SMTP 发送邮件（端口 25）
func (s *MailService) sendPlain(addr string, auth smtp.Auth, mailFrom string, fromEmail string, to []string, msg string) error {
	conn, err := dialSMTP(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 认证
	if auth != nil {
		if err = conn.Auth(auth); err != nil {
			return err
		}
	}

	// 设置发件人（使用认证用户，避免被静默丢弃）
	if err = conn.Mail(mailFrom); err != nil {
		return err
	}

	// 添加收件人
	for _, addr := range to {
		if err = conn.Rcpt(addr); err != nil {
			return err
		}
	}

	// 写入邮件内容
	writer, err := conn.Data()
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	return conn.Quit()
}

// TestEmail 发送测试邮件
func (s *MailService) TestEmail(recipient string) error {
	// 获取邮件配置
	var config models.MailConfig
	if err := s.db.Last(&config).Error; err != nil {
		return fmt.Errorf("failed to get mail config: %v", err)
	}

	if !config.IsEnabled {
		return fmt.Errorf("mail service is disabled")
	}

	subject := "【系统加固平台】测试邮件"
	body := `
<html>
<body style="font-family: Arial, sans-serif; padding: 20px;">
<h2 style="color: #059669;">欢迎使用系统加固平台的邮件通知功能！</h2>
<p>这是一封来自 <strong>系统加固平台</strong> 的测试邮件。</p>
<p>如果您收到此邮件，说明邮件服务器配置正确。</p>
<hr/>
<p style="color: #999; font-size: 12px;">这是自动生成的测试邮件，请勿回复。</p>
</body>
</html>`

	to := []string{recipient}
	err := s.SendEmail(to, subject, body)
	if err != nil {
		// 更新测试结果
		s.db.Model(&config).Updates(models.MailConfig{
			TestResult: fmt.Sprintf("失败：%v", err),
			UpdatedAt:  time.Now(),
		})
		return err
	}

	// 更新测试结果
	s.db.Model(&config).Updates(models.MailConfig{
		TestResult: "成功",
		UpdatedAt:  time.Now(),
	})
	return nil
}

// GenerateReportHTML 生成 HTML 格式的加固报告
func (s *MailService) GenerateReportHTML(plan models.ReportSchedule) string {
	now := time.Now()
	
	// 1. 获取所有客户端
	var clients []models.Client
	s.db.Order("id DESC").Find(&clients)
	
	// 2. 获取 Linux 检查记录
	var linuxChecks []models.SystemCheck
	limit := len(clients)
	s.db.Limit(limit).Order("client_uuid DESC, id DESC").Find(&linuxChecks)
	
	// 3. 获取 Windows 检查记录
	var windowsChecks []models.WindowsSystemCheck
	s.db.Limit(limit).Order("client_uuid DESC, id DESC").Find(&windowsChecks)
	
	// 4. 按 UUID 映射最近的检查记录
	checkMap := make(map[string]interface{})
	for _, check := range linuxChecks {
		if _, exists := checkMap[check.ClientUUID]; !exists {
			checkMap[check.ClientUUID] = check
		}
	}
	for _, check := range windowsChecks {
		if _, exists := checkMap[check.ClientUUID]; !exists {
			checkMap[check.ClientUUID] = check
		}
	}
	
	// 5. 计算合规状态
	clientInfos := make([]ClientCheckInfo, 0)
	var totalCompliant, totalNonCompliant int
	
	// 6. 获取标准配置
	var linuxStandards []models.LinuxStandard
	s.db.Model(&models.LinuxStandard{}).Where("deleted_at IS NULL").Find(&linuxStandards)
	linuxStandardMap := make(map[string]string)
	for _, std := range linuxStandards {
		linuxStandardMap[std.FieldName] = std.StandardValue
	}
	
	var windowsStandards []models.WindowsStandard
	s.db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&windowsStandards)
	windowsStandardMap := make(map[string]string)
	for _, std := range windowsStandards {
		windowsStandardMap[std.FieldName] = std.StandardValue
	}
	
	// 7. 区域统计
	var regions []models.Region
	s.db.Preload("Clients").Find(&regions)
	
	regionStats := make([]gin.H, 0)
	overallOnline, overallOffline := 0, 0
	
	for _, region := range regions {
		onlineCount := 0
		offlineCount := 0
		for _, client := range region.Clients {
			if client.Status == "active" {
				onlineCount++
			} else {
				offlineCount++
			}
		}
		overallOnline += onlineCount
		overallOffline += offlineCount
		
		regionStats = append(regionStats, gin.H{
			"name":        region.Name,
			"total":       len(region.Clients),
			"online":      onlineCount,
			"offline":     offlineCount,
		})
	}
	
	// 8. 为每个客户端计算合规状态
	for uuid, check := range checkMap {
		client := models.Client{}
		s.db.Where("client_uuid = ?", uuid).First(&client)
		
		var compliance *models.ComplianceResult
		var systemType string
		var clientCheck interface{}
		
		switch c := check.(type) {
		case models.SystemCheck:
			systemType = "Linux"
			clientCheck = c
			result := models.CompareCompliance(&c, linuxStandardMap)
			compliance = result
			
			if compliance.Status == "compliant" {
				totalCompliant++
			} else {
				totalNonCompliant++
			}
			
		case models.WindowsSystemCheck:
			systemType = "Windows"
			clientCheck = c
			result := models.CompareWindowsCompliance(&c, windowsStandardMap)
			compliance = result
			
			if compliance.Status == "compliant" {
				totalCompliant++
			} else {
				totalNonCompliant++
			}
		}
		
		clientInfos = append(clientInfos, ClientCheckInfo{
			Client:    client,
			Check:     clientCheck,
			Compliance: compliance,
			SystemType: systemType,
		})
	}
	
	// 9. 生成 HTML
	html := s.renderReportHTML(now, totalCompliant, totalNonCompliant, regionStats, clientInfos, plan.Subject)
	
	return html
}

// renderReportHTML 渲染 HTML 报告
func (s *MailService) renderReportHTML(now time.Time, compliantCount, nonCompliantCount int, regionStats []gin.H, clientInfos []ClientCheckInfo, subject string) string {
	rand.Seed(time.Now().UnixNano())
	
	// 获取当前日期字符串
	dateStr := now.Format("2006 年 01 月 02 日 15:04:05")
	
	// 生成唯一 ID
	randomID := fmt.Sprintf("%x", rand.Intn(0xffffffff))
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background: #f5f7fa; padding: 20px; }
.container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 12px rgba(0,0,0,0.1); }
.header { background: linear-gradient(135deg, #059669 0%%, #10b981 100%%); color: white; padding: 30px; text-align: center; }
.header h1 { margin-bottom: 10px; font-size: 28px; }
.header .subtitle { opacity: 0.9; font-size: 14px; }
.cards { display: flex; justify-content: space-around; padding: 30px; background: #fafafa; border-bottom: 1px solid #eee; }
.card { text-align: center; min-width: 150px; }
.card .number { font-size: 36px; font-weight: bold; margin-bottom: 8px; }
.card.success .number { color: #10b981; }
.card.failure .number { color: #ef4444; }
.card .label { color: #666; font-size: 14px; }
.regions { padding: 20px 30px; background: #fff; border-top: 1px solid #eee; }
.regions h2 { color: #333; margin-bottom: 15px; font-size: 18px; }
.region-item { padding: 10px; background: #fafafa; margin-bottom: 8px; border-radius: 4px; }
.region-name { font-weight: 600; color: #333; }
.region-stats { font-size: 13px; color: #666; margin-top: 4px; }
.table-container { padding: 0 30px 30px; }
.el-table { width: 100%%; border-collapse: collapse; }
.el-table th, .el-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; font-size: 14px; }
.el-table th { background: #fafafa; font-weight: 600; color: #333; }
.el-table tr:hover { background: #f5f7fa; }
.status-tag { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }
.tag-success { background: #d1fae5; color: #059669; }
.tag-danger { background: #fee2e2; color: #ef4444; }
.details-row { display: table-row; }
.details-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; }
.detail-item { padding: 8px; background: white; border-radius: 4px; border: 1px solid #e5e7eb; }
.detail-label { color: #666; font-size: 13px; margin-bottom: 4px; }
.detail-value { color: #333; font-weight: 500; font-size: 14px; }
.detail-value.non-compliant { color: #ef4444; font-weight: 600; }
.standard-hint { color: #d97706; font-size: 12px; margin-left: 8px; background: #fffbeb; padding: 2px 6px; border-radius: 4px; }
.footer { text-align: center; padding: 20px; color: #999; font-size: 12px; border-top: 1px solid #eee; background: #fafafa; }
</style>
</head>
<body>
<div class="container" id="%s">
<div class="header">
<h1>🛡️ 系统加固合规报告</h1>
<p class="subtitle">%s</p>
</div>

<div class="cards">
<div class="card success">
<div class="number">%d</div>
<div class="label">合规数量</div>
</div>
<div class="card failure">
<div class="number">%d</div>
<div class="label">不合规数量</div>
</div>
</div>

<div class="regions">
<h2>各区域客户端统计</h2>
`+generateRegionHTML(regionStats)+`
</div>

<div class="table-container">
<h2>客户端加固情况</h2>
<table class="el-table">
<thead>
<tr>
<th>序号</th>
<th>计算机名</th>
<th>IP 地址</th>
<th>系统类型</th>
<th>合规状态</th>
<th>操作</th>
</tr>
</thead>
<tbody id="client-table-body">
`+generateClientTableHTML(clientInfos)+`
</tbody>
</table>
</div>

<div class="footer">
本报告由系统加固平台自动生成<br>
© 2024 System Hardening Platform. All rights reserved.
</div>
</div>
</body>
</html>`, 
		subject, randomID, dateStr, compliantCount, nonCompliantCount, randomID)
	
	return html
}

// generateRegionHTML 生成区域统计 HTML
func generateRegionHTML(stats []gin.H) string {
	html := "<div class=\"region-list\">\n"
	for _, stat := range stats {
		name := getString(stat, "name")
		total := getInt(stat, "total")
		online := getInt(stat, "online")
		offline := getInt(stat, "offline")
		html += fmt.Sprintf("<div class=\"region-item\"><div class=\"region-name\">📍 %s</div><div class=\"region-stats\">总计：%d | 在线：%d | 离线：%d</div></div>\n", name, total, online, offline)
	}
	html += "</div>"
	return html
}

// generateClientTableHTML 生成客户端表格 HTML（无 JavaScript 版本）
func generateClientTableHTML(infos []ClientCheckInfo) string {
	html := ""
	rand.Seed(time.Now().UnixNano())
	
	for i, info := range infos {
		rowID := fmt.Sprintf("row-%d", i)
		detailsID := fmt.Sprintf("details-%d", i)
		statusClass := "tag-success"
		statusLabel := "✅ 合规"
		if info.Compliance != nil && info.Compliance.Status == "non_compliant" {
			statusClass = "tag-danger"
			statusLabel = "❌ 不合规"
		}
		
		// 使用 <details><summary> 标签实现纯 HTML 折叠详情（无需 JavaScript）
		html += fmt.Sprintf(`<tr id="%s">
<td>%d</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td><span class="status-tag %s">%s</span></td>
<td>——</td>
</tr>
<tr id="%s" class="details-row"><td colspan="6">
<details>
<summary style="cursor: pointer; padding: 10px; background: #ecfdf5; color: #059669; font-weight: 500; border-radius: 4px;">👁️ 点击查看完整详情</summary>
<div class="details" id="%s">%s</div>
</details>
</td>
</tr>`, rowID, i+1,
			info.Client.DeviceName,
			info.Client.IPAddress,
			info.SystemType,
			statusClass, statusLabel, detailsID,
			detailsID, renderClientDetails(info))
	}
	
	return html
}

// renderClientDetails 渲染客户端详情
func renderClientDetails(info ClientCheckInfo) string {
	html := "<div class=\"details-grid\">"
	
	switch check := info.Check.(type) {
	case models.SystemCheck:
		html += renderLinuxDetails(check, info.Compliance)
	case models.WindowsSystemCheck:
		html += renderWindowsDetails(check, info.Compliance)
	}
	
	html += "</div>"
	return html
}

// renderLinuxDetails 渲染 Linux 详情
func renderLinuxDetails(check models.SystemCheck, compliance *models.ComplianceResult) string {
	html := ""
	fieldMap := map[string]string{
		"hostname":              check.Hostname,
		"ip":                    check.IP,
		"operasystem":           check.Operasystem,
		"kernel":                check.Kernel,
		"dnf_conf_gpgcheck":     check.DnfConfGpgcheck,
		"redhat_repo_gpgcheck":  check.RedhatRepoGpgcheck,
		"pass_max_days":         check.PassMaxDays,
		"pass_min_days":         check.PassMinDays,
		"pass_min_len":          check.PassMinLen,
		"pass_warn_age":         check.PassWarnAge,
		"inactive":              check.Inactive,
		"gid":                   check.GID,
		"tmout":                 check.Tmout,
		"cron":                  check.Cron,
		"crontab":               check.Crontab,
		"cron_hourly":           check.CronHourly,
		"cron_daily":            check.CronDaily,
		"cron_weekly":           check.CronWeekly,
		"cron_monthly":          check.CronMonthly,
		"cron_deny":             check.CronDeny,
		"at_deny":               check.AtDeny,
		"cron_allow":            check.CronAllow,
		"at_allow":              check.AtAllow,
		"sshd_config":           check.SshdConfig,
		"log_level":             check.LogLevel,
		"x11_forwarding":        check.X11Forwarding,
		"max_auth_tries":        check.MaxAuthTries,
		"ignore_rhosts":         check.IgnoreRhosts,
		"hostbased_authentication": check.HostbasedAuthentication,
		"permit_root_login":     check.PermitRootLogin,
		"permit_empty_passwords": check.PermitEmptyPasswords,
		"permit_user_environment": check.PermitUserEnvironment,
		"client_alive_interval": check.ClientAliveInterval,
		"client_alive_count_max": check.ClientAliveCountMax,
		"login_grace_time":      check.LoginGraceTime,
		"minlen":                check.Minlen,
		"minclass":              check.Minclass,
		"dcredit":               check.Dcredit,
		"ucredit":               check.Ucredit,
		"lcredit":               check.Lcredit,
		"ocredit":               check.Ocredit,
		"password_remember":     check.PasswordRemember,
		"passwd":                check.Passwd,
		"passwd_minus":          check.PasswdMinus,
		"group":                 check.GroupCol,
		"group_minus":           check.GroupMinus,
		"shadow":                check.Shadow,
		"shadow_minus":          check.ShadowMinus,
		"gshadow":               check.Gshadow,
		"gshadow_minus":         check.GshadowMinus,
		"crypto_policies":       check.CryptoPolicies,
		"ntp_server":            check.NtpServer,
	}
	
	fieldLabels := getFieldLabels()
	for fieldName, value := range fieldMap {
		if value == "" {
			continue
		}
		label := fieldLabels[fieldName]
		isNonCompliant := false
		standardValue := ""
		
		if compliance != nil {
			for _, nf := range compliance.NonCompliantFields {
				if nf.Field == fieldName {
					isNonCompliant = true
					standardValue = nf.Standard
					break
				}
			}
		}
		
		valueClass := "detail-value"
		if isNonCompliant {
			valueClass += " non-compliant"
		}
		
		html += fmt.Sprintf("<div class=\"detail-item\"><div class=\"detail-label\">%s</div><div class=\"%s\">%s%s</div></div>", 
			label, valueClass, value, 
			func() string {
				if standardValue != "" {
					return fmt.Sprintf("<span class=\"standard-hint\">标准：%s</span>", standardValue)
				}
				return ""
			}())
	}
	
	return html
}

// renderWindowsDetails 渲染 Windows 详情
func renderWindowsDetails(check models.WindowsSystemCheck, compliance *models.ComplianceResult) string {
	html := ""
	fieldMap := map[string]string{
		"hostname":           check.Hostname,
		"domainname":         check.Domainname,
		"ip":                 check.IP,
		"operasystem":        check.Operasystem,
		"license_result":     check.LicenseResult,
		"minimum_password_age":        check.MinimumPasswordAge,
		"maximum_password_age":        check.MaximumPasswordAge,
		"minimum_password_length":     check.MinimumPasswordLength,
		"password_complexity":         check.PasswordComplexity,
		"password_history_size":       check.PasswordHistorySize,
		"lockout_bad_count":           check.LockoutBadCount,
		"lockout_duration":            check.LockoutDuration,
		"reset_lockout_count":         check.ResetLockoutCount,
		"require_logon_to_change_password": check.RequireLogonToChangePwd,
		"new_administrator_name":        check.NewAdministratorName,
		"new_guest_name":                check.NewGuestName,
		"clear_text_password":           check.ClearTextPassword,
		"lsa_anonymous_name_lookup":     check.LSAAnonymousNameLookup,
		"enable_admin_account":          check.EnableAdminAccount,
		"enable_guest_account":          check.EnableGuestAccount,
		"audit_system_events":           check.AuditSystemEvents,
		"audit_logon_events":            check.AuditLogonEvents,
		"audit_object_access":           check.AuditObjectAccess,
		"audit_privilege_use":           check.AuditPrivilegeUse,
		"audit_policy_change":           check.AuditPolicyChange,
		"audit_account_manage":          check.AuditAccountManage,
		"audit_process_tracking":        check.AuditProcessTracking,
		"audit_ds_access":               check.AuditDSAccess,
		"audit_account_logon":           check.AuditAccountLogon,
		"storage_devices":               check.RemovableStorageDenied,
		"screen_saver_active":           check.ScreenSaverActive,
		"screen_saver_secure":           check.ScreenSaverIsSecure,
		"screen_save_timeout":           check.ScreenSaveTimeOut,
	}
	
	labels := getWindowsFieldLabels()
	for fieldName, value := range fieldMap {
		if value == "" {
			continue
		}
		label := labels[fieldName]
		isNonCompliant := false
		standardValue := ""
		
		if compliance != nil {
			for _, nf := range compliance.NonCompliantFields {
				if nf.Field == fieldName {
					isNonCompliant = true
					standardValue = nf.Standard
					break
				}
			}
		}
		
		valueClass := "detail-value"
		if isNonCompliant {
			valueClass += " non-compliant"
		}
		
		html += fmt.Sprintf("<div class=\"detail-item\"><div class=\"detail-label\">%s</div><div class=\"%s\">%s%s</div></div>", 
			label, valueClass, value,
			func() string {
				if standardValue != "" {
					return fmt.Sprintf("<span class=\"standard-hint\">标准：%s</span>", standardValue)
				}
				return ""
			}())
	}
	
	return html
}

// EncodeBase64 简单 Base64 编码辅助函数
func encodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// extractDomain 从邮箱地址中提取域名
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return "localhost"
}

// 字段标签映射
var linuxFieldLabels = map[string]string{
	"hostname": "计算机名",
	"ip": "IP 地址",
	"operasystem": "操作系统",
	"kernel": "内核版本",
	"dnf_conf_gpgcheck": "dnf.conf_gpgcheck",
	"redhat_repo_gpgcheck": "redhat.repo_gpgcheck",
	"pass_max_days": "PASS_MAX_DAYS",
	"pass_min_days": "PASS_MIN_DAYS",
	"pass_min_len": "PASS_MIN_LEN",
	"pass_warn_age": "PASS_WARN_AGE",
	"inactive": "INACTIVE",
	"gid": "GID (root)",
	"tmout": "TMOUT",
	"cron": "Cron",
	"crontab": "Crontab",
	"cron_hourly": "CronHourly",
	"cron_daily": "CronDaily",
	"cron_weekly": "CronWeekly",
	"cron_monthly": "CronMonthly",
	"cron_deny": "CronDeny",
	"at_deny": "AtDeny",
	"cron_allow": "CronAllow",
	"at_allow": "AtAllow",
	"sshd_config": "sshd_config",
	"log_level": "LogLevel",
	"x11_forwarding": "X11Forwarding",
	"max_auth_tries": "MaxAuthTries",
	"ignore_rhosts": "IgnoreRhosts",
	"hostbased_authentication": "HostbasedAuthentication",
	"permit_root_login": "PermitRootLogin",
	"permit_empty_passwords": "PermitEmptyPasswords",
	"permit_user_environment": "PermitUserEnvironment",
	"client_alive_interval": "ClientAliveInterval",
	"client_alive_count_max": "ClientAliveCountMax",
	"login_grace_time": "LoginGraceTime",
	"minlen": "minlen",
	"minclass": "minclass",
	"dcredit": "dcredit",
	"ucredit": "ucredit",
	"lcredit": "lcredit",
	"ocredit": "ocredit",
	"password_remember": "password_remember",
	"passwd": "passwd",
	"passwd_minus": "passwd_minus",
	"group": "group",
	"group_minus": "group_minus",
	"shadow": "shadow",
	"shadow_minus": "shadow_minus",
	"gshadow": "gshadow",
	"gshadow_minus": "gshadow_minus",
	"crypto_policies": "CryptoPolicies",
	"ntp_server": "NTPServer",
}

func getFieldLabels() map[string]string {
	return linuxFieldLabels
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		default:
			return 0
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Windows 字段标签映射
var windowsFieldLabels = map[string]string{
	"hostname": "计算机名",
	"domainname": "域名",
	"ip": "IP 地址",
	"operasystem": "操作系统",
	"license_result": "激活状态",
	"minimum_password_age": "密码最短使用天数",
	"maximum_password_age": "密码最长使用天数",
	"minimum_password_length": "密码最小长度",
	"password_complexity": "密码复杂度",
	"password_history_size": "密码历史记录数",
	"lockout_bad_count": "账户锁定阈值",
	"lockout_duration": "锁定持续时间 (分钟)",
	"reset_lockout_count": "重置锁定计数 (分钟)",
	"require_logon_to_change_password": "登录更改密码",
	"new_administrator_name": "管理员名称",
	"new_guest_name": "来宾名称",
	"clear_text_password": "明文密码存储",
	"lsa_anonymous_name_lookup": "LSA 匿名查找",
	"enable_admin_account": "启用管理员账户",
	"enable_guest_account": "启用来宾账户",
	"audit_system_events": "系统事件",
	"audit_logon_events": "登录事件",
	"audit_object_access": "对象访问",
	"audit_privilege_use": "特权使用",
	"audit_policy_change": "策略更改",
	"audit_account_manage": "账户管理",
	"audit_process_tracking": "进程跟踪",
	"audit_ds_access": "DS 访问",
	"audit_account_logon": "账户登录",
	"storage_devices": "移动存储设备",
	"screen_saver_active": "屏保启用",
	"screen_saver_secure": "屏保安全",
	"screen_save_timeout": "屏保超时 (秒)",
}

func getWindowsFieldLabels() map[string]string {
	return windowsFieldLabels
}
