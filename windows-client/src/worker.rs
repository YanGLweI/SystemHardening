//! 客户端业务工作循环（与 Linux 客户端逻辑保持一致：
//! 注册/加载 Token → 心跳（2 分钟）→ 按服务端下发的检查计划执行加固检查）

use std::sync::mpsc::{Receiver, RecvTimeoutError};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use chrono::Local;

use crate::api;
use crate::collector;
use crate::config::Config;
use crate::schedule;
use crate::token::TokenManager;

// 自动更新模块
use crate::checkupdate;

// 待执行任务处理模块
use crate::task_fetch;

/// 心跳间隔：2 分钟（与 Linux 客户端一致）
pub const HEARTBEAT_INTERVAL: Duration = Duration::from_secs(120);
/// 启动宽限期：服务启动后首次检查延迟 60s，避开更新重启后的半就绪窗口
const STARTUP_GRACE: Duration = Duration::from_secs(60);
/// 降级采集/失败后的重试间隔
const DEGRADED_RETRY_INTERVAL: Duration = Duration::from_secs(300);
/// 停止信号轮询间隔（保证服务停止响应及时）
const POLL_INTERVAL: Duration = Duration::from_secs(10);

/// 标记是否已启动版本检查线程
static UPDATE_CHECK_STARTED: AtomicBool = AtomicBool::new(false);

/// 判断错误是否为认证失败（HTTP 401/403）
fn is_auth_error(err: &str) -> bool {
    err.contains("HTTP 401") || err.contains("HTTP 403")
}

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
            
            // 【新增】如果 tokens.json 中没有 hardware_uuid，首次采集并保存
            if token_manager.hardware_uuid().is_empty() {
                log::info!("HardwareUUID 缺失，正在采集...");
                let hw_uuid = collector::collect_hardware_uuid();
                if !hw_uuid.is_empty() {
                    token_manager.set_hardware_uuid(&hw_uuid);
                    // 立即保存，避免心跳发送空值
                    let short_token = token_manager.short_token().to_string();
                    let refresh_token = token_manager.refresh_token().to_string();
                    let expires_at = token_manager.expires_at_str();
                    
                    if let Err(e) = token_manager.save(&short_token, &refresh_token, &expires_at) {
                        log::warn!("保存 hardware_uuid 失败：{}", e);
                    } else {
                        log::info!("✅ 已保存 hardware_uuid: {}", hw_uuid);
                    }
                }
            }
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
            
            // 2. 采集操作系统信息和硬件 UUID
            let os_version = collector::get_os_version();
            let hardware_uuid = collector::collect_hardware_uuid();
                        
            // 3. 注册客户端
            let reg_resp = api::register(
                &config.server_url,
                &temp_resp.temp_token,
                &config.device_name,
                &config.ip_address,
                &os_version,
                env!("CARGO_PKG_VERSION"),
                &hardware_uuid, // 【新增】传递硬件 UUID
            )?;
            
            // 4. 保存 Tokens (先设置 UUID，确保 save 时一并持久化)
            token_manager.set_client_uuid(&reg_resp.client_uuid);
            token_manager.set_hardware_uuid(&hardware_uuid); // 【新增}
            token_manager.save(
                &reg_resp.short_token,
                &reg_resp.refresh_token,
                &reg_resp.expires_at,
            )?;
            log::info!("注册成功：client_uuid={}, hardware_uuid={}", 
                reg_resp.client_uuid, hardware_uuid);
        }
    }
    Ok(())
}

