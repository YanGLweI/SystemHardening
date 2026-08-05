package models

import (
	"regexp"
)

// NonCompliantField 不合规字段信息
type NonCompliantField struct {
	Field    string `json:"field"`    // 字段名
	Label    string `json:"label"`    // 字段标签
	Actual   string `json:"actual"`   // 实际值
	Standard string `json:"standard"` // 标准值
}

// ComplianceResult 合规比对结果
type ComplianceResult struct {
	Status             string              `json:"status"` // "compliant" / "non_compliant"
	NonCompliantFields []NonCompliantField `json:"non_compliant_fields"`
}

// CompareCompliance 比对 Linux 系统检查记录与标准配置
func CompareCompliance(check *SystemCheck, standardMap map[string]string) *ComplianceResult {
	result := &ComplianceResult{
		Status:             "compliant",
		NonCompliantFields: []NonCompliantField{},
	}

	// 定义字段名到实际值的映射
	fieldValues := map[string]string{
		"dnf_conf_gpgcheck":        check.DnfConfGpgcheck,
		"redhat_repo_gpgcheck":     check.RedhatRepoGpgcheck,
		"pass_max_days":            check.PassMaxDays,
		"pass_min_days":            check.PassMinDays,
		"pass_min_len":             check.PassMinLen,
		"pass_warn_age":            check.PassWarnAge,
		"inactive":                 check.Inactive,
		"gid":                      check.GID,
		"tmout":                    check.Tmout,
		"cron":                     check.Cron,
		"crontab":                  check.Crontab,
		"cron_hourly":              check.CronHourly,
		"cron_daily":               check.CronDaily,
		"cron_weekly":              check.CronWeekly,
		"cron_monthly":             check.CronMonthly,
		"cron_deny":                check.CronDeny,
		"at_deny":                  check.AtDeny,
		"cron_allow":               check.CronAllow,
		"at_allow":                 check.AtAllow,
		"sshd_config":              check.SshdConfig,
		"log_level":                check.LogLevel,
		"x11_forwarding":           check.X11Forwarding,
		"max_auth_tries":           check.MaxAuthTries,
		"ignore_rhosts":            check.IgnoreRhosts,
		"hostbased_authentication": check.HostbasedAuthentication,
		"permit_root_login":        check.PermitRootLogin,
		"permit_empty_passwords":   check.PermitEmptyPasswords,
		"permit_user_environment":  check.PermitUserEnvironment,
		"client_alive_interval":    check.ClientAliveInterval,
		"client_alive_count_max":   check.ClientAliveCountMax,
		"login_grace_time":         check.LoginGraceTime,
		"minlen":                   check.Minlen,
		"minclass":                 check.Minclass,
		"dcredit":                  check.Dcredit,
		"ucredit":                  check.Ucredit,
		"lcredit":                  check.Lcredit,
		"ocredit":                  check.Ocredit,
		"password_remember":        check.PasswordRemember,
		"passwd":                   check.Passwd,
		"passwd_minus":             check.PasswdMinus,
		"group":                    check.GroupCol,
		"group_minus":              check.GroupMinus,
		"shadow":                   check.Shadow,
		"shadow_minus":             check.ShadowMinus,
		"gshadow":                  check.Gshadow,
		"gshadow_minus":            check.GshadowMinus,
		"crypto_policies":          check.CryptoPolicies,
		"ntp_server":               check.NtpServer,
	}

	// 遍历所有字段的实际值，而不是标准值
	// 这样可以确保所有有实际值的字段都会被检查
	for fieldName, actualValue := range fieldValues {
		if actualValue == "" {
			// 空值也视为不合规，如果有标准值的话
			if standardValue, ok := standardMap[fieldName]; ok && standardValue != "" {
				result.Status = "non_compliant"
				result.NonCompliantFields = append(result.NonCompliantFields, NonCompliantField{
					Field:    fieldName,
					Label:    GetLinuxFieldLabel(fieldName),
					Actual:   "(empty)",
					Standard: standardValue,
				})
			}
			continue
		}

		if standardValue, ok := standardMap[fieldName]; ok && standardValue != "" {
			if !matchStandard(actualValue, standardValue) {
				result.Status = "non_compliant"
				result.NonCompliantFields = append(result.NonCompliantFields, NonCompliantField{
					Field:    fieldName,
					Label:    GetLinuxFieldLabel(fieldName),
					Actual:   actualValue,
					Standard: standardValue,
				})
			}
		}
	}

	return result
}

