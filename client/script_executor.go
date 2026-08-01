package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// SystemCheckData 加固检查结果数据结构
type SystemCheckData struct {
	Date                      string `json:"date"`
	Hostname                  string `json:"hostname"`
	Operasystem               string `json:"operasystem"`
	Kernel                    string `json:"kernel"`
	IP                        string `json:"ip"`
	DnfConfGpgcheck         string `json:"dnf_conf_gpgcheck"`
	RedhatRepoGpgcheck      string `json:"redhat_repo_gpgcheck"`
	PassMaxDays               string `json:"pass_max_days"`
	PassMinDays               string `json:"pass_min_days"`
	PassMinLen                string `json:"pass_min_len"`
	PassWarnAge               string `json:"pass_warn_age"`
	Inactive                  string `json:"inactive"`
	GID                       string `json:"gid"`
	Tmout                     string `json:"tmout"`
	Cron                      string `json:"cron"`
	Crontab                   string `json:"crontab"`
	CronHourly                string `json:"cron_hourly"`
	CronDaily                 string `json:"cron_daily"`
	CronWeekly                string `json:"cron_weekly"`
	CronMonthly               string `json:"cron_monthly"`
	CronDeny                  string `json:"cron_deny"`
	AtDeny                    string `json:"at_deny"`
	CronAllow                 string `json:"cron_allow"`
	AtAllow                   string `json:"at_allow"`
	SshdConfig                string `json:"sshd_config"`
	LogLevel                  string `json:"log_level"`
	X11Forwarding             string `json:"x11_forwarding"`
	MaxAuthTries              string `json:"max_auth_tries"`
	IgnoreRhosts              string `json:"ignore_rhosts"`
	HostbasedAuthentication   string `json:"hostbased_authentication"`
	PermitRootLogin           string `json:"permit_root_login"`
	PermitEmptyPasswords      string `json:"permit_empty_passwords"`
	PermitUserEnvironment     string `json:"permit_user_environment"`
	ClientAliveInterval       string `json:"client_alive_interval"`
	ClientAliveCountMax       string `json:"client_alive_count_max"`
	LoginGraceTime            string `json:"login_grace_time"`
	Minlen                    string `json:"minlen"`
	Minclass                  string `json:"minclass"`
	Dcredit                   string `json:"dcredit"`
	Ucredit                   string `json:"ucredit"`
	Lcredit                   string `json:"lcredit"`
	Ocredit                   string `json:"ocredit"`
	PasswordRemember          string `json:"password_remember"`
	Passwd                    string `json:"passwd"`
	PasswdMinus               string `json:"passwd_minus"`
	Group                     string `json:"group"`
	GroupMinus                string `json:"group_minus"`
	Shadow                    string `json:"shadow"`
	ShadowMinus               string `json:"shadow_minus"`
	Gshadow                   string `json:"gshadow"`
	GshadowMinus              string `json:"gshadow_minus"`
	CryptoPolicies            string `json:"crypto_policies"`
	NtpServer                 string `json:"ntp_server"`
}

