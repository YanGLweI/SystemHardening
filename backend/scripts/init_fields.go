package scripts

import (
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

// InitFieldsAndGroups 初始化字段定义和分组数据
func InitFieldsAndGroups() {
	db := database.DB
	
	var count int64
	db.Model(&models.LinuxField{}).Count(&count)
	if count > 0 {
		return // 已初始化，跳过
	}
	
	// 所有 Linux 加固字段的定义
	fields := []models.LinuxField{
		// 系统更新 (2 个)
		{FieldName: "dnf_conf_gpgcheck", FieldLabel: "dnf.conf_gpgcheck", FieldGroup: "系统更新", Category: "system-update", SortOrder: 1, DataType: "string", Description: "dnf.conf 中的 gpgcheck 设置"},
		{FieldName: "redhat_repo_gpgcheck", FieldLabel: "redhat.repo_gpgcheck", FieldGroup: "系统更新", Category: "system-update", SortOrder: 2, DataType: "string", Description: "redhat.repo 中的 gpgcheck 设置"},
		
		// 用户账户策略 (6 个)
		{FieldName: "pass_max_days", FieldLabel: "PASS_MAX_DAYS", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 3, DataType: "string", Description: "密码最长有效天数"},
		{FieldName: "pass_min_days", FieldLabel: "PASS_MIN_DAYS", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 4, DataType: "string", Description: "密码最短有效天数"},
		{FieldName: "pass_min_len", FieldLabel: "PASS_MIN_LEN", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 5, DataType: "string", Description: "密码最小长度"},
		{FieldName: "pass_warn_age", FieldLabel: "PASS_WARN_AGE", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 6, DataType: "string", Description: "密码过期前警告天数"},
		{FieldName: "inactive", FieldLabel: "INACTIVE", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 7, DataType: "string", Description: "账户失效前的宽限期"},
		{FieldName: "gid", FieldLabel: "GID", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 8, DataType: "string", Description: "root 用户的 GID"},
		{FieldName: "tmout", FieldLabel: "TMOUT", FieldGroup: "用户账户策略", Category: "user-policy", SortOrder: 9, DataType: "string", Description: "Shell 超时时间"},
		
		// 计划任务 (10 个)
		{FieldName: "cron", FieldLabel: "Cron", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 10, DataType: "string", Description: "Cron 守护进程状态"},
		{FieldName: "crontab", FieldLabel: "Crontab", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 11, DataType: "string", Description: "crontab 文件权限"},
		{FieldName: "cron_hourly", FieldLabel: "CronHourly", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 12, DataType: "string", Description: "cron.hourly 目录权限"},
		{FieldName: "cron_daily", FieldLabel: "CronDaily", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 13, DataType: "string", Description: "cron.daily 目录权限"},
		{FieldName: "cron_weekly", FieldLabel: "CronWeekly", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 14, DataType: "string", Description: "cron.weekly 目录权限"},
		{FieldName: "cron_monthly", FieldLabel: "CronMonthly", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 15, DataType: "string", Description: "cron.monthly 目录权限"},
		{FieldName: "cron_deny", FieldLabel: "CronDeny", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 16, DataType: "string", Description: "cron.deny 文件内容"},
		{FieldName: "at_deny", FieldLabel: "AtDeny", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 17, DataType: "string", Description: "at.deny 文件内容"},
		{FieldName: "cron_allow", FieldLabel: "CronAllow", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 18, DataType: "string", Description: "cron.allow 文件内容"},
		{FieldName: "at_allow", FieldLabel: "AtAllow", FieldGroup: "计划任务", Category: "cron-config", SortOrder: 19, DataType: "string", Description: "at.allow 文件内容"},
		
		// SSH 配置 (12 个)
		{FieldName: "sshd_config", FieldLabel: "sshd_config", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 20, DataType: "string", Description: "sshd_config 文件权限"},
		{FieldName: "log_level", FieldLabel: "LogLevel", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 21, DataType: "string", Description: "SSH 日志级别"},
		{FieldName: "x11_forwarding", FieldLabel: "X11Forwarding", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 22, DataType: "string", Description: "X11 转发设置"},
		{FieldName: "max_auth_tries", FieldLabel: "MaxAuthTries", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 23, DataType: "string", Description: "最大认证尝试次数"},
		{FieldName: "ignore_rhosts", FieldLabel: "IgnoreRhosts", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 24, DataType: "string", Description: "忽略 rhosts 设置"},
		{FieldName: "hostbased_authentication", FieldLabel: "HostbasedAuthentication", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 25, DataType: "string", Description: "基于主机的认证设置"},
		{FieldName: "permit_root_login", FieldLabel: "PermitRootLogin", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 26, DataType: "string", Description: "允许 root 登录"},
		{FieldName: "permit_empty_passwords", FieldLabel: "PermitEmptyPasswords", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 27, DataType: "string", Description: "允许空密码"},
		{FieldName: "permit_user_environment", FieldLabel: "PermitUserEnvironment", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 28, DataType: "string", Description: "允许用户环境变量"},
		{FieldName: "client_alive_interval", FieldLabel: "ClientAliveInterval", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 29, DataType: "string", Description: "客户端 alive 间隔"},
		{FieldName: "client_alive_count_max", FieldLabel: "ClientAliveCountMax", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 30, DataType: "string", Description: "客户端 alive 最大计数"},
		{FieldName: "login_grace_time", FieldLabel: "LoginGraceTime", FieldGroup: "SSH 配置", Category: "ssh-config", SortOrder: 31, DataType: "string", Description: "登录宽限时间"},
		
		// 密码策略 (7 个)
		{FieldName: "minlen", FieldLabel: "minlen", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 32, DataType: "string", Description: "密码最小长度"},
		{FieldName: "minclass", FieldLabel: "minclass", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 33, DataType: "string", Description: "密码最小字符类数量"},
		{FieldName: "dcredit", FieldLabel: "dcredit", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 34, DataType: "string", Description: "数字信用度"},
		{FieldName: "ucredit", FieldLabel: "ucredit", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 35, DataType: "string", Description: "小写字母信用度"},
		{FieldName: "lcredit", FieldLabel: "lcredit", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 36, DataType: "string", Description: "大写字母信用度"},
		{FieldName: "ocredit", FieldLabel: "ocredit", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 37, DataType: "string", Description: "特殊字符信用度"},
		{FieldName: "password_remember", FieldLabel: "password_remember", FieldGroup: "密码策略", Category: "password-policy", SortOrder: 38, DataType: "string", Description: "密码历史记住数量"},
		
		// 文件权限 (8 个)
		{FieldName: "passwd", FieldLabel: "passwd", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 39, DataType: "string", Description: "passwd 文件权限"},
		{FieldName: "passwd_minus", FieldLabel: "passwd_minus", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 40, DataType: "string", Description: "passwd- 文件权限"},
		{FieldName: "group", FieldLabel: "group", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 41, DataType: "string", Description: "group 文件权限"},
		{FieldName: "group_minus", FieldLabel: "group_minus", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 42, DataType: "string", Description: "group- 文件权限"},
		{FieldName: "shadow", FieldLabel: "shadow", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 43, DataType: "string", Description: "shadow 文件权限"},
		{FieldName: "shadow_minus", FieldLabel: "shadow_minus", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 44, DataType: "string", Description: "shadow- 文件权限"},
		{FieldName: "gshadow", FieldLabel: "gshadow", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 45, DataType: "string", Description: "gshadow 文件权限"},
		{FieldName: "gshadow_minus", FieldLabel: "gshadow_minus", FieldGroup: "文件权限", Category: "file-permission", SortOrder: 46, DataType: "string", Description: "gshadow- 文件权限"},
		
		// 加密与时钟 (2 个)
		{FieldName: "crypto_policies", FieldLabel: "CryptoPolicies", FieldGroup: "加密与时钟", Category: "crypto-sync", SortOrder: 47, DataType: "string", Description: "加密策略"},
		{FieldName: "ntp_server", FieldLabel: "NTPServer", FieldGroup: "加密与时钟", Category: "crypto-sync", SortOrder: 48, DataType: "string", Description: "NTP 服务器地址"},
	}
	
	// 批量插入所有字段
	err := db.Create(&fields).Error
	if err != nil {
		panic("Failed to insert fields: " + err.Error())
	}
	
	println("✅ Successfully initialized " + string(rune(len(fields)+'0')) + " fields into linux_fields table")
}
