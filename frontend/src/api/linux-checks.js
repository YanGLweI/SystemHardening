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