// executeScript 执行 Shell 脚本
func executeScript() (string, error) {
	scriptPath := config.ScriptPath
	
	cmd := exec.Command("/bin/bash", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	if err != nil {
		return "", fmt.Errorf("script execution failed: %w, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// parseOutput 解析脚本输出为 Struct
func parseOutput(output string) *SystemCheckData {
	checkData := &SystemCheckData{}
	
	// 正则表达式匹配两种格式:
	// 1. valuesXX=值 (旧格式)
	// 2. field_name=value (新格式，更具可读性)
	reValueFormat := regexp.MustCompile(`^values(\d+)=([\S\s]+)$`)
	reFieldFormat := regexp.MustCompile(`^(\w+)=(.*)$`)
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 尝试匹配新的字段名称格式 (preferred)
		if matches := reFieldFormat.FindStringSubmatch(line); len(matches) == 3 {
			fieldName := strings.ToLower(matches[1])
			value := strings.TrimSpace(matches[2])
			
			switch fieldName {
			case "day":
				checkData.Date = value
			case "name":
				checkData.Hostname = value
			case "centos":
				checkData.Operasystem = value
			case "kernel":
				checkData.Kernel = value
			case "local_ip":
				checkData.IP = value
			case "dnf_conf_gpgcheck":
				checkData.DnfConfGpgcheck = value
			case "redhat_repo_gpgcheck":
				checkData.RedhatRepoGpgcheck = value
			case "pass_max_days":
				checkData.PassMaxDays = value
			case "pass_min_days":
				checkData.PassMinDays = value
			case "pass_min_len":
				checkData.PassMinLen = value
			case "pass_warn_age":
				checkData.PassWarnAge = value
			case "inactive":
				checkData.Inactive = value
			case "gid":
				checkData.GID = value
			case "tmout":
				checkData.Tmout = value
			case "cron":
				checkData.Cron = value
			case "crontab":
				checkData.Crontab = value
			case "cron_hourly":
				checkData.CronHourly = value
			case "cron_daily":
				checkData.CronDaily = value
			case "cron_weekly":
				checkData.CronWeekly = value
			case "cron_monthly":
				checkData.CronMonthly = value
			case "cron_deny":
				checkData.CronDeny = value
			case "at_deny":
				checkData.AtDeny = value
			case "cron_allow":
				checkData.CronAllow = value
			case "at_allow":
				checkData.AtAllow = value
			case "sshd_config":
				checkData.SshdConfig = value
			case "log_level":
				checkData.LogLevel = value
			case "x11_forwarding":
				checkData.X11Forwarding = value
			case "max_auth_tries":
				checkData.MaxAuthTries = value
			case "ignore_rhosts":
				checkData.IgnoreRhosts = value
			case "hostbased_authentication":
				checkData.HostbasedAuthentication = value
			case "permit_root_login":
				checkData.PermitRootLogin = value
			case "permit_empty_passwords":
				checkData.PermitEmptyPasswords = value
			case "permit_user_environment":
				checkData.PermitUserEnvironment = value
			case "client_alive_interval":
				checkData.ClientAliveInterval = value
			case "client_alive_count_max":
				checkData.ClientAliveCountMax = value
			case "login_grace_time":
				checkData.LoginGraceTime = value
			case "minlen":
				checkData.Minlen = value
			case "minclass":
				checkData.Minclass = value
			case "dcredit":
				checkData.Dcredit = value
			case "ucredit":
				checkData.Ucredit = value
			case "lcredit":
				checkData.Lcredit = value
			case "ocredit":
				checkData.Ocredit = value
			case "password_remember":
				checkData.PasswordRemember = value
			case "passwd":
				checkData.Passwd = value
			case "passwd_minus":
				checkData.PasswdMinus = value
			case "group":
				checkData.Group = value
			case "group_minus":
				checkData.GroupMinus = value
			case "shadow":
				checkData.Shadow = value
			case "shadow_minus":
				checkData.ShadowMinus = value
			case "gshadow":
				checkData.Gshadow = value
			case "gshadow_minus":
				checkData.GshadowMinus = value
			case "crypto_policies":
				checkData.CryptoPolicies = value
			case "ntp_server":
				checkData.NtpServer = value
			}
		} else if matches := reValueFormat.FindStringSubmatch(line); len(matches) == 3 {
			// 兼容旧的 valuesXX 格式
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			
			switch key {
			case "5":
				checkData.DnfConfGpgcheck = value
			case "6":
				checkData.RedhatRepoGpgcheck = value
			case "7":
				checkData.PassMaxDays = value
			case "8":
				checkData.PassMinDays = value
			case "9":
				checkData.PassMinLen = value
			case "10":
				checkData.PassWarnAge = value
			case "11":
				checkData.Inactive = value
			case "12":
				checkData.GID = value
			case "13":
				checkData.Tmout = value
			case "14":
				checkData.Cron = value
			case "15":
				checkData.Crontab = value
			case "16":
				checkData.CronHourly = value
			case "17":
				checkData.CronDaily = value
			case "18":
				checkData.CronWeekly = value
			case "19":
				checkData.CronMonthly = value
			case "20":
				checkData.CronDeny = value
			case "21":
				checkData.AtDeny = value
			case "22":
				checkData.CronAllow = value
			case "23":
				checkData.AtAllow = value
			case "24":
				checkData.SshdConfig = value
			case "25":
				checkData.LogLevel = value
			case "26":
				checkData.X11Forwarding = value
			case "27":
				checkData.MaxAuthTries = value
			case "28":
				checkData.IgnoreRhosts = value
			case "29":
				checkData.HostbasedAuthentication = value
			case "30":
				checkData.PermitRootLogin = value
			case "31":
				checkData.PermitEmptyPasswords = value
			case "32":
				checkData.PermitUserEnvironment = value
			case "33":
				checkData.ClientAliveInterval = value
			case "34":
				checkData.ClientAliveCountMax = value
			case "35":
				checkData.LoginGraceTime = value
			case "36":
				checkData.Minlen = value
			case "37":
				checkData.Minclass = value
			case "38":
				checkData.Dcredit = value
			case "39":
				checkData.Ucredit = value
			case "40":
				checkData.Lcredit = value
			case "41":
				checkData.Ocredit = value
			case "42":
				checkData.PasswordRemember = value
			case "43":
				checkData.Passwd = value
			case "44":
				checkData.PasswdMinus = value
			case "45":
				checkData.Group = value
			case "46":
				checkData.GroupMinus = value
			case "47":
				checkData.Shadow = value
			case "48":
				checkData.ShadowMinus = value
			case "49":
				checkData.Gshadow = value
			case "50":
				checkData.GshadowMinus = value
			case "51":
				checkData.CryptoPolicies = value
			case "52":
				checkData.NtpServer = value
			}
		}
	}
	
	return checkData
}
