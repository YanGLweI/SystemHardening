use std::collections::HashMap;
use std::path::PathBuf;
use std::process::Command;

use chrono::Local;
use regex::Regex;
use serde::Deserialize;
use winreg::enums::*;
use winreg::RegKey;
use wmi::WMIConnection;

use crate::models::WindowsSystemCheckData;

// ==================== WMI 查询结构体 ====================

// WMI 查询结构体命名必须与 WMI 类名/属性名一致，忽略命名风格警告
#[derive(Deserialize)]
#[allow(dead_code, non_camel_case_types, non_snake_case)]
struct Win32_OperatingSystem {
    Caption: Option<String>,
    Version: Option<String>,
}

#[derive(Deserialize)]
#[allow(dead_code, non_camel_case_types, non_snake_case)]
struct Win32_ComputerSystem {
    Domain: Option<String>,
    DNSHostName: Option<String>,
}

#[derive(Deserialize)]
#[allow(dead_code, non_camel_case_types, non_snake_case)]
struct Win32_NetworkAdapterConfiguration {
    IPAddress: Option<Vec<String>>,
    IPEnabled: Option<bool>,
    DefaultIPGateway: Option<Vec<String>>,
    IPSubnet: Option<Vec<String>>,
}

#[derive(Deserialize)]
#[allow(dead_code, non_camel_case_types, non_snake_case)]
struct Win32_UserAccount {
    Name: Option<String>,
    SID: Option<String>,
    Disabled: Option<bool>,
}

#[derive(Deserialize)]
#[allow(dead_code, non_camel_case_types, non_snake_case)]
struct SoftwareLicensingProduct {
    LicenseStatus: Option<i32>,
}

// ==================== 主采集函数 ====================

/// 获取操作系统版本字符串（注册客户端时使用）
pub fn get_os_version() -> String {
    let wmi_con = match WMIConnection::new() {
        Ok(con) => con,
        Err(e) => {
            log::warn!("WMI 初始化失败: {}", e);
            return String::new();
        }
    };

    match wmi_con.query::<Win32_OperatingSystem>() {
        Ok(results) => {
            if let Some(os) = results.first() {
                let caption = os.Caption.as_deref().unwrap_or("");
                let version = os.Version.as_deref().unwrap_or("");
                if !version.is_empty() {
                    return format!("{} {}", caption, version);
                }
                return caption.to_string();
            }
            String::new()
        }
        Err(e) => {
            log::warn!("操作系统版本查询失败: {}", e);
            String::new()
        }
    }
}

// ==================== 辅助函数 ====================

/// 将 IPv4 地址字符串转为 u32
fn ip_to_u32(ip: &str) -> Option<u32> {
    let parts: Vec<&str> = ip.split('.').collect();
    if parts.len() != 4 {
        return None;
    }
    let mut result: u32 = 0;
    for part in &parts {
        let octet: u32 = part.parse().ok()?;
        if octet > 255 {
            return None;
        }
        result = (result << 8) | octet;
    }
    Some(result)
}

/// 计算网络地址：ip & mask
fn network_address(ip: u32, mask: u32) -> u32 {
    ip & mask
}

/// 获取本机主 IP 地址（优先取与默认网关同网段的 IP）
pub fn get_ip_address() -> String {
    let wmi_con = match WMIConnection::new() {
        Ok(con) => con,
        Err(e) => {
            log::warn!("WMI 初始化失败: {}", e);
            return String::new();
        }
    };

    match wmi_con.query::<Win32_NetworkAdapterConfiguration>() {
        Ok(results) => {
            let mut fallback_ip = String::new();
            for adapter in &results {
                if !adapter.IPEnabled.unwrap_or(false) {
                    continue;
                }
                // 没有网关的适配器跳过（不是主网卡）
                let gateways = match &adapter.DefaultIPGateway {
                    Some(gw) if !gw.is_empty() => gw,
                    _ => continue,
                };
                let gateway = &gateways[0];
                let Some(gw_ip) = ip_to_u32(gateway) else { continue };

                let ips = match &adapter.IPAddress {
                    Some(ips) => ips,
                    None => continue,
                };
                let subnets = adapter.IPSubnet.as_ref();

                // 遍历该适配器的所有 IP，找与网关同网段的那个
                for (i, ip_str) in ips.iter().enumerate() {
                    if ip_str.starts_with("127.") || !ip_str.contains('.') {
                        continue;
                    }
                    let Some(ip_val) = ip_to_u32(ip_str) else { continue };

                    // 尝试用对应索引的子网掩码计算
                    let mask_val = if let Some(subs) = subnets {
                        if i < subs.len() {
                            ip_to_u32(&subs[i])
                        } else {
                            None
                        }
                    } else {
                        None
                    };

                    // 如果有子网掩码，验证是否真的与网关同网段
                    if let Some(mask) = mask_val {
                        if network_address(ip_val, mask) == network_address(gw_ip, mask) {
                            return ip_str.clone();
                        }
                    } else {
                        // 没有子网掩码信息时，作为备选
                        if fallback_ip.is_empty() {
                            fallback_ip = ip_str.clone();
                        }
                    }
                }
            }
            fallback_ip
        }
        Err(e) => {
            log::warn!("IP 地址查询失败: {}", e);
            String::new()
        }
    }
}

