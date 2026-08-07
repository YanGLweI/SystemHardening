//! 加固检查计划模块（与 Linux 客户端 schedule.go 逻辑保持一致：
//! 每 5 分钟从服务端拉取计划，计划变化时立即重算下次检查时刻，
//! 叠加一次性随机抖动 0~15 分钟错开上报洪峰）

use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

use chrono::{Datelike, Duration as ChronoDuration, Local, NaiveDate, NaiveDateTime, TimeZone};

use crate::api;
use crate::config::Config;
use crate::token::TokenManager;

/// 计划拉取间隔：5 分钟
const SCHEDULE_POLL_INTERVAL: Duration = Duration::from_secs(5 * 60);

/// 服务端下发的加固检查计划
#[derive(Debug, Clone, serde::Deserialize)]
pub struct CheckSchedule {
    pub schedule_type: String, // daily/weekly/monthly
    pub check_time: String,    // HH:mm（半小时粒度）
    #[serde(default)]
    pub weekday: i32,          // 1-7（周一到周日，weekly）
    #[serde(default)]
    pub day_of_month: i32,     // 1-31（monthly）
    #[serde(default)]
    pub updated_at: Option<String>, // 计划最后更新时间（用于变更检测）
}

/// 计划应用状态（调度循环与计划拉取线程共享）
pub struct ScheduleState {
    pub applied_key: Option<String>,           // 已应用计划的变更标识
    pub applied: Option<CheckSchedule>,        // 已应用的计划
    pub jitter_secs: i64,                      // 随机抖动（进程生命周期内固定）
    pub next_check_time: Option<chrono::DateTime<Local>>, // 下次检查时刻
}

impl ScheduleState {
    pub fn new() -> Self {
        ScheduleState {
            applied_key: None,
            applied: None,
            jitter_secs: 0,
            next_check_time: None,
        }
    }
}

/// 启动计划拉取线程（每 5 分钟一次），返回共享的下次检查时刻
/// state 由调用方传入，供检查执行完成后重算下次时刻使用
pub fn spawn_schedule_loop(
    config: Config,
    token_manager: TokenManager,
    state: Arc<Mutex<ScheduleState>>,
) -> Arc<Mutex<Option<chrono::DateTime<Local>>>> {
    let next_check_time: Arc<Mutex<Option<chrono::DateTime<Local>>>> = Arc::new(Mutex::new(None));
    let next_clone = Arc::clone(&next_check_time);

    thread::spawn(move || {
        loop {
            apply_schedule_if_changed(&config, &token_manager, &state, &next_clone);
            thread::sleep(SCHEDULE_POLL_INTERVAL);
        }
    });

    next_check_time
}

/// 计划变更标识：计划内容 + updated_at（后端兜底默认计划无 updated_at 时用内容区分）
fn schedule_change_key(s: &CheckSchedule) -> String {
    format!(
        "{}|{}|{}|{}|{}",
        s.schedule_type,
        s.check_time,
        s.weekday,
        s.day_of_month,
        s.updated_at.as_deref().unwrap_or("")
    )
}

/// 拉取计划并比对变更标识，仅在变化时重新生效
fn apply_schedule_if_changed(
    config: &Config,
    token_manager: &TokenManager,
    state: &Arc<Mutex<ScheduleState>>,
    next_check_time: &Arc<Mutex<Option<chrono::DateTime<Local>>>>,
) {
    let token = token_manager.short_token().to_string();
    if token.is_empty() {
        log::warn!("[SCHEDULE] 无可用 Token，跳过计划拉取");
        return;
    }

    let schedule = match api::get_check_schedule(&config.server_url, &token) {
        Ok(s) => s,
        Err(e) => {
            log::warn!("[SCHEDULE] 拉取检查计划失败: {}", e);
            return;
        }
    };

    let key = schedule_change_key(&schedule);

    let mut st = state.lock().unwrap();
    if st.applied_key.as_deref() == Some(key.as_str()) {
        return; // 计划未变化，无需重新生效
    }

    // 计划变化（或首次应用）：生成一次性随机抖动（0~15 分钟）并重算下次检查时刻
    let is_first = st.applied_key.is_none();
    if is_first {
        st.jitter_secs = generate_jitter_secs();
    }
    let now = Local::now();
    let next = calc_next_check(&schedule, now) + ChronoDuration::seconds(st.jitter_secs);
    log::info!(
        "[SCHEDULE] 应用新检查计划: type={} time={} weekday={} day_of_month={}，下次检查: {}（抖动 {}s）",
        schedule.schedule_type,
        schedule.check_time,
        schedule.weekday,
        schedule.day_of_month,
        next.format("%Y-%m-%d %H:%M:%S"),
        st.jitter_secs
    );
    st.applied_key = Some(key);
    st.applied = Some(schedule);
    st.next_check_time = Some(next);
    drop(st);

    *next_check_time.lock().unwrap() = Some(next);
}