// CompareWindowsCompliance 比对 Windows 系统检查记录与标准配置
func CompareWindowsCompliance(check *WindowsSystemCheck, standardMap map[string]string) *ComplianceResult {
	result := &ComplianceResult{
		Status:             "compliant",
		NonCompliantFields: []NonCompliantField{},
	}

	// 定义字段名到实际值的映射
	fieldValues := map[string]string{
		// 基本设置
		"license_result": check.LicenseResult,

		// 账户密码策略
		"minimum_password_age":             check.MinimumPasswordAge,
		"maximum_password_age":             check.MaximumPasswordAge,
		"minimum_password_length":          check.MinimumPasswordLength,
		"password_complexity":              check.PasswordComplexity,
		"password_history_size":            check.PasswordHistorySize,
		"lockout_bad_count":                check.LockoutBadCount,
		"lockout_duration":                 check.LockoutDuration,
		"reset_lockout_count":              check.ResetLockoutCount,
		"require_logon_to_change_password": check.RequireLogonToChangePwd,
		"new_administrator_name":           check.NewAdministratorName,
		"new_guest_name":                   check.NewGuestName,
		"clear_text_password":              check.ClearTextPassword,
		"lsa_anonymous_name_lookup":        check.LSAAnonymousNameLookup,
		"enable_admin_account":             check.EnableAdminAccount,
		"enable_guest_account":             check.EnableGuestAccount,

		// 审计策略
		"audit_system_events":    check.AuditSystemEvents,
		"audit_logon_events":     check.AuditLogonEvents,
		"audit_object_access":    check.AuditObjectAccess,
		"audit_privilege_use":    check.AuditPrivilegeUse,
		"audit_policy_change":    check.AuditPolicyChange,
		"audit_account_manage":   check.AuditAccountManage,
		"audit_process_tracking": check.AuditProcessTracking,
		"audit_ds_access":        check.AuditDSAccess,
		"audit_account_logon":    check.AuditAccountLogon,

		// 设备控制与屏幕保护
		"storage_devices":     check.RemovableStorageDenied,
		"screen_saver_active": check.ScreenSaverActive,
		"screen_saver_secure": check.ScreenSaverIsSecure,
		"screen_save_timeout": check.ScreenSaveTimeOut,
	}

	// 遍历所有字段的实际值
	for fieldName, actualValue := range fieldValues {
		if actualValue == "" {
			// 空值也视为不合规，如果有标准值的话
			if standardValue, ok := standardMap[fieldName]; ok && standardValue != "" {
				result.Status = "non_compliant"
				result.NonCompliantFields = append(result.NonCompliantFields, NonCompliantField{
					Field:    fieldName,
					Label:    GetWindowsFieldLabel(fieldName),
					Actual:   "(empty)",
					Standard: standardValue,
				})
			}
			continue
		}

		if standardValue, ok := standardMap[fieldName]; ok && standardValue != "" {
			if !matchStandard(actualValue, standardValue) {
				result.Status = "non_compliant"
				result.NonCompliantFields = append(result.NonCompliantFields, NonCompliantField{
					Field:    fieldName,
					Label:    GetWindowsFieldLabel(fieldName),
					Actual:   actualValue,
					Standard: standardValue,
				})
			}
		}
	}

	return result
}

// matchStandard 判断实际值是否符合标准值（支持正则表达式）
// 当标准值以 / 开头和结尾时，视为正则表达式进行匹配；否则使用精确匹配
func matchStandard(actual, standard string) bool {
	if len(standard) >= 2 && standard[0] == '/' && standard[len(standard)-1] == '/' {
		pattern := standard[1 : len(standard)-1]
		matched, err := regexp.MatchString(pattern, actual)
		if err != nil {
			// 正则表达式无效时，降级为精确匹配
			return actual == standard
		}
		return matched
	}
	return actual == standard
}