/// 采集 Windows 加固信息
pub fn collect_windows_info() -> Result<WindowsSystemCheckData, String> {
    log::info!("Starting Windows hardening info collection...");

    let wmi_con = WMIConnection::new()
        .map_err(|e| format!("WMI init failed: {}", e))?;

    let mut data = WindowsSystemCheckData::default();

    // 1. 基本信息
    collect_system_info(&wmi_con, &mut data);
    collect_network_info(&wmi_con, &mut data);
    collect_license_info(&wmi_con, &mut data);

    // 2. 密码策略（secedit）
    collect_password_policy(&mut data);

    // 3. 审计策略（注册表）
    collect_audit_policy(&mut data);

    // 4. 设备控制（注册表）
    collect_device_control(&mut data);

    // 5. 屏幕保护（registry.pol → SYSVOL → HKU → 注册表）
    collect_screen_saver(&mut data);

    // 6. 管理员/来宾账户（WMI 实际状态，覆盖 GPO 配置值）
    collect_admin_accounts(&wmi_con, &mut data);

    // 7. 日期和时间
    data.date = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
    data.client_version = env!("CARGO_PKG_VERSION").to_string();

    log::info!("Collection completed: hostname={}, domain={}, ip={}", data.hostname, data.domainname, data.ip);
    Ok(data)
}

// ==================== 采集子函数 ====================

/// 采集系统基本信息（hostname, OS, domain）
fn collect_system_info(wmi: &WMIConnection, data: &mut WindowsSystemCheckData) {
    // 主机名
    if let Ok(name) = std::env::var("COMPUTERNAME") {
        data.hostname = name;
    }

    // 操作系统信息
    if let Ok(results) = wmi.query::<Win32_OperatingSystem>() {
        if let Some(os) = results.first() {
            let caption = os.Caption.as_deref().unwrap_or("");
            let version = os.Version.as_deref().unwrap_or("");
            data.operasystem = if !version.is_empty() {
                format!("{} {}", caption, version)
            } else {
                caption.to_string()
            };
        }
    }

    // 域信息
    if let Ok(results) = wmi.query::<Win32_ComputerSystem>() {
        if let Some(cs) = results.first() {
            data.domainname = cs.Domain.as_deref().unwrap_or("").to_string();
            if data.hostname.is_empty() {
                data.hostname = cs.DNSHostName.as_deref().unwrap_or("").to_string();
            }
        }
    }
}

/// 采集网络信息（IP 地址，优先取与默认网关同网段的 IP）
fn collect_network_info(wmi: &WMIConnection, data: &mut WindowsSystemCheckData) {
    if let Ok(results) = wmi.query::<Win32_NetworkAdapterConfiguration>() {
        let mut fallback_ip = String::new();
        for adapter in &results {
            if !adapter.IPEnabled.unwrap_or(false) {
                continue;
            }
            // 没有网关的适配器跳过
            let gateways = match &adapter.DefaultIPGateway {
                Some(gw) if !gw.is_empty() => gw,
                _ => continue,
            };
            let gateway = &gateways[0];
            let Some(gw_ip) = ip_to_u32(gateway) else { continue };

            let ips = match &adapter.IPAddress {
                Some(ips) => ips,
                None => continue,
            };
            let subnets = adapter.IPSubnet.as_ref();

            for (i, ip_str) in ips.iter().enumerate() {
                if ip_str.starts_with("127.") || !ip_str.contains('.') {
                    continue;
                }
                let Some(ip_val) = ip_to_u32(ip_str) else { continue };

                let mask_val = if let Some(subs) = subnets {
                    if i < subs.len() {
                        ip_to_u32(&subs[i])
                    } else {
                        None
                    }
                } else {
                    None
                };

                if let Some(mask) = mask_val {
                    if network_address(ip_val, mask) == network_address(gw_ip, mask) {
                        data.ip = ip_str.clone();
                        return;
                    }
                } else if fallback_ip.is_empty() {
                    fallback_ip = ip_str.clone();
                }
            }
        }
        if !fallback_ip.is_empty() {
            data.ip = fallback_ip;
        }
    }
}

/// 采集许可证状态
fn collect_license_info(wmi: &WMIConnection, data: &mut WindowsSystemCheckData) {
    // WMI 查询：获取已激活的 Windows 许可证状态
    let wql = "SELECT LicenseStatus FROM SoftwareLicensingProduct WHERE PartialProductKey IS NOT NULL";
    if let Ok(results) = wmi.raw_query::<SoftwareLicensingProduct>(wql) {
        if let Some(prod) = results.first() {
            let status = prod.LicenseStatus.unwrap_or(0);
            data.license_result = status.to_string();
            // LicenseStatus: 1=已授权(已激活), 其他=未激活
        }
    }
    if data.license_result.is_empty() {
        data.license_result = "0".to_string();
    }
}

/// 采集密码策略（secedit /export 导出本地安全策略）
/// 注意：secedit 导出的 INF 文件为 UTF-16 LE 编码，必须按字节读取并解码
fn collect_password_policy(data: &mut WindowsSystemCheckData) {
    let temp_dir = std::env::temp_dir();
    let secpol_path = temp_dir.join("secpol_win.inf");

    // 执行 secedit /export（只导出文件，不通过管道输出，避免编码问题）
    let output = Command::new("cmd")
        .args([
            "/c",
            &format!(
                "secedit /export /areas SECURITYPOLICY /cfg \"{}\" /quiet",
                secpol_path.display()
            ),
        ])
        .output();

    match output {
        Ok(out) if out.status.success() => {
            // 直接读取文件字节并解码（UTF-16 LE with BOM）
            match std::fs::read(&secpol_path) {
                Ok(bytes) => {
                    let content = decode_text(&bytes);
                    parse_secedit_output(&content, data);
                }
                Err(e) => log::warn!("读取 secedit 导出文件失败: {}", e),
            }
            // 清理临时文件
            let _ = std::fs::remove_file(&secpol_path);
        }
        Ok(out) => {
            log::warn!(
                "secedit /export 失败（stdout: {}），降级解析 GptTmpl.inf",
                String::from_utf8_lossy(&out.stdout).trim().chars().take(200).collect::<String>()
            );
            collect_password_policy_from_gpttmpl(data);
        }
        Err(e) => {
            log::warn!("secedit 执行失败: {}，降级解析 GptTmpl.inf", e);
            collect_password_policy_from_gpttmpl(data);
        }
    }
}

