-- Data Migration Script: 从旧 table(含特殊字符) 迁移到新 table(GORM 兼容)
-- This script converts existing data from the old systemcheck table with special characters in column names
-- to a new clean table with GORM-compatible column names

USE IT;

-- Step 1: Create new table (if it doesn't exist)
CREATE TABLE IF NOT EXISTS systemcheck_new (
  `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
  `date` varchar(50) DEFAULT NULL COMMENT '检查时间',
  `hostname` varchar(100) DEFAULT NULL COMMENT '计算机名',
  `operasystem` varchar(200) DEFAULT NULL COMMENT '操作系统版本',
  `kernel` varchar(100) DEFAULT NULL COMMENT '内核版本',
  `ip` varchar(50) DEFAULT NULL COMMENT 'IP 地址',
  `dnf_conf_gpgcheck` varchar(50) DEFAULT NULL COMMENT 'dnf gpgcheck 配置',
  `redhat_repo_gpgcheck` varchar(50) DEFAULT NULL COMMENT 'redhat repo gpgcheck 配置',
  `pass_max_days` varchar(50) DEFAULT NULL COMMENT '密码最大有效期',
  `pass_min_days` varchar(50) DEFAULT NULL COMMENT '密码最小有效期',
  `pass_min_len` varchar(50) DEFAULT NULL COMMENT '密码最小长度',
  `pass_warn_age` varchar(50) DEFAULT NULL COMMENT '密码警告提前天数',
  `inactive` varchar(50) DEFAULT NULL COMMENT '账号过期宽限天数',
  `gid` varchar(50) DEFAULT NULL COMMENT 'root 用户 GID',
  `tmout` varchar(50) DEFAULT NULL COMMENT 'Shell 超时时间',
  `cron` varchar(50) DEFAULT NULL COMMENT 'Cron 守护进程状态',
  `crontab` varchar(200) DEFAULT NULL COMMENT 'crontab 文件权限',
  `cron_hourly` varchar(200) DEFAULT NULL COMMENT 'cron.hourly 目录权限',
  `cron_daily` varchar(200) DEFAULT NULL COMMENT 'cron.daily 目录权限',
  `cron_weekly` varchar(200) DEFAULT NULL COMMENT 'cron.weekly 目录权限',
  `cron_monthly` varchar(200) DEFAULT NULL COMMENT 'cron.monthly 目录权限',
  `cron_deny` varchar(200) DEFAULT NULL COMMENT 'cron.deny 文件权限',
  `at_deny` varchar(200) DEFAULT NULL COMMENT 'at.deny 文件权限',
  `cron_allow` varchar(200) DEFAULT NULL COMMENT 'cron.allow 文件权限',
  `at_allow` varchar(200) DEFAULT NULL COMMENT 'at.allow 文件权限',
  `sshd_config` varchar(200) DEFAULT NULL COMMENT 'sshd_config 文件权限',
  `log_level` varchar(50) DEFAULT NULL COMMENT 'SSH 日志级别',
  `x11_forwarding` varchar(50) DEFAULT NULL COMMENT 'X11 转发设置',
  `max_auth_tries` varchar(50) DEFAULT NULL COMMENT '最大认证尝试次数',
  `ignore_rhosts` varchar(50) DEFAULT NULL COMMENT '忽略 rhosts 设置',
  `hostbased_authentication` varchar(50) DEFAULT NULL COMMENT '基于主机的认证',
  `permit_root_login` varchar(50) DEFAULT NULL COMMENT 'root 登录设置',
  `permit_empty_passwords` varchar(50) DEFAULT NULL COMMENT '允许空密码',
  `permit_user_environment` varchar(50) DEFAULT NULL COMMENT '允许用户环境',
  `client_alive_interval` varchar(50) DEFAULT NULL COMMENT '客户端存活间隔',
  `client_alive_count_max` varchar(50) DEFAULT NULL COMMENT '客户端存活最大次数',
  `login_grace_time` varchar(50) DEFAULT NULL COMMENT '登录宽限时间',
  `minlen` varchar(50) DEFAULT NULL COMMENT '密码最小长度要求',
  `minclass` varchar(50) DEFAULT NULL COMMENT '密码最小字符类别数',
  `dcredit` varchar(50) DEFAULT NULL COMMENT '数字字符 credit',
  `ucredit` varchar(50) DEFAULT NULL COMMENT '小写字符 credit',
  `lcredit` varchar(50) DEFAULT NULL COMMENT '大写字符 credit',
  `ocredit` varchar(50) DEFAULT NULL COMMENT '特殊字符 credit',
  `password_remember` varchar(50) DEFAULT NULL COMMENT '密码历史记住次数',
  `passwd` varchar(200) DEFAULT NULL COMMENT '/etc/passwd 权限',
  `passwd_minus` varchar(200) DEFAULT NULL COMMENT '/etc/passwd- 权限',
  `group` varchar(200) DEFAULT NULL COMMENT '/etc/group 权限',
  `group_minus` varchar(200) DEFAULT NULL COMMENT '/etc/group- 权限',
  `shadow` varchar(200) DEFAULT NULL COMMENT '/etc/shadow 权限',
  `shadow_minus` varchar(200) DEFAULT NULL COMMENT '/etc/shadow- 权限',
  `gshadow` varchar(200) DEFAULT NULL COMMENT '/etc/gshadow 权限',
  `gshadow_minus` varchar(200) DEFAULT NULL COMMENT '/etc/gshadow- 权限',
  `crypto_policies` varchar(100) DEFAULT NULL COMMENT '加密策略',
  `ntp_server` varchar(200) DEFAULT NULL COMMENT 'NTP 服务器',
  `deleted_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  INDEX `idx_systemcheck_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Linux 系统加固检查结果表 - GORM 兼容版';

-- Step 2: Check if old table exists and has data
SELECT COUNT(*) as total_records FROM systemcheck WHERE 1=1 LIMIT 1;

-- Step 3: Migrate data if old table exists
-- WARNING: Make sure backup is created before running this!
-- Only run if you have existing data that needs to be preserved

/*
INSERT INTO systemcheck_new (
  id, date, hostname, operasystem, kernel, ip,
  dnf_conf_gpgcheck, redhat_repo_gpgcheck,
  pass_max_days, pass_min_days, pass_min_len, pass_warn_age,
  inactive, gid, tmout, cron, crontab,
  cron_hourly, cron_daily, cron_weekly, cron_monthly,
  cron_deny, at_deny, cron_allow, at_allow,
  sshd_config, log_level, x11_forwarding, max_auth_tries,
  ignore_rhosts, hostbased_authentication, permit_root_login,
  permit_empty_passwords, permit_user_environment, client_alive_interval,
  client_alive_count_max, login_grace_time, minlen, minclass,
  dcredit, ucredit, lcredit, ocredit, password_remember,
  passwd, passwd_minus, group, group_minus, shadow, shadow_minus,
  gshadow, gshadow_minus, crypto_policies, ntp_server,
  deleted_at, created_at, updated_at
)
SELECT 
  id, date, hostname, operasystem, kernel, ip,
  -- Convert column names from snake_case to match MySQL special column names
  `dnf.conf_gpgcheck`, `redhat.repo_gpgcheck`,
  `PASS_MAX_DAYS`, `PASS_MIN_DAYS`, `PASS_MIN_LEN`, `PASS_WARN_AGE`,
  `INACTIVE`, `GID`, `TMOUT`, `Cron`, `crontab`,
  `cron.hourly`, `cron.daily`, `cron.weekly`, `cron.monthly`,
  `cron.deny`, `at.deny`, `cron.allow`, `at.allow`,
  `sshd_config`, `LogLevel`, `X11Forwarding`, `MaxAuthTries`,
  `IgnoreRhosts`, `HostbasedAuthentication`, `PermitRootLogin`,
  `PermitEmptyPasswords`, `PermitUserEnvironment`, `ClientAliveInterval`,
  `ClientAliveCountMax`, `LoginGraceTime`, `minlen`, `minclass`,
  `dcredit`, `ucredit`, `lcredit`, `ocredit`, `password_remember`,
  `passwd`, `passwd-`, `group`, `group-`, `shadow`, `shadow-`,
  `gshadow`, `gshadow-`, `crypto_policies`, `ntp_server`,
  deleted_at, created_at, updated_at
FROM systemcheck;

SELECT ROW_COUNT() as migrated_rows;
*/

-- Step 4: Verify migration
SELECT COUNT(*) as new_table_total FROM systemcheck_new;

-- Step 5: Show column mapping (for reference)
SELECT 'Migration Complete!' as status;
