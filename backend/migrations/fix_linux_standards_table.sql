-- 修复 linux_standards 表结构错误

-- 删除旧表并重建
DROP TABLE IF EXISTS `linux_standards`;

-- 创建正确的 linux_standards 表
CREATE TABLE `linux_standards` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `field_name` VARCHAR(50) NOT NULL COMMENT '字段名：pass_max_days',
  `field_label` VARCHAR(50) NOT NULL COMMENT '字段标签：PASS_MAX_DAYS',
  `standard_value` VARCHAR(200) NOT NULL COMMENT '标准值',
  `group_name` VARCHAR(50) NOT NULL COMMENT '分组：系统更新/用户账户策略等',
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  INDEX `idx_field_name` (`field_name`),
  INDEX `idx_group_name` (`group_name`),
  INDEX `idx_linux_standards_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Linux 标准配置表';
