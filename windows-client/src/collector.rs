use std::collections::HashMap;
use std::io::Read;
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

use chrono::Local;
use regex::Regex;
use serde::Deserialize;
use winreg::enums::*;
use winreg::RegKey;
use wmi::WMIConnection;

use crate::models::WindowsSystemCheckData;

/// 本机实际应用的 GPO 列表：(appliedOrder, GUID 大写)
/// None 表示 gpresult 不可用（安全降级：只解析本地策略文件，不解析域 GPO 缓存）
type AppliedGpos = Option<Vec<(u32, String)>>;

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
    
    // 【新增】采集硬件 UUID（复用已有 WMI 连接，避免重复初始化）
    data.hardware_uuid = collect_hardware_uuid_with(&wmi_con);
    
    collect_network_info(&wmi_con, &mut data);
    collect_license_info(&wmi_con, &mut data);

    // 获取本机实际应用的 GPO 列表（RSOP）：DataStore/SYSVOL 中的 GPO 文件含全部域 GPO 定义，
    // 被安全筛选等排除的 GPO 文件仍在，必须先过滤再解析，否则误判合规；整个采集周期只查一次
    let applied_gpos = query_applied_gpos(&data.domainname);

    // 2. 密码策略（secedit）
    collect_password_policy(&mut data, &applied_gpos);

    // 3. 审计策略（注册表）
    collect_audit_policy(&mut data, &applied_gpos);

    // 4. 设备控制（注册表）
    collect_device_control(&mut data, &applied_gpos);

    // 5. 屏幕保护（registry.pol → SYSVOL → HKU → 注册表）
    // SYSVOL 降级仅用计算机侧列表作排序优先级，不作过滤条件（Session 0 无用户侧 RSOP，见函数注释）
    collect_screen_saver(&mut data, &applied_gpos);

    // 6. 管理员/来宾账户（WMI 实际状态，覆盖 GPO 配置值）
    collect_admin_accounts(&wmi_con, &mut data);

    // 7. 日期和时间
    data.date = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
    data.client_version = env!("CARGO_PKG_VERSION").to_string();

    log::info!(
        "Collection completed: hostname={}, domain={}, ip={}, hardware_uuid={}",
        data.hostname, 
        data.domainname, 
        data.ip,
        data.hardware_uuid
    );
    Ok(data)
}

/// 降级采集判定：基础信息存在但密码/审计/屏保三大策略组同时全空。
/// 典型场景为更新重启后环境半就绪（SYSVOL 不可达、secedit/auditpol 失败）。
/// 此状态的数据不得上传，否则会用空值覆盖服务端历史健康数据。
/// 注意：域控豁免场景仅屏保为空、密码/审计有值，不会误判为降级。
pub fn is_degraded_collection(data: &WindowsSystemCheckData) -> bool {
    !data.hostname.is_empty()
        && data.minimum_password_length.is_empty()
        && data.audit_system_events.is_empty()
        && data.screen_saver_active.is_empty()
}

// ==================== 采集子函数 ====================

/// 采集 Windows System UUID (BIOS SerialNumber)
pub fn collect_hardware_uuid() -> String {
    let wmi_con = match WMIConnection::new() {
        Ok(con) => con,
        Err(e) => {
            log::warn!("WMI init failed for hardware UUID: {}", e);
            return String::new();
        }
    };
    collect_hardware_uuid_with(&wmi_con)
}

/// 复用已有 WMI 连接采集硬件 UUID（避免重复的 COM 初始化和命名空间绑定开销）
/// 优先取 Win32_ComputerSystemProduct.UUID（SMBIOS UUID，每台机器/虚拟机全局唯一）；
/// BIOS 序列号仅作兜底，且排除常见占位值（"Default string" 等，白牌机普遍相同会导致去重冲突）
fn collect_hardware_uuid_with(wmi_con: &WMIConnection) -> String {
    log::info!("Collecting hardware UUID from SMBIOS (Win32_ComputerSystemProduct.UUID)...");

    #[derive(Deserialize)]
    #[allow(dead_code, non_camel_case_types, non_snake_case)]
    struct Win32_ComputerSystemProduct {
        UUID: Option<String>,
    }

    match wmi_con.query::<Win32_ComputerSystemProduct>() {
        Ok(results) => {
            if let Some(product) = results.first() {
                let uuid = product.UUID.clone().unwrap_or_default();
                if is_valid_hardware_uuid(&uuid) {
                    log::info!("✅ Hardware UUID collected: {}", uuid);
                    return uuid;
                }
                log::warn!("⚠️ Win32_ComputerSystemProduct.UUID 无效（{}），尝试 BIOS 序列号兜底", uuid);
            }
        }
        Err(e) => {
            log::warn!("WMI query Win32_ComputerSystemProduct failed: {}", e);
        }
    }

    // 兜底：BIOS 序列号（排除占位值）
    #[derive(Deserialize)]
    #[allow(dead_code, non_camel_case_types, non_snake_case)]
    struct Win32_BIOS {
        SerialNumber: Option<String>,
    }

    match wmi_con.query::<Win32_BIOS>() {
        Ok(results) => {
            if let Some(bios) = results.first() {
                let serial = bios.SerialNumber.clone().unwrap_or_default();
                let trimmed = serial.trim();
                if !trimmed.is_empty() && !is_placeholder_serial(trimmed) {
                    log::info!("✅ Hardware UUID collected (BIOS serial fallback): {}", trimmed);
                    return trimmed.to_string();
                }
                log::warn!("⚠️ BIOS 序列号为占位值或空: {}", serial);
            }
        }
        Err(e) => {
            log::warn!("WMI query Win32_BIOS failed: {}", e);
        }
    }

    log::warn!("⚠️ Hardware UUID empty, using fallback");
    String::new()
}

/// 校验标准硬件 UUID 格式（8-4-4-4-12 十六进制，大小写不敏感），
/// 并拒绝全 0 / 全 F 等无效 SMBIOS 值
pub fn is_valid_hardware_uuid(s: &str) -> bool {
    let t = s.trim();
    let re = Regex::new(r"(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$").unwrap();
    if !re.is_match(t) {
        return false;
    }
    let hex_only: String = t.chars().filter(|c| *c != '-').collect();
    let all_zero = hex_only.chars().all(|c| c == '0');
    let all_f = hex_only.chars().all(|c| c.to_ascii_lowercase() == 'f');
    !all_zero && !all_f
}