/// 降级方案：解析 GPO 的 GptTmpl.inf 获取密码策略配置值
/// 来源：本地组策略、GPO DataStore 缓存、域 SYSVOL 源文件（UTF-16 LE 编码）
/// secedit 在服务上下文（Session 0）中可能失败，GptTmpl.inf 是 GPO 应用的实际配置
fn collect_password_policy_from_gpttmpl(data: &mut WindowsSystemCheckData) {
    let mut merged: HashMap<String, String> = HashMap::new();

    for file in find_gpttmpl_files(&data.domainname) {
        match std::fs::read(&file) {
            Ok(bytes) => {
                let content = decode_text(&bytes);
                if let Some(map) = parse_system_access_section(&content) {
                    log::info!("从 GptTmpl.inf 解析密码策略: {}", file.display());
                    for (k, v) in map {
                        merged.insert(k, v);
                    }
                }
            }
            Err(e) => log::warn!("读取 GptTmpl.inf 失败 ({}): {}", file.display(), e),
        }
    }

    if merged.is_empty() {
        log::warn!("未找到可用的 GptTmpl.inf，密码策略留空");
        return;
    }
    apply_password_policy_map(&merged, data);
}

/// 查找所有可用的 GptTmpl.inf 文件（本地策略 → DataStore 缓存 → SYSVOL 源）
fn find_gpttmpl_files(domain: &str) -> Vec<PathBuf> {
    find_policy_files(domain, "GptTmpl.inf")
}

/// 查找所有可用的 Registry.pol 文件（DataStore 缓存 → SYSVOL 源）
fn find_registry_pol_files(domain: &str) -> Vec<PathBuf> {
    let mut files = find_policy_files(domain, "Registry.pol");
    // 本地用户策略缓存
    files.push(PathBuf::from(r"C:\Windows\System32\GroupPolicy\User\registry.pol"));
    files
}

/// 按文件名递归查找策略文件（本地 → DataStore 缓存 → SYSVOL 源）
fn find_policy_files(domain: &str, file_name: &str) -> Vec<PathBuf> {
    let mut files = Vec::new();
    let mut roots = vec![r"C:\Windows\System32\GroupPolicy".to_string()];
    // 域环境：SYSVOL 上的 GPO 源文件（SYSTEM 域成员可读）
    if !domain.is_empty()
        && !domain.eq_ignore_ascii_case("WORKGROUP")
        && !domain.eq_ignore_ascii_case("WORKSTATION")
    {
        roots.push(format!(r"\\{}\SysVol\{}\Policies", domain, domain));
    }

    for root in roots {
        collect_policy_files_recursive(&root, file_name, &mut files, 0);
    }
    files
}

/// 递归查找指定文件名的策略文件（限制深度，避免 SYSVOL 大目录过慢）
fn collect_policy_files_recursive(dir: &str, file_name: &str, files: &mut Vec<PathBuf>, depth: usize) {
    if depth > 10 {
        return;
    }
    let Ok(entries) = std::fs::read_dir(dir) else {
        return;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if path.is_dir() {
            collect_policy_files_recursive(&path.to_string_lossy(), file_name, files, depth + 1);
        } else if entry.file_name().to_string_lossy().eq_ignore_ascii_case(file_name) {
            files.push(path);
        }
    }
}

/// 解码文本字节：secedit/auditpol 输出为 UTF-16 LE（带 BOM），部分系统为 UTF-8
fn decode_text(bytes: &[u8]) -> String {
    if bytes.starts_with(&[0xFF, 0xFE]) {
        // UTF-16 LE with BOM
        let units: Vec<u16> = bytes[2..]
            .chunks_exact(2)
            .map(|c| u16::from_le_bytes([c[0], c[1]]))
            .collect();
        String::from_utf16_lossy(&units)
    } else if bytes.starts_with(&[0xEF, 0xBB, 0xBF]) {
        // UTF-8 with BOM
        String::from_utf8_lossy(&bytes[3..]).into_owned()
    } else {
        // 默认 UTF-8
        String::from_utf8_lossy(bytes).into_owned()
    }
}

/// 解析 secedit / GptTmpl.inf 输出的 INF 内容，提取 [System Access] 节为键值映射
fn parse_system_access_section(content: &str) -> Option<HashMap<String, String>> {
    let start = content.find("[System Access]")?;
    let section = match content.rfind('[') {
        Some(end) if end > start => &content[start..end],
        _ => &content[start..],
    };

    // 解析 key = value 行
    let re = Regex::new(r"^\s*(\w+)\s*=\s*(.*)").unwrap();
    let mut map = HashMap::new();
    for line in section.lines() {
        if let Some(caps) = re.captures(line) {
            let key = caps.get(1).unwrap().as_str().trim().to_string();
            let value = caps.get(2).unwrap().as_str().trim().trim_matches('\"').to_string();
            map.insert(key, value);
        }
    }
    if map.is_empty() {
        None
    } else {
        Some(map)
    }
}

/// 将 [System Access] 键值映射应用到采集结果
fn apply_password_policy_map(map: &HashMap<String, String>, data: &mut WindowsSystemCheckData) {
    if let Some(v) = map.get("MinimumPasswordAge") {
        data.minimum_password_age = v.clone();
    }
    if let Some(v) = map.get("MaximumPasswordAge") {
        data.maximum_password_age = v.clone();
    }
    if let Some(v) = map.get("MinimumPasswordLength") {
        data.minimum_password_length = v.clone();
    }
    if let Some(v) = map.get("PasswordComplexity") {
        data.password_complexity = v.clone();
    }
    if let Some(v) = map.get("PasswordHistorySize") {
        data.password_history_size = v.clone();
    }
    if let Some(v) = map.get("LockoutBadCount") {
        data.lockout_bad_count = v.clone();
    }
    if let Some(v) = map.get("LockoutDuration") {
        data.lockout_duration = v.clone();
    }
    if let Some(v) = map.get("ResetLockoutCount") {
        data.reset_lockout_count = v.clone();
    }
    if let Some(v) = map.get("RequireLogonToChangePassword") {
        data.require_logon_to_change_password = v.clone();
    }
    if let Some(v) = map.get("ClearTextPassword") {
        data.clear_text_password = v.clone();
    }
    if let Some(v) = map.get("LSAAnonymousNameLookup") {
        data.lsa_anonymous_name_lookup = v.clone();
    }
    if let Some(v) = map.get("NewAdministratorName") {
        data.new_administrator_name = v.clone();
    }
    if let Some(v) = map.get("NewGuestName") {
        data.new_guest_name = v.clone();
    }
    if let Some(v) = map.get("EnableAdminAccount") {
        data.enable_admin_account = v.clone();
    }
    if let Some(v) = map.get("EnableGuestAccount") {
        data.enable_guest_account = v.clone();
    }
}

