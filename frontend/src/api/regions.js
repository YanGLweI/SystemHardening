import request from './request'

/**
 * 创建区域
 * @param {Object} data - 创建数据
 * @param {String} data.name - 区域名称
 * @returns {Promise}
 */
export function createRegion(data) {
  return request({
    url: '/regions',
    method: 'post',
    data
  })
}

/**
 * 获取所有区域列表
 * @returns {Promise}
 */
export function listRegions() {
  return request({
    url: '/regions',
    method: 'get'
  })
}

/**
 * 更新区域客户端关联
 * @param {Number} id - 区域 ID
 * @param {Array} clientIds - 客户端 ID 数组
 * @returns {Promise}
 */
export function updateRegionClients(id, clientIds) {
  return request({
    url: `/regions/${id}/clients`,
    method: 'put',
    data: { client_ids: clientIds }
  })
}

/**
 * 删除区域
 * @param {Number} id - 区域 ID
 * @returns {Promise}
 */
export function deleteRegion(id) {
  return request({
    url: `/regions/${id}`,
    method: 'delete'
  })
}
