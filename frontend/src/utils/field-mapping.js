/**
 * 字段分组和标签映射
 */

export const fieldGroups = {
  // 系统更新
  dnf_conf_gpgcheck: { Group: '系统更新' },
  redhat_repo_gpgcheck: { Group: '系统更新' },
  
  // 用户账户策略
  pass_max_days: { Group: '用户账户策略' },
  pass_min_days: { Group: '用户账户策略' },
  pass_min_len: { Group: '用户账户策略' },
  pass_warn_age: { Group: '用户账户策略' },
  inactive: { Group: '用户账户策略' },
  gid: { Group: '用户账户策略' },
  
  // 计划任务
  cron: { Group: '计划任务' },
  crontab: { Group: '计划任务' },
  cron_hourly: { Group: '计划任务' },
  cron_daily: { Group: '计划任务' },
  cron_weekly: { Group: '计划任务' },
  cron_monthly: { Group: '计划任务' },
  cron_deny: { Group: '计划任务' },
  at_deny: { Group: '计划任务' },
  cron_allow: { Group: '计划任务' },
  at_allow: { Group: '计划任务' },
  
  // SSH 配置
  sshd_config: { Group: 'SSH 配置' },
  log_level: { Group: 'SSH 配置' },
  x11_forwarding: { Group: 'SSH 配置' },
  max_auth_tries: { Group: 'SSH 配置' },
  ignore_rhosts: { Group: 'SSH 配置' },
  hostbased_authentication: { Group: 'SSH 配置' },
  permit_root_login: { Group: 'SSH 配置' },
  permit_empty_passwords: { Group: 'SSH 配置' },
  permit_user_environment: { Group: 'SSH 配置' },
  client_alive_interval: { Group: 'SSH 配置' },
  client_alive_count_max: { Group: 'SSH 配置' },
  login_grace_time: { Group: 'SSH 配置' },
  
  // 密码策略
  minlen: { Group: '密码策略' },
  minclass: { Group: '密码策略' },
  dcredit: { Group: '密码策略' },
  ucredit: { Group: '密码策略' },
  lcredit: { Group: '密码策略' },
  ocredit: { Group: '密码策略' },
  password_remember: { Group: '密码策略' },
  
  // 文件权限
  passwd: { Group: '文件权限' },
  passwd_minus: { Group: '文件权限' },
  group: { Group: '文件权限' },
  group_minus: { Group: '文件权限' },
  shadow: { Group: '文件权限' },
  shadow_minus: { Group: '文件权限' },
  gshadow: { Group: '文件权限' },
  gshadow_minus: { Group: '文件权限' },
  
  // 加密与时钟
  crypto_policies: { Group: '加密与时钟' },
  ntp_server: { Group: '加密与时钟' }
}

// Tab 顺序定义
export const TABS_ORDER = [
  { key: 'system-update', label: '系统更新', fields: ['dnf_conf_gpgcheck', 'redhat_repo_gpgcheck'] },
  { key: 'user-policy', label: '用户账户策略', fields: ['pass_max_days', 'pass_min_days', 'pass_min_len', 'pass_warn_age', 'inactive', 'gid'] },
  { key: 'cron-config', label: '计划任务', fields: ['cron', 'crontab', 'cron_hourly', 'cron_daily', 'cron_weekly', 'cron_monthly', 'cron_deny', 'at_deny', 'cron_allow', 'at_allow'] },
  { key: 'ssh-config', label: 'SSH 配置', fields: ['sshd_config', 'log_level', 'x11_forwarding', 'max_auth_tries', 'ignore_rhosts', 'hostbased_authentication', 'permit_root_login', 'permit_empty_passwords', 'permit_user_environment', 'client_alive_interval', 'client_alive_count_max', 'login_grace_time'] },
  { key: 'password-policy', label: '密码策略', fields: ['minlen', 'minclass', 'dcredit', 'ucredit', 'lcredit', 'ocredit', 'password_remember'] },
  { key: 'file-permission', label: '文件权限', fields: ['passwd', 'passwd_minus', 'group', 'group_minus', 'shadow', 'shadow_minus', 'gshadow', 'gshadow_minus'] },
  { key: 'crypto-sync', label: '加密与时钟', fields: ['crypto_policies', 'ntp_server'] }
]

export function getFieldLabel(fieldName) {
  const labels = {
    dnf_conf_gpgcheck: 'dnf.conf_gpgcheck',
    redhat_repo_gpgcheck: 'redhat.repo_gpgcheck',
    pass_max_days: 'PASS_MAX_DAYS',
    pass_min_days: 'PASS_MIN_DAYS',
    pass_min_len: 'PASS_MIN_LEN',
    pass_warn_age: 'PASS_WARN_AGE',
    inactive: 'INACTIVE',
    gid: 'GID',
    tmout: 'TMOUT',
    cron: 'Cron',
    crontab: 'Crontab',
    cron_hourly: 'CronHourly',
    cron_daily: 'CronDaily',
    cron_weekly: 'CronWeekly',
    cron_monthly: 'CronMonthly',
    cron_deny: 'CronDeny',
    at_deny: 'AtDeny',
    cron_allow: 'CronAllow',
    at_allow: 'AtAllow',
    sshd_config: 'sshd_config',
    log_level: 'LogLevel',
    x11_forwarding: 'X11Forwarding',
    max_auth_tries: 'MaxAuthTries',
    ignore_rhosts: 'IgnoreRhosts',
    hostbased_authentication: 'HostbasedAuthentication',
    permit_root_login: 'PermitRootLogin',
    permit_empty_passwords: 'PermitEmptyPasswords',
    permit_user_environment: 'PermitUserEnvironment',
    client_alive_interval: 'ClientAliveInterval',
    client_alive_count_max: 'ClientAliveCountMax',
    login_grace_time: 'LoginGraceTime',
    minlen: 'minlen',
    minclass: 'minclass',
    dcredit: 'dcredit',
    ucredit: 'ucredit',
    lcredit: 'lcredit',
    ocredit: 'ocredit',
    password_remember: 'password_remember',
    passwd: 'passwd',
    passwd_minus: 'passwd_minus',
    group: 'group',
    group_minus: 'group_minus',
    shadow: 'shadow',
    shadow_minus: 'shadow_minus',
    gshadow: 'gshadow',
    gshadow_minus: 'gshadow_minus',
    crypto_policies: 'CryptoPolicies',
    ntp_server: 'NTPServer'
  }
  
  return labels[fieldName] || fieldName
}