/// 解析 secedit 输出的 INF 文件（兼容 GptTmpl.inf）
fn parse_secedit_output(content: &str, data: &mut WindowsSystemCheckData) {
    if let Some(map) = parse_system_access_section(content) {
        apply_password_policy_map(&map, data);
    }
}

/// 采集审计策略：优先解析 GptTmpl.inf 的 [Event Audit] 节（GPO 配置值，0=无/1=成功/2=失败/3=成功和失败），
/// 无配置时回退 auditpol（Win10/11 高级审计，旧注册表路径仅 Win7/2008 有效）
fn collect_audit_policy(data: &mut WindowsSystemCheckData) {
    // 1. 优先：GptTmpl.inf 的 [Event Audit] 节（权威配置值，不依赖策略刷新时序）
    let mut merged: HashMap<String, String> = HashMap::new();
    for file in find_gpttmpl_files(&data.domainname) {
        if let Ok(bytes) = std::fs::read(&file) {
            if let Some(map) = parse_event_audit_section(&decode_text(&bytes)) {
                log::info!("从 GptTmpl.inf 解析审计策略: {}", file.display());
                for (k, v) in map {
                    merged.insert(k, v);
                }
            }
        }
    }
    if !merged.is_empty() {
        apply_audit_policy_map(&merged, data);
        return;
    }

    // 2. 回退：auditpol /get /category:* /r 输出 CSV，兼容 Win10/11
    collect_audit_policy_auditpol(data);
}

/// 解析 GptTmpl.inf 的 [Event Audit] 节为键值映射
fn parse_event_audit_section(content: &str) -> Option<HashMap<String, String>> {
    let start = content.find("[Event Audit]")?;
    let section = match content.rfind('[') {
        Some(end) if end > start => &content[start..end],
        _ => &content[start..],
    };

    let re = Regex::new(r"^\s*(\w+)\s*=\s*(.*)").unwrap();
    let mut map = HashMap::new();
    for line in section.lines() {
        if let Some(caps) = re.captures(line) {
            let key = caps.get(1).unwrap().as_str().trim().to_string();
            let value = caps.get(2).unwrap().as_str().trim().to_string();
            map.insert(key, value);
        }
    }
    if map.is_empty() {
        None
    } else {
        Some(map)
    }
}

/// 将 [Event Audit] 键值映射应用到采集结果
fn apply_audit_policy_map(map: &HashMap<String, String>, data: &mut WindowsSystemCheckData) {
    if let Some(v) = map.get("AuditSystemEvents") {
        data.audit_system_events = v.clone();
    }
    if let Some(v) = map.get("AuditLogonEvents") {
        data.audit_logon_events = v.clone();
    }
    if let Some(v) = map.get("AuditObjectAccess") {
        data.audit_object_access = v.clone();
    }
    if let Some(v) = map.get("AuditPrivilegeUse") {
        data.audit_privilege_use = v.clone();
    }
    if let Some(v) = map.get("AuditPolicyChange") {
        data.audit_policy_change = v.clone();
    }
    if let Some(v) = map.get("AuditAccountManage") {
        data.audit_account_manage = v.clone();
    }
    if let Some(v) = map.get("AuditProcessTracking") {
        data.audit_process_tracking = v.clone();
    }
    if let Some(v) = map.get("AuditDSAccess") {
        data.audit_ds_access = v.clone();
    }
    if let Some(v) = map.get("AuditAccountLogon") {
        data.audit_account_logon = v.clone();
    }
}

