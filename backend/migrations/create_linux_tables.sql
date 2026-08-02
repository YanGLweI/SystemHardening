-- 创建 Linux 标准配置相关表

-- Linux 字段定义表
CREATE TABLE IF NOT EXISTS `linux_fields` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `field_name` VARCHAR(50) NOT NULL UNIQUE COMMENT '字段名：pass_max_days',
  `field_label` VARCHAR(50) NOT NULL COMMENT '字段标签：PASS_MAX_DAYS',
  `field_group` VARCHAR(50) COMMENT '所属分组',
  `category` VARCHAR(50) COMMENT '业务分类（用于前端 Tab 分组）',
  `sort_order` INT DEFAULT 0 COMMENT '排序顺序',
  `description` VARCHAR(500) COMMENT '字段描述',
  `is_required` TINYINT(1) DEFAULT FALSE COMMENT '是否必填',
  `data_type` VARCHAR(20) DEFAULT 'string' COMMENT '数据类型：string/number/int/bool',
  `default_value` VARCHAR(100) COMMENT '默认值',
  `created_at` DATETIME(3) NULL,
  `updated_at` DATETIME(3) NULL,
  `deleted_at` DATETIME(3) NULL,
  INDEX `idx_field_name` (`field_name`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Linux 加固字段定义表';
