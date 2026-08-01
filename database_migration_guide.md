# 数据库自动迁移指南

## 概述

本项目使用 GORM 的 `AutoMigrate` 功能在应用启动时自动创建或更新数据库表结构，无需手动执行 SQL 语句。

## 支持的模型

当前系统包含以下两个数据模型，会自动迁移到数据库：

### 1. User（用户模型）
- **表名**: `users`
- **字段**:
  - ID (主键)
  - Username (唯一，非空)
  - Password (非空，JSON 序列化时隐藏)
  - Email
  - Name
  - Role (默认值："user")
  - Status (默认值:1)
  - CreatedAt
  - UpdatedAt
  - DeletedAt (软删除支持)

### 2. SystemCheck（Linux 加固检查模型）
- **表名**: `systemcheck`
- **字段**:
  - ID (主键，自增)
  - Date (检查时间)
  - Hostname (计算机名)
  - Operasystem (操作系统版本)
  - Kernel (内核版本)
  - IP (IP 地址)
  - DnfConfGpgcheck
  - RedhatRepoGpgcheck
  - PassMaxDays
  - PassMinDays
  - PassMinLen
  - PassWarnAge
  - Inactive
  - GID
  - Tmout
  - Cron
  - Crontab
  - CronHourly
  - CronDaily
  - CronWeekly
  - CronMonthly
  - CronDeny
  - AtDeny
  - CronAllow
  - AtAllow
  - SshdConfig
  - LogLevel
  - X11Forwarding
  - MaxAuthTries
  - IgnoreRhosts
  - HostbasedAuthentication
  - PermitRootLogin
  - PermitEmptyPasswords
  - PermitUserEnvironment
  - ClientAliveInterval
  - ClientAliveCountMax
  - LoginGraceTime
  - Minlen
  - Minclass
  - Dcredit
  - Ucredit
  - Lcredit
  - Ocredit
  - PasswordRemember
  - Passwd
  - PasswdMinus
  - Group
  - GroupMinus
  - Shadow
  - ShadowMinus
  - Gshadow
  - GshadowMinus
  - CryptoPolicies
  - NtpServer
  - CreatedAt
  - UpdatedAt
  - DeletedAt (软删除支持)

## 工作原理

### 启动流程

当应用启动时，会按以下顺序执行：

```
1. 加载配置文件 (configs.LoadConfig())
2. 连接数据库 (database.ConnectDB())
3. ⭐ 自动迁移数据表 (database.AutoMigrate()) ← 新增
4. 初始化 LDAP 服务
5. 设置路由
6. 启动 HTTP 服务器
```

### AutoMigrate 行为

GORM 的 `AutoMigrate` 会根据 Go 结构体定义自动：

1. **创建新表**: 如果表不存在，则创建该表
2. **添加新列**: 如果表中缺少某些列，则添加这些列
3. **修改列类型**: 如果列类型不匹配，则修改列类型
4. **添加索引**: 根据 struct tag 中的 `gorm:"index"` 添加索引
5. **安全操作**: 不会删除已有数据或列（除非明确指定）

**注意**: AutoMigrate 适合开发和测试环境。在生产环境中，建议使用专门的迁移工具（如 golang-migrate）进行受控的版本控制迁移。

## 使用方法

### 首次部署

首次运行时，AutoMigrate 会自动创建所有表：

```bash
cd backend
go run cmd/main.go
```

您将看到类似以下的日志：

```
Database connection established successfully
Database tables migrated successfully
[info] created tables, count: 2
...其他详细信息...
Server starting on port 8080
```

### 添加新模型

如果您添加了新的模型，只需将其注册到 `AutoMigrate` 中：

```go
// database/db.go
func AutoMigrate() {
    err := DB.AutoMigrate(
        &models.User{},           // 现有
        &models.SystemCheck{},    // 现有
        &models.NewModel{},       // 新增
    )
    // ...
}
```

下次启动应用时，新表的表结构会被自动创建。

### 修改现有模型

当修改 model 结构体时，AutoMigrate 会自动更新表结构：

例如，为 `SystemCheck` 添加新字段：

```go
type SystemCheck struct {
    // 现有字段...
    
    NewField string `gorm:"column:new_field;size:100" json:"new_field"`
}
```

