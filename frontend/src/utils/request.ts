import axios from 'axios'
import { createDiscreteApi } from 'naive-ui'
// 注意：这里引入 store 是为了在拦截器中使用，但不要在文件顶部直接实例化
import { useUserStore } from '../stores/user'

// 1. 创建 Naive UI 的独立 API
// message: 用于轻量提示
// dialog: 用于 403 这种需要用户确认的“模态框”警告
const { message, dialog } = createDiscreteApi(['message', 'dialog'])

// 2. 创建 axios 实例
const service = axios.create({
  // 配合 vite.config.ts 的 proxy
  baseURL: '/api/v1',
  // 建议改长一点，题库导入导出或大列表查询可能耗时
  timeout: 15000
})

// 3. 请求拦截器
service.interceptors.request.use(
  (config) => {
    // 只有在请求真正发起时才获取 store，避免 Pinia 未初始化报错
    const userStore = useUserStore()
    const token = userStore.token || localStorage.getItem('token')

    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 4. 响应拦截器
service.interceptors.response.use(
  (response) => {
    // 成功直接返回数据
    return response.data
  },
  (error) => {
    const status = error.response?.status
    // 优先取后端返回的 error 字段，其次是 message，最后是默认文案
    const errorMsg = error.response?.data?.error || error.response?.data?.message || '请求服务异常'

    // --- A. Token 过期 / 未登录 (401) ---
    if (status === 401) {
      // 如果是在登录页（登录失败），不进行跳转逻辑
      if (!window.location.pathname.includes('/login')) {
        message.warning('登录状态已过期，请重新登录')
        const userStore = useUserStore()
        userStore.logout()
        setTimeout(() => {
          window.location.href = '/login'
        }, 1000)
      } else {
        message.error(errorMsg)
      }
    }

    // --- B. 权限不足 (403) - 配合后端 checkAccess 使用 ---
    else if (status === 403) {
      // 使用 Dialog 模态框，比 Message 更醒目，强制用户看到
      dialog.warning({
        title: '🔒 访问受限',
        content: errorMsg || '您暂无该内容的访问权限，请联系管理员或获取授权。',
        positiveText: '知道了',
        maskClosable: false // 禁止点击遮罩关闭
      })
    }

    // --- C. 资源不存在 (404) ---
    else if (status === 404) {
      message.error('请求的资源不存在')
    }

    // --- D. 服务器内部错误 (500) ---
    else if (status >= 500) {
      message.error('服务器繁忙，请稍后重试')
    }

    // --- E. 网络超时或其他 ---
    else if (error.code === 'ECONNABORTED' || error.message.includes('timeout')) {
      message.error('请求超时，请检查网络')
    }

    // --- F. 其他错误 ---
    else {
      message.error(errorMsg)
    }

    return Promise.reject(error)
  }
)

export default service