/// auditpol 采集（Win10/11 高级审计子类别 GUID 映射）
fn collect_audit_policy_auditpol(data: &mut WindowsSystemCheckData) {
    let output = Command::new("auditpol")
        .args(["/get", "/category:*", "/r"])
        .output();

    let content = match output {
        Ok(out) if out.status.success() => decode_text(&out.stdout),
        Ok(out) => {
            log::warn!(
                "auditpol /get 失败: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            return;
        }
        Err(e) => {
            log::warn!("auditpol 执行失败: {}", e);
            return;
        }
    };

    // CSV 格式：MachineName,Policy Target,Subcategory,Subcategory GUID,Inclusion Setting,Exclusion Setting,Setting Value
    // 用子类别 GUID 匹配（不随系统语言变化），取各旧式类别下子类别设置的最大值
    let mut by_guid: HashMap<String, u32> = HashMap::new();
    for line in content.lines().skip(1) {
        let cols: Vec<&str> = line.split(',').collect();
        if cols.len() >= 7 {
            let guid = cols[3].trim().to_string();
            if let Ok(val) = cols[6].trim().parse::<u32>() {
                by_guid.insert(guid, val);
            }
        }
    }

    // 子类别 GUID 常量（Windows 10/11 稳定）
    const SYS_SYSTEM: &str = "{0CCE921D-69AE-11D9-BED3-505054503030}";
    const SYS_LOGON: &str = "{0CCE9215-69AE-11D9-BED3-505054503030}";
    const SYS_ACCOUNT_LOGOFF: &str = "{0CCE9216-69AE-11D9-BED3-505054503030}";
    const SYS_OBJECT_ACCESS: [&str; 13] = [
        "{0CCE921B-69AE-11D9-BED3-505054503030}", // File System
        "{0CCE921C-69AE-11D9-BED3-505054503030}", // Registry
        "{0CCE921E-69AE-11D9-BED3-505054503030}", // Kernel Object
        "{0CCE921F-69AE-11D9-BED3-505054503030}", // SAM
        "{0CCE9220-69AE-11D9-BED3-505054503030}", // Certification Services
        "{0CCE9221-69AE-11D9-BED3-505054503030}", // Application Generated
        "{0CCE9222-69AE-11D9-BED3-505054503030}", // Handle Manipulation
        "{0CCE9224-69AE-11D9-BED3-505054503030}", // File Share
        "{0CCE9225-69AE-11D9-BED3-505054503030}", // Filtering Platform Packet Drop
        "{0CCE9226-69AE-11D9-BED3-505054503030}", // Filtering Platform Connection
        "{0CCE9227-69AE-11D9-BED3-505054503030}", // Other Object Access Events
        "{0CCE9245-69AE-11D9-BED3-505054503030}", // Removable Storage
        "{0CCE9246-69AE-11D9-BED3-505054503030}", // Central Access Policy Staging
    ];
    const SYS_PRIVILEGE_USE: [&str; 3] = [
        "{0CCE9228-69AE-11D9-BED3-505054503030}", // Sensitive Privilege Use
        "{0CCE9229-69AE-11D9-BED3-505054503030}", // Non Sensitive Privilege Use
        "{0CCE922A-69AE-11D9-BED3-505054503030}", // Other Privilege Use Events
    ];
    const SYS_POLICY_CHANGE: [&str; 6] = [
        "{0CCE922F-69AE-11D9-BED3-505054503030}", // Audit Policy Change
        "{0CCE9230-69AE-11D9-BED3-505054503030}", // Authentication Policy Change
        "{0CCE9231-69AE-11D9-BED3-505054503030}", // Authorization Policy Change
        "{0CCE9232-69AE-11D9-BED3-505054503030}", // MPSSVC Rule-Level Policy Change
        "{0CCE9233-69AE-11D9-BED3-505054503030}", // Filtering Platform Policy Change
        "{0CCE9234-69AE-11D9-BED3-505054503030}", // Other Policy Change Events
    ];
    const SYS_ACCOUNT_MANAGE: [&str; 6] = [
        "{0CCE9235-69AE-11D9-BED3-505054503030}", // User Account Management
        "{0CCE9236-69AE-11D9-BED3-505054503030}", // Computer Account Management
        "{0CCE9237-69AE-11D9-BED3-505054503030}", // Security Group Management
        "{0CCE9238-69AE-11D9-BED3-505054503030}", // Distribution Group Management
        "{0CCE9239-69AE-11D9-BED3-505054503030}", // Application Group Management
        "{0CCE923A-69AE-11D9-BED3-505054503030}", // Other Account Management Events
    ];
    const SYS_PROCESS_TRACKING: [&str; 6] = [
        "{0CCE922B-69AE-11D9-BED3-505054503030}", // Process Creation
        "{0CCE922C-69AE-11D9-BED3-505054503030}", // Process Termination
        "{0CCE922D-69AE-11D9-BED3-505054503030}", // DPAPI Activity
        "{0CCE922E-69AE-11D9-BED3-505054503030}", // RPC Events
        "{0CCE9248-69AE-11D9-BED3-505054503030}", // Plug and Play Events
        "{0CCE9249-69AE-11D9-BED3-505054503030}", // Token Right Adjusted Events
    ];
    const SYS_DS_ACCESS: [&str; 4] = [
        "{0CCE923B-69AE-11D9-BED3-505054503030}", // Directory Service Access
        "{0CCE923C-69AE-11D9-BED3-505054503030}", // Directory Service Changes
        "{0CCE923D-69AE-11D9-BED3-505054503030}", // Directory Service Replication
        "{0CCE923E-69AE-11D9-BED3-505054503030}", // Detailed Directory Service Replication
    ];
    const SYS_ACCOUNT_LOGON: [&str; 4] = [
        "{0CCE923F-69AE-11D9-BED3-505054503030}", // Credential Validation
        "{0CCE9240-69AE-11D9-BED3-505054503030}", // Kerberos Service Ticket Operations
        "{0CCE9241-69AE-11D9-BED3-505054503030}", // Kerberos Authentication Service
        "{0CCE9242-69AE-11D9-BED3-505054503030}", // Other Account Logon Events
    ];

    // 取一组子类别 GUID 中的最大值（表示最严格的审计配置）
    let max_of = |guids: &[&str], map: &HashMap<String, u32>| -> String {
        guids
            .iter()
            .filter_map(|g| map.get(*g))
            .copied()
            .max()
            .unwrap_or(0)
            .to_string()
    };

    data.audit_system_events = max_of(&[SYS_SYSTEM], &by_guid);
    data.audit_logon_events = max_of(&[SYS_LOGON], &by_guid);
    data.audit_object_access = max_of(&SYS_OBJECT_ACCESS, &by_guid);
    data.audit_privilege_use = max_of(&SYS_PRIVILEGE_USE, &by_guid);
    data.audit_policy_change = max_of(&SYS_POLICY_CHANGE, &by_guid);
    data.audit_account_manage = max_of(&SYS_ACCOUNT_MANAGE, &by_guid);
    data.audit_process_tracking = max_of(&SYS_PROCESS_TRACKING, &by_guid);
    data.audit_ds_access = max_of(&SYS_DS_ACCESS, &by_guid);
    data.audit_account_logon = max_of(&SYS_ACCOUNT_LOGON, &by_guid);

    // 兼容旧版系统（Server 2008 风格）：
    // 若 auditpol 未返回任何子类别，回退读取注册表 Policies\Audit
    if by_guid.is_empty() {
        collect_audit_policy_legacy(data);
    }
    let _ = SYS_ACCOUNT_LOGOFF; // Logoff 子类别暂未映射到模型字段
}

/// 旧版审计策略采集（Win7/Server 2008 注册表路径，作为回退）
fn collect_audit_policy_legacy(data: &mut WindowsSystemCheckData) {
    let path = r"SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Audit";
    let hklm = RegKey::predef(HKEY_LOCAL_MACHINE);

    if let Ok(key) = hklm.open_subkey_with_flags(path, KEY_READ) {
        let audit_fields = [
            ("System", &mut data.audit_system_events),
            ("Logon", &mut data.audit_logon_events),
            ("ObjectAccess", &mut data.audit_object_access),
            ("PrivilegeUse", &mut data.audit_privilege_use),
            ("PolicyChange", &mut data.audit_policy_change),
            ("AccountManage", &mut data.audit_account_manage),
            ("ProcessTracking", &mut data.audit_process_tracking),
            ("DsAccess", &mut data.audit_ds_access),
            ("AccountLogon", &mut data.audit_account_logon),
        ];

        for (name, field) in audit_fields {
            if let Ok(val) = key.get_value::<u32, _>(name) {
                *field = val.to_string();
            } else {
                *field = "0".to_string();
            }
        }
    }
}

/// 采集设备控制（移动存储）
/// "所有可移动存储类: 拒绝所有权限" → 根键 Deny_All=1；
/// 各存储类拒绝 → 子键（{GUID}）下的 Deny_Read/Deny_Write/Deny_Execute=1
fn collect_device_control(data: &mut WindowsSystemCheckData) {
    let path = r"SOFTWARE\Policies\Microsoft\Windows\RemovableStorageDevices";
    let hklm = RegKey::predef(HKEY_LOCAL_MACHINE);

    let mut denied = false;
    if let Ok(key) = hklm.open_subkey_with_flags(path, KEY_READ) {
        // 1. 根键 Deny_All（所有可移动存储类：拒绝所有访问）
        if key.get_value::<u32, _>("Deny_All").unwrap_or(0) == 1 {
            denied = true;
        }

        // 2. 各存储类子键（{GUID}）下的 Deny_Read/Deny_Write/Deny_Execute
        if !denied {
            for subkey_name in key.enum_keys().filter_map(|k| k.ok()) {
                if let Ok(subkey) = key.open_subkey_with_flags(&subkey_name, KEY_READ) {
                    for value_name in ["Deny_All", "Deny_Read", "Deny_Write", "Deny_Execute"] {
                        if subkey.get_value::<u32, _>(value_name).unwrap_or(0) == 1 {
                            denied = true;
                            break;
                        }
                    }
                }
                if denied {
                    break;
                }
            }
        }
    }

    // 3. 降级：GPO 策略文件（registry.pol）中的 Deny_* 配置
    if !denied {
        let target_root = "Software\\Policies\\Microsoft\\Windows\\RemovableStorageDevices";
        'outer: for file in find_registry_pol_files(&data.domainname) {
            if let Ok(bytes) = std::fs::read(&file) {
                for (pol_path, pol_name, vtype, value) in parse_registry_pol(&bytes) {
                    let is_root = pol_path.eq_ignore_ascii_case(target_root);
                    let is_sub = pol_path.len() > target_root.len()
                        && pol_path[..target_root.len()].eq_ignore_ascii_case(target_root);
                    if !is_root && !is_sub {
                        continue;
                    }
                    let name_ok = ["Deny_All", "Deny_Read", "Deny_Write", "Deny_Execute"]
                        .iter()
                        .any(|n| pol_name.eq_ignore_ascii_case(n));
                    if name_ok && pol_value(&value, vtype) == "1" {
                        log::info!("从 registry.pol 解析移动存储策略: {}", file.display());
                        denied = true;
                        break 'outer;
                    }
                }
            }
        }
    }

    data.storage_devices = if denied { "1" } else { "0" }.to_string();
}

