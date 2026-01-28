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

  // ==========================================
  // 2. Action: 登录
  // ==========================================
  const login = async (loginForm: any) => {
    try {
      const res: any = await request.post('/auth/login', loginForm)
      
      console.log('Login Response:', res)

      if (res.token) {
        // --- 更新 State ---
        token.value = res.token
        
        // 🔥🔥🔥 核心修复：后端现在直接返回 id 了，直接拿即可！无需解析 Token 🔥🔥🔥
        // 这里的 res.id 对应后端返回的 data.id
        id.value = String(res.id || '') 
        
        username.value = res.username || ''
        role.value = res.role || 'user' 
        nickname.value = res.nickname || ''

        // --- 持久化 (存入浏览器缓存) ---
        localStorage.setItem('token', res.token)
        localStorage.setItem('id', String(id.value)) // 存入 ID
        localStorage.setItem('username', username.value)
        localStorage.setItem('role', role.value)
        localStorage.setItem('nickname', nickname.value)
        
        console.log('✅ 登录成功，当前用户 ID:', id.value, '角色:', role.value)
        return true
      }
      return false
    } catch (error) {
      console.error('登录请求失败:', error)
      return false
    }
  }

  // ==========================================
  // 3. Action: 登出
  // ==========================================
  const logout = () => {
    // 清空 State
    id.value = ''
    token.value = ''
    username.value = ''
    role.value = ''
    nickname.value = ''

    // 清空 LocalStorage
    localStorage.removeItem('token')
    localStorage.removeItem('id')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    localStorage.removeItem('nickname')
  }

  // 4. 导出给组件使用
  return { 
    id, 
    token, 
    username, 
    nickname,
    role, 
    login, 
    logout 
  }
})