/// BIOS 序列号占位值黑名单（白牌机/虚拟机常见，多台机器相同，不可用作去重依据）
pub fn is_placeholder_serial(s: &str) -> bool {
    let lower = s.trim().to_lowercase();
    matches!(
        lower.as_str(),
        "default string"
            | "to be filled by o.e.m."
            | "to be filled by oem"
            | "not specified"
            | "none"
            | "null"
            | "n/a"
            | "unknown"
            | "o.e.m."
            | "system serial number"
            | "chassis serial number"
            | "baseboard serial number"
            | "serial"
            | "0"
            | "0000000000"
            | "123456789"
            | "0123456789"
    )
}

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
fn collect_password_policy(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
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
            // secedit 在 update resume 等阶段可能"退出码 0 但内容为空"，
            // 解析为空时必须同样降级到 GptTmpl，否则密码策略被静默留空
            if data.minimum_password_length.is_empty() {
                log::warn!("secedit /export 成功但解析结果为空（可能处于 update resume 阶段），降级解析 GptTmpl.inf");
                collect_password_policy_from_gpttmpl(data, applied_gpos);
            }
        }
        Ok(out) => {
            log::warn!(
                "secedit /export 失败（退出码: {:?}, stdout: {}, stderr: {}），降级解析 GptTmpl.inf",
                out.status.code(),
                String::from_utf8_lossy(&out.stdout).trim().chars().take(200).collect::<String>(),
                String::from_utf8_lossy(&out.stderr).trim().chars().take(200).collect::<String>()
            );
            collect_password_policy_from_gpttmpl(data, applied_gpos);
        }
        Err(e) => {
            log::warn!("secedit 执行失败: {}，降级解析 GptTmpl.inf", e);
            collect_password_policy_from_gpttmpl(data, applied_gpos);
        }
    }
}

