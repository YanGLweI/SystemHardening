use serde::Deserialize;

/// 客户端配置
#[derive(Debug, Clone, Deserialize, Default)]
pub struct Config {
    #[serde(default = "default_server_url")]
    pub server_url: String,
    #[serde(default = "default_local_db_path")]
    pub local_db_path: String,
    #[serde(default = "default_device_name")]
    pub device_name: String,
    #[serde(default = "default_ip_address")]
    pub ip_address: String,
}

fn default_server_url() -> String {
    "http://localhost:8080".to_string()
}
fn default_local_db_path() -> String {
    "C:\\ProgramData\\SystemHardening\\WindowsClient\\tokens.json".to_string()
}
fn default_device_name() -> String {
    "localhost".to_string()
}
fn default_ip_address() -> String {
    "127.0.0.1".to_string()
}

impl Config {
    pub fn load(path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let content = std::fs::read_to_string(path)?;
        let config: Config = serde_yaml::from_str(&content)?;
        Ok(config)
    }
}