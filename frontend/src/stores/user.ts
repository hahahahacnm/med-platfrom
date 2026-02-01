import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '../utils/request'

export const useUserStore = defineStore('user', () => {
  // ==========================================
  // 1. State: 状态定义
  // ==========================================
  const id = ref(localStorage.getItem('id') || '')
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')
  const role = ref(localStorage.getItem('role') || '')
  const nickname = ref(localStorage.getItem('nickname') || '')
  // 头像状态
  const avatar = ref(localStorage.getItem('avatar') || '')

  // ==========================================
  // 🔥🔥🔥 新增核心方法：手动设置并持久化头像 🔥🔥🔥
  // ==========================================
  const setAvatar = (newUrl: string) => {
    // 1. 更新内存状态 (Pinia)
    avatar.value = newUrl
    // 2. 更新硬盘缓存 (LocalStorage)
    localStorage.setItem('avatar', newUrl)
  }

  // ==========================================
  // 2. Action: 登录
  // ==========================================
  const login = async (loginForm: any) => {
    try {
      const res: any = await request.post('/auth/login', loginForm)

      if (res.token) {
        token.value = res.token
        id.value = String(res.id || '')
        username.value = res.username || ''
        role.value = res.role || 'user'
        nickname.value = res.nickname || ''

        // 使用封装的方法设置头像
        setAvatar(res.avatar || '')

        localStorage.setItem('token', res.token)
        localStorage.setItem('id', String(id.value))
        localStorage.setItem('username', username.value)
        localStorage.setItem('role', role.value)
        localStorage.setItem('nickname', nickname.value)

        return true
      }
      return false
    } catch (error) {
      console.error('登录请求失败:', error)
      throw error // 🔥 Rethrow needed to handle logic in UI (e.g. Captcha)
    }
  }

  // ==========================================
  // 3. Action: 从后端拉取最新资料
  // ==========================================
  const fetchProfile = async () => {
    try {
      const res: any = await request.get('/user/profile')
      if (res.data) {
        nickname.value = res.data.nickname || username.value
        role.value = res.data.role || 'user'

        // 🔥 这里也调用 setAvatar 确保同步
        setAvatar(res.data.avatar || '')

        localStorage.setItem('nickname', nickname.value)
        localStorage.setItem('role', role.value)
      }
    } catch (e) {
      console.error('刷新用户信息失败', e)
    }
  }

  // ==========================================
  // 4. Action: 登出
  // ==========================================
  const logout = () => {
    id.value = ''
    token.value = ''
    username.value = ''
    role.value = ''
    nickname.value = ''
    avatar.value = ''

    localStorage.clear() // 简单粗暴清空所有
  }

  // 5. 导出
  return {
    id, token, username, nickname, avatar, role,
    login, logout, fetchProfile, setAvatar // 🔥 别忘了导出 setAvatar
  }
})