/// 检查执行完成后，基于已应用的计划重算下次检查时刻
pub fn recompute_next_check(state: &Arc<Mutex<ScheduleState>>, next_check_time: &Arc<Mutex<Option<chrono::DateTime<Local>>>>) {
    let mut st = state.lock().unwrap();
    let applied = match &st.applied {
        Some(s) => s.clone(),
        // 从未成功拉取计划时使用兜底默认计划（每天 02:00）
        None => CheckSchedule {
            schedule_type: "daily".to_string(),
            check_time: "02:00".to_string(),
            weekday: 1,
            day_of_month: 1,
            updated_at: None,
        },
    };
    let now = Local::now();
    let next = calc_next_check(&applied, now) + ChronoDuration::seconds(st.jitter_secs);
    st.next_check_time = Some(next);
    drop(st);

    *next_check_time.lock().unwrap() = Some(next);
    log::info!("[SCHEDULE] 下次检查时刻已重算: {}", next.format("%Y-%m-%d %H:%M:%S"));
}

/// 基于计划计算下一个计划时刻（本地时区）
/// monthly 遇当月无该日期时取当月最后一天
pub fn calc_next_check(s: &CheckSchedule, from: chrono::DateTime<Local>) -> chrono::DateTime<Local> {
    let (hour, minute) = parse_check_time(&s.check_time);

    match s.schedule_type.as_str() {
        "weekly" => {
            // 服务端 weekday 1-7 对应周一到周日；chrono Weekday 周一=0
            let target_naive = from.date_naive().and_hms_opt(hour, minute, 0).unwrap();
            let delta = (s.weekday - 1 - from.weekday().num_days_from_monday() as i32 + 7) % 7;
            let mut target = target_naive + ChronoDuration::days(delta as i64);
            if target <= from.naive_local() {
                target += ChronoDuration::days(7);
            }
            naive_to_local(target)
        }
        "monthly" => {
            // 向后最多找 13 个月，取下一个出现的日期
            let mut cursor = from.date_naive();
            for _ in 0..13 {
                let last = days_in_month(cursor.year(), cursor.month());
                let day = if s.day_of_month > last { last } else { s.day_of_month };
                let target = NaiveDate::from_ymd_opt(cursor.year(), cursor.month(), day as u32)
                    .unwrap()
                    .and_hms_opt(hour, minute, 0)
                    .unwrap();
                if target > from.naive_local() {
                    return naive_to_local(target);
                }
                // 移到下个月
                let (y, m) = if cursor.month() == 12 {
                    (cursor.year() + 1, 1)
                } else {
                    (cursor.year(), cursor.month() + 1)
                };
                cursor = NaiveDate::from_ymd_opt(y, m, 1).unwrap();
            }
            // 理论上不可达，兜底按 daily
            calc_daily(hour, minute, from)
        }
        _ => calc_daily(hour, minute, from), // daily（未知类型也按 daily 兜底）
    }
}

/// daily 计划：今天的计划时刻已过则顺延到明天
fn calc_daily(hour: u32, minute: u32, from: chrono::DateTime<Local>) -> chrono::DateTime<Local> {
    let mut target = from.date_naive().and_hms_opt(hour, minute, 0).unwrap();
    if target <= from.naive_local() {
        target += ChronoDuration::days(1);
    }
    naive_to_local(target)
}

/// 解析 HH:mm，失败时兜底 02:00
fn parse_check_time(check_time: &str) -> (u32, u32) {
    let parts: Vec<&str> = check_time.split(':').collect();
    if parts.len() != 2 {
        return (2, 0);
    }
    match (parts[0].parse::<u32>(), parts[1].parse::<u32>()) {
        (Ok(h), Ok(m)) if h < 24 && m < 60 => (h, m),
        _ => (2, 0),
    }
}

/// 返回指定年月的天数
fn days_in_month(year: i32, month: u32) -> i32 {
    let (y, m) = if month == 12 { (year + 1, 1) } else { (year, month + 1) };
    NaiveDate::from_ymd_opt(y, m, 1).unwrap().pred_opt().unwrap().day() as i32
}

/// NaiveDateTime 转本地时区 DateTime（DST 春季前移导致该时刻不存在时顺延 1 小时）
fn naive_to_local(naive: NaiveDateTime) -> chrono::DateTime<Local> {
    use chrono::LocalResult;
    match Local.from_local_datetime(&naive) {
        LocalResult::Single(dt) => dt,
        LocalResult::Ambiguous(dt, _) => dt,
        LocalResult::None => {
            let shifted = naive + ChronoDuration::hours(1);
            Local
                .from_local_datetime(&shifted)
                .earliest()
                .unwrap_or_else(Local::now)
        }
    }
}

/// 生成 0~15 分钟的随机抖动（秒）；项目无 rand 依赖，用纳秒时间做简单散列
fn generate_jitter_secs() -> i64 {
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.subsec_nanos() as u64)
        .unwrap_or(0);
    ((nanos % 900) * 37 % 900) as i64
}
