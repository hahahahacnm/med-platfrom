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
  {
    path: '/verify-email',
    name: 'VerifyEmail',
    component: () => import('../views/auth/VerifyEmail.vue'),
    meta: { title: '邮箱验证' }
  },

  // ============================
  // 🟠 2. 用户页面 (需登录) - 使用 MainLayout
  // ============================
  {
    path: '/',
    component: () => import('../layout/MainLayout.vue'),
    redirect: '/home', 
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
      {
        path: '/payment-test',
        name: 'PaymentTest',
        component: PaymentTest,
        meta: { title: '订阅中心' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/personal/Profile.vue'),
        meta: { title: '个人中心' }
      },
      {
        path: 'feedback', // 访问路径 /feedback
        name: 'PlatformFeedback',
        component: () => import('../views/PlatformFeedback.vue'),
        meta: { title: '意见反馈' }
      },
      {
        path: '/forum',
        component: () => import('../views/forum/ForumHome.vue')
      },
      // 🔥🔥🔥 论坛路由 🔥🔥🔥
      {
        path: '/forum/board/:id',
        name: 'BoardDetail',
        component: () => import('../views/forum/BoardDetail.vue'),
        meta: { title: '板块详情' }
      },
      {
        path: 'post/:id',
        name: 'PostDetail',
        component: () => import('../views/forum/PostDetail.vue'),
        meta: { title: '帖子详情' }
      },
      // 🔥🔥🔥 新增：题目详情页 (用于通知跳转) 🔥🔥🔥
      {
        path: 'question/:id',
        name: 'QuestionDetail',
        component: () => import('../views/QuestionDetail.vue'),
        meta: { title: '题目详情' }
      },
    ]
  },

  // ============================
  // 🔴 3. 管理员后台 (嵌套路由 + 布局)
  // ============================
  {
    path: '/admin',
    component: AdminLayout, 
    meta: { requiresAuth: true, requiresAdmin: true }, 
    children: [
      // 🔥🔥🔥 默认跳转到控制台 🔥🔥🔥
      {
        path: '',
        name: 'AdminDashboard', 
        component: () => import('../views/admin/Dashboard.vue'),
        meta: { title: '控制台' }
      },
      {
        path: 'configs',
        name: 'SystemConfig',
        component: () => import('../views/admin/SystemConfig.vue'),
        meta: { title: '平台参数管理', roles: ['admin'] }
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('../views/admin/UserManagement.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: 'resources',
        name: 'ResourceManager',
        component: () => import('../views/admin/ResourceManager.vue'),
        meta: { title: '资源管理' }
      },
      {
        path: 'feedbacks',
        name: 'FeedbackManager',
        component: () => import('../views/admin/FeedbackManager.vue'),
        meta: { title: '题目纠错' }
      },
      {
        path: 'user-auths',
        name: 'UserAuthManager',
        component: UserAuthManager,
        meta: { title: '业务授权' }
      },
      {
        path: 'codes',
        name: 'CodeManager',
        component: () => import('../views/admin/CodeManager.vue'),
        meta: { title: '卡密管理', roles: ['admin'] } // 这里注意：只有超管能发卡密，普通代理不能进
      },
      {
        path: 'products',
        name: 'ProductManager',
        component: () => import('../views/admin/ProductManager.vue'), 
        meta: { title: '商品配置' }
      },
      {
        path: 'discount-settings',
        name: 'DiscountSettings',
        component: () => import('../views/admin/DiscountSettings.vue'),
        meta: { title: '优惠策略配置' }
      },
      {
        path: 'audit-logs', 
        name: 'AuditLogs',
        component: () => import('../views/admin/AuditLogManager.vue'),
        meta: { title: '授权审计' }
      },
      {
        path: 'notes',
        name: 'NoteManagement',
        component: () => import('../views/admin/NoteManagement.vue'),
        meta: { title: '评论管理' }
      },
      {
        path: 'platform-feedbacks',
        name: 'PlatformFeedbackManager',
        component: () => import('../views/admin/PlatformFeedbackManager.vue'),
        meta: { title: '平台反馈管理' }
      },
      {
        path: 'forum', // 访问 /admin/forum
        name: 'AdminForum',
        component: () => import('../views/admin/Forum.vue'),
        meta: { title: '论坛管理' }
      },
      {
        path: 'mail-center', // 或者你喜欢的路径
        name: 'AdminMailCenter',
        component: () => import('../views/admin/MailCenter.vue'),
        meta: { requiresAuth: true, roles: ['admin', 'superadmin'] }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 🔥🔥🔥 增强版路由守卫 🔥🔥🔥
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const userRole = localStorage.getItem('role')

  // 🔥 核心修改：将 VerifyEmail 加入白名单，允许无 token 访问
  const whiteList = ['Login', 'Register', 'VerifyEmail']

  // 1. 检查 Token
  if (!token && !whiteList.includes(to.name as string)) {
    return next({ name: 'Login' })
  }

  // 2. 已登录防回退 (如果已登录且尝试访问登录、注册等页面，重定向到首页)
  if (token && whiteList.includes(to.name as string)) {
    // 允许已登录用户重新验证邮箱（换绑场景）
    if (to.name !== 'VerifyEmail') {
      return next({ name: 'Home' })
    }
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