重启应用后，`new_field` 列会被自动添加到表中。

## 配置示例

### 开发环境

直接运行即可：

```bash
go run cmd/main.go
```

### 生产环境

建议先编译再运行：

```bash
# 编译
go build -o server ./cmd/main.go

# 运行
./server
```

## 常见问题

### Q1: 如何查看迁移过程中的详细信息？

AutoMigrate 会输出详细的日志信息，包括：
- 创建的表数量
- 添加的列
- 修改的索引等

确保 GORM 的日志级别设置为 Info 或 Debug：

```go
// database/db.go
gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // 或 logger.Info
}
```

### Q2: 是否会自动覆盖现有数据？

不会。AutoMigrate 是安全的表结构变更工具，不会覆盖或丢失数据。

### Q3: 如何回滚迁移？

当前的 AutoMigrate 不支持回滚。如需回滚，需要：
1. 备份数据库
2. 手动修改表结构
3. 或使用专业的迁移工具（如 golang-migrate）

### Q4: 可以只迁移特定表吗？

可以，只需要在 `AutoMigrate` 中列出要迁移的模型：

```go
err := DB.AutoMigrate(&models.SystemCheck{}) // 只迁移 systemcheck 表
```

### Q5: MySQL 连接失败怎么办？

检查以下几点：
1. 数据库服务是否运行
2. 用户名密码是否正确
3. 数据库是否存在
4. 网络是否可达

错误示例：
```
Failed to connect to database: dial tcp 10.66.254.155:3306: i/o timeout
```

### Q6: 迁移失败怎么办？

常见原因：
1. 权限不足（需要 CREATE、ALTER 权限）
2. 表已被占用
3. 约束冲突

解决方案：
```sql
-- 检查并授予必要权限
GRANT CREATE, ALTER, INDEX ON IT.* TO 'it'@'%';
FLUSH PRIVILEGES;
```

## 数据库表结构详情

以下是系统自动创建的表结构：

### users 表

```sql
CREATE TABLE `users` (
  `id` int UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `username` varchar(255) NOT NULL UNIQUE,
  `password` varchar(255) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `role` varchar(50) DEFAULT 'user',
  `status` int DEFAULT 1,
  `created_at` datetime(6) DEFAULT NULL,
  `updated_at` datetime(6) DEFAULT NULL,
  `deleted_at` datetime(6) DEFAULT NULL,
  INDEX `idx_users_deleted_at` (`deleted_at`)
);
```

### systemcheck 表

主要字段结构：

```sql
CREATE TABLE `systemcheck` (
  `id` int UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `date` varchar(50) DEFAULT NULL,
  `hostname` varchar(100) DEFAULT NULL,
  `operasystem` varchar(200) DEFAULT NULL,
  `kernel` varchar(100) DEFAULT NULL,
  `ip` varchar(50) DEFAULT NULL,
  -- ... 其他 45 个字段
  `created_at` datetime(6) DEFAULT NULL,
  `updated_at` datetime(6) DEFAULT NULL,
  `deleted_at` datetime(6) DEFAULT NULL,
  INDEX `idx_systemcheck_deleted_at` (`deleted_at`)
);
```

完整字段列表请查看 [`backend/models/linux_check.go`](file:///Users/yeung/Projects/system_hardening/backend/models/linux_check.go)。

## 最佳实践

### ✅ 推荐做法

1. **开发环境**: 直接使用 AutoMigrate，快速迭代
2. **代码注释**: 在 model 中添加字段注释，便于理解
3. **版本控制**: 将 model 文件提交到 Git
4. **定期备份**: 即使是自动迁移，也要定期备份数据

### ❌ 避免的做法

1. **生产环境盲目使用**: 生产环境建议使用专业迁移工具
2. **依赖自动更改 schema**: 应该通过文档记录预期的表结构
3. **跳过错误日志**: 始终检查迁移日志确认成功

## 相关文档

- [GORM 官方文档 - 迁移](https://gorm.io/migration.html)
- [Go Migrate - 专业迁移工具](https://github.com/golang-migrate/migrate)
- [项目 API 文档](./backend/LINUX_HARDENING_README.md)

---

**最后更新**: 2026 年 8 月  
**维护者**: 系统加固平台团队
