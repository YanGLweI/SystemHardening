//! 待执行任务处理模块（与 Linux 客户端保持一致的 HTTP Polling 模式）

use std::collections::HashMap;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use crate::config::Config;
use crate::token::TokenManager;

/// 待执行任务列表响应
#[derive(Debug, Deserialize)]
pub struct TaskListResponse {
    pub tasks: Vec<CheckTask>,
}

/// 待执行任务
#[derive(Debug, Deserialize)]
pub struct CheckTask {
    pub task_id: String,
    pub client_uuid: String,
    pub status: String,
    #[serde(default)]
    pub issued_at: Option<String>,
    pub created_at: String,
}

/// 上报任务结果请求
#[derive(Debug, Serialize)]
struct TaskResultRequest {
    status: String,
    error_message: Option<String>,
    result_data: Option<HashMap<String, serde_json::Value>>,
}

/// 拉取待执行任务
pub async fn fetch_pending_tasks(server_url: &str, short_token: &str, client_uuid: &str) -> Result<Vec<CheckTask>, String> {
    // 去除尾部斜杠后拼接，避免双斜杠导致 404
    let url = format!("{}/api/client/tasks/pending", server_url.trim_end_matches('/'));
    
    log::info!("[TASK] Fetching pending tasks from: {}", url);
    log::info!("[TASK] Sending X-Client-UUID: {}", client_uuid);
    
    let client = Client::new();
    let resp = client
        .get(&url)
        .header("X-Client-Token", short_token)
        .header("X-Client-UUID", client_uuid)  // ✅ 添加 UUID 请求头
        .send()
        .await
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().await.unwrap_or_default();
        return Err(format!("Get pending tasks failed: HTTP {}, body: {}", status, body));
    }

    let result: TaskListResponse = resp.json()
        .await
        .map_err(|e| format!("Parse response failed: {}", e))?;

    log::info!("[TASK] Fetched {} pending tasks", result.tasks.len());
    Ok(result.tasks)
}

/// 上报任务执行结果
pub async fn submit_task_result(
    server_url: &str, 
    task_id: &str, 
    status: &str, 
    error_message: Option<&str>,
    result_data: Option<HashMap<String, serde_json::Value>>
) -> Result<(), String> {
    let url = format!("{}/api/client/tasks/{}/result", server_url.trim_end_matches('/'), task_id);
    
    let payload = TaskResultRequest {
        status: status.to_string(),
        error_message: error_message.map(|s| s.to_string()),
        result_data,
    };

    let client = Client::new();
    let resp = client
        .put(&url)
        .header("Content-Type", "application/json")
        .json(&payload)
        .send()
        .await
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().await.unwrap_or_default();
        return Err(format!("Submit task result failed: HTTP {}, body: {}", status, body));
    }

    log::info!("[TASK] Result submitted for task {}: status={}", task_id, status);
    Ok(())
}

/// 处理待执行任务
pub async fn process_pending_tasks(config: &Config, token_manager: &TokenManager, _client_uuid: &str) {
    let token = match token_manager.short_token() {
        t if t.is_empty() => {
            log::warn!("[TASK] No token available, skipping task processing");
            return;
        }
        t => t,
    };

    let tasks = match fetch_pending_tasks(&config.server_url, token, _client_uuid).await {
        Ok(tasks) => tasks,
        Err(e) => {
            log::warn!("[TASK] Failed to fetch pending tasks: {}", e);
            return;
        }
    };

    if tasks.is_empty() {
        return;
    }

    for task in tasks {
        process_single_task(config, token_manager, _client_uuid, task).await;
    }
}

