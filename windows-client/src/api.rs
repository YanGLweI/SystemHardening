use crate::models::{
    HeartbeatResponse, RefreshTokenResponse, RegisterRequest, RegisterResponse,
    TempTokenResponse, UploadResponse, WindowsSystemCheckData,
};
use reqwest::blocking::Client;

/// 请求临时安装 Token
pub fn request_temp_token(
    server_url: &str,
    device_name: &str,
    ip_address: &str,
) -> Result<TempTokenResponse, String> {
    let payload = serde_json::json!({
        "device_name": device_name,
        "ip_address": ip_address,
    });

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/client/request-temp-token", server_url))
        .json(&payload)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("Request temp token failed: {}", body));
    }

    resp.json::<TempTokenResponse>()
        .map_err(|e| format!("Parse response failed: {}", e))
}

/// 使用临时 Token 注册客户端
pub fn register(
    server_url: &str,
    temp_token: &str,
    device_name: &str,
    ip_address: &str,
    os_version: &str,
    client_version: &str, // 新增：客户端版本
    hardware_uuid: &str, // 【新增】硬件 UUID
) -> Result<RegisterResponse, String> {
    let req = RegisterRequest {
        temp_token: temp_token.to_string(),
        device_name: device_name.to_string(),
        ip_address: ip_address.to_string(),
        os_version: os_version.to_string(),
        client_version: client_version.to_string(), // 新增：客户端版本
        hardware_uuid: hardware_uuid.to_string(), // 【新增}
    };

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/client/register", server_url))
        .json(&req)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("Register failed: {}", body));
    }

    resp.json::<RegisterResponse>()
        .map_err(|e| format!("Parse response failed: {}", e))
}

/// 发送心跳
pub fn send_heartbeat(
    server_url: &str, 
    short_token: &str, 
    hardware_uuid: &str,
    device_name: &str,  // 【新增】当前设备名
    ip_address: &str,   // 【新增】当前 IP 地址
) -> Result<HeartbeatResponse, String> {
    let client = Client::new();
    // 【关键改进】心跳携带完整设备身份信息，确保平台实时同步 hostname 和 IP 变化
    let body = serde_json::json!({
        "client_version": env!("CARGO_PKG_VERSION"),
        "hardware_uuid": hardware_uuid,
        "device_name": device_name,      // 新增字段
        "ip_address": ip_address,        // 新增字段
    });
    let resp = client
        .post(format!("{}/api/client/heartbeat", server_url))
        .header("X-Client-Token", short_token)
        .json(&body)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().unwrap_or_default();
        return Err(format!("Heartbeat failed: HTTP {}, body: {}", status, body));
    }

    resp.json::<HeartbeatResponse>()
        .map_err(|e| format!("Parse response failed: {}", e))
}

/// 刷新 Token
pub fn refresh_token(
    server_url: &str,
    refresh_token: &str,
) -> Result<RefreshTokenResponse, String> {
    let payload = serde_json::json!({
        "refresh_token": refresh_token,
    });

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/client/refresh-token", server_url))
        .json(&payload)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("Refresh token failed: {}", body));
    }

    resp.json::<RefreshTokenResponse>()
        .map_err(|e| format!("Parse response failed: {}", e))
}

/// 上传 Windows 加固检查数据
pub fn upload_windows_data(
    server_url: &str,
    short_token: &str,
    data: &WindowsSystemCheckData,
) -> Result<UploadResponse, String> {
    let payload = serde_json::json!({
        "data": data,
    });

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/client/upload-data-windows", server_url))
        .header("X-Client-Token", short_token)
        .json(&payload)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    let status = resp.status();
    let body = resp.text().map_err(|e| format!("Read response failed: {}", e))?;

    if !status.is_success() {
        return Err(format!("Upload failed: HTTP {}, body: {}", status, body));
    }

    serde_json::from_str(&body)
        .map_err(|e| format!("Parse response failed: {}, body: {}", e, body))
}

/// 检查更新 (同步阻塞版本)
pub fn check_update_blocking(
    server_url: &str,
    short_token: &str,
) -> Result<crate::checkupdate::CheckUpdateResponse, String> {
    let client = Client::new();
    let resp = client
        .get(format!("{}/api/client/check-update", server_url))
        .header("X-Client-Token", short_token)
        // 【关键】携带本地实际运行版本，供后端比对并同步更新记录，避免重复更新死循环
        .header("X-Client-Version", env!("CARGO_PKG_VERSION"))
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    let status = resp.status();
    if !status.is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("Check update failed: HTTP {}, body: {}", status, body));
    }

    resp.json::<crate::checkupdate::CheckUpdateResponse>()
        .map_err(|e| format!("Parse response failed: {}", e))
}

/// 获取加固检查计划 (同步阻塞版本)
pub fn get_check_schedule(
    server_url: &str,
    short_token: &str,
) -> Result<crate::schedule::CheckSchedule, String> {
    let client = Client::new();
    let resp = client
        .get(format!("{}/api/client/check-schedule", server_url))
        .header("X-Client-Token", short_token)
        .send()
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    let status = resp.status();
    if !status.is_success() {
        let body = resp.text().unwrap_or_default();
        return Err(format!("Get check schedule failed: HTTP {}, body: {}", status, body));
    }

    resp.json::<crate::schedule::CheckSchedule>()
        .map_err(|e| format!("Parse response failed: {}", e))
}