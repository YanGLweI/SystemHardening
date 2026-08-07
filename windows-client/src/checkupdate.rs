//! Windows 客户端版本检查模块

use crate::{api, config::Config, token::TokenManager, downloader, installer};
use std::thread;
use std::time::Duration;

/// 检查更新响应结构 (与服务端 API 匹配)
#[derive(Debug, serde::Deserialize)]
pub struct CheckUpdateResponse {
    pub has_update: bool,
    pub current_version: String,
    pub new_version: String,
    pub download_url: String,
    pub hash: String,
    #[serde(default)]
    pub size: i64,
    pub filename: String,
}

/// 版本检查间隔：5 分钟
const UPDATE_CHECK_INTERVAL: Duration = Duration::from_secs(5 * 60);

/// 检查是否有新版本
pub fn check_for_update(config: &Config, token_manager: &TokenManager) -> Result<CheckUpdateResponse, String> {
    log::info!("[UPDATE] Checking for new version...");
    
    let short_token = token_manager.short_token();
    if short_token.is_empty() {
        return Err("Failed to get short token".to_string());
    }
    
    api::check_update_blocking(&config.server_url, short_token)
}

/// 版本检查循环（非阻塞 - 立即启动后台线程）
pub fn version_check_loop(
    config: Config,
    token_manager: TokenManager,
) -> Result<(), String> {
    log::info!("[UPDATE] Starting automatic update checker...");
    
    // 启动后台定时线程（首次立即检查 + 每 5 分钟定时检查）
    let config_clone = config.clone();
    let token_manager_clone = token_manager.clone();
    
    thread::spawn(move || {
        // 1. 立即执行第一次检查
        log::info!("[UPDATE] Performing initial version check...");
        match run_check_and_install(&config_clone, &token_manager_clone) {
            Ok(true) => log::info!("[UPDATE] Initial update installed successfully"),
            Ok(false) => {} // 没有更新或失败，继续循环
            Err(e) => log::warn!("[UPDATE] Initial check failed: {}", e),
        }
        
        let interval = UPDATE_CHECK_INTERVAL;
        
        loop {
            // 等待下一个检查时间
            std::thread::sleep(interval);
            
            log::info!("[UPDATE] Performing scheduled version check...");
            
            match run_check_and_install(&config_clone, &token_manager_clone) {
                Ok(true) => log::info!("[UPDATE] Update installed successfully"),
                Ok(false) => {} // 没有更新或失败，继续循环
                Err(e) => log::warn!("[UPDATE] Check failed: {}", e),
            }
        }
    });
    
    Ok(())
}

/// 运行检查并安装更新
fn run_check_and_install(
    config: &Config,
    token_manager: &TokenManager,
) -> Result<bool, String> {
    let response = match check_for_update(config, token_manager) {
        Ok(resp) => resp,
        Err(e) => {
            log::warn!("[UPDATE] Check failed: {}", e);
            return Ok(false);
        }
    };
    
    if !response.has_update {
        log::info!("[UPDATE] No new version available. Current: {}", response.current_version);
        return Ok(false);
    }
    
    log::info!("[UPDATE] New version found: {} -> {}", 
        response.current_version, response.new_version);

    // 【关键】本地防护：后端返回的目标版本与本地实际版本一致时跳过，避免重复更新死循环
    if response.new_version == env!("CARGO_PKG_VERSION") {
        log::info!("[UPDATE] Target version equals local version {}, skip installation", env!("CARGO_PKG_VERSION"));
        return Ok(false);
    }
    
    // 下载更新包
    let temp_path = match downloader::download_update(&response.download_url, &response.filename, &response.hash) {
        Ok(path) => path,
        Err(e) => {
            log::error!("[UPDATE] Download failed: {}", e);
            return Ok(false);
        }
    };
    
    log::info!("[UPDATE] Downloaded to: {}", temp_path);
    
    // 安装更新
    match installer::install_update(&temp_path) {
        Ok(_) => {
            log::info!("[UPDATE] ✅ Update completed!");
            Ok(true)
        }
        Err(e) => {
            log::error!("[UPDATE] Installation failed: {}", e);
            Ok(false)
        }
    }
}