/// 工作主循环：心跳（每 2 分钟）+ 加固检查（按服务端下发的检查计划）
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

    // 启动检查计划拉取线程（每 5 分钟从服务端获取计划）
    let schedule_state = Arc::new(Mutex::new(schedule::ScheduleState::new()));
    let next_check_time = schedule::spawn_schedule_loop(
        config.clone(),
        token_manager.clone(),
        Arc::clone(&schedule_state),
    );

    // 启动任务轮询协程（每 5 分钟检查 pending tasks）
    // 【关键】优先使用注册时服务端分配的真实 UUID（后端按此 UUID 过滤任务）
    let client_uuid = if !token_manager.client_uuid().is_empty() {
        token_manager.client_uuid().to_string()
    } else {
        std::env::var("COMPUTERNAME").unwrap_or_else(|_| "localhost".to_string())
    };
    let task_config = config.clone();
    let task_token_manager = token_manager.clone();
    thread::spawn(move || {
        use tokio::runtime::Runtime;
        if let Ok(rt) = Runtime::new() {
            rt.block_on(async {
                task_fetch::spawn_task_poller(task_config, task_token_manager, client_uuid).await
            });
        }
    });

    // ✅ 启动版本检查线程（首次立即检查 + 每 5 分钟定时检查）
    if !UPDATE_CHECK_STARTED.swap(true, Ordering::SeqCst) {
        log::info!("[UPDATE] Starting automatic update checker...");
        let checkupdate_thread = checkupdate::version_check_loop(config.clone(), token_manager.clone());
        if let Err(e) = checkupdate_thread {
            log::error!("[UPDATE] Failed to start version check loop: {}", e);
        } else {
            log::info!("[UPDATE] Version check thread started successfully");
        }
    } else {
        log::warn!("[UPDATE] Version check already started, skipping duplicate initialization");
    }

    // 首次启动检查与心跳（兜底补检，与 Linux 客户端一致；带启动宽限期）
    let boot_instant = Instant::now();
    let mut last_heartbeat = Instant::now() - HEARTBEAT_INTERVAL;
    let mut initial_check_done = false;
    let mut next_retry: Option<Instant> = None;

    log::info!("工作循环启动：心跳间隔 {}s，检查按服务端计划执行", HEARTBEAT_INTERVAL.as_secs());

    loop {
        // 加固检查：宽限期后首次执行；之后到达服务端计划的检查时刻时执行
        let due = if !initial_check_done {
            boot_instant.elapsed() >= STARTUP_GRACE
                && next_retry.map(|t| Instant::now() >= t).unwrap_or(true)
        } else {
            match *next_check_time.lock().unwrap() {
                Some(t) => Local::now() >= t,
                None => false, // 尚未应用计划，等待拉取线程
            }
        };
        if due {
            let uploaded = run_daily_check(config, token_manager);
            if uploaded {
                initial_check_done = true;
            } else {
                // 降级采集/失败：不标记完成，间隔重试，避免空值覆盖服务端健康数据
                next_retry = Some(Instant::now() + DEGRADED_RETRY_INTERVAL);
                log::warn!("[CHECK] 首次检查未完成（降级采集或采集/上传失败），{}s 后重试", DEGRADED_RETRY_INTERVAL.as_secs());
            }
            schedule::recompute_next_check(&schedule_state, &next_check_time);
        }

        // 心跳
        if last_heartbeat.elapsed() >= HEARTBEAT_INTERVAL {
            log::info!("[HEARTBEAT] 检查点：last_heartbeat={}s ago", last_heartbeat.elapsed().as_secs());
                    
            // 【关键】检测 tokens.json 文件是否被删除，若丢失则自动重新注册（无需重启服务）
            if !token_manager.file_exists() {
                log::warn!("[TOKEN] tokens.json 文件丢失，自动重新注册...");
                match ensure_registered(config, token_manager) {
                    Ok(()) => log::info!("[TOKEN] ✅ 重新注册成功，tokens.json 已重新生成"),
                    Err(e) => log::error!("[TOKEN] 重新注册失败：{}", e),
                }
            } else if let Err(e) = send_heartbeat(config, token_manager) {
                if is_auth_error(&e) {
                    log::warn!("[AUTH] 心跳认证失败，清除本地 Token 并重新注册...");
                    token_manager.clear();
                    match ensure_registered(config, token_manager) {
                        Ok(()) => log::info!("[AUTH] 重新注册成功"),
                        Err(e) => log::error!("[AUTH] 重新注册失败：{}", e),
                    }
                } else {
                    log::error!("[HEARTBEAT] 心跳失败：{}", e);
                }
            } else {
                log::info!("[HEARTBEAT] ✅ 心跳正常");
            }
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

/// 每日加固检查：刷新 Token → 采集 → 降级门禁 → 上传
/// 返回是否成功完成上传；false 时由调用方决定重试（首次检查不标记完成）
fn run_daily_check(config: &Config, token_manager: &mut TokenManager) -> bool {
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
                    return false;
                }
                log::info!("[TOKEN] Token 刷新成功");
            }
            Err(e) => {
                log::error!("[TOKEN] Token 刷新失败: {}（可能需要重新安装客户端）", e);
                return false;
            }
        }
    }

    // 2. 采集 Windows 加固信息
    let data = match collector::collect_windows_info() {
        Ok(d) => d,
        Err(e) => {
            log::error!("[CHECK] 数据采集失败: {}", e);
            return false;
        }
    };

    // 2.5 降级采集门禁：基础信息存在但密码/审计/屏保三大策略组全空时，
    // 视为环境半就绪（如更新重启后数十秒内），跳过本次上传，避免空值覆盖服务端历史健康数据
    if collector::is_degraded_collection(&data) {
        log::warn!("[CHECK] ⚠️ 检测到降级采集：hostname={} 但密码/审计/屏保三组策略全空（环境可能半就绪），跳过本次上传", data.hostname);
        return false;
    }

    // 3. 上传数据
    match api::upload_windows_data(&config.server_url, token_manager.short_token(), &data) {
        Ok(resp) => {
            log::info!("[CHECK] 数据上传成功: record_id={}, status={}", resp.record_id, resp.status);
            true
        }
        Err(e) => {
            log::error!("[CHECK] 数据上传失败: {}", e);
            false
        }
    }
}

/// 发送心跳
fn send_heartbeat(config: &Config, token_manager: &TokenManager) -> Result<(), String> {
    if !token_manager.has_token() {
        log::warn!("[HEARTBEAT] 无可用 Token，跳过心跳");
        return Ok(());
    }
    
    // 【新增】采集当前的设备名和 IP 地址用于心跳上报
    let data = match collector::collect_windows_info() {
        Ok(d) => d,
        Err(e) => {
            log::warn!("采集系统信息失败：{}，使用缓存的设备名和 IP", e);
            // 如果采集失败，继续执行心跳但记录警告
            api::send_heartbeat(
                &config.server_url, 
                token_manager.short_token(), 
                token_manager.hardware_uuid(),
                "unknown",     // 无法采集时使用占位符
                "0.0.0.0",     // 无法采集时使用占位符
            )?;
            log::info!(
                "[HEARTBEAT] 心跳发送成功 (降级模式), hardware_uuid={}",
                token_manager.hardware_uuid()
            );
            return Ok(());
        }
    };
    
    let device_name = data.hostname.clone();
    let ip_address = data.ip.clone();
    
    api::send_heartbeat(
        &config.server_url, 
        token_manager.short_token(), 
        token_manager.hardware_uuid(),
        &device_name,     // 新增参数
        &ip_address,      // 新增参数
    )?;
    
    log::info!(
        "[HEARTBEAT] 心跳发送成功，device_name={}, ip={}, hardware_uuid={}",
        device_name, ip_address, token_manager.hardware_uuid()
    );
    Ok(())
}
