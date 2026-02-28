import router from '../router'
import { useUserStore } from '../stores/user'
import { createDiscreteApi } from 'naive-ui'

// 为了在路由守卫里使用 Message，需要独立创建实例
const { message } = createDiscreteApi(['message'])

// 白名单：不需要登录，或者“即使信息不全”也能访问的页面

// 1. 修改白名单数组，加入 '/verify-email'
const whiteList = ['/login', '/register', '/profile', '/verify-email']

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const token = userStore.token

  // 1. 如果没有 Token
  if (!token) {
    // 现在 to.path 为 '/verify-email' 时，会进入这个 if 分支并成功执行 next()
    if (whiteList.includes(to.path)) {
      next() 
    } else {
      next('/login') 
    }
    return
  }
  // 2. 如果有 Token (已登录)
  if (token) {
    // 如果已登录还想去登录页，直接踢回首页 (或让后续逻辑判断去 profile)
    if (to.path === '/login') {
      next('/')
      return
    }

    // 🔥🔥🔥 核心拦截：检查信息是否完整 🔥🔥🔥
    if (userStore.isProfileIncomplete) {
      // 只有去“个人资料页”才放行，去其他任何页面都拦截
      if (to.path !== '/profile') {
        
        // 避免在跳转过程中重复弹窗
        if (from.path !== '/profile') { 
           message.warning('新用户请先完善 所在学校、专业及年级 信息！')
        }
        
        next('/profile') // 强制重定向
        return
      }
    }
  }

  // 3. 其他情况正常放行
  next()
})