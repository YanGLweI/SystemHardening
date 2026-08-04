import axios from 'axios'
import { Message } from 'element-ui'

const service = axios.create({
  baseURL: process.env.VUE_APP_BASE_API || 'http://localhost:8080/api',
  timeout: 15000
})

// 请求拦截器
service.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    // 如果没有响应对象（比如网络错误），不要尝试访问它
    if (!error.response) {
      console.error('Network error or no response:', error)
      Message.error('无法连接到服务器')
      return Promise.reject(error)
    }
    
    const status = error.response.status
    
    switch (status) {
      case 401:
        // Token 失效或过期
        localStorage.removeItem('token')
        localStorage.removeItem('username')
        localStorage.removeItem('rememberMe')
        Message.error({
          message: '会话已过期，请重新登录',
          duration: 2000
        })
        break
        
      case 403:
        Message.error('无权访问此资源')
        break
        
      case 404:
        Message.error('接口不存在')
        break
        
      case 500:
        Message.error('服务器内部错误')
        break
        
      default:
        (function() {
          const errorMsg = error.response.data?.error || 
                          error.response.data?.message || 
                          '网络错误或服务器异常'
          Message.error(errorMsg)
        })()
    }
    
    return Promise.reject(error)
  }
)

export default service