// GetLinuxFieldLabel 获取 Linux 字段的显示标签
func GetLinuxFieldLabel(fieldName string) string {
	labels := map[string]string{
		"dnf_conf_gpgcheck":        "dnf.conf_gpgcheck",
		"redhat_repo_gpgcheck":     "redhat.repo_gpgcheck",
		"pass_max_days":            "PASS_MAX_DAYS",
		"pass_min_days":            "PASS_MIN_DAYS",
		"pass_min_len":             "PASS_MIN_LEN",
		"pass_warn_age":            "PASS_WARN_AGE",
		"inactive":                 "INACTIVE",
		"gid":                      "GID",
		"tmout":                    "TMOUT",
		"cron":                     "Cron",
		"crontab":                  "Crontab",
		"cron_hourly":              "CronHourly",
		"cron_daily":               "CronDaily",
		"cron_weekly":              "CronWeekly",
		"cron_monthly":             "CronMonthly",
		"cron_deny":                "CronDeny",
		"at_deny":                  "AtDeny",
		"cron_allow":               "CronAllow",
		"at_allow":                 "AtAllow",
		"sshd_config":              "sshd_config",
		"log_level":                "LogLevel",
		"x11_forwarding":           "X11Forwarding",
		"max_auth_tries":           "MaxAuthTries",
		"ignore_rhosts":            "IgnoreRhosts",
		"hostbased_authentication": "HostbasedAuthentication",
		"permit_root_login":        "PermitRootLogin",
		"permit_empty_passwords":   "PermitEmptyPasswords",
		"permit_user_environment":  "PermitUserEnvironment",
		"client_alive_interval":    "ClientAliveInterval",
		"client_alive_count_max":   "ClientAliveCountMax",
		"login_grace_time":         "LoginGraceTime",
		"minlen":                   "minlen",
		"minclass":                 "minclass",
		"dcredit":                  "dcredit",
		"ucredit":                  "ucredit",
		"lcredit":                  "lcredit",
		"ocredit":                  "ocredit",
		"password_remember":        "password_remember",
		"passwd":                   "passwd",
		"passwd_minus":             "passwd_minus",
		"group":                    "group",
		"group_minus":              "group_minus",
		"shadow":                   "shadow",
		"shadow_minus":             "shadow_minus",
		"gshadow":                  "gshadow",
		"gshadow_minus":            "gshadow_minus",
		"crypto_policies":          "CryptoPolicies",
		"ntp_server":               "NTPServer",
	}

	if label, ok := labels[fieldName]; ok {
		return label
	}
	return fieldName
}

// GetWindowsFieldLabel 获取 Windows 字段的显示标签
func GetWindowsFieldLabel(fieldName string) string {
	labels := map[string]string{
		// 基本设置
		"license_result": "激活状态",

		// 账户密码策略
		"minimum_password_age":             "密码最短使用天数",
		"maximum_password_age":             "密码最长使用天数",
		"minimum_password_length":          "密码最小长度",
		"password_complexity":              "密码复杂度",
		"password_history_size":            "密码历史记录数",
		"lockout_bad_count":                "账户锁定阈值",
		"lockout_duration":                 "锁定持续时间(分钟)",
		"reset_lockout_count":              "重置锁定计数(分钟)",
		"require_logon_to_change_password": "登录更改密码",
		"new_administrator_name":           "管理员名称",
		"new_guest_name":                   "来宾名称",
		"clear_text_password":              "明文密码存储",
		"lsa_anonymous_name_lookup":        "LSA 匿名查找",
		"enable_admin_account":             "启用管理员账户",
		"enable_guest_account":             "启用来宾账户",

		// 审计策略
		"audit_system_events":    "系统事件",
		"audit_logon_events":     "登录事件",
		"audit_object_access":    "对象访问",
		"audit_privilege_use":    "特权使用",
		"audit_policy_change":    "策略更改",
		"audit_account_manage":   "账户管理",
		"audit_process_tracking": "进程跟踪",
		"audit_ds_access":        "DS 访问",
		"audit_account_logon":    "账户登录",

		// 设备控制与屏幕保护
		"storage_devices":     "移动存储设备",
		"screen_saver_active": "屏保启用",
		"screen_saver_secure": "屏保安全",
		"screen_save_timeout": "屏保超时(秒)",
	}

	if label, ok := labels[fieldName]; ok {
		return label
	}
	return fieldName
}
