//! 客户端业务工作循环（与 Linux 客户端逻辑保持一致：
//! 注册/加载 Token → 心跳（2 分钟）→ 每日加固检查（24 小时））

use std::sync::mpsc::{Receiver, RecvTimeoutError};
use std::time::{Duration, Instant};

use crate::api;
use crate::collector;
use crate::config::Config;
use crate::token::TokenManager;

/// 心跳间隔：2 分钟（与 Linux 客户端一致）
pub const HEARTBEAT_INTERVAL: Duration = Duration::from_secs(120);
/// 加固检查间隔：24 小时（与 Linux 客户端一致）
pub const CHECK_INTERVAL: Duration = Duration::from_secs(24 * 60 * 60);
/// 停止信号轮询间隔（保证服务停止响应及时）
const POLL_INTERVAL: Duration = Duration::from_secs(10);

/// 初始化：填充缺失的设备信息，加载现有 Tokens，不存在则走注册流程
pub fn ensure_registered(config: &mut Config, token_manager: &mut TokenManager) -> Result<(), String> {
    // 配置为空时自动采集设备信息
    if config.device_name.is_empty() {
        config.device_name = std::env::var("COMPUTERNAME").unwrap_or_else(|_| "localhost".to_string());
        log::info!("自动采集设备名: {}", config.device_name);
    }
    if config.ip_address.is_empty() {
        config.ip_address = collector::get_ip_address();
        if config.ip_address.is_empty() {
            config.ip_address = "127.0.0.1".to_string();
        }
        log::info!("自动采集 IP 地址: {}", config.ip_address);
    }

    match token_manager.load() {
        Ok(()) => {
            log::info!("已从本地加载现有 Tokens");
        }
        Err(e) => {
            log::info!("没有现有 Tokens（{}），开始注册流程...", e);

            // 1. 请求临时 Token
            let temp_resp = api::request_temp_token(
                &config.server_url,
                &config.device_name,
                &config.ip_address,
            )?;
            log::info!("获取临时 Token 成功");

            // 2. 注册客户端
            let os_version = collector::get_os_version();
            let reg_resp = api::register(
                &config.server_url,
                &temp_resp.temp_token,
                &config.device_name,
                &config.ip_address,
                &os_version,
            )?;

            // 3. 保存 Tokens
            token_manager.save(
                &reg_resp.short_token,
                &reg_resp.refresh_token,
                &reg_resp.expires_at,
            )?;
            log::info!("注册成功: client_uuid={}", reg_resp.client_uuid);
        }
    }
    Ok(())
}

/// 工作主循环：心跳（每 2 分钟）+ 每日加固检查（每 24 小时）
/// 收到停止信号（Ok 或 Disconnected）时退出
pub fn worker_loop(
    config: &mut Config,
    token_manager: &mut TokenManager,
    stop_rx: &Receiver<()>,
) -> Result<(), String> {
    // 初始化：加载或注册 Tokens（失败时循环重试，保证服务保持运行状态）
    loop {
        match ensure_registered(config, token_manager) {
            Ok(()) => break,
            Err(e) => {
                log::error!("注册/初始化失败: {}，60 秒后重试...", e);
                // 等待期间响应停止信号
                match stop_rx.recv_timeout(Duration::from_secs(60)) {
                    Ok(()) | Err(RecvTimeoutError::Disconnected) => {
                        log::info!("收到停止信号，退出注册重试循环");
                        return Ok(());
                    }
                    Err(RecvTimeoutError::Timeout) => {}
                }
            }
        }
    }

    // 首次启动立即执行检查与心跳（与 Linux 客户端一致）
    let mut last_heartbeat = Instant::now() - HEARTBEAT_INTERVAL;
    let mut last_check = Instant::now() - CHECK_INTERVAL;

    log::info!(
        "工作循环启动：心跳间隔 {}s，检查间隔 {}s",
        HEARTBEAT_INTERVAL.as_secs(),
        CHECK_INTERVAL.as_secs()
    );

    loop {
        // 每日加固检查
        if last_check.elapsed() >= CHECK_INTERVAL {
            run_daily_check(config, token_manager);
            last_check = Instant::now();
        }

        // 心跳
        if last_heartbeat.elapsed() >= HEARTBEAT_INTERVAL {
            send_heartbeat(config, token_manager);
            last_heartbeat = Instant::now();
        }

        // 等待停止信号（轮询间隔 10 秒，保证停止响应及时）
        match stop_rx.recv_timeout(POLL_INTERVAL) {
            Ok(()) | Err(RecvTimeoutError::Disconnected) => {
                log::info!("收到停止信号，工作循环退出");
                break;
            }
            Err(RecvTimeoutError::Timeout) => {}
        }
    }

    Ok(())
}

/// 每日加固检查：刷新 Token → 采集 → 上传
fn run_daily_check(config: &Config, token_manager: &mut TokenManager) {
    log::info!("[CHECK] 开始每日加固检查...");

    // 1. Token 过期检查与刷新
    if token_manager.is_expired() {
        log::info!("[TOKEN] Token 即将过期，尝试刷新...");
        match api::refresh_token(&config.server_url, token_manager.refresh_token()) {
            Ok(resp) => {
                // 刷新后 refresh_token 不变，仅更新 short_token 和过期时间
                let current_short = token_manager.short_token().to_string();
                if let Err(e) = token_manager.save(&current_short, &resp.short_token, &resp.expires_at) {
                    log::error!("[TOKEN] 保存刷新结果失败: {}", e);
                    return;
                }
                log::info!("[TOKEN] Token 刷新成功");
            }
            Err(e) => {
                log::error!("[TOKEN] Token 刷新失败: {}（可能需要重新安装客户端）", e);
                return;
            }
        }
    }

    // 2. 采集 Windows 加固信息
    let data = match collector::collect_windows_info() {
        Ok(d) => d,
        Err(e) => {
            log::error!("[CHECK] 数据采集失败: {}", e);
            return;
        }
    };

    // 3. 上传数据
    match api::upload_windows_data(&config.server_url, token_manager.short_token(), &data) {
        Ok(resp) => {
            log::info!("[CHECK] 数据上传成功: record_id={}, status={}", resp.record_id, resp.status);
        }
        Err(e) => {
            log::error!("[CHECK] 数据上传失败: {}", e);
        }
    }
}

/// 发送心跳
fn send_heartbeat(config: &Config, token_manager: &TokenManager) {
    if !token_manager.has_token() {
        log::warn!("[HEARTBEAT] 无可用 Token，跳过心跳");
        return;
    }
    match api::send_heartbeat(&config.server_url, token_manager.short_token()) {
        Ok(resp) => log::info!("[HEARTBEAT] 心跳发送成功: status={}", resp.status),
        Err(e) => log::warn!("[HEARTBEAT] 心跳发送失败: {}", e),
    }
}
