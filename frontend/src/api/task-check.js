import request from '@/api/request'

/**
 * 触发立即检查任务
 * @param {string} clientUuid - 客户端 UUID
 * @returns {Promise}
 */
export function triggerCheck(clientUuid) {
  return request({
    url: '/tasks/trigger',
    method: 'post',
    data: {
      client_uuid: clientUuid
    }
  })
}

/**
 * 查询任务状态
 * @param {string} taskId - 任务 ID
 * @returns {Promise}
 */
export function getTaskStatus(taskId) {
  return request({
    url: `/tasks/${taskId}`,
    method: 'get'
  })
}

/**
 * 删除任务（用于卡死任务重试）
 * @param {string} taskId - 任务 ID
 * @returns {Promise}
 */
export function deleteTask(taskId) {
  return request({
    url: `/tasks/${taskId}`,
    method: 'delete'
  })
}

/**
 * 获取客户端的最新任务（如果有）
 * @param {string} clientUuid - 客户端 UUID
 * @returns {Promise}
 */
export function getClientLatestTask(clientUuid) {
  return request({
    url: `/tasks/client/${clientUuid}`,
    method: 'get'
  })
}