/// 采集屏幕保护设置：
/// 优先级 1: HKEY_USERS 下已登录用户的真实配置（最准确的实际生效值）
/// 优先级 2: SYSVOL 上的 GPO 源文件（域控制器下发的权威策略，仅在无活跃用户时使用）
fn collect_screen_saver(data: &mut WindowsSystemCheckData) {
    // 优先级 1: 尝试从 HKEY_USERS 读取已登录用户的实际配置
    if collect_screen_saver_from_hku(data) {
        return; // 成功读取到用户配置，直接返回
    }
    
    // 优先级 2: 没有用户登录时，降级读取 SYSVOL GPO 源文件
    // SYSVOL 包含所有组策略对象的原始定义，可作为最终备份方案
    collect_screen_saver_from_sysvol(data);
}

/// 枚举 HKEY_USERS 下已登录用户（域/本地账户 SID），读取其策略注册表屏保设置
/// 返回值含义：是否枚举到真实用户（S-1-5-21- 开头）。
/// 只要存在真实用户即视为“用户层已处理”（未读到数据 = 用户未配置/被豁免），
/// 仅当完全没有任何真实用户（无人登录）时才返回 false 触发 SYSVOL 降级
fn collect_screen_saver_from_hku(data: &mut WindowsSystemCheckData) -> bool {
    let hku = RegKey::predef(HKEY_USERS);
    let mut has_user = false; // 是否枚举到真实用户
    for sid in hku.enum_keys().filter_map(|k| k.ok()) {
        // 跳过内置账户（.DEFAULT / SYSTEM / LocalService / NetworkService）
        // 只处理真实用户账户（SID 以 S-1-5-21-开头）
        if !sid.starts_with("S-1-5-21-") {
            continue;
        }
        has_user = true; // 枚举到真实用户，用户层已处理
        let path = format!(r"{}\Software\Policies\Microsoft\Windows\Control Panel\Desktop", sid);
        if let Ok(key) = hku.open_subkey_with_flags(&path, KEY_READ) {
            if read_screen_saver_from_key(&key, data) {
                log::info!("从 HKEY_USERS\\{} 读取屏保策略", sid);
            }
        }
    }
    has_user
}

