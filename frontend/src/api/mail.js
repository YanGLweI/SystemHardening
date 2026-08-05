import request from './request'

/**
 * 获取邮件配置
 * @returns {Promise}
 */
export function getMailConfig() {
  return request({
    url: '/mail-config',
    method: 'get'
  })
}

/**
 * 保存邮件配置
 * @param {Object} data - 配置数据
 * @param {String} data.smtp_host - SMTP 服务器地址
 * @param {Number} data.smtp_port - SMTP 端口
 * @param {String} data.username - 账号
 * @param {String} data.password - 密码
 * @param {String} data.from_email - 发件人邮箱（可选）
 * @param {Boolean} data.is_enabled - 是否启用
 * @returns {Promise}
 */
export function saveMailConfig(data) {
  return request({
    url: '/mail-config',
    method: 'put',
    data
  })
}

/**
 * 发送测试邮件
 * @param {String} recipient - 收件人邮箱
 * @returns {Promise}
 */
export function testEmail(recipient) {
  return request({
    url: '/mail/test',
    method: 'post',
    data: { recipient }
  })
}

/**
 * 获取报告计划列表
 * @param {Object} params - 查询参数
 * @param {Number} params.page - 页码
 * @param {Number} params.page_size - 每页数量
 * @returns {Promise}
 */
export function listSchedules(params) {
  return request({
    url: '/report-schedules',
    method: 'get',
    params
  })
}

/**
 * 创建报告计划
 * @param {Object} data - 计划数据
 * @param {String} data.name - 报告名称
 * @param {String} data.schedule_type - 频率类型（daily/every_n_days/weekly/every_n_weeks/monthly/every_n_months）
 * @param {String} data.send_time - 发送时间（HH:mm）
 * @param {Number} data.interval_days - 间隔天数（every_n_days 时有效）
 * @param {Number} data.weekday - 星期几（1-7，weekly 时有效）
 * @param {Number} data.day_of_month - 每月日期（monthly 时有效）
 * @param {Number} data.interval_weeks - 间隔周数（every_n_weeks 时有效）
 * @param {Number} data.interval_months - 间隔月数（every_n_months 时有效）
 * @param {String} data.recipients - 收件人（逗号分隔）
 * @param {String} data.subject - 邮件主题
 * @param {Boolean} data.is_enabled - 启用状态
 * @param {String} data.created_by - 创建人
 * @param {String} data.last_updated_by - 最后更新人
 * @returns {Promise}
 */
export function createSchedule(data) {
  return request({
    url: '/report-schedules',
    method: 'post',
    data
  })
}

/**
 * 更新报告计划
 * @param {Number} id - 计划 ID
 * @param {Object} data - 更新数据
 * @returns {Promise}
 */
export function updateSchedule(id, data) {
  return request({
    url: `/report-schedules/${id}`,
    method: 'put',
    data
  })
}

/**
 * 删除报告计划
 * @param {Number} id - 计划 ID
 * @returns {Promise}
 */
export function deleteSchedule(id) {
  return request({
    url: `/report-schedules/${id}`,
    method: 'delete'
  })
}

/**
 * 立即发送报告
 * @param {Number} id - 计划 ID
 * @returns {Promise}
 */
export function immediateSend(id) {
  return request({
    url: `/report-schedules/${id}/send`,
    method: 'post'
  })
}
