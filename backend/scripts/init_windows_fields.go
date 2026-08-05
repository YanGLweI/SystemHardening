package scripts

import (
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

// InitWindowsFieldsAndGroups 初始化 Windows 字段定义和分组数据
func InitWindowsFieldsAndGroups() {
	db := database.DB

	var count int64
	db.Model(&models.WindowsField{}).Count(&count)
	if count > 0 {
		return // 已初始化，跳过
	}

	// 所有 Windows 加固字段的定义
	fields := []models.WindowsField{
		// 基本设置
		{FieldName: "license_result", FieldLabel: "激活状态", FieldGroup: "基本设置", Category: "basic", SortOrder: 1, DataType: "string", Description: "Windows 激活状态"},

		// 账户密码策略 (15 个)
		{FieldName: "minimum_password_age", FieldLabel: "密码最短使用天数", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 10, DataType: "string", Description: "密码最短使用天数设置"},
		{FieldName: "maximum_password_age", FieldLabel: "密码最长使用天数", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 11, DataType: "string", Description: "密码最长使用天数设置"},
		{FieldName: "minimum_password_length", FieldLabel: "密码最小长度", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 12, DataType: "string", Description: "密码最小长度设置"},
		{FieldName: "password_complexity", FieldLabel: "密码复杂度", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 13, DataType: "string", Description: "密码复杂度要求"},
		{FieldName: "password_history_size", FieldLabel: "密码历史记录数", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 14, DataType: "string", Description: "记住的密码数量"},
		{FieldName: "lockout_bad_count", FieldLabel: "账户锁定阈值", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 15, DataType: "string", Description: "登录失败锁定阈值"},
		{FieldName: "lockout_duration", FieldLabel: "锁定持续时间(分钟)", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 16, DataType: "string", Description: "账户锁定持续时间"},
		{FieldName: "reset_lockout_count", FieldLabel: "重置锁定计数(分钟)", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 17, DataType: "string", Description: "锁定计数器重置时间"},
		{FieldName: "require_logon_to_change_password", FieldLabel: "登录更改密码", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 18, DataType: "string", Description: "要求登录后更改密码"},
		{FieldName: "new_administrator_name", FieldLabel: "管理员名称", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 19, DataType: "string", Description: "重命名管理员账户"},
		{FieldName: "new_guest_name", FieldLabel: "来宾名称", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 20, DataType: "string", Description: "重命名来宾账户"},
		{FieldName: "clear_text_password", FieldLabel: "明文密码存储", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 21, DataType: "string", Description: "禁止明文存储密码"},
		{FieldName: "lsa_anonymous_name_lookup", FieldLabel: "LSA 匿名查找", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 22, DataType: "string", Description: "LSA 匿名名称查找"},
		{FieldName: "enable_admin_account", FieldLabel: "启用管理员账户", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 23, DataType: "string", Description: "管理员账户状态"},
		{FieldName: "enable_guest_account", FieldLabel: "启用来宾账户", FieldGroup: "账户密码策略", Category: "password_policy", SortOrder: 24, DataType: "string", Description: "来宾账户状态"},

		// 审计策略 (9 个)
		{FieldName: "audit_system_events", FieldLabel: "系统事件", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 30, DataType: "string", Description: "审核系统事件"},
		{FieldName: "audit_logon_events", FieldLabel: "登录事件", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 31, DataType: "string", Description: "审核登录事件"},
		{FieldName: "audit_object_access", FieldLabel: "对象访问", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 32, DataType: "string", Description: "审核对象访问"},
		{FieldName: "audit_privilege_use", FieldLabel: "特权使用", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 33, DataType: "string", Description: "审核特权使用"},
		{FieldName: "audit_policy_change", FieldLabel: "策略更改", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 34, DataType: "string", Description: "审核策略更改"},
		{FieldName: "audit_account_manage", FieldLabel: "账户管理", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 35, DataType: "string", Description: "审核账户管理"},
		{FieldName: "audit_process_tracking", FieldLabel: "进程跟踪", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 36, DataType: "string", Description: "审核进程跟踪"},
		{FieldName: "audit_ds_access", FieldLabel: "DS 访问", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 37, DataType: "string", Description: "审核目录服务访问"},
		{FieldName: "audit_account_logon", FieldLabel: "账户登录", FieldGroup: "审计策略", Category: "audit_policy", SortOrder: 38, DataType: "string", Description: "审核账户登录"},

		// 设备控制 (1 个)
		{FieldName: "storage_devices", FieldLabel: "移动存储设备", FieldGroup: "设备控制", Category: "device_control", SortOrder: 40, DataType: "string", Description: "移动存储设备访问控制"},

		// 屏幕保护 (3 个)
		{FieldName: "screen_saver_active", FieldLabel: "屏保启用", FieldGroup: "屏幕保护", Category: "screensaver", SortOrder: 50, DataType: "string", Description: "屏幕保护程序启用"},
		{FieldName: "screen_saver_secure", FieldLabel: "屏保安全", FieldGroup: "屏幕保护", Category: "screensaver", SortOrder: 51, DataType: "string", Description: "屏幕保护程序安全"},
		{FieldName: "screen_save_timeout", FieldLabel: "屏保超时(秒)", FieldGroup: "屏幕保护", Category: "screensaver", SortOrder: 52, DataType: "string", Description: "屏幕保护超时时间"},
	}

	// 批量插入所有字段
	err := db.Create(&fields).Error
	if err != nil {
		panic("Failed to insert windows fields: " + err.Error())
	}

	println("✅ Successfully initialized " + string(rune(len(fields)+'0')) + " fields into windows_fields table")
}