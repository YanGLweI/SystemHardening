package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemCheck Linux 系统检查记录
type SystemCheck struct {
	ID                    uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientUUID            string    `gorm:"size:64;index:idx_systemcheck_client_uuid" json:"client_uuid"`              // 客户端 UUID - 无外键约束
	Date                  string    `gorm:"column:date;size:50" json:"date"`              // 检查时间
	Hostname              string    `gorm:"column:hostname;size:100" json:"hostname"`     // 计算机名
	Operasystem           string    `gorm:"column:operasystem;size:200" json:"operasystem"`  // 操作系统版本
	Kernel                string    `gorm:"column:kernel;size:100" json:"kernel"`         // 内核版本
	IP                    string    `gorm:"column:ip;size:50" json:"ip"`                  // IP 地址
	DnfConfGpgcheck       string    `gorm:"column:dnf_conf_gpgcheck;size:50" json:"dnf_conf_gpgcheck"`      // dnf.conf 中的 gpgcheck
	RedhatRepoGpgcheck    string    `gorm:"column:redhat_repo_gpgcheck;size:50" json:"redhat_repo_gpgcheck"` // redhat.repo 中的 gpgcheck
	PassMaxDays           string    `gorm:"column:pass_max_days;size:50" json:"pass_max_days"`
	PassMinDays           string    `gorm:"column:pass_min_days;size:50" json:"pass_min_days"`
	PassMinLen            string    `gorm:"column:pass_min_len;size:50" json:"pass_min_len"`
	PassWarnAge           string    `gorm:"column:pass_warn_age;size:50" json:"pass_warn_age"`
	Inactive              string    `gorm:"column:inactive;size:50" json:"inactive"`
	GID                   string    `gorm:"column:gid;size:50" json:"gid"`
	Tmout                 string    `gorm:"column:tmout;size:50" json:"tmout"`
	Cron                  string    `gorm:"column:cron;size:50" json:"cron"`
	Crontab               string    `gorm:"column:crontab;size:200" json:"crontab"`
	CronHourly            string    `gorm:"column:cron_hourly;size:200" json:"cron_hourly"`
	CronDaily             string    `gorm:"column:cron_daily;size:200" json:"cron_daily"`
	CronWeekly            string    `gorm:"column:cron_weekly;size:200" json:"cron_weekly"`
	CronMonthly           string    `gorm:"column:cron_monthly;size:200" json:"cron_monthly"`
	CronDeny              string    `gorm:"column:cron_deny;size:200" json:"cron_deny"`
	AtDeny                string    `gorm:"column:at_deny;size:200" json:"at_deny"`
	CronAllow             string    `gorm:"column:cron_allow;size:200" json:"cron_allow"`
	AtAllow               string    `gorm:"column:at_allow;size:200" json:"at_allow"`
	SshdConfig            string    `gorm:"column:sshd_config;size:200" json:"sshd_config"`
	LogLevel              string    `gorm:"column:log_level;size:50" json:"log_level"`
	X11Forwarding         string    `gorm:"column:x11_forwarding;size:50" json:"x11_forwarding"`
	MaxAuthTries          string    `gorm:"column:max_auth_tries;size:50" json:"max_auth_tries"`
	IgnoreRhosts          string    `gorm:"column:ignore_rhosts;size:50" json:"ignore_rhosts"`
	HostbasedAuthentication string    `gorm:"column:hostbased_authentication;size:50" json:"hostbased_authentication"`
	PermitRootLogin       string    `gorm:"column:permit_root_login;size:50" json:"permit_root_login"`
	PermitEmptyPasswords  string    `gorm:"column:permit_empty_passwords;size:50" json:"permit_empty_passwords"`
	PermitUserEnvironment string    `gorm:"column:permit_user_environment;size:50" json:"permit_user_environment"`
	ClientAliveInterval   string    `gorm:"column:client_alive_interval;size:50" json:"client_alive_interval"`
	ClientAliveCountMax   string    `gorm:"column:client_alive_count_max;size:50" json:"client_alive_count_max"`
	LoginGraceTime        string    `gorm:"column:login_grace_time;size:50" json:"login_grace_time"`
	Minlen                string    `gorm:"column:minlen;size:50" json:"minlen"`
	Minclass              string    `gorm:"column:minclass;size:50" json:"minclass"`
	Dcredit               string    `gorm:"column:dcredit;size:50" json:"dcredit"`
	Ucredit               string    `gorm:"column:ucredit;size:50" json:"ucredit"`
	Lcredit               string    `gorm:"column:lcredit;size:50" json:"lcredit"`
	Ocredit               string    `gorm:"column:ocredit;size:50" json:"ocredit"`
	PasswordRemember      string    `gorm:"column:password_remember;size:50" json:"password_remember"`
	Passwd                string    `gorm:"column:passwd;size:200" json:"passwd"`
	PasswdMinus           string    `gorm:"column:passwd_minus;size:200" json:"passwd_minus"`
	GroupCol              string    `gorm:"column:group;size:200" json:"group"`
	GroupMinus            string    `gorm:"column:group_minus;size:200" json:"group_minus"`
	Shadow                string    `gorm:"column:shadow;size:200" json:"shadow"`
	ShadowMinus           string    `gorm:"column:shadow_minus;size:200" json:"shadow_minus"`
	Gshadow               string    `gorm:"column:gshadow;size:200" json:"gshadow"`
	GshadowMinus          string    `gorm:"column:gshadow_minus;size:200" json:"gshadow_minus"`
	CryptoPolicies        string    `gorm:"column:crypto_policies;size:100" json:"crypto_policies"`
	NtpServer             string    `gorm:"column:ntp_server;size:200" json:"ntp_server"`
	DeletedAt             gorm.DeletedAt  `gorm:"index" json:"-"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

// TableName specify table name
func (SystemCheck) TableName() string {
	return "systemcheck"
}
