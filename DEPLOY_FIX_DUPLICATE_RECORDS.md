# 🚀 修复数据库重复记录 - 部署指南

## 📋 问题概述

**原始问题**：每次客户端上传系统检查数据时，后端都使用 `CREATE` 插入新记录，导致同一个客户端有多条历史记录。

**影响**：
- 数据库中同一 `client_uuid` 对应多条 `systemcheck` 记录
- 前端显示时每个客户端出现多次（重复条目）
- 无法正确显示最新的系统加固状态

---

## ✅ 解决方案

### 1. 代码修改

**修改文件**：`backend/controllers/client_controller.go`  
**修改函数**：`UploadData`

**核心逻辑变更**：
```go
// 之前：每次都 CREATE 新记录 ❌
if err := cc.db.Create(&req.Data).Error; err != nil {
    // ...
}

// 现在：先查询，有则 UPDATE，无则 CREATE ✅
var existingRecord models.SystemCheck
result := cc.db.Where("client_uuid = ?", token.ClientUUID).Order("id DESC").First(&existingRecord)

if result.Error == nil {
    // 记录存在 → UPDATE 操作
    req.Data.ID = existingRecord.ID
    cc.db.Model(&models.SystemCheck{}).Where("id = ?", existingRecord.ID).Updates(req.Data)
} else {
    // 记录不存在 → CREATE 操作
    cc.db.Create(&req.Data)
}
```

### 2. 清理现有重复数据

在服务器上执行以下 SQL：

```bash
mysql -u root -p'!Qw2!Qw2' system_hardening < /path/to/cleanup_duplicate_records.sql
```

或直接在 MySQL 中运行：
```sql
DELETE FROM systemcheck 
WHERE id NOT IN (
    SELECT MAX(id) 
    FROM systemcheck 
    GROUP BY client_uuid
);
```

---

## 🛠️ 部署步骤

### Step 1: 停止后端服务

根据您服务器的实际部署方式选择：

#### 方式 A：Docker 部署
```bash
ssh root@10.60.254.127
cd /path/to/system-hardening
docker-compose stop backend
```

#### 方式 B：直接运行二进制文件（本地已修复）
```bash
# 本地已编译好修复版本
# 在服务器上手动替换并重启
ssh root@10.60.254.127
pkill -f backend_app
mv backend_app_old backend_app  # 如果有备份

# 或者使用本地脚本启动（推荐）
cd /Users/yeung/Projects/system_hardening/backend
./start_backend.sh
```

### Step 2: 清理数据库重复记录

```bash
# 方法 1: 使用 SQL 脚本
mysql -u root -p'!Qw2!Qw2' system_hardening < /Users/yeung/Projects/system_hardening/backend/cleanup_duplicate_records.sql

# 方法 2: 直接在 MySQL 中执行
mysql -u root -p'!Qw2!Qw2' system_hardening
> DELETE FROM systemcheck WHERE id NOT IN (SELECT MAX(id) FROM systemcheck GROUP BY client_uuid);
> EXIT;
```

### Step 3: 重新编译并替换后端程序

```bash
# 在本地计算机上
cd /Users/yeung/Projects/system_hardening/backend/cmd
go build -o ../backend_app_new .

# 上传到服务器
scp /Users/yeung/Projects/system_hardening/backend/backend_app_new root@10.60.254.127:/tmp/

# 或者直接在服务器上编译（如果能访问）
ssh root@10.60.254.127
cd /home/system-hardening/backend/cmd
go build -o ../backend_app_new .
```

### Step 4: 启动后端服务

#### 方式 A：Docker 部署
```bash
# 将新的二进制文件复制到容器
docker cp /tmp/backend_app_new container_name:/app/backend_app
docker-compose start backend
```

#### 方式 B：直接运行二进制文件
```bash
mv /path/to/old_backend_app /path/to/old_backend_app.backup
cp /tmp/backend_app_new /path/to/backend_app
nohup ./backend_app > backend.log 2>&1 &
```

### Step 5: 验证修复

```bash
# 检查总记录数
mysql -u root -p'!Qw2!Qw2' system_hardening -e "SELECT COUNT(*) as total_records FROM systemcheck;"

# 检查唯一客户端数
mysql -u root -p'!Qw2!Qw2' system_hardening -e "SELECT COUNT(DISTINCT client_uuid) as unique_clients FROM systemcheck;"

# 查看 test-it 客户端的记录数
mysql -u root -p'!Qw2!Qw2' system_hardening -e "SELECT client_uuid, COUNT(*) as count FROM systemcheck GROUP BY client_uuid;"
```

**预期结果**：
- ✅ 总记录数 = 唯一客户端数（每条 client_uuid 只有一条记录）
- ✅ 前端页面只显示一个 `test-it` 条目

---

## 📊 修复前后对比

### 修复前（❌）
```
client_uuid: B98D3472-1F9E-44F2-A177-40D115F2A6BB (test-it)
├── ID=9   2026/08/01_22:37:57
├── ID=10  2026/08/01_22:56:50
├── ID=11  2026/08/01_23:04:47
├── ID=12  2026/08/01_23:11:46
├── ID=13  2026/08/01_23:24:40
├── ID=14  2026/08/01_23:31:41
└── ID=15  2026/08/01_23:37:52
```
**前端显示**: 7 个重复的 `test-it` 条目

### 修复后（✅）
```
client_uuid: B98D3472-1F9E-44F2-A177-40D115F2A6BB (test-it)
└── ID=15  2026/08/01_23:37:52 (最新一次更新到这里)
```
**前端显示**: 1 个 `test-it` 条目，包含最新的加固状态

---

## 🔍 日志验证

更新后的行为会在日志中显示：

**首次注册（新建）**：
```
✅ Created new systemcheck record for client: B98D3472-...
```

**后续更新（更新）**：
```
✅ Updated existing systemcheck record for client: B98D3472-... (ID=15)
```

---

## ⚙️ 未来优化建议

1. **考虑保留历史记录**：如果需要审计功能，可以创建一个单独的 `systemcheck_history` 表，主表保持最新，历史表保留所有记录。

2. **添加索引**：在 `client_uuid` 字段上添加索引以提高查询性能：
   ```sql
   CREATE INDEX idx_systemcheck_client_uuid ON systemcheck(client_uuid);
   ```

3. **定时清理**：定期清理超过一定时间（如 90 天）的历史记录。

---

## 📞 支持

如有问题，请查看：
- `/Users/yeung/Projects/system_hardening/backend/controllers/client_controller.go`
- 后端日志文件：`/path/to/backend.log`
- SQL 清理脚本：`/Users/yeung/Projects/system_hardening/backend/cleanup_duplicate_records.sql`
