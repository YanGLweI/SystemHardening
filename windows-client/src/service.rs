//! Windows 服务集成（windows-service crate）
//! 服务生命周期：注册控制处理器 → 运行状态 → 工作循环 → 停止

use std::ffi::OsString;
use std::sync::mpsc::channel;
use std::thread;
use std::time::Duration;

use windows_service::service::{
    ServiceControl, ServiceControlAccept, ServiceExitCode, ServiceState, ServiceStatus,
    ServiceType,
};
use windows_service::service_control_handler::{self, ServiceControlHandlerResult};
use windows_service::{define_windows_service, service_dispatcher};

use crate::config::Config;
use crate::token::TokenManager;
use crate::worker;

/// 服务名称（安装/卸载时需与 NSIS 脚本一致）
pub const SERVICE_NAME: &str = "SystemHardeningWinClient";
/// 服务显示名称（供 NSIS 安装脚本使用）
#[allow(dead_code)]
pub const SERVICE_DISPLAY_NAME: &str = "系统加固 Windows 客户端";
/// 服务描述（供 NSIS 安装脚本使用）
#[allow(dead_code)]
pub const SERVICE_DESCRIPTION: &str =
    "采集 Windows 系统加固检查数据并上报管理平台（只读采集，不修改系统配置）";

/// 默认配置文件路径（与 NSIS 安装脚本一致）
pub const DEFAULT_CONFIG_PATH: &str = "C:\\ProgramData\\SystemHardening\\WindowsClient\\config.yaml";

define_windows_service!(ffi_service_main, service_main);

/// 服务主入口（由 Windows 服务控制管理器调用）
fn service_main(arguments: Vec<OsString>) {
    log::info!("服务主入口启动，参数: {:?}", arguments);
    if let Err(e) = run_service() {
        log::error!("服务运行失败: {}", e);
        std::process::exit(1);
    }
}

/// 以 Windows 服务方式启动（由 main 调用）
pub fn run() -> windows_service::Result<()> {
    service_dispatcher::start(SERVICE_NAME, ffi_service_main)
}

/// 服务运行主体
fn run_service() -> Result<(), String> {
    // 1. 先注册服务控制处理器（立即执行，确保 SCM 能收到停止信号）
    let (shutdown_tx, shutdown_rx) = channel::<()>();
    let handler_tx = shutdown_tx.clone();
    let handler = service_control_handler::register(SERVICE_NAME, move |control_event| {
        match control_event {
            ServiceControl::Stop | ServiceControl::Shutdown => {
                log::info!("收到停止/关闭控制事件");
                let _ = handler_tx.send(());
                ServiceControlHandlerResult::NoError
            }
            _ => ServiceControlHandlerResult::NotImplemented,
        }
    })
    .map_err(|e| format!("注册服务控制处理器失败: {}", e))?;

    // 2. 报告服务进入启动中状态（避免 SCM 超时）
    handler
        .set_service_status(ServiceStatus {
            service_type: ServiceType::OWN_PROCESS,
            current_state: ServiceState::StartPending,
            controls_accepted: ServiceControlAccept::STOP | ServiceControlAccept::SHUTDOWN,
            exit_code: ServiceExitCode::Win32(0),
            checkpoint: 0,
            wait_hint: Duration::from_secs(10),
            process_id: None,
        })
        .map_err(|e| format!("设置服务状态失败: {}", e))?;

    // 3. 加载配置（失败不退出进程，改用默认配置并记录错误，保证服务可启动）
    let config = Config::load(DEFAULT_CONFIG_PATH).unwrap_or_else(|e| {
        log::error!("加载配置文件失败（{}）: {}", DEFAULT_CONFIG_PATH, e);
        log::warn!("使用默认配置继续运行，请检查配置文件");
        Config::default()
    });
    log::info!(
        "配置加载：server={}, device={} ({})",
        config.server_url,
        config.device_name,
        config.ip_address
    );

    // 4. 报告服务进入运行状态
    handler
        .set_service_status(ServiceStatus {
            service_type: ServiceType::OWN_PROCESS,
            current_state: ServiceState::Running,
            controls_accepted: ServiceControlAccept::STOP | ServiceControlAccept::SHUTDOWN,
            exit_code: ServiceExitCode::Win32(0),
            checkpoint: 0,
            wait_hint: Duration::default(),
            process_id: None,
        })
        .map_err(|e| format!("设置服务状态失败: {}", e))?;

    log::info!("服务已启动");

    // 5. 在独立线程中运行工作循环（占用唯一 Receiver，收到停止信号后退出）
    let mut config = config;
    let mut token_manager = TokenManager::new(&config.local_db_path);
    let worker = thread::spawn(move || {
        if let Err(e) = worker::worker_loop(&mut config, &mut token_manager, &shutdown_rx) {
            log::error!("工作循环异常退出: {}", e);
            // 注册/初始化失败时通知控制处理器线程结束
            let _ = shutdown_tx.send(());
        }
    });

    // 等待工作线程退出（收到停止信号或异常退出）
    let _ = worker.join();

    // 报告服务停止
    let _ = handler.set_service_status(ServiceStatus {
        service_type: ServiceType::OWN_PROCESS,
        current_state: ServiceState::Stopped,
        controls_accepted: ServiceControlAccept::empty(),
        exit_code: ServiceExitCode::Win32(0),
        checkpoint: 0,
        wait_hint: Duration::default(),
        process_id: None,
    });

    log::info!("服务已停止");
    Ok(())
}
