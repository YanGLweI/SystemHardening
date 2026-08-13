import request from './request'

/**
 * 获取客户端列表
 * @param {Object} params - 查询参数
 * @returns {Promise}
 */
export function getClientList(params) {
  return request({
    url: '/clients',
    method: 'get',
    params: params
  })
}

/**
 * 获取全部客户端列表（循环分页拉取全量，避免分页遗漏）
 * @param {Object} params - 查询参数（如 os_type）
 * @returns {Promise<Array>}
 */
export async function getAllClients(params = {}) {
  const pageSize = 100
  const all = []
  let page = 1
  while (true) {
    const res = await getClientList({ ...params, page, pageSize })
    const items = res.list || res || []
    all.push(...items)
    const total = res.total !== undefined ? res.total : all.length
    if (all.length >= total || items.length === 0) break
    page++
  }
  return all
}

/**
 * 删除客户端
 * @param {Number} id - 客户端 ID
 * @returns {Promise}
 */
export function deleteClient(id) {
  return request({
    url: `/clients/${id}`,
    method: 'delete'
  })
}

/**
 * 获取加固检查计划
 * @returns {Promise}
 */
export function getCheckSchedule() {
  return request({
    url: '/check-schedule',
    method: 'get'
  })
}

/**
 * 保存加固检查计划
 * @param {Object} data - 计划数据（schedule_type/check_time/weekday/day_of_month）
 * @returns {Promise}
 */
export function saveCheckSchedule(data) {
  return request({
    url: '/check-schedule',
    method: 'put',
    data: data
  })
}
