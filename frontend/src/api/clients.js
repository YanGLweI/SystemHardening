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
