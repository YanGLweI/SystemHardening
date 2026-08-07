package models

import (
	"time"
	"gorm.io/gorm"
)

// WindowsSystemCheck Windows 系统检查记录
type WindowsSystemCheck struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientUUID           string    `gorm:"size:64;index:idx_client_uuid" json:"client_uuid"`
	Date                 string    `gorm:"column:date;size:50" json:"date"` // 检查时间（格式：YYYY-MM-DD HH:mm:ss）
	Hostname             string    `gorm:"column:hostname;size:100" json:"hostname"`
	Domainname           string    `gorm:"column:domainname;size:100" json:"domainname"`
	IP                   string    `gorm:"column:ip;size:50" json:"ip"`
	Operasystem          string    `gorm:"column:operasystem;size:200" json:"operasystem"`
	LicenseResult        string    `gorm:"column:LicenseResult;size:50" json:"LicenseResult"`

	// 账户密码策略
	MinimumPasswordAge         string `gorm:"column:minimumpasswordage;size:50" json:"minimum_password_age"`
	MaximumPasswordAge         string `gorm:"column:maximumpasswordage;size:50" json:"maximum_password_age"`
	MinimumPasswordLength      string `gorm:"column:minimumpasswordlength;size:50" json:"minimum_password_length"`
	PasswordComplexity         string `gorm:"column:passwordcomplexity;size:50" json:"password_complexity"`
	PasswordHistorySize        string `gorm:"column:passwordhistorysize;size:50" json:"password_history_size"`
	LockoutBadCount            string `gorm:"column:lockoutbadcount;size:50" json:"lockout_bad_count"`
	LockoutDuration            string `gorm:"column:lockoutduration;size:50" json:"lockout_duration"`
	ResetLockoutCount          string `gorm:"column:resetlockoutcount;size:50" json:"reset_lockout_count"`
	RequireLogonToChangePwd    string `gorm:"column:requirelogontochangepassword;size:50" json:"require_logon_to_change_password"`
	NewAdministratorName       string `gorm:"column:newadministratorname;size:100" json:"new_administrator_name"`
	NewGuestName               string `gorm:"column:newguestname;size:100" json:"new_guest_name"`
	ClearTextPassword          string `gorm:"column:cleartextpassword;size:50" json:"clear_text_password"`
	LSAAnonymousNameLookup     string `gorm:"column:lsaanonymousnamelookup;size:50" json:"lsa_anonymous_name_lookup"`
	EnableAdminAccount         string `gorm:"column:enableadminaccount;size:50" json:"enable_admin_account"`
	EnableGuestAccount         string `gorm:"column:enableguestaccount;size:50" json:"enable_guest_account"`

	// 审计策略
	AuditSystemEvents   string `gorm:"column:AuditSystemEvents;size:50" json:"audit_system_events"`
	AuditLogonEvents    string `gorm:"column:AuditLogonEvents;size:50" json:"audit_logon_events"`
	AuditObjectAccess   string `gorm:"column:AuditObjectAccess;size:50" json:"audit_object_access"`
	AuditPrivilegeUse   string `gorm:"column:AuditPrivilegeUse;size:50" json:"audit_privilege_use"`
	AuditPolicyChange   string `gorm:"column:AuditPolicyChange;size:50" json:"audit_policy_change"`
	AuditAccountManage  string `gorm:"column:AuditAccountManage;size:50" json:"audit_account_manage"`
	AuditProcessTracking string `gorm:"column:AuditProcessTracking;size:50" json:"audit_process_tracking"`
	AuditDSAccess       string `gorm:"column:AuditDSAccess;size:50" json:"audit_ds_access"`
	AuditAccountLogon   string `gorm:"column:AuditAccountLogon;size:50" json:"audit_account_logon"`

	// 设备控制与屏幕保护
	RemovableStorageDenied string `gorm:"column:StorageDevices;size:50" json:"storage_devices"`
	ScreenSaverActive      string `gorm:"column:ScreenSaveActive;size:50" json:"screen_saver_active"`
	ScreenSaverIsSecure    string `gorm:"column:ScreenSaverIsSecure;size:50" json:"screen_saver_secure"`
	ScreenSaveTimeOut     string `gorm:"column:ScreenSaveTimeOut;size:50" json:"screen_save_timeout"`

	// 合规状态（不存储到数据库）
	ComplianceStatus string `gorm:"-" json:"compliance_status,omitempty"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TableName specify table name
func (WindowsSystemCheck) TableName() string {
	return "systemcheck_windows"
}

// WindowsStandard Windows 标准配置记录
type WindowsStandard struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FieldName     string         `gorm:"size:200;not null;uniqueIndex:idx_windows_field_name" json:"field_name"`
	FieldLabel    string         `gorm:"size:200;not null" json:"field_label"`
	StandardValue string         `gorm:"size:500;not null" json:"standard_value"`
	GroupName     string         `gorm:"size:100;not null;index:idx_windows_group_name" json:"group_name"`
	Description   string         `gorm:"size:500" json:"description"`
	SortOrder     int            `gorm:"default:1" json:"sort_order"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (WindowsStandard) TableName() string {
	return "windows_standard"
}

// WindowsField Windows 加固字段定义表
type WindowsField struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FieldName    string         `gorm:"size:200;not null;uniqueIndex" json:"field_name"`
	FieldLabel   string         `gorm:"size:200;not null" json:"field_label"`
	FieldGroup   string         `gorm:"size:100" json:"field_group"`
	Category     string         `gorm:"size:50" json:"category"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	Description  string         `gorm:"size:500" json:"description"`
	DataType     string         `gorm:"size:20;default:'string'" json:"data_type"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specify table name
func (WindowsField) TableName() string {
	return "windows_fields"
}