import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
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
  const avatar = ref(localStorage.getItem('avatar') || '')
  
  // 🔥 新增：积分余额 (存储为数字)
  const points = ref(Number(localStorage.getItem('points') || 0))

  // 完善信息相关字段
  const school = ref(localStorage.getItem('school') || '')
  const major = ref(localStorage.getItem('major') || '')
  const grade = ref(localStorage.getItem('grade') || '')

  // ==========================================
  // 🔥 Getter: 计算属性
  // ==========================================
  // 判断“是否未完善个人信息”
  const isProfileIncomplete = computed(() => {
    // 只有普通用户 (user) 需要强制完善，管理员/代理不需要
    if (role.value === 'admin' || role.value === 'agent') {
      return false
    }
    return !school.value || !major.value || !grade.value
  })

  // ==========================================
  // Actions
  // ==========================================
  
  const setAvatar = (newUrl: string) => {
      avatar.value = newUrl
      localStorage.setItem('avatar', newUrl)
  }

  const updateUserInfo = (info: any) => {
    if (info.school) { school.value = info.school; localStorage.setItem('school', info.school) }
    if (info.major) { major.value = info.major; localStorage.setItem('major', info.major) }
    if (info.grade) { grade.value = info.grade; localStorage.setItem('grade', info.grade) }
    if (info.nickname) { nickname.value = info.nickname; localStorage.setItem('nickname', info.nickname) }
  }

  // 登录
  const login = async (loginForm: any) => {
    try {
      const res: any = await request.post('/auth/login', loginForm)
      
      if (res.token) {
        token.value = res.token
        id.value = String(res.id || '') 
        username.value = res.username || ''
        role.value = res.role || 'user' 
        nickname.value = res.nickname || ''
        avatar.value = res.avatar || ''
        
        // 🔥 保存积分
        points.value = res.points || 0
        
        school.value = res.school || ''
        major.value = res.major || ''
        grade.value = res.grade || ''

        // 持久化存储
        localStorage.setItem('token', res.token)
        localStorage.setItem('id', String(id.value))
        localStorage.setItem('username', username.value)
        localStorage.setItem('role', role.value)
        localStorage.setItem('nickname', nickname.value)
        localStorage.setItem('avatar', avatar.value)
        localStorage.setItem('points', String(points.value)) // 🔥
        
        localStorage.setItem('school', school.value)
        localStorage.setItem('major', major.value)
        localStorage.setItem('grade', grade.value)
        
        return true
      }
      return false
    } catch (error) {
      console.error('登录请求失败:', error)
      return false
    }
  }

  // 拉取最新资料 (包含积分)
  const fetchProfile = async () => {
      try {
          const res: any = await request.get('/user/profile')
          if (res.data) {
              const d = res.data
              nickname.value = d.nickname || username.value
              role.value = d.role || 'user'
              setAvatar(d.avatar || '')
              
              // 🔥 同步更新积分
              points.value = d.points || 0
              localStorage.setItem('points', String(points.value))

              // 同步更新状态
              school.value = d.school || ''
              major.value = d.major || ''
              grade.value = d.grade || ''

              localStorage.setItem('nickname', nickname.value)
              localStorage.setItem('role', role.value)
              localStorage.setItem('school', school.value)
              localStorage.setItem('major', major.value)
              localStorage.setItem('grade', grade.value)
          }
      } catch (e) {
          console.error('刷新用户信息失败', e)
      }
  }

  // 登出
  const logout = () => {
    id.value = ''
    token.value = ''
    username.value = ''
    role.value = ''
    nickname.value = ''
    avatar.value = '' 
    points.value = 0 // 🔥
    school.value = ''
    major.value = ''
    grade.value = ''

    localStorage.clear() 
  }

  return { 
    id, token, username, nickname, avatar, role, points, // 🔥 导出 points
    school, major, grade, isProfileIncomplete,
    login, logout, fetchProfile, setAvatar, updateUserInfo
  }
})