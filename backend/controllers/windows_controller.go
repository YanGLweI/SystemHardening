package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/database"
	"github.com/yeung/system-hardening/backend/models"
)

type WindowsController struct{}

// NewWindowsController create new WindowsController instance
func NewWindowsController() *WindowsController {
	return &WindowsController{}
}

// List 获取 Windows 加固检查列表（分页）
func (wc *WindowsController) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	keyword := c.Query("keyword")
	complianceStatus := c.Query("compliance_status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var checks []models.WindowsSystemCheck
	var total int64

	db := database.DB

	// 获取所有 Windows 标准配置
	var standards []models.WindowsStandard
	db.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&standards)

	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}

	// 如果需要按合规状态过滤，需要先获取全部数据再内存过滤
	if complianceStatus != "" {
		var allChecks []models.WindowsSystemCheck
		query := db.Model(&models.WindowsSystemCheck{})
		if keyword != "" {
			query = query.Where("hostname LIKE ? OR ip LIKE ? OR domainname LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
		if err := query.Order("id DESC").Find(&allChecks).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch data"})
			return
		}
		// 计算合规状态并过滤
		for i := range allChecks {
			result := CompareWindowsCompliance(&allChecks[i], standardMap)
			allChecks[i].ComplianceStatus = result.Status
		}
		for _, check := range allChecks {
			if check.ComplianceStatus == complianceStatus {
				checks = append(checks, check)
			}
		}
		total = int64(len(checks))
		// 手动分页
		offset := (page - 1) * pageSize
		end := offset + pageSize
		if offset > len(checks) {
			checks = []models.WindowsSystemCheck{}
		} else {
			if end > len(checks) {
				end = len(checks)
			}
			checks = checks[offset:end]
		}
	} else {
		// 无需合规过滤，正常 SQL 分页
		query := db.Model(&models.WindowsSystemCheck{})
		if keyword != "" {
			query = query.Where("hostname LIKE ? OR ip LIKE ? OR domainname LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
		query.Count(&total)
		if err := query.Order("id DESC").Limit(pageSize).Offset((page-1)*pageSize).Find(&checks).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch data"})
			return
		}
		for i := range checks {
			result := CompareWindowsCompliance(&checks[i], standardMap)
			checks[i].ComplianceStatus = result.Status
		}
	}

	c.JSON(200, gin.H{
		"list":     checks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Detail 获取单个 Windows 主机加固检查详情
func (wc *WindowsController) Detail(c *gin.Context) {
	id := c.Param("id")

	var check models.WindowsSystemCheck
	if err := database.DB.First(&check, id).Error; err != nil {
		c.JSON(404, gin.H{
			"error": "Record not found",
		})
		return
	}

	// 获取所有标准配置
	var standards []models.WindowsStandard
	database.DB.Model(&models.WindowsStandard{}).Where("deleted_at IS NULL").Find(&standards)

	// 构建标准值映射
	standardMap := make(map[string]string)
	for _, std := range standards {
		standardMap[std.FieldName] = std.StandardValue
	}

	// 计算合规状态
	result := CompareWindowsCompliance(&check, standardMap)
	check.ComplianceStatus = result.Status

	c.JSON(200, gin.H{
		"check":      check,
		"compliance": result,
	})
}

// CompareWindowsCompliance 比对 Windows 系统检查记录与标准配置
func CompareWindowsCompliance(check *models.WindowsSystemCheck, standardMap map[string]string) *ComplianceResult {
	result := &ComplianceResult{
		Status:             "compliant",
		NonCompliantFields: []NonCompliantField{},
	}

	// 定义字段名到实际值的映射
	fieldValues := map[string]string{
		// 基本设置
		"LicenseResult": check.LicenseResult,

		// 账户密码策略
		"minimum_password_age":              check.MinimumPasswordAge,
		"maximum_password_age":              check.MaximumPasswordAge,
		"minimum_password_length":           check.MinimumPasswordLength,
		"password_complexity":               check.PasswordComplexity,
		"password_history_size":             check.PasswordHistorySize,
		"lockout_bad_count":                 check.LockoutBadCount,
		"lockout_duration":                  check.LockoutDuration,
		"reset_lockout_count":               check.ResetLockoutCount,
		"require_logon_to_change_password":  check.RequireLogonToChangePwd,
		"new_administrator_name":            check.NewAdministratorName,
		"new_guest_name":                    check.NewGuestName,
		"clear_text_password":               check.ClearTextPassword,
		"lsa_anonymous_name_lookup":         check.LSAAnonymousNameLookup,
		"enable_admin_account":              check.EnableAdminAccount,
		"enable_guest_account":              check.EnableGuestAccount,

		// 审计策略
		"audit_system_events":   check.AuditSystemEvents,
		"audit_logon_events":    check.AuditLogonEvents,
		"audit_object_access":   check.AuditObjectAccess,
		"audit_privilege_use":   check.AuditPrivilegeUse,
		"audit_policy_change":   check.AuditPolicyChange,
		"audit_account_manage":  check.AuditAccountManage,
		"audit_process_tracking": check.AuditProcessTracking,
		"audit_ds_access":       check.AuditDSAccess,
		"audit_account_logon":   check.AuditAccountLogon,

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
					Label:    getWindowsFieldLabel(fieldName),
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
					Label:    getWindowsFieldLabel(fieldName),
					Actual:   actualValue,
					Standard: standardValue,
				})
			}
		}
	}

	return result
}

// getWindowsFieldLabel 获取 Windows 字段的显示标签
func getWindowsFieldLabel(fieldName string) string {
	labels := map[string]string{
		// 基本设置
		"LicenseResult": "激活状态",

		// 账户密码策略
		"minimum_password_age":              "密码最短使用天数",
		"maximum_password_age":              "密码最长使用天数",
		"minimum_password_length":           "密码最小长度",
		"password_complexity":               "密码复杂度",
		"password_history_size":             "密码历史记录数",
		"lockout_bad_count":                 "账户锁定阈值",
		"lockout_duration":                  "锁定持续时间(分钟)",
		"reset_lockout_count":               "重置锁定计数(分钟)",
		"require_logon_to_change_password":  "登录更改密码",
		"new_administrator_name":            "管理员名称",
		"new_guest_name":                    "来宾名称",
		"clear_text_password":               "明文密码存储",
		"lsa_anonymous_name_lookup":         "LSA 匿名查找",
		"enable_admin_account":              "启用管理员账户",
		"enable_guest_account":              "启用来宾账户",

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