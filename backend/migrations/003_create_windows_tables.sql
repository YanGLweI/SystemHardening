-- 创建 Windows 相关表
-- 包括 Windows 加固检查记录表、标准配置表和字段定义表

-- Windows 加固检查记录表
CREATE TABLE IF NOT EXISTS `systemcheck_windows` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `client_uuid` VARCHAR(64) NOT NULL,
  `date` VARCHAR(50) NOT NULL,
  `hostname` VARCHAR(100) DEFAULT NULL,
  `domainname` VARCHAR(100) DEFAULT NULL,
  `ip` VARCHAR(50) DEFAULT NULL,
  `operasystem` VARCHAR(200) DEFAULT NULL,
  `LicenseResult` VARCHAR(50) DEFAULT NULL,

  -- 账户密码策略 (15 项)
  `minimumpasswordage` VARCHAR(50) DEFAULT NULL,
  `maximumpasswordage` VARCHAR(50) DEFAULT NULL,
  `minimumpasswordlength` VARCHAR(50) DEFAULT NULL,
  `passwordcomplexity` VARCHAR(50) DEFAULT NULL,
  `passwordhistorysize` VARCHAR(50) DEFAULT NULL,
  `lockoutbadcount` VARCHAR(50) DEFAULT NULL,
  `lockoutduration` VARCHAR(50) DEFAULT NULL,
  `resetlockoutcount` VARCHAR(50) DEFAULT NULL,
  `requirelogontochangepassword` VARCHAR(50) DEFAULT NULL,
  `newadministratorname` VARCHAR(100) DEFAULT NULL,
  `newguestname` VARCHAR(100) DEFAULT NULL,
  `cleartextpassword` VARCHAR(50) DEFAULT NULL,
  `lsaanonymousnamelookup` VARCHAR(50) DEFAULT NULL,
  `enableadminaccount` VARCHAR(50) DEFAULT NULL,
  `enableguestaccount` VARCHAR(50) DEFAULT NULL,

  -- 审计策略 (9 项)
  `AuditSystemEvents` VARCHAR(50) DEFAULT NULL,
  `AuditLogonEvents` VARCHAR(50) DEFAULT NULL,
  `AuditObjectAccess` VARCHAR(50) DEFAULT NULL,
  `AuditPrivilegeUse` VARCHAR(50) DEFAULT NULL,
  `AuditPolicyChange` VARCHAR(50) DEFAULT NULL,
  `AuditAccountManage` VARCHAR(50) DEFAULT NULL,
  `AuditProcessTracking` VARCHAR(50) DEFAULT NULL,
  `AuditDSAccess` VARCHAR(50) DEFAULT NULL,
  `AuditAccountLogon` VARCHAR(50) DEFAULT NULL,

  -- 设备控制与屏幕保护
  `StorageDevices` VARCHAR(50) DEFAULT NULL,
  `ScreenSaveActive` VARCHAR(50) DEFAULT NULL,
  `ScreenSaverIsSecure` VARCHAR(50) DEFAULT NULL,
  `ScreenSaveTimeOut` VARCHAR(50) DEFAULT NULL,

  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,

  INDEX `idx_client_uuid` (`client_uuid`),
  INDEX `idx_date` (`date`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Windows 加固检查记录表';

-- Windows 标准配置表
CREATE TABLE IF NOT EXISTS `windows_standard` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `field_name` VARCHAR(200) NOT NULL COMMENT '字段名',
  `field_label` VARCHAR(200) NOT NULL COMMENT '字段标签',
  `standard_value` VARCHAR(500) NOT NULL COMMENT '标准值（支持正则）',
  `group_name` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '说明',
  `sort_order` INT DEFAULT 1 COMMENT '排序序号',
  `is_active` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,

  UNIQUE KEY `uk_windows_field_name` (`field_name`),
  INDEX `idx_windows_group_name` (`group_name`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Windows 标准配置表';

-- Windows 字段定义表
CREATE TABLE IF NOT EXISTS `windows_fields` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `field_name` VARCHAR(200) NOT NULL COMMENT '字段名',
  `field_label` VARCHAR(200) NOT NULL COMMENT '字段标签',
  `field_group` VARCHAR(100) DEFAULT NULL COMMENT '所属分组',
  `category` VARCHAR(50) DEFAULT NULL COMMENT '业务分类',
  `sort_order` INT DEFAULT 0 COMMENT '排序顺序',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '字段描述',
  `data_type` VARCHAR(20) DEFAULT 'string' COMMENT '数据类型',
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,

  UNIQUE KEY `uk_win_field_name` (`field_name`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Windows 加固字段定义表';

-- 初始化 Windows 字段定义数据
INSERT INTO `windows_fields` (`field_name`, `field_label`, `field_group`, `category`, `sort_order`, `description`) VALUES
-- 基本设置
('LicenseResult', '激活状态', '基本设置', 'basic', 1, 'Windows 激活状态'),

-- 账户密码策略
('minimum_password_age', '密码最短使用天数', '账户密码策略', 'password_policy', 10, '密码最短使用天数设置'),
('maximum_password_age', '密码最长使用天数', '账户密码策略', 'password_policy', 11, '密码最长使用天数设置'),
('minimum_password_length', '密码最小长度', '账户密码策略', 'password_policy', 12, '密码最小长度设置'),
('password_complexity', '密码复杂度', '账户密码策略', 'password_policy', 13, '密码复杂度要求'),
('password_history_size', '密码历史记录数', '账户密码策略', 'password_policy', 14, '记住的密码数量'),
('lockout_bad_count', '账户锁定阈值', '账户密码策略', 'password_policy', 15, '登录失败锁定阈值'),
('lockout_duration', '锁定持续时间(分钟)', '账户密码策略', 'password_policy', 16, '账户锁定持续时间'),
('reset_lockout_count', '重置锁定计数(分钟)', '账户密码策略', 'password_policy', 17, '锁定计数器重置时间'),
('require_logon_to_change_password', '登录更改密码', '账户密码策略', 'password_policy', 18, '要求登录更改密码'),
('new_administrator_name', '管理员名称', '账户密码策略', 'password_policy', 19, '重命名管理员账户'),
('new_guest_name', '来宾名称', '账户密码策略', 'password_policy', 20, '重命名来宾账户'),
('clear_text_password', '明文密码存储', '账户密码策略', 'password_policy', 21, '禁止明文存储密码'),
('lsa_anonymous_name_lookup', 'LSA 匿名查找', '账户密码策略', 'password_policy', 22, 'LSA 匿名名称查找'),
('enable_admin_account', '启用管理员账户', '账户密码策略', 'password_policy', 23, '管理员账户状态'),
('enable_guest_account', '启用来宾账户', '账户密码策略', 'password_policy', 24, '来宾账户状态'),

-- 审计策略
('audit_system_events', '系统事件', '审计策略', 'audit_policy', 30, '审核系统事件'),
('audit_logon_events', '登录事件', '审计策略', 'audit_policy', 31, '审核登录事件'),
('audit_object_access', '对象访问', '审计策略', 'audit_policy', 32, '审核对象访问'),
('audit_privilege_use', '特权使用', '审计策略', 'audit_policy', 33, '审核特权使用'),
('audit_policy_change', '策略更改', '审计策略', 'audit_policy', 34, '审核策略更改'),
('audit_account_manage', '账户管理', '审计策略', 'audit_policy', 35, '审核账户管理'),
('audit_process_tracking', '进程跟踪', '审计策略', 'audit_policy', 36, '审核进程跟踪'),
('audit_ds_access', 'DS 访问', '审计策略', 'audit_policy', 37, '审核目录服务访问'),
('audit_account_logon', '账户登录', '审计策略', 'audit_policy', 38, '审核账户登录'),

-- 设备控制
('storage_devices', '移动存储设备', '设备控制', 'device_control', 40, '移动存储设备访问控制'),

-- 屏幕保护
('screen_saver_active', '屏保启用', '屏幕保护', 'screensaver', 50, '屏幕保护程序启用'),
('screen_saver_secure', '屏保安全', '屏幕保护', 'screensaver', 51, '屏幕保护程序安全'),
('screen_save_timeout', '屏保超时(秒)', '屏幕保护', 'screensaver', 52, '屏幕保护超时时间');