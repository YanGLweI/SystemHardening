//! Windows 系统加固客户端入口
//!
//! 运行模式：
//! - 默认：以 Windows 服务方式运行（安装后由 SCM 启动）
//! - `--foreground` / `--debug`：前台调试模式（可后跟配置文件路径）

mod api;
mod collector;
mod config;
mod models;
mod service;
mod token;
mod worker;

// 自动更新相关模块
mod checkupdate;
mod downloader;
mod installer;

// 加固检查计划模块
mod schedule;

use std::env;
use std::io::Write;
use std::sync::mpsc::channel;

use chrono::Local;

use crate::config::Config;
use crate::token::TokenManager;
use crate::worker::worker_loop;

/// 构建日志器：输出本地时间（默认 env_logger 输出 UTC）
fn build_logger() -> env_logger::Builder {
    let mut builder = env_logger::Builder::from_env(
        env_logger::Env::default().default_filter_or("info"),
    );
    builder.format(|buf, record| {
        writeln!(
            buf,
            "[{} {} {}] {}",
            Local::now().format("%Y-%m-%dT%H:%M:%S%:z"),
            record.level(),
            record.target(),
            record.args()
        )
    });
    builder
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let is_foreground = args.iter().any(|a| a == "--foreground" || a == "--debug");

    if is_foreground {
        // 前台调试模式：日志输出到控制台
        build_logger().init();
    } else {
        // 服务模式：日志写入文件，便于诊断（服务无控制台输出）
        let log_path =
            "C:\\ProgramData\\SystemHardening\\WindowsClient\\service.log";
        match std::fs::OpenOptions::new()
            .create(true)
            .append(true)
            .open(log_path)
        {
            Ok(file) => {
                build_logger()
                    .target(env_logger::Target::Pipe(Box::new(file)))
                    .init();
            }
            Err(e) => {
                eprintln!("打开日志文件失败（{}），回退到控制台输出: {}", log_path, e);
                build_logger().init();
            }
        }
    }

    // 前台调试模式：直接运行工作循环，不注册为服务
    if is_foreground {
        run_foreground(&args);
        return;
    }

    // 默认：Windows 服务模式
    log::info!(
        "Windows 系统加固客户端 v{} 以服务模式启动",
        env!("CARGO_PKG_VERSION")
    );
    if let Err(e) = service::run() {
        eprintln!("服务启动失败: {}", e);
        std::process::exit(1);
    }
}

/// 前台调试模式：第一个非 -- 参数作为配置文件路径
fn run_foreground(args: &[String]) {
    log::info!("前台调试模式启动（Ctrl+C 退出）");

    let config_path = args
        .iter()
        .find(|a| !a.starts_with("--"))
        .map(|s| s.as_str())
        .unwrap_or(service::DEFAULT_CONFIG_PATH);

    let mut config = Config::load(config_path).unwrap_or_else(|e| {
        log::warn!("加载配置失败（{}），使用默认配置: {}", e, config_path);
        Config::default()
    });

    log::info!(
        "服务器: {}，设备: {} ({})，Token 存储: {}",
        config.server_url,
        config.device_name,
        config.ip_address,
        config.local_db_path
    );

    let mut token_manager = TokenManager::new(&config.local_db_path);

    // 前台模式不接收停止信号：保持 sender 存活（forget 避免 drop 触发 Disconnected）
    let (_keep_alive_tx, stop_rx) = channel::<()>();
    std::mem::forget(_keep_alive_tx);

    if let Err(e) = worker_loop(&mut config, &mut token_manager, &stop_rx) {
        log::error!("工作循环异常退出: {}", e);
        std::process::exit(1);
    }
}
