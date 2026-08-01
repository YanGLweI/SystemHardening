# Linux 加固检查模块说明

## 功能概述

Linux 加固检查模块用于展示从数据库获取的 Linux 系统加固检查结果。该模块提供了分页列表展示和详情查看功能，帮助用户了解各主机的安全加固状态。

## 数据库字段说明

根据 `mysql-insert.sh`脚本中的数据表结构，`systemcheck` 表包含以下字段：

### 基本信息
- `id`: 记录 ID（自增）
- `date`: 检查时间
- `hostname`: 计算机名
- `operasystem`: 操作系统版本
- `kernel`: 内核版本
- `ip`: IP 地址

### 系统更新配置
- `dnf.conf_gpgcheck`: /etc/dnf/dnf.conf 中的 gpgcheck 设置
- `redhat.repo_gpgcheck`: /etc/yum.repos.d/redhat.repo 中的 gpgcheck 设置

### 用户账户策略
- `PASS_MAX_DAYS`: 密码最大有效期
- `PASS_MIN_DAYS`: 密码最小有效期
- `PASS_MIN_LEN`: 密码最小长度
- `PASS_WARN_AGE`: 密码警告提前天数
- `INACTIVE`: 账号过期宽限天数
- `GID`: root 用户的 GID
- `TMOUT`: Shell 超时时间（秒）

### 计划任务配置
- `Cron`: Cron 守护进程启用状态
- `crontab`: crontab 文件权限信息
- `cron.hourly`: cron.hourly 目录权限信息
- `cron.daily`: cron.daily 目录权限信息
- `cron.weekly`: cron.weekly 目录权限信息
- `cron.monthly`: cron.monthly 目录权限信息
- `cron.deny`: cron.deny 文件权限信息
- `at.deny`: at.deny 文件权限信息
- `cron.allow`: cron.allow 文件权限信息
- `at.allow`: at.allow 文件权限信息

### SSH 服务器配置
- `sshd_config`: sshd_config 文件权限信息
- `LogLevel`: SSH 日志级别
- `X11Forwarding`: X11 转发设置
- `MaxAuthTries`: 最大认证尝试次数
- `IgnoreRhosts`: 是否忽略 rhosts 文件
- `HostbasedAuthentication`: 基于主机的认证设置
- `PermitRootLogin`: 是否允许 root 登录
- `PermitEmptyPasswords`: 是否允许空密码
- `PermitUserEnvironment`: 是否允许用户环境
- `ClientAliveInterval`: 客户端存活检测间隔
- `ClientAliveCountMax`: 客户端存活检测最大次数
- `LoginGraceTime`: 登录宽限时间

### 密码策略与复杂度
- `minlen`: 密码最小长度要求
- `minclass`: 密码最小字符类别数
- `dcredit`: 数字字符 credits
- `ucredit`: 小写字符 credits
- `lcredit`: 大写字符 credits
- `ocredit`: 特殊字符 credits
- `password_remember`: 密码历史记住次数

### 文件系统权限
- `passwd`: /etc/passwd 文件权限信息
- `passwd-`: /etc/passwd- 文件权限信息
- `group`: /etc/group 文件权限信息
- `group-`: /etc/group- 文件权限信息
- `shadow`: /etc/shadow 文件权限信息
- `shadow-`: /etc/shadow- 文件权限信息
- `gshadow`: /etc/gshadow 文件权限信息
- `gshadow-`: /etc/gshadow- 文件权限信息

### 加密与时钟同步
- `crypto_policies`: 加密策略设置
- `ntp_server`: NTP 服务器配置

## API 接口说明

### 1. 获取 Linux 加固检查列表（分页）
**请求:**
```
GET /api/linux-checks?page=1&pageSize=10
```

**响应:**
```json
{
  "list": [...],
  "total": 100,
  "page": 1,
  "pageSize": 10
}
```

**参数:**
- `page`: 页码（默认值为 1）
- `pageSize`: 每页数量（默认值为 10）

### 2. 获取单个主机加固检查详情
**请求:**
```
GET /api/linux-checks/1
```

**响应:**
返回完整的 SystemCheck 对象，包含所有字段。

## 前端页面

### 页面位置
在左侧菜单中：**安全加固 > Linux 加固**

### 页面功能
1. **表格展示**: 显示计算机名、IP、系统和合规状态（目前显示为 "-"）
2. **分页**: 支持切换每页数量和翻页
3. **详情查看**: 点击"详情"按钮可查看所有字段的详细信息
4. **标签页分组**: 详情弹窗使用多个标签页将不同类别的配置项分组展示

### 详情弹窗分组
- **基本信息**: 主机名、IP、系统版本、内核版本、检查时间
- **系统更新**: gpgcheck 相关配置
- **用户账户策略**: 密码策略、账户超时设置等
- **计划任务**: Cron 相关配置
- **SSH 配置**: SSH 服务器配置
- **密码策略**: 密码复杂度要求
- **文件权限**: 关键系统文件权限
- **加密与时钟**: 加密策略和时间同步配置

## 后续扩展

### 合规状态判断
当前合规状态列显示为"-"，后续可以：
1. 在后端增加合规性判断逻辑
2. 根据关键字段计算合规得分或状态
3. 在数据库中新增 `compliance_status` 字段或使用计算字段

### 数据迁移
如果需要从旧数据库迁移数据到新的 `systemcheck` 表，可以参考 `mysql-insert.sh`中的 SQL 语句结构。

## 测试

使用提供的测试脚本验证 API:
```bash
cd backend
./test_linux_checks.sh <your_jwt_token>
```

## 注意事项
1. 该模块仅用于展示数据，不支持增删改操作
2. 详情页面的所有字段都直接来自数据库查询
3. 分页数据按 ID 降序排列（最新添加的记录在前）
4. 需要 JWT 认证才能访问（通过路由中间件保护）
