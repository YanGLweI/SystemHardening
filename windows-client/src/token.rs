use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::Path;

/// Token 持久化数据
#[derive(Debug, Serialize, Deserialize)]
struct TokenData {
    short_token: String,
    refresh_token: String,
    expires_at: String,
}

/// Token 管理器（使用 JSON 文件持久化）
#[derive(Clone)]
pub struct TokenManager {
    db_path: String,
    short_token: String,
    refresh_token: String,
    expires_at: DateTime<Utc>,
}

impl TokenManager {
    pub fn new(db_path: &str) -> Self {
        Self {
            db_path: db_path.to_string(),
            short_token: String::new(),
            refresh_token: String::new(),
            expires_at: DateTime::default(),
        }
    }

    /// 从文件加载 Token
    pub fn load(&mut self) -> Result<(), String> {
        let path = Path::new(&self.db_path);
        if !path.exists() {
            return Err("Token file not found".to_string());
        }
        let content = fs::read_to_string(path).map_err(|e| format!("Read failed: {}", e))?;
        let data: TokenData =
            serde_json::from_str(&content).map_err(|e| format!("Parse failed: {}", e))?;

        self.short_token = data.short_token;
        self.refresh_token = data.refresh_token;
        self.expires_at = data
            .expires_at
            .parse::<DateTime<Utc>>()
            .map_err(|e| format!("Parse expires_at failed: {}", e))?;

        Ok(())
    }

    /// 保存 Token 到文件
    pub fn save(&mut self, short_token: &str, refresh_token: &str, expires_at: &str) -> Result<(), String> {
        let data = TokenData {
            short_token: short_token.to_string(),
            refresh_token: refresh_token.to_string(),
            expires_at: expires_at.to_string(),
        };

        let json = serde_json::to_string_pretty(&data).map_err(|e| format!("Serialize failed: {}", e))?;

        // 确保目录存在
        if let Some(parent) = Path::new(&self.db_path).parent() {
            fs::create_dir_all(parent).map_err(|e| format!("Create dir failed: {}", e))?;
        }

        fs::write(&self.db_path, &json).map_err(|e| format!("Write failed: {}", e))?;

        self.short_token = short_token.to_string();
        self.refresh_token = refresh_token.to_string();
        self.expires_at = expires_at
            .parse::<DateTime<Utc>>()
            .map_err(|e| format!("Parse expires_at failed: {}", e))?;

        log::info!("Tokens saved to {}", self.db_path);
        Ok(())
    }

    /// 获取短 Token
    pub fn short_token(&self) -> &str {
        &self.short_token
    }

    /// 获取刷新 Token
    pub fn refresh_token(&self) -> &str {
        &self.refresh_token
    }

    /// 检查 Token 是否过期（提前 24 小时预警）
    pub fn is_expired(&self) -> bool {
        if self.expires_at.timestamp() == 0 {
            return true;
        }
        let now = Utc::now();
        // 剩余时间少于 24 小时视为即将过期
        self.expires_at < now || (self.expires_at - now).num_hours() < 24
    }

    /// 是否有 Token
    pub fn has_token(&self) -> bool {
        !self.short_token.is_empty()
    }

    /// 清除本地 Token（删除文件并重置内存状态）
    pub fn clear(&mut self) {
        self.short_token = String::new();
        self.refresh_token = String::new();
        self.expires_at = DateTime::default();

        let path = Path::new(&self.db_path);
        if path.exists() {
            if let Err(e) = fs::remove_file(path) {
                log::warn!("删除 Token 文件失败: {}", e);
            } else {
                log::info!("已清除本地 Token 文件: {}", self.db_path);
            }
        }
    }
}