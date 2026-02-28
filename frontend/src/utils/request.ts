import axios from 'axios'
import { createDiscreteApi } from 'naive-ui'
import { useUserStore } from '../stores/user'

// 1. 创建 Naive UI 的独立 API
const { message, dialog } = createDiscreteApi(['message', 'dialog'])

// 2. 创建 axios 实例
const service = axios.create({
  baseURL: '/api/v1', 
  timeout: 15000 
})

// 3. 请求拦截器
service.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    const token = userStore.token || localStorage.getItem('token')

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 4. 响应拦截器
service.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const status = error.response?.status
    const data = error.response?.data
    const backendMsg = data?.message || data?.error || '请求服务异常'

    // --- A. 身份校验相关 (401) ---
    if (status === 401) {
      // 🔥 核心修复点：判断是否是登录请求
      // 通过 error.config.url 来识别请求路径
      const isLoginRequest = error.config.url?.includes('/auth/login')

      if (isLoginRequest) {
        // 如果是登录接口报 401，说明是账号或密码错误
        message.error(backendMsg) // 这里会显示 "账号或密码错误"
      } else {
        // 如果是其他接口报 401，说明 Token 失效
        message.warning('登录状态已过期，请重新登录')
        const userStore = useUserStore()
        userStore.logout()
        // 只有当前不在登录页时才重定向，避免重复跳转
        if (window.location.pathname !== '/login') {
          setTimeout(() => {
            window.location.href = '/login'
          }, 1000)
        }
      }
    } 
    
    // --- B. 权限不足 (403) ---
    else if (status === 403) {
      const isForbidden = data?.error === 'FORBIDDEN'
      dialog.warning({
        title: '🔒 访问受限',
        content: backendMsg,
        positiveText: isForbidden ? '去商城获取授权' : '知道了',
        negativeText: isForbidden ? '先等等' : undefined,
        maskClosable: false,
        onPositiveClick: () => {
           if (isForbidden) {
             window.location.href = '/payment-test'
           }
        }
      })
    }

    // --- 其他错误处理保持不变 ---
    else if (status === 404) {
      message.error('请求的资源不存在')
    }
    else if (status >= 500) {
      message.error('服务器繁忙，请稍后重试')
    }
    else {
      message.error(backendMsg)
    }

    return Promise.reject(error)
  }
)

export default service