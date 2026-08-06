import request from './request'

/**
 * 上传安装包
 * @param {string} type - 包类型：linux 或 windows
 * @param {File} file - 上传的文件对象
 */
export function uploadPackage(type, file) {
  const formData = new FormData()
  formData.append('type', type)
  formData.append('package', file)
  
  return request({
    url: '/packages/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 获取包信息（大小和哈希值）
 * @param {string} type - 包类型：linux 或 windows
 */
export function getPackageInfo(type) {
  return request({
    url: `/packages/${type}/info`,
    method: 'get'
  })
}

/**
 * 下载安装包
 * @param {string} type - 包类型：linux 或 windows
 */
export function downloadPackage(type) {
  return request({
    url: `/packages/${type}/download`,
    method: 'get',
    responseType: 'blob' // 关键：指定返回类型为 blob
  })
}
