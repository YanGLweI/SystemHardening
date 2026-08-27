use serde::{Deserialize, Serialize};

/// 默认客户端版本（用于避免显示 unknown）
fn default_client_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

/// Windows 加固检查数据（对应后端 upload-data-windows 接口的 data 字段）
#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct WindowsSystemCheckData {
    pub date: String,
    pub hostname: String,
    pub domainname: String,
    pub ip: String,
    pub operasystem: String,
    
    #[serde(rename = "LicenseResult")]
    pub license_result: String,
    
    #[serde(rename = "client_version")]
    pub client_version: String,
    
    // 【新增】硬件 UUID
    #[serde(rename = "hardware_uuid", default)]
    pub hardware_uuid: String,

    // 账户密码策略 (15 项)
    pub minimum_password_age: String,
    pub maximum_password_age: String,
    pub minimum_password_length: String,
    pub password_complexity: String,
    pub password_history_size: String,
    pub lockout_bad_count: String,
    pub lockout_duration: String,
    pub reset_lockout_count: String,
    pub require_logon_to_change_password: String,
    pub new_administrator_name: String,
    pub new_guest_name: String,
    pub clear_text_password: String,
    pub lsa_anonymous_name_lookup: String,
    pub enable_admin_account: String,
    pub enable_guest_account: String,

    // 审计策略 (9 项)
    pub audit_system_events: String,
    pub audit_logon_events: String,
    pub audit_object_access: String,
    pub audit_privilege_use: String,
    pub audit_policy_change: String,
    pub audit_account_manage: String,
    pub audit_process_tracking: String,
    pub audit_ds_access: String,
    pub audit_account_logon: String,

    // 设备控制与屏幕保护
    pub storage_devices: String,
    pub screen_saver_active: String,
    pub screen_saver_secure: String,
    pub screen_save_timeout: String,
}

/// API 响应：临时 Token
#[derive(Debug, Deserialize)]
#[allow(dead_code)] // 部分字段仅用于反序列化验证
pub struct TempTokenResponse {
    pub temp_token: String,
    pub expires_in: i32,
    pub expires_at: String,
    // 后端 request-temp-token 接口不返回以下字段，使用默认值兼容
    #[serde(default)]
    pub device_name: String,
    #[serde(default)]
    pub ip_address: String,
}

/// 注册请求
#[derive(Debug, Serialize)]
pub struct RegisterRequest {
    pub temp_token: String,
    pub device_name: String,
    pub ip_address: String,
    pub os_version: String,
    pub client_version: String, // 新增：客户端版本
    
    // 【新增】硬件 UUID (空值不发送)
    #[serde(skip_serializing_if = "String::is_empty")]
    pub hardware_uuid: String,
}

/// 注册响应
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct RegisterResponse {
    pub client_uuid: String,
    pub short_token: String,
    pub refresh_token: String,
    pub expires_at: String,
    pub device_name: String,
    pub ip_address: String,
}

/// 心跳响应
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct HeartbeatResponse {
    pub status: String,
    pub client_uuid: String,
}

/// Token 刷新响应
#[derive(Debug, Deserialize)]
pub struct RefreshTokenResponse {
    pub short_token: String,
    pub expires_at: String,
}

/// 上传响应
#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct UploadResponse {
    pub message: String,
    pub record_id: u64,
    pub status: String,
}