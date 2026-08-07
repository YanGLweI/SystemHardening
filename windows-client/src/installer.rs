//! Windows 客户端安装模块 (NSIS exe 静默安装)
//!
//! 设计要点：安装更新必须直接运行 NSIS 安装包，由 NSIS 自行管理服务生命周期
//! （停止旧服务 → 删除旧服务 → 复制文件 → 注册新服务 → 启动服务）。
//! 不能在客户端进程内先停止自身服务，否则服务进程退出会中断安装流程。

use std::fs;
use std::path::PathBuf;
use std::process::Command;

/// 安装更新包 (运行 NSIS 静默安装)
pub fn install_update(temp_exe_path: &str) -> Result<(), String> {
    log::info!("[INSTALLER] Starting installation from: {}", temp_exe_path);
    
    // 1. 【关键】备份配置文件（双保险；NSIS 脚本同样会保留已有配置）
    let config_path = PathBuf::from(r"C:\ProgramData\SystemHardening\WindowsClient\config.yaml");
    let config_backup = format!(r"C:\ProgramData\SystemHardening\WindowsClient\config.backup.{}.yaml", 
        chrono::Local::now().format("%Y%m%d%H%M%S"));
    
    if config_path.exists() {
        match fs::copy(&config_path, &config_backup) {
            Ok(_) => log::info!("[INSTALLER] Config backed up to: {}", config_backup),
            Err(e) => return Err(format!("Backup config failed: {}", e)),
        }
    } else {
        log::warn!("[INSTALLER] Config file not found: {:?}", config_path);
    }
    
    // 2. 【关键】直接运行 NSIS 安装包（静默模式）
    //    注意：不能在服务进程内等待安装完成，否则服务停止时会中断安装。
    //    解决方案：以分离进程方式启动安装器，立即返回，让服务安全退出。
    //    NSIS 脚本已内置：net stop → sc delete → 复制文件 → sc create → net start
    log::info!("[INSTALLER] Running installer in silent mode (NSIS handles service lifecycle)...");
    
    use std::os::windows::process::CommandExt;
    const CREATE_NEW_PROCESS_GROUP: u32 = 0x00000200;
    const DETACHED_PROCESS: u32 = 0x00000008;
    
    let child = Command::new(temp_exe_path)
        .args(&["/S"])
        .creation_flags(CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS)
        .spawn()
        .map_err(|e| format!("Failed to start installer: {}", e))?;
    
    log::info!("[INSTALLER] Installer started as detached process (PID: {}), service will exit now", child.id());
    log::info!("[INSTALLER] ✅ Update initiated! Installation will continue in background.");
    Ok(())
}
