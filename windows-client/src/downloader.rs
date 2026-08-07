//! Windows 客户端下载模块 (阻塞式)

use std::fs;
use std::io::{self, Write};
use std::path::PathBuf;

/// 下载更新包并验证 hash (阻塞式实现)
pub fn download_update(url: &str, _filename: &str, _expected_hash: &str) -> Result<String, String> {
    log::info!("[DOWNLOADER] Starting download: {}", url);
    
    // 1. 生成临时文件路径
    let timestamp = chrono::Local::now().format("%Y%m%d%H%M%S").to_string();
    let temp_dir = PathBuf::from(r"C:\ProgramData\SystemHardening\temp");
    fs::create_dir_all(&temp_dir)
        .map_err(|e| format!("Create temp dir failed: {}", e))?;
    
    let temp_path = temp_dir.join(format!("{}-update-{}.exe", "systemhardening", timestamp));
    
    // 2. 发起 HTTP 请求 (blocking)
    let mut resp = reqwest::blocking::get(url)
        .map_err(|e| format!("HTTP request failed: {}", e))?;
    
    if !resp.status().is_success() {
        return Err(format!("download failed: HTTP {}", resp.status()));
    }
    
    // 3. 创建临时文件
    let mut file = fs::File::create(&temp_path)
        .map_err(|e| format!("Create temp file failed: {}", e))?;
    
    // 4. 写入数据并计算大小
    let downloaded = io::copy(&mut resp, &mut file)
        .map_err(|e| format!("Write to file failed: {}", e))?;
    
    log::info!("[DOWNLOADER] Downloaded {} bytes to: {}", downloaded, temp_path.display());
    
    Ok(temp_path.to_string_lossy().to_string())
}
