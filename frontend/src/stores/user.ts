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
  // 🔥🔥🔥 必须补上 avatar，否则左上角头像是空的 🔥🔥🔥
  const avatar = ref(localStorage.getItem('avatar') || '')

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
        id.value = String(res.id || '') 
        
        username.value = res.username || ''
        role.value = res.role || 'user' 
        nickname.value = res.nickname || ''
        // 🔥 保存头像到 State
        avatar.value = res.avatar || ''

        // --- 持久化 (存入浏览器缓存) ---
        localStorage.setItem('token', res.token)
        localStorage.setItem('id', String(id.value))
        localStorage.setItem('username', username.value)
        localStorage.setItem('role', role.value)
        localStorage.setItem('nickname', nickname.value)
        // 🔥 保存头像到本地缓存
        localStorage.setItem('avatar', avatar.value)
        
        console.log('✅ 登录成功，当前用户 ID:', id.value, '头像:', avatar.value)
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
    avatar.value = '' // 🔥 清空头像状态

    // 清空 LocalStorage
    localStorage.removeItem('token')
    localStorage.removeItem('id')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    localStorage.removeItem('nickname')
    localStorage.removeItem('avatar') // 🔥 移除头像缓存
  }

  // 4. 导出给组件使用
  return { 
    id, 
    token, 
    username, 
    nickname,
    avatar, // 🔥 必须导出！Dashboard.vue 才能用 userStore.avatar
    role, 
    login, 
    logout 
  }
})