/// 从 registry.pol 字节数组解析屏保三项（ScreenSaveActive, ScreenSaverIsSecure, ScreenSaveTimeOut）
fn apply_registry_pol_screen_saver(bytes: &[u8], data: &mut WindowsSystemCheckData) {
    let target_path = "Software\\Policies\\Microsoft\\Windows\\Control Panel\\Desktop";
    for (path, name, vtype, value) in parse_registry_pol(bytes) {
        if !path.eq_ignore_ascii_case(target_path) {
            continue;
        }
        match name.as_str() {
            "ScreenSaveActive" => data.screen_saver_active = pol_value(&value, vtype),
            "ScreenSaverIsSecure" => data.screen_saver_secure = pol_value(&value, vtype),
            "ScreenSaveTimeOut" => data.screen_save_timeout = pol_value(&value, vtype),
            _ => {}
        }
    }
}

/// 从域控 SYSVOL 读取 GPO 源文件中的屏保策略配置
/// 仅在没有用户登录时降级使用（优先级 2），因为 SYSVOL 反映的是域控制器下发的权威策略
fn collect_screen_saver_from_sysvol(data: &mut WindowsSystemCheckData) {
    // 仅在非工作组模式下尝试 SYSVOL（有域名的机器）
    if data.domainname.is_empty() || 
       data.domainname.eq_ignore_ascii_case("WORKGROUP") || 
       data.domainname.eq_ignore_ascii_case("WORKSTATION") {
        log::debug!("当前机器未加入域或为工作组模式，跳过 SYSVOL 解析");
        return;
    }
    
    let sysvol_root = format!(r"\\{}\SysVol\{}\Policies", data.domainname, data.domainname);
    log::info!("开始从 SYSVOL 读取屏保策略：{}", sysvol_root);
    
    match std::fs::read_dir(&sysvol_root) {
        Ok(entries) => {
            for entry in entries.flatten() {
                if !entry.path().is_dir() {
                    continue;
                }
                // 查找所有 GPO 的 User\registry.pol 文件
                let pol = entry.path().join("User").join("registry.pol");
                if let Ok(bytes) = std::fs::read(&pol) {
                    apply_registry_pol_screen_saver(&bytes, data);
                    log::info!("成功从 SYSVOL GPO 源解析屏保策略：{:?}", pol.display());
                }
            }
        }
        Err(e) => log::warn!("无法访问 SYSVOL ({}): {}", sysvol_root, e),
    }
}

/// 采集管理员/来宾账户（WMI 实际账户状态，覆盖 GPO 配置值）
/// 根据 SID 结尾 -500（Administrator）/-501（Guest）识别
fn collect_admin_accounts(wmi: &WMIConnection, data: &mut WindowsSystemCheckData) {
    match wmi.query::<Win32_UserAccount>() {
        Ok(accounts) => {
            for acc in accounts {
                let sid = acc.SID.unwrap_or_default();
                let name = acc.Name.unwrap_or_default();
                let state = if acc.Disabled.unwrap_or(false) { "0" } else { "1" };
                if sid.ends_with("-500") {
                    data.new_administrator_name = name;
                    data.enable_admin_account = state.to_string();
                } else if sid.ends_with("-501") {
                    data.new_guest_name = name;
                    data.enable_guest_account = state.to_string();
                }
            }
        }
        Err(e) => log::warn!("Win32_UserAccount 查询失败: {}", e),
    }
}

/// 解析 GPO registry.pol 文件（用户/计算机配置策略注册表文件）
/// 支持两种格式（根据文件头自动识别）：
/// 1. PReg 二进制格式（XP+ 标准）：PReg 头 + 记录 [record_size][0x1E][type][path_len][name_len]...
/// 2. GPT 文本格式（老式 GPO）：PReg 头 + [路径;值名;类型;长度;数据]（UTF-16LE 字符流，数值为二进制 u32）
/// 返回 (注册表路径, 值名, 值类型, 值数据)
fn parse_registry_pol(bytes: &[u8]) -> Vec<(String, String, u32, Vec<u8>)> {
    // 二进制格式：offset 12 处为记录签名 0x1E
    if bytes.len() >= 16 && bytes[12] == 0x1E && bytes[13] == 0 && bytes[14] == 0 && bytes[15] == 0 {
        parse_registry_pol_binary(bytes)
    } else if bytes.len() >= 10 && bytes[8] == 0x5b && bytes[9] == 0 {
        // 文本格式：offset 8 处为 '['（UTF-16LE）
        parse_registry_pol_text(bytes)
    } else {
        Vec::new()
    }
}

/// PReg 二进制格式解析
fn parse_registry_pol_binary(bytes: &[u8]) -> Vec<(String, String, u32, Vec<u8>)> {
    let mut entries = Vec::new();
    // 文件头：PReg 01 00 00 00（8 字节）
    if bytes.len() < 8 || &bytes[0..4] != b"PReg" {
        return entries;
    }

    let mut offset = 8usize;
    while offset + 24 <= bytes.len() {
        let signature = u32::from_le_bytes([
            bytes[offset + 4],
            bytes[offset + 5],
            bytes[offset + 6],
            bytes[offset + 7],
        ]);
        if signature != 0x1E {
            break;
        }

        let vtype = u32::from_le_bytes([
            bytes[offset + 8],
            bytes[offset + 9],
            bytes[offset + 10],
            bytes[offset + 11],
        ]);
        let path_len = u32::from_le_bytes([
            bytes[offset + 12],
            bytes[offset + 13],
            bytes[offset + 14],
            bytes[offset + 15],
        ]) as usize;
        let name_len = u32::from_le_bytes([
            bytes[offset + 16],
            bytes[offset + 17],
            bytes[offset + 18],
            bytes[offset + 19],
        ]) as usize;

        let mut pos = offset + 20;
        if pos + path_len + name_len + 4 > bytes.len() {
            break;
        }

        // 路径（UTF-16 LE，含结尾 null）
        let path = decode_utf16(&bytes[pos..pos + path_len]);
        pos += path_len;

        // 值名（UTF-16 LE，含结尾 null）
        let name = decode_utf16(&bytes[pos..pos + name_len]);
        pos += name_len;

        // 值大小与数据
        let data_len = u32::from_le_bytes([
            bytes[pos],
            bytes[pos + 1],
            bytes[pos + 2],
            bytes[pos + 3],
        ]) as usize;
        pos += 4;
        if pos + data_len > bytes.len() {
            break;
        }
        let value = bytes[pos..pos + data_len].to_vec();

        entries.push((path, name, vtype, value));
        offset = pos + data_len;
    }
    entries
}

