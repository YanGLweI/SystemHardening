# Windows 系统加固客户端（Rust）

跨平台客户端体系的 Windows 端，使用 Rust 编写。与 Linux 客户端（Go）共用同一套
后端协议：临时 Token 注册 → Token 认证 → 心跳 → 定期上传加固检查数据。

## 功能特性

- **只读采集**：通过 WMI / 注册表 / secedit 采集 Windows 加固状态，不修改任何系统配置（加固由域控 GPO 完成）
- **29 项检查数据**：
  - 系统信息：主机名、域名、IP、操作系统、许可证状态
  - 账户密码策略（15 项）：密码长度/复杂度/有效期/锁定策略等
  - 审计策略（9 项）：系统/登录/对象访问等审计事件
  - 设备控制：移动存储设备禁用状态
  - 屏幕保护：启用状态、密码保护、超时时间
- **Windows 服务**：以 `SystemHardeningWinClient` 服务运行，开机自启，自动重启（故障恢复策略）
- **Token 管理**：与 Linux 客户端一致的注册/刷新协议，JSON 文件持久化
- **运行周期**：心跳每 2 分钟，加固检查每 24 小时（与 Linux 客户端一致）

## 项目结构

```
windows-client/
├── Cargo.toml              # 依赖：wmi / winreg / windows-service / reqwest / serde
├── config.example.yaml     # 示例配置
├── installer/
│   └── windows_client.nsi  # NSIS 安装包脚本
└── src/
    ├── main.rs             # 入口（服务模式 / --foreground 调试模式）
    ├── service.rs          # Windows 服务生命周期
    ├── worker.rs           # 业务循环（注册/心跳/每日检查）
    ├── collector.rs        # 信息采集（WMI + 注册表 + secedit）
    ├── api.rs              # HTTP API 通信（reqwest blocking）
    ├── token.rs            # Token 管理器（JSON 持久化）
    ├── models.rs           # 数据模型
    └── config.rs           # YAML 配置加载
```

## 编译

需要 Rust 工具链 + `x86_64-pc-windows-gnu` 交叉编译目标：

```bash
rustup target add x86_64-pc-windows-gnu
cargo build --release --target x86_64-pc-windows-gnu
# 产物：target/x86_64-pc-windows-gnu/release/windows_hardening_client.exe
```

## 安装与部署

### 方式一：NSIS 安装包（推荐）

在 Windows 上安装 [NSIS](https://nsis.sourceforge.io/) 后编译：

```bash
makensis installer/windows_client.nsi
# 产物：installer/dist/SystemHardening_WindowsClient_Setup_1.0.0.exe
```

安装包会：注册 Windows 服务 → 复制配置 → 启动服务。

### 方式二：手动部署

1. 将 `windows_hardening_client.exe` 复制到 `C:\Program Files\SystemHardening\WindowsClient\`
2. 将 `config.example.yaml` 复制为 `C:\ProgramData\SystemHardening\WindowsClient\config.yaml` 并修改 `server_url`
3. 以管理员身份注册并启动服务：

```bat
sc create SystemHardeningWinClient binPath= "C:\Program Files\SystemHardening\WindowsClient\windows_hardening_client.exe" start= auto DisplayName= "系统加固 Windows 客户端"
net start SystemHardeningWinClient
```

### 前台调试

```bash
windows_hardening_client.exe --foreground config.yaml
```

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server_url` | 管理服务器地址（必填） | `http://localhost:8080` |
| `local_db_path` | Token 持久化路径 | `C:\ProgramData\SystemHardening\WindowsClient\tokens.json` |
| `device_name` | 设备名（留空自动采集主机名） | 自动 |
| `ip_address` | IP 地址（留空自动采集） | 自动 |

## 运行要求

- Windows Server 2016+ / Windows 10 x64
- 管理员权限（采集审计策略等需要）
- 可访问管理服务器（HTTP 8080）

## 常用命令

```bat
:: 查看服务状态
sc query SystemHardeningWinClient

:: 重启服务（修改配置后）
sc stop SystemHardeningWinClient
sc start SystemHardeningWinClient

:: 卸载服务
sc delete SystemHardeningWinClient
```
