package controllers

// FieldGroup 字段分组信息
type FieldGroup struct {
	Group string // 分组名称
}

// FieldGroups 所有字段的分组映射
var FieldGroups = map[string]FieldGroup{
	// 系统更新
	"dnf_conf_gpgcheck":       {Group: "系统更新"},
	"redhat_repo_gpgcheck":    {Group: "系统更新"},
	
	// 用户账户策略
	"pass_max_days":     {Group: "用户账户策略"},
	"pass_min_days":     {Group: "用户账户策略"},
	"pass_min_len":      {Group: "用户账户策略"},
	"pass_warn_age":     {Group: "用户账户策略"},
	"inactive":          {Group: "用户账户策略"},
	"gid":               {Group: "用户账户策略"},
	"tmout":             {Group: "用户账户策略"},
	
	// 计划任务
	"cron":              {Group: "计划任务"},
	"crontab":           {Group: "计划任务"},
	"cron_hourly":       {Group: "计划任务"},
	"cron_daily":        {Group: "计划任务"},
	"cron_weekly":       {Group: "计划任务"},
	"cron_monthly":      {Group: "计划任务"},
	"cron_deny":         {Group: "计划任务"},
	"at_deny":           {Group: "计划任务"},
	"cron_allow":        {Group: "计划任务"},
	"at_allow":          {Group: "计划任务"},
	
	// SSH 配置
	"sshd_config":             {Group: "SSH 配置"},
	"log_level":               {Group: "SSH 配置"},
	"x11_forwarding":          {Group: "SSH 配置"},
	"max_auth_tries":          {Group: "SSH 配置"},
	"ignore_rhosts":           {Group: "SSH 配置"},
	"hostbased_authentication":{Group: "SSH 配置"},
	"permit_root_login":       {Group: "SSH 配置"},
	"permit_empty_passwords":  {Group: "SSH 配置"},
	"permit_user_environment": {Group: "SSH 配置"},
	"client_alive_interval":   {Group: "SSH 配置"},
	"client_alive_count_max":  {Group: "SSH 配置"},
	"login_grace_time":        {Group: "SSH 配置"},
	
	// 密码策略
	"minlen":            {Group: "密码策略"},
	"minclass":          {Group: "密码策略"},
	"dcredit":           {Group: "密码策略"},
	"ucredit":           {Group: "密码策略"},
	"lcredit":           {Group: "密码策略"},
	"ocredit":           {Group: "密码策略"},
	"password_remember": {Group: "密码策略"},
	
	// 文件权限
	"passwd":          {Group: "文件权限"},
	"passwd_minus":    {Group: "文件权限"},
	"group":           {Group: "文件权限"},
	"group_minus":     {Group: "文件权限"},
	"shadow":          {Group: "文件权限"},
	"shadow_minus":    {Group: "文件权限"},
	"gshadow":         {Group: "文件权限"},
	"gshadow_minus":   {Group: "文件权限"},
	
	// 加密与时钟
	"crypto_policies": {Group: "加密与时钟"},
	"ntp_server":      {Group: "加密与时钟"},
}
