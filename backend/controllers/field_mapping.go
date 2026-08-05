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

// WindowsFieldGroups Windows 字段分组映射（对应 windows_fields 表）
var WindowsFieldGroups = map[string]string{
	// 账户密码策略
	"minimum_password_age":             "账户密码策略",
	"maximum_password_age":             "账户密码策略",
	"minimum_password_length":          "账户密码策略",
	"password_complexity":              "账户密码策略",
	"password_history_size":            "账户密码策略",
	"lockout_bad_count":                "账户密码策略",
	"lockout_duration":                 "账户密码策略",
	"reset_lockout_count":              "账户密码策略",
	"require_logon_to_change_password": "账户密码策略",
	"new_administrator_name":           "账户密码策略",
	"new_guest_name":                   "账户密码策略",
	"clear_text_password":              "账户密码策略",
	"lsa_anonymous_name_lookup":        "账户密码策略",
	"enable_admin_account":             "账户密码策略",
	"enable_guest_account":             "账户密码策略",

	// 审计策略
	"audit_system_events":     "审计策略",
	"audit_logon_events":      "审计策略",
	"audit_object_access":     "审计策略",
	"audit_privilege_use":     "审计策略",
	"audit_policy_change":     "审计策略",
	"audit_account_manage":    "审计策略",
	"audit_process_tracking":  "审计策略",
	"audit_ds_access":         "审计策略",
	"audit_account_logon":     "审计策略",

	// 设备控制
	"storage_devices": "设备控制",

	// 屏幕保护
	"screen_saver_active": "屏幕保护",
	"screen_saver_secure": "屏幕保护",
	"screen_save_timeout": "屏幕保护",
}

// WindowsFieldLabels Windows 字段标签映射
var WindowsFieldLabels = map[string]string{
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
	"audit_system_events":     "系统事件",
	"audit_logon_events":      "登录事件",
	"audit_object_access":     "对象访问",
	"audit_privilege_use":     "特权使用",
	"audit_policy_change":     "策略更改",
	"audit_account_manage":    "账户管理",
	"audit_process_tracking":  "进程跟踪",
	"audit_ds_access":         "DS 访问",
	"audit_account_logon":     "账户登录",

	// 设备控制
	"storage_devices": "移动存储设备",

	// 屏幕保护
	"screen_saver_active": "屏保启用",
	"screen_saver_secure": "屏保安全",
	"screen_save_timeout": "屏保超时(秒)",
}
