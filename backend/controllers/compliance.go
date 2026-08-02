package controllers

import (
	"github.com/yeung/system-hardening/backend/models"
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

// CompareCompliance 比对系统检查记录与标准配置
func CompareCompliance(check *models.SystemCheck, standardMap map[string]string) *ComplianceResult {
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

	// 遍历标准配置，只比对已配置的标准值（非空）
	for fieldName, standardValue := range standardMap {
		if standardValue == "" {
			continue // 跳过未配置标准值的字段
		}

		if actualValue, ok := fieldValues[fieldName]; ok {
			if actualValue != standardValue {
				result.Status = "non_compliant"
				result.NonCompliantFields = append(result.NonCompliantFields, NonCompliantField{
					Field:    fieldName,
					Label:    getFieldLabel(fieldName),
					Actual:   actualValue,
					Standard: standardValue,
				})
			}
		}
	}

	return result
}

// getFieldLabel 获取字段的显示标签
func getFieldLabel(fieldName string) string {
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