/// GPT 文本格式解析：每条记录为 [路径\0;值名\0;类型:u32;长度:u32;数据]（UTF-16LE）
fn parse_registry_pol_text(bytes: &[u8]) -> Vec<(String, String, u32, Vec<u8>)> {
    let mut entries = Vec::new();
    if bytes.len() < 8 || &bytes[0..4] != b"PReg" {
        return entries;
    }

    let mut pos = 8usize;
    while pos + 10 <= bytes.len() {
        // 定位记录起始 '['
        if bytes[pos] != 0x5b || bytes[pos + 1] != 0 {
            pos += 1;
            continue;
        }
        let mut cur = pos + 2;

        // 路径：UTF-16LE 直到双 null 终止
        let Some((path, next)) = read_utf16_until_null(bytes, cur) else {
            break;
        };
        if next + 2 > bytes.len() || bytes[next] != 0x3b || bytes[next + 1] != 0 {
            break;
        }
        cur = next + 2;

        // 值名：UTF-16LE 直到双 null 终止
        let Some((name, next2)) = read_utf16_until_null(bytes, cur) else {
            break;
        };
        if next2 + 2 > bytes.len() || bytes[next2] != 0x3b || bytes[next2 + 1] != 0 {
            break;
        }
        cur = next2 + 2;
        if cur + 10 > bytes.len() {
            break;
        }

        // 值类型 u32
        let vtype = u32::from_le_bytes([bytes[cur], bytes[cur + 1], bytes[cur + 2], bytes[cur + 3]]);
        if bytes[cur + 4] != 0x3b || bytes[cur + 5] != 0 {
            break;
        }
        cur += 6;

        // 数据长度 u32
        let data_len = u32::from_le_bytes([bytes[cur], bytes[cur + 1], bytes[cur + 2], bytes[cur + 3]]) as usize;
        if bytes[cur + 4] != 0x3b || bytes[cur + 5] != 0 {
            break;
        }
        cur += 6;
        if cur + data_len > bytes.len() {
            break;
        }

        let value = bytes[cur..cur + data_len].to_vec();
        entries.push((path, name, vtype, value));

        cur += data_len;
        // 跳过记录结束 ']'
        if cur + 2 <= bytes.len() && bytes[cur] == 0x5d && bytes[cur + 1] == 0 {
            cur += 2;
        }
        pos = cur;
    }
    entries
}

/// 读取 UTF-16LE 字符串直到双 null 终止，返回 (字符串, 终止后位置)
fn read_utf16_until_null(bytes: &[u8], pos: usize) -> Option<(String, usize)> {
    let mut end = pos;
    while end + 2 <= bytes.len() {
        if bytes[end] == 0 && bytes[end + 1] == 0 {
            break;
        }
        end += 2;
    }
    if end >= bytes.len() || end < pos + 2 {
        return None;
    }
    let s = decode_utf16(&bytes[pos..end]);
    Some((s, end + 2))
}

/// 解码 UTF-16 LE 字节为字符串（去除结尾 null）
fn decode_utf16(bytes: &[u8]) -> String {
    let units: Vec<u16> = bytes
        .chunks_exact(2)
        .map(|c| u16::from_le_bytes([c[0], c[1]]))
        .collect();
    String::from_utf16_lossy(&units)
        .trim_end_matches('\0')
        .to_string()
}

/// registry.pol 值数据转字符串（REG_DWORD 按数字，REG_SZ 按文本）
fn pol_value(data: &[u8], vtype: u32) -> String {
    match vtype {
        4 => {
            // REG_DWORD
            if data.len() >= 4 {
                u32::from_le_bytes([data[0], data[1], data[2], data[3]]).to_string()
            } else {
                "0".to_string()
            }
        }
        _ => {
            // REG_SZ / REG_EXPAND_SZ 等文本类型
            decode_utf16(data)
        }
    }
}

fn read_screen_saver_from_key(key: &RegKey, data: &mut WindowsSystemCheckData) -> bool {
    let mut found = false;
    // GPO 策略写入了 REG_SZ 字符串值（如 "1"），手动配置可能为 REG_DWORD，两种类型均需支持
    if let Ok(val) = key.get_value::<String, _>("ScreenSaveActive") {
        data.screen_saver_active = val;
        found = true;
    } else if let Ok(val) = key.get_value::<u32, _>("ScreenSaveActive") {
        data.screen_saver_active = val.to_string();
        found = true;
    }
    if let Ok(val) = key.get_value::<String, _>("ScreenSaverIsSecure") {
        data.screen_saver_secure = val;
        found = true;
    } else if let Ok(val) = key.get_value::<u32, _>("ScreenSaverIsSecure") {
        data.screen_saver_secure = val.to_string();
        found = true;
    }
    if let Ok(val) = key.get_value::<String, _>("ScreenSaveTimeOut") {
        data.screen_save_timeout = val;
        found = true;
    } else if let Ok(val) = key.get_value::<u32, _>("ScreenSaveTimeOut") {
        data.screen_save_timeout = val.to_string();
        found = true;
    }
    found
}