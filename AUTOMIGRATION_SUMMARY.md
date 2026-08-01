# 数据库自动迁移功能实现

## 修改内容

### 1. 增强数据库模块 ([`backend/database/db.go`](file:///Users/yeung/Projects/system_hardening/backend/database/db.go))

添加了 `AutoMigrate()` 函数，用于自动创建或更新数据表结构：

```go
func AutoMigrate() {
    err := DB.AutoMigrate(
        &models.User{},           // 用户表
        &models.SystemCheck{},    // Linux 加固检查表
    )
    
    if err != nil {
        log.Fatalf("AutoMigration failed: %v", err)
    }
    
    log.Println("Database tables migrated successfully")
}
```

### 2. 更新主程序入口 ([`backend/cmd/main.go`](file:///Users/yeung/Projects/system_hardening/backend/cmd/main.go))

在数据库连接成功后立即调用 `AutoMigrate()`，确保表结构在应用启动时自动初始化：

```go
// 初始化数据库连接
database.ConnectDB(config.Database)

// 自动迁移数据表（创建或更新表结构）✨
database.AutoMigrate()
```

## 工作流程

```
应用启动 → 加载配置 → 连接数据库 → ⭐ 自动迁移表结构 → 初始化服务 → 启动 HTTP 服务器
```

## 效果

启动时会看到类似以下的日志输出：

```
Database connection established successfully
[info] created table                [table] systemcheck
[info] create index                [index] idx_systemcheck_deleted_at   [table] systemcheck
[info] created table                [table] users
[info] create index                [index] idx_users_deleted_at         [table] users
Database tables migrated successfully
Server starting on port 8080
```

## 支持的表

- **users** - 用户管理表
- **systemcheck** - Linux 系统加固检查表（48 个字段）

## 优势

✅ **零手动操作**: 首次部署无需执行 SQL  
✅ **开发效率高**: 快速迭代，修改 model 自动同步表结构  
✅ **安全可靠**: 不会覆盖现有数据  
✅ **易于维护**: 新增模型只需注册到 AutoMigrate  

## 注意事项

⚠️ **生产环境建议**: 
- 使用专业迁移工具（如 golang-migrate）进行版本控制
- 定期备份数据库
- 记录预期的表结构变更

详细指南请查看：[`database_migration_guide.md`](file:///Users/yeung/Projects/system_hardening/database_migration_guide.md)

## 测试验证

```bash
# 1. 启动应用（首次运行会自动创建表）
cd backend
go run cmd/main.go

# 2. 检查数据库中的表
mysql -u your_user -p
USE IT;
SHOW TABLES;

# 3. 查看表结构
DESCRIBE systemcheck;
DESCRIBE users;
```

---

**实现时间**: 2026 年 8 月  
**状态**: ✅ 已完成并编译通过
