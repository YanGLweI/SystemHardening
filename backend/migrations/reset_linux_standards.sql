-- 重置并初始化标准配置模块数据库表

-- 删除旧表（按依赖关系顺序）
DROP TABLE IF EXISTS `linux_field_groups`;
DROP TABLE IF EXISTS `linux_groups`;
DROP TABLE IF EXISTS `linux_fields`;
DROP TABLE IF EXISTS `linux_standards`;
