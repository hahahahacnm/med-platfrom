import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import MyNotes from '../views/MyNotes.vue'
import UserAuthManager from '../views/admin/UserAuthManager.vue'

// 引入布局组件
import AdminLayout from '../layout/AdminLayout.vue'
import PaymentTest from '../views/PaymentTest.vue'

const routes = [
  // ============================
  // 🟢 1. 公共页面 (无需登录)
  // ============================
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/register',
    name: 'Register',
    component: Register
  },

  // ============================
  // 🟠 2. 用户页面 (需登录)
  // ============================
  // ============================
  // 🟠 2. 用户页面 (需登录) - 使用 MainLayout
  // ============================
  {
    path: '/',
    component: () => import('../layout/MainLayout.vue'),
    redirect: '/home', // 可选：如果希望默认路径清晰显示
    meta: { requiresAuth: true },
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('../views/Home.vue'),
        meta: { title: '总览' }
      },
      {
        path: 'quiz',
        name: 'QuizBank',
        component: () => import('../views/QuizBank.vue'),
        meta: { title: '题库' }
      },
      {
        path: 'mistakes',
        name: 'Mistakes',
        component: () => import('../views/Mistakes.vue'),
        meta: { title: '错题集' }
      },
      {
        path: 'favorites',
        name: 'Favorites',
        component: () => import('../views/Favorites.vue'),
        meta: { title: '收藏夹' }
      },
      {
        path: 'my-notes',
        name: 'MyNotes',
        component: MyNotes,
        meta: { title: '我的笔记' }
      },
      // 把它放在 Home 的同一级，或者根据你的布局需求放置
      {
        path: '/payment-test',
        name: 'PaymentTest',
        component: PaymentTest,
        meta: { title: '订阅中心' } // 需要登录才能买
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/personal/Profile.vue'),
        meta: { title: '个人中心' }
      },
    ]
  },

  // ============================
  // 🔴 3. 管理员后台 (嵌套路由 + 布局)
  // ============================
  {
    path: '/admin',
    component: AdminLayout, // 🔥 使用带侧边栏的布局
    meta: { requiresAuth: true, requiresAdmin: true }, // 只有管理员能进
    children: [
      {
        path: '',
        redirect: '/admin/users' // 默认跳到用户管理
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('../views/admin/UserManagement.vue'),
        meta: { title: '用户管理' }
      },

      // 🔥 资源管理器
      {
        path: 'resources',
        name: 'ResourceManager',
        component: () => import('../views/admin/ResourceManager.vue'),
        meta: { title: '资源管理' }
      },

      // 🔥 业务授权
      {
        path: 'user-auths',
        name: 'UserAuthManager',
        component: UserAuthManager,
        meta: { title: '业务授权' }
      },

      // 🔥🔥🔥 [新增] 商品管理入口 🔥🔥🔥
      {
        path: 'products',
        name: 'ProductManager',
        component: () => import('../views/admin/ProductManager.vue'),
        meta: { title: '商品配置' }
      },
      {
        path: '/admin/audit-logs',
        name: 'AuditLogs',
        component: () => import('../views/admin/AuditLogManager.vue'),
        meta: { title: '授权审计' }
      },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 🔥🔥🔥 增强版路由守卫 🔥🔥🔥
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const userRole = localStorage.getItem('role')

  const whiteList = ['Login', 'Register']

  // 1. 检查 Token
  if (!token && !whiteList.includes(to.name as string)) {
    return next({ name: 'Login' })
  }

  // 2. 已登录防回退
  if (token && whiteList.includes(to.name as string)) {
    return next({ name: 'Home' })
  }

  // 3. 🛡️ 权限检查
  if (to.meta.requiresAdmin) {
    if (userRole !== 'admin' && userRole !== 'agent') {
      alert('权限不足：非管理员禁止访问')
      return next({ name: 'Home' })
    }
  }

  // 4. 放行
  next()
})

export default router
