import request from '@/api/request'

/**
 * 获取 Linux 加固检查列表（分页）
 * @param {Object} params - 查询参数
 * @param {Number} params.page - 页码
 * @param {Number} params.pageSize - 每页数量
 */
export function getList(params) {
  return request({
    url: '/linux-checks',
    method: 'get',
    params
  })
}

/**
 * 获取单个主机加固检查详情
 * @param {String} id - 记录 ID
 */
export function getDetail(id) {
  return request({
    url: `/linux-checks/${id}`,
    method: 'get'
  })
}

/**
 * 批量创建标准配置
 * @param {Array} data - 标准配置数组
 */
export function createStandards(data) {
  return request({
    url: '/linux-standards',
    method: 'post',
    data
  })
}

/**
 * 获取标准配置列表
 */
export function listStandards() {
  return request({
    url: '/linux-standards',
    method: 'get'
  })
}

/**
 * 更新标准配置
 * @param {String} id - 记录 ID
 * @param {Object} data - 更新数据
 */
export function updateStandard(id, data) {
  return request({
    url: `/linux-standards/${id}`,
    method: 'put',
    data
  })
}

/**
 * 删除标准配置
 * @param {String} id - 记录 ID
 */
export function deleteStandard(id) {
  return request({
    url: `/linux-standards/${id}`,
    method: 'delete'
  })
}

/**
 * 获取可用的 Linux 加固字段列表（未配置的）
 */
export function getAvailableFields() {
  return request({
    url: '/linux-standards/fields',
    method: 'get'
  })
}
