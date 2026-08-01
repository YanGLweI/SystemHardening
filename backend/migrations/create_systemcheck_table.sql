-- SQL Migration Script for systemcheck table
-- This script creates/recreates the systemcheck table with proper MySQL-compatible column names

USE IT;

DROP TABLE IF EXISTS systemcheck;

CREATE TABLE `systemcheck` (
  `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
  `date` varchar(50) DEFAULT NULL COMMENT '检查时间',
  `hostname` varchar(100) DEFAULT NULL COMMENT '计算机名',
  `operasystem` varchar(200) DEFAULT NULL COMMENT '操作系统版本',
  `kernel` varchar(100) DEFAULT NULL COMMENT '内核版本',
  `ip` varchar(50) DEFAULT NULL COMMENT 'IP 地址',
  `dnf.conf_gpgcheck` varchar(50) DEFAULT NULL COMMENT 'dnf gpgcheck 配置',
  `redhat.repo_gpgcheck` varchar(50) DEFAULT NULL COMMENT 'redhat repo gpgcheck 配置',
  `PASS_MAX_DAYS` varchar(50) DEFAULT NULL COMMENT '密码最大有效期',
  `PASS_MIN_DAYS` varchar(50) DEFAULT NULL COMMENT '密码最小有效期',
  `PASS_MIN_LEN` varchar(50) DEFAULT NULL COMMENT '密码最小长度',
  `PASS_WARN_AGE` varchar(50) DEFAULT NULL COMMENT '密码警告提前天数',
  `INACTIVE` varchar(50) DEFAULT NULL COMMENT '账号过期宽限天数',
  `GID` varchar(50) DEFAULT NULL COMMENT 'root 用户 GID',
  `TMOUT` varchar(50) DEFAULT NULL COMMENT 'Shell 超时时间',
  `Cron` varchar(50) DEFAULT NULL COMMENT 'Cron 守护进程状态',
  `crontab` varchar(200) DEFAULT NULL COMMENT 'crontab 文件权限',
  `cron.hourly` varchar(200) DEFAULT NULL COMMENT 'cron.hourly 目录权限',
  `cron.daily` varchar(200) DEFAULT NULL COMMENT 'cron.daily 目录权限',
  `cron.weekly` varchar(200) DEFAULT NULL COMMENT 'cron.weekly 目录权限',
  `cron.monthly` varchar(200) DEFAULT NULL COMMENT 'cron.monthly 目录权限',
  `cron.deny` varchar(200) DEFAULT NULL COMMENT 'cron.deny 文件权限',
  `at.deny` varchar(200) DEFAULT NULL COMMENT 'at.deny 文件权限',
  `cron.allow` varchar(200) DEFAULT NULL COMMENT 'cron.allow 文件权限',
  `at.allow` varchar(200) DEFAULT NULL COMMENT 'at.allow 文件权限',
  `sshd_config` varchar(200) DEFAULT NULL COMMENT 'sshd_config 文件权限',
  `LogLevel` varchar(50) DEFAULT NULL COMMENT 'SSH 日志级别',
  `X11Forwarding` varchar(50) DEFAULT NULL COMMENT 'X11 转发设置',
  `MaxAuthTries` varchar(50) DEFAULT NULL COMMENT '最大认证尝试次数',
  `IgnoreRhosts` varchar(50) DEFAULT NULL COMMENT '忽略 rhosts 设置',
  `HostbasedAuthentication` varchar(50) DEFAULT NULL COMMENT '基于主机的认证',
  `PermitRootLogin` varchar(50) DEFAULT NULL COMMENT 'root 登录设置',
  `PermitEmptyPasswords` varchar(50) DEFAULT NULL COMMENT '允许空密码',
  `PermitUserEnvironment` varchar(50) DEFAULT NULL COMMENT '允许用户环境',
  `ClientAliveInterval` varchar(50) DEFAULT NULL COMMENT '客户端存活间隔',
  `ClientAliveCountMax` varchar(50) DEFAULT NULL COMMENT '客户端存活最大次数',
  `LoginGraceTime` varchar(50) DEFAULT NULL COMMENT '登录宽限时间',
  `minlen` varchar(50) DEFAULT NULL COMMENT '密码最小长度要求',
  `minclass` varchar(50) DEFAULT NULL COMMENT '密码最小字符类别数',
  `dcredit` varchar(50) DEFAULT NULL COMMENT '数字字符 credit',
  `ucredit` varchar(50) DEFAULT NULL COMMENT '小写字符 credit',
  `lcredit` varchar(50) DEFAULT NULL COMMENT '大写字符 credit',
  `ocredit` varchar(50) DEFAULT NULL COMMENT '特殊字符 credit',
  `password_remember` varchar(50) DEFAULT NULL COMMENT '密码历史记住次数',
  `passwd` varchar(200) DEFAULT NULL COMMENT '/etc/passwd 权限',
  `passwd-` varchar(200) DEFAULT NULL COMMENT '/etc/passwd- 权限',
  `group` varchar(200) DEFAULT NULL COMMENT '/etc/group 权限',
  `group-` varchar(200) DEFAULT NULL COMMENT '/etc/group- 权限',
  `shadow` varchar(200) DEFAULT NULL COMMENT '/etc/shadow 权限',
  `shadow-` varchar(200) DEFAULT NULL COMMENT '/etc/shadow- 权限',
  `gshadow` varchar(200) DEFAULT NULL COMMENT '/etc/gshadow 权限',
  `gshadow-` varchar(200) DEFAULT NULL COMMENT '/etc/gshadow- 权限',
  `crypto_policies` varchar(100) DEFAULT NULL COMMENT '加密策略',
  `ntp_server` varchar(200) DEFAULT NULL COMMENT 'NTP 服务器',
  `deleted_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  INDEX `idx_systemcheck_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Linux 系统加固检查结果表';

-- 验证表创建
SHOW TABLES LIKE 'systemcheck';
DESCRIBE systemcheck LIMIT 10;
