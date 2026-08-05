import request from './request'

/**
 * 获取看板统计数据
 * @returns {Promise}
 */
export function getDashboardStats() {
  return request({
    url: '/dashboard/stats',
    method: 'get'
  })
}