/// 处理单个任务
pub async fn process_single_task(
    config: &Config, 
    token_manager: &TokenManager,
    client_uuid: &str,
    task: CheckTask
) {
    use std::time::Instant;
    
    log::info!("[TASK] Processing task {} for client {}", task.task_id, task.client_uuid);
    
    let start_time = Instant::now();
    
    // 1. 更新任务状态为 executing
    if let Err(e) = submit_task_result(
        &config.server_url,
        &task.task_id,
        "executing",
        None,
        None
    ).await {
        log::error!("[TASK] Failed to update task status to executing: {}", e);
        return;
    }
    
    // 2. 执行加固检查（调用现有的 collector 模块）
    let check_output = match run_immediate_check(config, token_manager).await {
        Ok(output) => output,
        Err(e) => {
            log::error!("[TASK] Task {} execution failed: {}", task.task_id, e);
            
            // 3. 上报失败结果
            let _ = submit_task_result(
                &config.server_url,
                &task.task_id,
                "failed",
                Some(&format!("执行失败：{}", e)),
                None
            ).await;
            return;
        }
    };
    
    let duration = start_time.elapsed();
    
    // 3. 上报成功结果
    let mut result_data = HashMap::new();
    result_data.insert("duration_seconds".to_string(), serde_json::json!(duration.as_secs_f64()));
    result_data.insert("output_length".to_string(), serde_json::json!(check_output.len()));
    
    log::info!("[TASK] Task {} completed successfully in {:?}", task.task_id, duration);
    
    let _ = submit_task_result(
        &config.server_url,
        &task.task_id,
        "completed",
        None,
        Some(result_data)
    ).await;
}

/// 执行加固检查（复用现有的 collector 模块）
pub async fn run_immediate_check(config: &Config, token_manager: &TokenManager) -> Result<String, String> {
    use crate::collector;
    
    log::info!("[CHECK] Running immediate security check...");
    
    // 1. 采集系统信息
    match collector::collect_windows_info() {
        Ok(data) => {
            log::info!("[CHECK] Collection completed successfully");
            
            // 2. 上报数据到服务器
            let result = report_check_data(config, token_manager, data).await;
            
            if result.is_ok() {
                return Ok("加固检查数据采集成功".to_string());
            } else {
                return Err(format!("数据采集成功但上报失败：{}", result.unwrap_err()));
            }
        }
        Err(e) => {
            log::error!("[CHECK] Collection failed: {}", e);
            return Err(format!("数据采集失败：{}", e));
        }
    }
}

/// 上报检查数据到服务器
async fn report_check_data(
    config: &Config,
    token_manager: &TokenManager,
    data: crate::models::WindowsSystemCheckData
) -> Result<(), String> {
    let short_token = token_manager.short_token();
    if short_token.is_empty() {
        return Err("No short token available".to_string());
    }

    let url = if config.server_url.ends_with('/') {
        format!("{}/api/client/upload-data-windows", config.server_url.trim_end_matches('/'))
    } else {
        format!("{}/api/client/upload-data-windows", config.server_url)
    };

    let client = Client::new();
    
    // 【关键】后端期望 {"data": {...}} 包裹格式（与每日检查 api::upload_windows_data 保持一致）
    let payload = serde_json::json!({ "data": data });
    let json_body = serde_json::to_string(&payload)
        .map_err(|e| format!("Serialize to JSON failed: {}", e))?;

    let resp = client
        .post(&url)
        .header("Content-Type", "application/json")
        .header("X-Client-Token", short_token)
        .body(json_body)
        .send()
        .await
        .map_err(|e| format!("HTTP request failed: {}", e))?;

    if !resp.status().is_success() {
        let status = resp.status().as_u16();
        let body = resp.text().await.unwrap_or_default();
        return Err(format!("Upload failed: HTTP {}, body: {}", status, body));
    }

    Ok(())
}

/// 启动任务轮询协程
pub async fn spawn_task_poller(config: Config, mut token_manager: TokenManager, client_uuid: String) {
    log::info!("[TASK] Starting task poller (every 5 minutes)...");
    
    loop {
        // 每轮重读 tokens.json：主循环刷新/重注册后本线程及时获取新 Token
        if let Err(e) = token_manager.load() {
            log::warn!("[TASK] 重载 tokens.json 失败: {}", e);
        }
        // client_uuid 实时取值（非空优先），防止复活重注册分配新 UUID 后任务过滤失配
        let current_uuid = if !token_manager.client_uuid().is_empty() {
            token_manager.client_uuid().to_string()
        } else {
            client_uuid.clone()
        };
        process_pending_tasks(&config, &token_manager, &current_uuid).await;
        
        // 等待 5 分钟
        tokio::time::sleep(std::time::Duration::from_secs(5 * 60)).await;
    }
}