/// 降级方案：解析 GPO 的 GptTmpl.inf 获取密码策略配置值（GPO 定义值，非实际生效值）
/// 来源：本地组策略、GPO DataStore 缓存、域 SYSVOL 源文件（UTF-16 LE 编码），
/// 域 GPO 文件按 gpresult 实际应用列表过滤，避免被筛选排除的 GPO 被当作已应用；
/// 按应用顺序合并（后应用的 GPO 覆盖先应用的，符合 GPO 优先级规则）
fn collect_password_policy_from_gpttmpl(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
    let mut merged: HashMap<String, String> = HashMap::new();

    for file in find_gpttmpl_files(&data.domainname, applied_gpos) {
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

/// 查找所有可用的 GptTmpl.inf 文件（本地策略 → 已应用域 GPO 的 DataStore 缓存 → SYSVOL 源）
fn find_gpttmpl_files(domain: &str, applied_gpos: &AppliedGpos) -> Vec<PathBuf> {
    find_policy_files(domain, "GptTmpl.inf", applied_gpos)
}

/// 查找所有可用的 Registry.pol 文件（本地策略 → 已应用域 GPO 的 DataStore 缓存 → SYSVOL 源）
fn find_registry_pol_files(domain: &str, applied_gpos: &AppliedGpos) -> Vec<PathBuf> {
    let mut files = find_policy_files(domain, "Registry.pol", applied_gpos);
    // 本地用户策略缓存（无 GUID 路径，不会被过滤）
    files.push(PathBuf::from(r"C:\Windows\System32\GroupPolicy\User\registry.pol"));
    files
}

/// 按文件名递归查找策略文件（本地 → DataStore 缓存 → SYSVOL 源），
/// 并按 gpresult 实际应用列表过滤域 GPO 文件、按应用优先级排序：
/// DataStore/SYSVOL 中缓存的是域内全部 GPO 的定义（含被安全筛选/作用域排除的），
/// 不过滤会把未应用的政策当作已应用（如被拒绝的 USB Deny 误判为合规）
fn find_policy_files(domain: &str, file_name: &str, applied_gpos: &AppliedGpos) -> Vec<PathBuf> {
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

    filter_and_order_policy_files(files, applied_gpos)
}

/// 按实际应用 GPO 列表过滤并排序策略文件（纯逻辑，可单测）：
/// - 路径不含 GPO GUID（本地策略文件）：保留，排最前（优先级最低，先被覆盖）
/// - 路径含 GPO GUID（DataStore/SYSVOL）：仅保留 GUID 在实际应用列表中的文件，
///   并按 appliedOrder 升序——调用方按序合并时自然实现“后应用的 GPO 覆盖先应用的”
/// - applied_gpos 为 None（gpresult 不可用）：只保留本地策略文件，宁可留空也不误报合规
pub(crate) fn filter_and_order_policy_files(files: Vec<PathBuf>, applied_gpos: &AppliedGpos) -> Vec<PathBuf> {
    let guid_re =
        Regex::new(r"(?i)\{[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}\}").unwrap();

    let mut local = Vec::new();
    let mut gpo_files: Vec<(u32, PathBuf)> = Vec::new();
    // 汇总被筛选排除的 GPO（GUID → 文件数）：每次调用只输出一条日志，避免逐文件刷屏
    let mut skipped: Vec<(String, usize)> = Vec::new();
    let mut skipped_total = 0usize;
    for file in files {
        let path_str = file.to_string_lossy();
        match guid_re.find(&path_str) {
            Some(m) => {
                let guid = m.as_str().to_uppercase();
                let order = applied_gpos
                    .as_ref()
                    .and_then(|list| list.iter().find(|(_, g)| *g == guid).map(|(o, _)| *o));
                match order {
                    Some(order) => gpo_files.push((order, file)),
                    None => {
                        skipped_total += 1;
                        match skipped.iter_mut().find(|(g, _)| *g == guid) {
                            Some((_, count)) => *count += 1,
                            None => skipped.push((guid, 1)),
                        }
                    }
                }
            }
            None => local.push(file),
        }
    }

    if !skipped.is_empty() {
        // info 级别：便于管理员在日志中确认被筛选排除的 GPO 未参与解析（合规审计可观测性）
        log::info!(
            "跳过被 RSOP 筛选排除的 GPO 策略文件 {} 个（共 {} 个 GPO）: {}",
            skipped_total,
            skipped.len(),
            skipped
                .iter()
                .map(|(g, n)| format!("{}×{}", g, n))
                .collect::<Vec<_>>()
                .join(", ")
        );
    }

    gpo_files.sort_by_key(|(order, _)| *order);
    local
        .into_iter()
        .chain(gpo_files.into_iter().map(|(_, f)| f))
        .collect()
}

/// 通过 gpresult /x（RSOP XML，语言无关）获取本机计算机配置实际应用的 GPO 列表。
/// /scope computer 保证服务 Session 0 下无需用户上下文；结果供本次采集周期的策略文件过滤使用。
/// 返回 None 表示不可用（调用方按“仅本地策略”安全降级）。
/// 注意：屏保属用户配置策略，其 SYSVOL 降级仅用本列表作排序优先级、不作过滤条件（见 collect_screen_saver_from_sysvol）
fn query_applied_gpos(domain: &str) -> AppliedGpos {
    // 工作组环境无域 GPO，DataStore/SYSVOL 不存在，空列表即“无域 GPO 可解析”
    if domain.is_empty()
        || domain.eq_ignore_ascii_case("WORKGROUP")
        || domain.eq_ignore_ascii_case("WORKSTATION")
    {
        return Some(Vec::new());
    }

    let computer = run_gpresult_query("computer");
    if let Some(list) = &computer {
        if !list.is_empty() {
            log::info!("gpresult 获取到 {} 个已应用 GPO: {:?}", list.len(), list);
        }
    }
    computer
}

/// 计算机/用户配置已应用 GPO 列表并集（预留：待后续能获得用户上下文时启用；
/// Session 0 下 /scope user 必然失败，当前屏保 SYSVOL 降级已改用全量解析方案）
#[allow(dead_code)]
fn union_applied_gpos(computer: &AppliedGpos, user: &AppliedGpos) -> AppliedGpos {
    match (computer.as_ref(), user.as_ref()) {
        (None, None) => None,
        (Some(c), None) => Some(c.clone()),
        (None, Some(u)) => Some(u.clone()),
        (Some(c), Some(u)) => {
            let mut merged = c.clone();
            for (order, guid) in u {
                if !merged.iter().any(|(_, g)| g == guid) {
                    merged.push((*order, guid.clone()));
                }
            }
            merged.sort_by_key(|(order, _)| *order);
            Some(merged)
        }
    }
}

/// 执行单一 scope 的 gpresult /x 并解析；不可用时返回 None。
/// - 临时文件名唯一（进程 id + 纳秒）且执行前预清理：避免并行实例互踩或残留旧文件被误读
/// - 60 秒超时看门狗：域控 RPC 异常时 gpresult 可能挂起，杀进程安全降级，不卡死每日检查
fn run_gpresult_query(scope: &str) -> AppliedGpos {
    let unique = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let xml_path = std::env::temp_dir()
        .join(format!("shc_rsop_{}_{}_{}.xml", std::process::id(), scope, unique));
    let _ = std::fs::remove_file(&xml_path); // 防御性预清理同名残留文件

    let xml_arg = xml_path.to_string_lossy().to_string();
    let applied = match Command::new("gpresult")
        .args(["/x", &xml_arg, "/scope", scope, "/f"])
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()
    {
        Ok(mut child) => {
            let start = Instant::now();
            // 轮询退出状态；超时杀进程（/x 模式控制台输出极小，进程退出后读管道不会阻塞）
            let status = loop {
                match child.try_wait() {
                    Ok(Some(status)) => break Some(status),
                    Ok(None) => {
                        if start.elapsed() > Duration::from_secs(60) {
                            log::warn!("gpresult /x /scope {} 超时（60 秒未退出），终止进程并安全降级", scope);
                            let _ = child.kill();
                            let _ = child.wait();
                            break None;
                        }
                        std::thread::sleep(Duration::from_millis(200));
                    }
                    Err(e) => {
                        log::warn!("gpresult 进程状态查询失败: {}，仅解析本地策略文件", e);
                        break None;
                    }
                }
            };

            let mut stderr_bytes = Vec::new();
            if let Some(mut err_pipe) = child.stderr.take() {
                let _ = err_pipe.read_to_end(&mut stderr_bytes);
            }
            let stderr = String::from_utf8_lossy(&stderr_bytes);

            match status {
                Some(status) if status.success() => match std::fs::read(&xml_path) {
                    Ok(bytes) => {
                        let list = extract_applied_gpos_section(
                            &decode_text(&bytes),
                            &format!("{}results", scope),
                        );
                        if list.is_empty() && scope == "computer" {
                            log::warn!("gpresult /x 成功但未解析到任何已应用 GPO（stderr: {}），仅解析本地策略文件",
                                stderr.trim());
                            None
                        } else {
                            Some(list)
                        }
                    }
                    Err(e) => {
                        log::warn!("读取 gpresult XML 失败: {}，仅解析本地策略文件", e);
                        None
                    }
                },
                Some(status) => {
                    if scope == "computer" {
                        log::warn!(
                            "gpresult /x 失败（退出码: {:?}, stderr: {}），仅解析本地策略文件",
                            status.code(),
                            stderr.trim()
                        );
                    } else {
                        // Session 0 无用户上下文属常态（无人登录才会触发屏保 SYSVOL 降级），debug 级别即可
                        log::debug!(
                            "gpresult 用户配置查询失败（退出码: {:?}, stderr: {}），本次无用户侧 GPO 列表",
                            status.code(),
                            stderr.trim()
                        );
                    }
                    None
                }
                None => None,
            }
        }
        Err(e) => {
            log::warn!("gpresult 执行失败: {}，仅解析本地策略文件", e);
            None
        }
    };

    let _ = std::fs::remove_file(&xml_path);
    applied
}

/// 提取计算机配置段实际应用的 GPO（见 extract_applied_gpos_section）
pub(crate) fn extract_applied_gpos(xml: &str) -> Vec<(u32, String)> {
    extract_applied_gpos_section(xml, "computerresults")
}

/// 从 RSOP XML 指定段（"computerresults"/"userresults"）提取实际应用的 GPO：
/// (appliedOrder, GUID 大写)，按应用顺序升序。
/// Schema（Rsop > ComputerResults/UserResults > GPO*）：GPO 块含 Identifier（GUID）与 Link 子节点，
/// Link 内的 AppliedOrder 为实际应用顺序；被安全筛选等排除的 GPO 仍会出现在列表中，
/// 但其 AppliedOrder 全为 0，据此剔除；指定段不存在时返回空列表（如 /scope user 输出无 ComputerResults）。
pub(crate) fn extract_applied_gpos_section(xml: &str, section: &str) -> Vec<(u32, String)> {
    // 段落定位忽略大小写：必须用 to_ascii_lowercase——Unicode to_lowercase 会改变字节长度
    // （如 İ→i̇ 2→3 字节、开尔文符→k 3→1），在小写文本上搜得的偏移切原文会错位甚至 panic，
    // 而 GPO 名称可为域管理员任意 Unicode 命名
    let lower = xml.to_ascii_lowercase();
    let open_tag = format!("<{}", section);
    let Some(start) = lower.find(&open_tag) else {
        return Vec::new();
    };
    // 段落终止边界：另一个 results 段；不存在则到文档末尾（双保险不混入另一侧 GPO）
    let end_tag = if section == "computerresults" {
        "<userresults"
    } else {
        "<computerresults"
    };
    let section_text = match lower[start..].find(end_tag) {
        Some(pos) => &xml[start..start + pos],
        None => &xml[start..],
    };

    let block_re =
        Regex::new(r"(?is)<[\w-]*:?(?:GPO|appliedGPO)\b[^>]*>(.*?)</[\w-]*:?(?:GPO|appliedGPO)\s*>")
            .unwrap();
    let guid_re =
        Regex::new(r"(?i)\{[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}\}")
            .unwrap();
    // 只匹配 AppliedOrder/Order，不会误匹配 LinkOrder 等前缀元素（正则要求紧跟 '<' 之后）
    let order_re = Regex::new(r"(?is)<(?:appliedOrder|Order)\b[^>]*>\s*(\d+)").unwrap();

    let mut result: Vec<(u32, String)> = Vec::new();
    for caps in block_re.captures_iter(section_text) {
        let block = caps.get(1).map(|m| m.as_str()).unwrap_or("");
        // 含过滤原因的块属于筛选排除列表（双保险）
        if block.to_ascii_lowercase().contains("filterreason") {
            continue;
        }
        let Some(guid_m) = guid_re.find(block) else {
            continue;
        };
        // 取块内所有 AppliedOrder 的最大值：被筛选排除的 GPO 所有 AppliedOrder 均为 0，据此剔除；
        // 无 AppliedOrder 元素的块（如 pivot 数据）同样不视为已应用，保守跳过
        let order = order_re
            .captures_iter(block)
            .filter_map(|c| c.get(1).and_then(|m| m.as_str().parse::<u32>().ok()))
            .max()
            .unwrap_or(0);
        if order == 0 {
            continue;
        }
        let guid = guid_m.as_str().to_uppercase();
        if !result.iter().any(|(_, g)| *g == guid) {
            result.push((order, guid));
        }
    }
    result.sort_by_key(|(order, _)| *order);
    result
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

/// 解码文本字节：secedit/auditpol/gpresult 输出为 UTF-16 LE（通常带 BOM），部分系统为 UTF-8；
/// 另嗅探无 BOM 的 UTF-16LE，避免被误当 UTF-8 产生乱码导致解析失败
fn decode_text(bytes: &[u8]) -> String {
    if bytes.starts_with(&[0xFF, 0xFE]) {
        decode_utf16le(&bytes[2..])
    } else if bytes.starts_with(&[0xEF, 0xBB, 0xBF]) {
        // UTF-8 with BOM
        String::from_utf8_lossy(&bytes[3..]).into_owned()
    } else if looks_like_utf16le(bytes) {
        // 无 BOM 的 UTF-16LE（部分重定向/第三方工具输出）
        decode_utf16le(bytes)
    } else {
        // 默认 UTF-8
        String::from_utf8_lossy(bytes).into_owned()
    }
}

/// 将 UTF-16LE 字节流解码为字符串（lossy）
fn decode_utf16le(bytes: &[u8]) -> String {
    let units: Vec<u16> = bytes
        .chunks_exact(2)
        .map(|c| u16::from_le_bytes([c[0], c[1]]))
        .collect();
    String::from_utf16_lossy(&units)
}

/// 无 BOM UTF-16LE 嗅探：采样前 512 字节，字节长度为偶数且 ≥90% 的高位字节（奇数偏移）为 0，
/// 即判定为 UTF-16LE（XML 等 ASCII 主导文本）；正常 UTF-8/ANSI 文本几乎不含 NUL，不会误判
fn looks_like_utf16le(bytes: &[u8]) -> bool {
    if bytes.len() < 4 || bytes.len() % 2 != 0 {
        return false;
    }
    let sample = &bytes[..bytes.len().min(512)];
    let pairs = sample.len() / 2;
    let zero_hi = sample[1..].iter().step_by(2).filter(|&&b| b == 0).count();
    zero_hi * 10 >= pairs * 9
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
fn collect_audit_policy(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
    // 1. 优先：已应用 GPO 的 GptTmpl.inf 的 [Event Audit] 节（权威配置值，不依赖策略刷新时序；
    //    文件列表已按 gpresult 实际应用列表过滤，按应用顺序合并，被筛选排除的 GPO 不参与）
    let mut merged: HashMap<String, String> = HashMap::new();
    for file in find_gpttmpl_files(&data.domainname, applied_gpos) {
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
fn collect_device_control(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
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

    // 3. 降级：已应用 GPO 的策略文件（registry.pol）中的 Deny_* 配置，
    //    文件列表已按 gpresult 实际应用列表过滤，被安全筛选等排除的 GPO 不会误判为已应用
    if !denied {
        let target_root = "Software\\Policies\\Microsoft\\Windows\\RemovableStorageDevices";
        'outer: for file in find_registry_pol_files(&data.domainname, applied_gpos) {
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

/// HKEY_USERS 屏保采集结果三态
/// 用于区分"已登录且读到数据"、"已登录但未配置(豁免)"、"无人登录"三种语义，
/// 避免豁免用户被 SYSVOL 降级误判为合规
enum HkuScreenSaverState {
    /// 存在当前登录用户，且读取到屏保配置
    LoggedInWithData,
    /// 存在当前登录用户，但未读到屏保配置（域控豁免/未配置），应保持为空，不降级
    LoggedInNoData,
    /// 无当前登录用户（如系统重启后停留在登录界面），需降级读取 SYSVOL
    NoLoggedInUser,
}

/// 采集屏幕保护设置：
/// 优先级 1: HKEY_USERS 下当前登录用户的真实配置（最准确的实际生效值）
/// 优先级 2: SYSVOL 上的 GPO 源文件（域控制器下发的权威策略，仅在无用户登录时使用）
/// 注意：已登录但被域控豁免的用户，屏保三项保持为空（不合规），不得降级到 SYSVOL 补数据
fn collect_screen_saver(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
    log::info!("[屏保采集] 开始采集屏保策略...");

    match collect_screen_saver_from_hku(data) {
        HkuScreenSaverState::LoggedInWithData => {
            log::info!("[屏保采集] ✅ 已从登录用户读取屏保策略，结束采集");
        }
        HkuScreenSaverState::LoggedInNoData => {
            // 有用户登录但未配置屏保（如域控豁免）：这是用户的真实状态，
            // 服务端应判为不合规，绝不能降级到 SYSVOL 用域策略"补"成合规
            log::warn!("[屏保采集] ⚠️ 存在已登录用户但未读到屏保策略配置（可能被域控豁免），三项保持为空，不降级 SYSVOL");
        }
        HkuScreenSaverState::NoLoggedInUser => {
            log::info!("[屏保采集] ⬇️ 无当前登录用户，开始 SYSVOL 降级...");
            collect_screen_saver_from_sysvol(data, applied_gpos);
        }
    }

    // 最终检查结果
    if data.screen_saver_active.is_empty() {
        log::warn!("[屏保采集] ❌ 未读取到屏保配置，三项字段均为空字符串");
    } else {
        log::info!("[屏保采集] ✅ 最终屏保配置：active={}, secure={}, timeout={}",
            data.screen_saver_active,
            data.screen_saver_secure,
            data.screen_save_timeout
        );
    }
}

/// 枚举 HKEY_USERS 下当前登录用户的屏保策略
/// 登录判定依据：HKEY_USERS\<SID>\Volatile Environment 子键仅在用户登录加载配置文件时创建，
/// 注销/重启后消失。历史登录残留的 SID（hive 被服务/计划任务等加载但无交互登录）
/// 以及 _Classes 后缀键均不视为"当前登录用户"
fn collect_screen_saver_from_hku(data: &mut WindowsSystemCheckData) -> HkuScreenSaverState {
    let hku = RegKey::predef(HKEY_USERS);
    let mut logged_in_sids: Vec<String> = Vec::new();

    log::debug!("[屏保采集] 开始枚举 HKEY_USERS 下的屏保策略...");

    for sid in hku.enum_keys().filter_map(|k| k.ok()) {
        // 只处理真实用户账户（SID 以 S-1-5-21-开头），跳过 _Classes 后缀与内置/服务账户
        if !sid.starts_with("S-1-5-21-") || sid.ends_with("_Classes") {
            log::trace!("[屏保采集] 跳过非真实用户 SID: {}", sid);
            continue;
        }
        // 判定是否当前登录：Volatile Environment 仅在用户登录加载配置文件时存在
        let volatile_path = format!(r"{}\Volatile Environment", sid);
        if hku.open_subkey_with_flags(&volatile_path, KEY_READ).is_err() {
            log::debug!("[屏保采集] 用户 {} 配置已加载但非当前登录（无 Volatile Environment），跳过", sid);
            continue;
        }
        logged_in_sids.push(sid);
    }

    if logged_in_sids.is_empty() {
        log::info!("[屏保采集] HKEY_USERS 中无当前登录的真实用户");
        return HkuScreenSaverState::NoLoggedInUser;
    }

    log::info!("[屏保采集] 检测到 {} 个当前登录用户: {:?}", logged_in_sids.len(), logged_in_sids);

    let mut found_data = false;
    for sid in &logged_in_sids {
        let path = format!(r"{}\Software\Policies\Microsoft\Windows\Control Panel\Desktop", sid);
        match hku.open_subkey_with_flags(&path, KEY_READ) {
            Ok(key) => {
                if read_screen_saver_from_key(&key, data) {
                    log::info!("[屏保采集] ✅ 从 HKEY_USERS\\{} 读取到屏保策略：active={}, secure={}, timeout={}",
                        sid,
                        data.screen_saver_active.clone(),
                        data.screen_saver_secure.clone(),
                        data.screen_save_timeout.clone()
                    );
                    found_data = true;
                } else {
                    log::warn!("[屏保采集] ⚠️ 已登录用户 {} 的屏保策略路径存在，但未读取到有效配置 (三项全为空)", sid);
                }
            }
            Err(_) => {
                log::info!("[屏保采集] 已登录用户 {} 无屏保策略注册表路径（可能被域控豁免）", sid);
            }
        }
    }

    if found_data {
        HkuScreenSaverState::LoggedInWithData
    } else {
        HkuScreenSaverState::LoggedInNoData
    }
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

/// SYSVOL GPO 目录解析排序（纯逻辑，可单测）：
/// 全部目录都参与解析（屏保属用户配置策略，Session 0 无用户侧 RSOP，无法甄别哪些 GPO
/// 对用户生效，域控下发的定义即权威降级源）；非应用列表内的目录在前（GUID 升序，结果确定），
/// 已应用 GPO 目录在后（AppliedOrder 升序）——配合 apply_registry_pol_screen_saver 的后写覆盖语义，
/// 保证 RSOP 确认应用的 GPO 定义值优先级最高；applied 为 None（gpresult 不可用）时全量按 GUID 升序
pub(crate) fn order_sysvol_gpo_dirs(dirs: Vec<(String, PathBuf)>, applied: &AppliedGpos) -> Vec<PathBuf> {
    let mut others: Vec<(String, PathBuf)> = Vec::new();
    let mut applied_dirs: Vec<(u32, PathBuf)> = Vec::new();
    for (guid, path) in dirs {
        let order = applied
            .as_ref()
            .and_then(|list| list.iter().find(|(_, g)| *g == guid).map(|(o, _)| *o));
        match order {
            Some(order) => applied_dirs.push((order, path)),
            None => others.push((guid, path)),
        }
    }
    others.sort_by(|a, b| a.0.cmp(&b.0));
    applied_dirs.sort_by_key(|(order, _)| *order);
    others
        .into_iter()
        .map(|(_, p)| p)
        .chain(applied_dirs.into_iter().map(|(_, p)| p))
        .collect()
}

/// 从域控 SYSVOL 读取 GPO 源文件中的屏保策略配置（GPO 定义值，非实际生效值）
/// 仅在 HKEY_USERS 未读到数据且无用户登录时降级使用（优先级 2）；
/// 全量解析域控下发的所有 GPO 目录的 User\registry.pol：屏保属用户配置策略，
/// Session 0 无用户上下文拿不到用户侧 RSOP，而下发屏保的 GPO 常不在计算机侧应用列表中，
/// 若按计算机侧列表过滤会误跳过屏保 GPO 导致屏保误判不合规（2.3.6 实测缺陷）；
/// 计算机侧应用列表可用时仅作排序优先级：已应用 GPO 最后解析，其定义值覆盖其他；
/// 豁免（已登录但未配置）场景由 LoggedInNoData 保证不进入本分支，无误判合规回归
fn collect_screen_saver_from_sysvol(data: &mut WindowsSystemCheckData, applied_gpos: &AppliedGpos) {
    // 仅在非工作组模式下尝试 SYSVOL（有域名的机器）
    if data.domainname.is_empty() || 
       data.domainname.eq_ignore_ascii_case("WORKGROUP") || 
       data.domainname.eq_ignore_ascii_case("WORKSTATION") {
        log::warn!("[SYSVOL 降级] 当前机器未加入域或为工作组模式 (domain={}), 无法降级读取 SYSVOL，屏保策略将保持为空", 
            data.domainname);
        return;
    }

    let sysvol_root = format!(r"\\{}\SysVol\{}\Policies", data.domainname, data.domainname);
    log::info!("[SYSVOL 降级] 开始从域控 SYSVOL 读取屏保策略：{}", sysvol_root);
    
    match std::fs::read_dir(&sysvol_root) {
        Ok(entries) => {
            let mut found_any_gpo = false;
            let mut found_valid_pol = false;

            // 收集全部 GPO 目录（目录名即 {GUID}），不跳过任何目录；排序见 order_sysvol_gpo_dirs
            let guid_re =
                Regex::new(r"(?i)\{[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}\}")
                    .unwrap();
            let mut dirs: Vec<(String, PathBuf)> = Vec::new();
            for entry in entries.flatten() {
                if !entry.path().is_dir() {
                    continue;
                }
                found_any_gpo = true;

                let name = entry.file_name().to_string_lossy().to_string();
                let Some(guid_m) = guid_re.find(&name) else { continue };
                dirs.push((guid_m.as_str().to_uppercase(), entry.path()));
            }

            let applied_count = applied_gpos
                .as_ref()
                .map(|list| {
                    dirs.iter()
                        .filter(|(guid, _)| list.iter().any(|(_, g)| g == guid))
                        .count()
                })
                .unwrap_or(0);
            log::info!(
                "[SYSVOL 降级] 解析 {} 个 GPO 目录（其中 {} 个为计算机侧已应用，优先覆盖）",
                dirs.len(),
                applied_count
            );

            for gpo_dir in order_sysvol_gpo_dirs(dirs, applied_gpos) {
                // 查找 GPO 的 User\registry.pol 文件（屏保属用户配置）
                let pol = gpo_dir.join("User").join("registry.pol");
                log::trace!("[SYSVOL 降级] 检查 SYSVOL GPO 目录中的 registry.pol: {:?}", pol.display());
                
                match std::fs::read(&pol) {
                    Ok(bytes) => {
                        let before_active = data.screen_saver_active.clone();
                        let before_secure = data.screen_saver_secure.clone();
                        let before_timeout = data.screen_save_timeout.clone();
                        
                        apply_registry_pol_screen_saver(&bytes, data);
                        
                        // 判断是否真的有数据更新
                        if before_active != data.screen_saver_active || 
                           before_secure != data.screen_saver_secure || 
                           before_timeout != data.screen_save_timeout {
                            found_valid_pol = true;
                            log::info!("[SYSVOL 降级] ✅ 成功从 GPO registry.pol 解析到屏保策略：{:?} (active={}, secure={}, timeout={})",
                                pol.display(),
                                data.screen_saver_active, 
                                data.screen_saver_secure, 
                                data.screen_save_timeout
                            );
                        } else {
                            log::trace!("[SYSVOL 降级] ℹ️ GPO registry.pol 中存在，但屏保配置与已有值一致或为空：{:?}", pol.display());
                        }
                    }
                    Err(e) => {
                        log::trace!("[SYSVOL 降级] 跳过不可读的 registry.pol 文件 ({}): {}", pol.display(), e);
                    }
                }
            }

            if !found_any_gpo {
                log::warn!("[SYSVOL 降级] SYSVOL 目录下未找到任何 GPO 子目录，屏保策略将保持为空");
            } else if !found_valid_pol {
                log::error!("[SYSVOL 降级] ❌ 找到 GPO 目录但所有 GPO 的 registry.pol 均未包含有效屏保配置，三项字段均为空！这可能是导致'屏保不合规'的根本原因！");
            }
        }
        Err(e) => {
            log::error!("[SYSVOL 降级] ❌ 无法访问域控 SYSVOL 路径 ({})，网络错误或权限不足：{} - 屏保策略将无法读取，可能导致不合规判定！",
                sysvol_root, e);
        }
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

// ==================== 单元测试 ====================
// 注意：仅覆盖纯逻辑（gpresult XML 解析、策略文件过滤/排序），不涉及注册表/文件系统。
// 完整测试需在 Windows 上运行（cargo test，见 GitHub Actions workflow）

#[cfg(test)]
mod tests {
    use super::*;

    /// 真实环境 gpresult /x 输出（TEST-WIN，USB Deny 被安全筛选拒绝）的忠实切片：
    /// GUID 位于 Path/Identifier（带默认命名空间）；本地组策略 Identifier 为 "LocalGPO"（无 GUID）；
    /// 被拒绝的 USB Deny 仍以 AppliedOrder=0 出现在列表中（必须被剔除）；
    /// ExtensionData 段内含大量嵌套 <GPO>（设置引用，无 AppliedOrder，必须被跳过）
    const RSOP_XML: &str = r#"<?xml version="1.0" encoding="utf-16"?>
<Rsop xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns="http://www.microsoft.com/GroupPolicy/Rsop">
  <ReadTime>2026-08-28T03:51:20.3310784Z</ReadTime>
  <DataType>LoggedData</DataType>
  <ComputerResults>
    <GPO>
      <Name>Default Domain Policy</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{31B2F340-016D-11D2-945F-00C04FB984F9}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <VersionDirectory>55</VersionDirectory>
      <VersionSysvol>55</VersionSysvol>
      <Enabled>true</Enabled>
      <IsValid>true</IsValid>
      <FilterAllowed>true</FilterAllowed>
      <AccessDenied>false</AccessDenied>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <SOMOrder>3</SOMOrder>
        <AppliedOrder>1</AppliedOrder>
        <LinkOrder>4</LinkOrder>
        <Enabled>true</Enabled>
        <NoOverride>false</NoOverride>
      </Link>
      <SecurityFilter>NT AUTHORITY\Authenticated Users</SecurityFilter>
      <ExtensionName>Security</ExtensionName>
      <ExtensionName>注册表</ExtensionName>
    </GPO>
    <GPO>
      <Name>本地组策略</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">LocalGPO</Identifier>
      </Path>
      <VersionDirectory>0</VersionDirectory>
      <VersionSysvol>0</VersionSysvol>
      <Enabled>true</Enabled>
      <IsValid>true</IsValid>
      <FilterAllowed>true</FilterAllowed>
      <AccessDenied>false</AccessDenied>
      <Link>
        <SOMPath>Local</SOMPath>
        <SOMOrder>1</SOMOrder>
        <AppliedOrder>0</AppliedOrder>
        <LinkOrder>1</LinkOrder>
        <Enabled>true</Enabled>
        <NoOverride>false</NoOverride>
      </Link>
    </GPO>
    <GPO>
      <Name>{881E827A-884C-4A1F-A8CA-55E1E413C98B}</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{881E827A-884C-4A1F-A8CA-55E1E413C98B}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <VersionDirectory>0</VersionDirectory>
      <VersionSysvol>0</VersionSysvol>
      <IsValid>false</IsValid>
      <FilterAllowed>false</FilterAllowed>
      <AccessDenied>false</AccessDenied>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <SOMOrder>2</SOMOrder>
        <AppliedOrder>0</AppliedOrder>
        <LinkOrder>3</LinkOrder>
        <Enabled>true</Enabled>
        <NoOverride>false</NoOverride>
      </Link>
    </GPO>
    <GPO>
      <Name>USB Deny</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <VersionDirectory>15</VersionDirectory>
      <VersionSysvol>65535</VersionSysvol>
      <Enabled>true</Enabled>
      <IsValid>true</IsValid>
      <FilterAllowed>false</FilterAllowed>
      <AccessDenied>true</AccessDenied>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <SOMOrder>1</SOMOrder>
        <AppliedOrder>0</AppliedOrder>
        <LinkOrder>2</LinkOrder>
        <Enabled>true</Enabled>
        <NoOverride>false</NoOverride>
      </Link>
      <SecurityFilter>NT AUTHORITY\Authenticated Users</SecurityFilter>
      <ExtensionName>注册表</ExtensionName>
    </GPO>
    <ExtensionData>
      <Extension xmlns:q1="http://www.microsoft.com/GroupPolicy/Settings/Security" xsi:type="q1:SecuritySettings" xmlns="http://www.microsoft.com/GroupPolicy/Settings">
        <q1:Account>
          <GPO xmlns="http://www.microsoft.com/GroupPolicy/Settings/Base">
            <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{31B2F340-016D-11D2-945F-00C04FB984F9}</Identifier>
            <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
          </GPO>
          <Precedence xmlns="http://www.microsoft.com/GroupPolicy/Settings/Base">1</Precedence>
          <q1:Name>MaximumPasswordAge</q1:Name>
          <q1:SettingNumber>30</q1:SettingNumber>
          <q1:Type>Password</q1:Type>
        </q1:Account>
      </Extension>
    </ExtensionData>
  </ComputerResults>
</Rsop>"#;

    /// 真实环境 gpresult /x 输出（OFS-DC01，全部策略已应用）的 GPO 段切片：
    /// 覆盖 GUID 大小写归一化（原始文件 DC Policy GUID 含小写 f）与 AppliedOrder≠LinkOrder 的场景
    const RSOP_XML_ALL_APPLIED: &str = r#"<?xml version="1.0" encoding="utf-16"?>
<Rsop xmlns="http://www.microsoft.com/GroupPolicy/Rsop">
  <ComputerResults>
    <GPO>
      <Name>Default Domain Policy</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{31B2F340-016D-11D2-945F-00C04FB984F9}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <SOMOrder>3</SOMOrder>
        <AppliedOrder>3</AppliedOrder>
        <LinkOrder>4</LinkOrder>
        <Enabled>true</Enabled>
      </Link>
    </GPO>
    <GPO>
      <Name>本地组策略</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">LocalGPO</Identifier>
      </Path>
      <Link>
        <SOMPath>Local</SOMPath>
        <AppliedOrder>1</AppliedOrder>
        <LinkOrder>1</LinkOrder>
      </Link>
    </GPO>
    <GPO>
      <Name>USB Deny</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <SOMOrder>1</SOMOrder>
        <AppliedOrder>2</AppliedOrder>
        <LinkOrder>2</LinkOrder>
        <Enabled>true</Enabled>
      </Link>
    </GPO>
    <GPO>
      <Name>{881E827A-884C-4A1F-A8CA-55E1E413C98B}</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{881E827A-884C-4A1F-A8CA-55E1E413C98B}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <IsValid>false</IsValid>
      <Link>
        <SOMPath>hot.local</SOMPath>
        <AppliedOrder>0</AppliedOrder>
        <LinkOrder>3</LinkOrder>
      </Link>
    </GPO>
    <GPO>
      <Name>Default Domain Controllers Policy</Name>
      <Path>
        <Identifier xmlns="http://www.microsoft.com/GroupPolicy/Types">{6AC1786C-016F-11D2-945F-00C04fB984F9}</Identifier>
        <Domain xmlns="http://www.microsoft.com/GroupPolicy/Types">hot.local</Domain>
      </Path>
      <Link>
        <SOMPath>hot.local/Domain Controllers</SOMPath>
        <SOMOrder>1</SOMOrder>
        <AppliedOrder>4</AppliedOrder>
        <LinkOrder>5</LinkOrder>
        <Enabled>true</Enabled>
      </Link>
    </GPO>
  </ComputerResults>
</Rsop>"#;

    #[test]
    fn test_extract_applied_gpos_real_output_usb_denied() {
        // 真实 TEST-WIN 输出：USB Deny 被安全筛选拒绝（AppliedOrder=0 + AccessDenied=true）
        let list = extract_applied_gpos(RSOP_XML);
        // 仅 Default Domain Policy 实际应用于计算机配置（本地组策略无 GUID，无效 GPO/USB Deny 均剔除）
        assert_eq!(list, vec![(1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string())]);
        // USB Deny 不得出现（本次缺陷的核心），ExtensionData 嵌套设置引用块也不得混入结果之外的噪音判定已通过列表等值断言覆盖
        assert!(!list.iter().any(|(_, g)| g == "{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}"));
    }

    #[test]
    fn test_extract_applied_gpos_real_output_all_applied() {
        // 真实 OFS-DC01 输出：全部策略应用，且 GUID 大小写需归一化、顺序取 AppliedOrder 而非 LinkOrder
        let list = extract_applied_gpos(RSOP_XML_ALL_APPLIED);
        assert_eq!(
            list,
            vec![
                (2, "{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}".to_string()),
                (3, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string()),
                (4, "{6AC1786C-016F-11D2-945F-00C04FB984F9}".to_string()),
            ]
        );
    }

    #[test]
    fn test_extract_applied_gpos_empty_input() {
        assert!(extract_applied_gpos("").is_empty());
        assert!(extract_applied_gpos("<rsop></rsop>").is_empty());
    }

    #[test]
    fn test_filter_and_order_policy_files_applied_list() {
        let applied: AppliedGpos = Some(vec![
            (2, "{9B883413-C9B1-4E31-8EF9-0966129B9CF1}".to_string()),
            (1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string()),
        ]);
        let files = vec![
            // 被筛选排除的 GPO（USB Deny）：必须被过滤掉（本次缺陷的核心）
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\DataStore\0\SysVol\hot.local\Policies\{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}\Machine\Registry.pol"),
            // 已应用 GPO（乱序传入，应按 appliedOrder 升序输出）
            PathBuf::from(r"\\hot.local\SysVol\hot.local\Policies\{9B883413-C9B1-4E31-8EF9-0966129B9CF1}\Machine\Microsoft\Windows NT\SecEdit\GptTmpl.inf"),
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\DataStore\0\SysVol\hot.local\Policies\{31B2F340-016D-11D2-945F-00C04FB984F9}\Machine\Registry.pol"),
            // 本地策略文件（无 GUID）：无条件保留且排最前（优先级最低，先被覆盖）
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\Machine\Registry.pol"),
        ];

        let result = filter_and_order_policy_files(files, &applied);
        assert_eq!(result.len(), 3);
        assert_eq!(result[0], PathBuf::from(r"C:\Windows\System32\GroupPolicy\Machine\Registry.pol"));
        // appliedOrder 1 先于 appliedOrder 2（合并时后者覆盖前者，符合 GPO 优先级）
        assert!(result[1].to_string_lossy().contains("31B2F340"));
        assert!(result[2].to_string_lossy().contains("9B883413"));
    }

    #[test]
    fn test_filter_and_order_policy_files_gpresult_unavailable() {
        // gpresult 不可用（None）：只保留本地策略文件，宁可留空也不误报合规
        let files = vec![
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\DataStore\0\SysVol\hot.local\Policies\{31B2F340-016D-11D2-945F-00C04FB984F9}\Machine\Registry.pol"),
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\Machine\Registry.pol"),
        ];
        let result = filter_and_order_policy_files(files, &None);
        assert_eq!(result.len(), 1);
        assert_eq!(result[0], PathBuf::from(r"C:\Windows\System32\GroupPolicy\Machine\Registry.pol"));
    }

    #[test]
    fn test_filter_and_order_policy_files_workgroup() {
        // 工作组（空已应用列表）：域 GPO 文件全部过滤，本地文件保留
        let files = vec![
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\DataStore\0\SysVol\x\Policies\{31B2F340-016D-11D2-945F-00C04FB984F9}\Machine\Registry.pol"),
            PathBuf::from(r"C:\Windows\System32\GroupPolicy\Machine\Registry.pol"),
        ];
        let result = filter_and_order_policy_files(files, &Some(Vec::new()));
        assert_eq!(result.len(), 1);
        assert!(result[0].to_string_lossy().contains("Machine\\Registry.pol"));
    }

    #[test]
    fn test_extract_applied_gpos_unicode_before_section_offset_safety() {
        // 段落前存在 Unicode 小写化会改变字节长度的字符（\u{0130} İ→i̇ +1 字节、\u{212A} 开尔文 K→k -2 字节）：
        // 旧 to_lowercase 实现下偏移错位会截取错误段落甚至 panic，ASCII 小写化实现必须正确提取
        let xml = "<?xml version=\"1.0\"?>\n<Rsop>\n<!-- \u{0130} \u{212A} 域管理员任意 Unicode 命名 -->\n<ComputerResults>\n<GPO>\n<Name>Test İ Policy</Name>\n<Path><Identifier>{31B2F340-016D-11D2-945F-00C04FB984F9}</Identifier></Path>\n<Link><AppliedOrder>1</AppliedOrder></Link>\n</GPO>\n</ComputerResults>\n</Rsop>";
        let list = extract_applied_gpos(xml);
        assert_eq!(
            list,
            vec![(1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string())]
        );
    }

    #[test]
    fn test_extract_applied_gpos_user_section() {
        // /scope user 输出（仅含 UserResults）：屏保并集需从用户段提取，且拒绝项同样剔除；
        // 对无 ComputerResults 的文档调 extract_applied_gpos 应返回空而非误解析全文档
        let xml = r#"<?xml version="1.0"?>
<Rsop>
  <UserResults>
    <GPO>
      <Name>ScreenSaver Policy</Name>
      <Path><Identifier>{9B883413-C9B1-4E31-8EF9-0966129B9CF1}</Identifier></Path>
      <Link><SOMPath>hot.local</SOMPath><AppliedOrder>1</AppliedOrder></Link>
    </GPO>
    <GPO>
      <Name>Denied User GPO</Name>
      <Path><Identifier>{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}</Identifier></Path>
      <AccessDenied>true</AccessDenied>
      <Link><SOMPath>hot.local</SOMPath><AppliedOrder>0</AppliedOrder></Link>
    </GPO>
  </UserResults>
</Rsop>"#;
        let user = extract_applied_gpos_section(xml, "userresults");
        assert_eq!(
            user,
            vec![(1, "{9B883413-C9B1-4E31-8EF9-0966129B9CF1}".to_string())]
        );
        assert!(extract_applied_gpos(xml).is_empty());
    }

    #[test]
    fn test_union_applied_gpos() {
        let computer: AppliedGpos =
            Some(vec![(1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string())]);
        let user: AppliedGpos = Some(vec![
            (2, "{9B883413-C9B1-4E31-8EF9-0966129B9CF1}".to_string()),
            // 与计算机侧重复的 GUID：并集不得重复收录
            (1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string()),
        ]);
        let union = union_applied_gpos(&computer, &user);
        let list = union.expect("两侧可用时并集应为 Some");
        assert_eq!(list.len(), 2);
        assert_eq!(list[0].1, "{31B2F340-016D-11D2-945F-00C04FB984F9}");
        assert_eq!(list[1].1, "{9B883413-C9B1-4E31-8EF9-0966129B9CF1}");
        // 两侧均不可用：保持 None（屏保 SYSVOL 安全降级）
        assert!(union_applied_gpos(&None, &None).is_none());
        // 单侧可用：透传（用户侧 Session 0 失败时不影响计算机列表）
        assert_eq!(union_applied_gpos(&computer, &None), computer);
        assert_eq!(union_applied_gpos(&None, &user), user);
    }

    #[test]
    fn test_decode_text_utf16le_variants() {
        let text = "<?xml version=\"1.0\"?><Rsop><Name>测试</Name></Rsop>";
        // 带 BOM 的 UTF-16LE（真实 gpresult 输出形态）
        let mut with_bom = vec![0xFF, 0xFE];
        with_bom.extend(text.encode_utf16().flat_map(u16::to_le_bytes));
        assert_eq!(decode_text(&with_bom), text);
        // 无 BOM 的 UTF-16LE：嗅探后同样正确解码（旧实现会误当 UTF-8 产生乱码）
        let no_bom: Vec<u8> = text.encode_utf16().flat_map(u16::to_le_bytes).collect();
        assert_eq!(decode_text(&no_bom), text);
        // 正常 UTF-8（含中文）不得被误判为 UTF-16
        let utf8 = "<?xml version=\"1.0\"?><名称>屏保策略</名称>";
        assert_eq!(decode_text(utf8.as_bytes()), utf8);
    }

    #[test]
    fn test_order_sysvol_gpo_dirs() {
        let dir = |g: &str| (g.to_string(), PathBuf::from(format!(r"\\d\SysVol\d\Policies\{}", g)));

        // gpresult 不可用（None）：全部目录保留、GUID 升序（乱序传入）
        let dirs = vec![
            dir("{9B883413-C9B1-4E31-8EF9-0966129B9CF1}"),
            dir("{17AB1402-8DFA-40D4-990E-ECD3094F3DA5}"),
        ];
        let result = order_sysvol_gpo_dirs(dirs.clone(), &None);
        assert_eq!(result.len(), 2);
        assert!(result[0].to_string_lossy().contains("17AB1402"));
        assert!(result[1].to_string_lossy().contains("9B883413"));

        // 部分命中应用列表：非应用目录在前（GUID 升序），应用目录在后（AppliedOrder 升序，优先覆盖）；
        // {9B883413} 即 2.3.6 实测被误跳过的屏保 GPO——新方案下必须保留在解析列表内
        let applied: AppliedGpos = Some(vec![
            (2, "{9B883413-C9B1-4E31-8EF9-0966129B9CF1}".to_string()),
            (1, "{31B2F340-016D-11D2-945F-00C04FB984F9}".to_string()),
        ]);
        let dirs = vec![
            dir("{9B883413-C9B1-4E31-8EF9-0966129B9CF1}"),
            dir("{881E827A-884C-4A1F-A8CA-55E1E413C98B}"),
            dir("{31B2F340-016D-11D2-945F-00C04FB984F9}"),
        ];
        let result = order_sysvol_gpo_dirs(dirs, &applied);
        assert_eq!(result.len(), 3);
        assert!(result[0].to_string_lossy().contains("881E827A")); // 非应用，最先解析（可被覆盖）
        assert!(result[1].to_string_lossy().contains("31B2F340")); // 应用 order=1
        assert!(result[2].to_string_lossy().contains("9B883413")); // 应用 order=2，最后解析（最高优先级）

        // 空目录列表
        assert!(order_sysvol_gpo_dirs(Vec::new(), &applied).is_empty());

        // 全部命中应用列表：无“其他”目录，纯按 AppliedOrder 升序
        let dirs = vec![
            dir("{9B883413-C9B1-4E31-8EF9-0966129B9CF1}"),
            dir("{31B2F340-016D-11D2-945F-00C04FB984F9}"),
        ];
        let result = order_sysvol_gpo_dirs(dirs, &applied);
        assert_eq!(result.len(), 2);
        assert!(result[0].to_string_lossy().contains("31B2F340"));
        assert!(result[1].to_string_lossy().contains("9B883413"));
    }
}