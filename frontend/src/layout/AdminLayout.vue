<script setup lang="ts">
import { h, computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { 
  NLayout, NLayoutSider, NLayoutHeader, NLayoutContent, NMenu, 
  NIcon, NDropdown, NAvatar, useMessage
} from 'naive-ui'
import { 
  PersonOutline, 
  LogOutOutline, 
  HomeOutline, 
  SettingsOutline, 
  FolderOpenOutline, 
  CardOutline,
  ShieldCheckmarkOutline,
  ChatboxEllipsesOutline,
  BuildOutline,
  AlertCircleOutline,
  NewspaperOutline,
  PricetagOutline, 
  PeopleOutline,
  SpeedometerOutline,
  TicketOutline,
  PaperPlaneOutline // 🔥 新增：邮件小飞机图标
} from '@vicons/ionicons5'
import { useUserStore } from '../stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const message = useMessage()

// 控制侧边栏折叠状态
const collapsed = ref(false)

// 计算头像地址（仅用于右上角）
const adminAvatar = computed(() => {
  if (userStore.avatar) return userStore.avatar.startsWith('http') ? userStore.avatar : `http://localhost:8080${userStore.avatar}`
  return undefined
})

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 🔥🔥🔥 核心：动态菜单过滤 🔥🔥🔥
const menuOptions = computed(() => {
  const role = userStore.role // 'admin' 或 'agent'

  const allMenus = [
    // --- 控制台 (所有人可见) ---
    {
      label: '控制台',
      key: 'dashboard',
      icon: renderIcon(SpeedometerOutline),
      roles: ['admin', 'agent']
    },
    {
      label: '平台参数管理',
      key: 'sys-config',
      icon: renderIcon(SettingsOutline),
      roles: ['admin']
    },

    // --- 仅 Admin 可见 ---
    { 
      label: '用户权限管理', 
      key: 'user-manage', 
      icon: renderIcon(PersonOutline),
      roles: ['admin'] 
    },
    {
      label: '卡密管理',
      key: 'code-manage',
      icon: renderIcon(TicketOutline),
      roles: ['admin']
    },
    // 🔥🔥🔥 新增：邮件营销中心 (仅 Admin 可见) 🔥🔥🔥
    {
      label: '系统群发与邮件',
      key: 'mail-center',
      icon: renderIcon(PaperPlaneOutline),
      roles: ['admin']
    },

    // --- Agent & Admin 通用 ---
    { 
      label: '业务授权管理', 
      key: 'auth-manage', 
      icon: renderIcon(CardOutline),
      roles: ['admin', 'agent']
    },
    { 
      label: '授权审计', 
      key: 'audit-log', 
      icon: renderIcon(ShieldCheckmarkOutline),
      roles: ['admin', 'agent']
    },
    
    // --- 优惠策略 (分角色显示) ---
    {
      label: '我的优惠策略',
      key: 'my-discount',
      icon: renderIcon(PricetagOutline),
      roles: ['agent'] // 代理看这个
    },
    {
      label: '代理定价监控',
      key: 'agent-discount-monitor',
      icon: renderIcon(PeopleOutline),
      roles: ['admin'] // 管理员看这个
    },

    // --- 仅 Admin 可见 ---
    { 
      label: '商品配置', 
      key: 'product-manage', 
      icon: renderIcon(SettingsOutline),
      roles: ['admin']
    },
    
    // --- Agent & Admin 通用 (内容审核) ---
    { 
      label: '评论与举报', 
      key: 'note-manage', 
      icon: renderIcon(ChatboxEllipsesOutline),
      roles: ['admin', 'agent']
    },
    
    // --- 仅 Admin 可见 (资源安全) ---
    { 
      label: '资源管理', 
      key: 'resource-manage', 
      icon: renderIcon(FolderOpenOutline),
      roles: ['admin']
    },
    
    // --- Agent & Admin 通用 (运营) ---
    {
      label: '论坛/公告管理',
      key: 'forum-manage',
      icon: renderIcon(NewspaperOutline),
      roles: ['admin', 'agent']
    },
    {
      label: '题目纠错',
      key: 'feedback-manage', 
      icon: renderIcon(BuildOutline),
      roles: ['admin', 'agent']
    },
    { 
      label: '平台反馈', 
      key: 'platform-feedback-manage', 
      icon: renderIcon(AlertCircleOutline),
      roles: ['admin', 'agent']
    },
    
    // --- 所有人 ---
    { 
      label: '返回前台刷题', 
      key: 'back-home', 
      icon: renderIcon(HomeOutline),
      roles: ['admin', 'agent']
    }
  ]

  // 过滤逻辑
  return allMenus.filter(item => item.roles.includes(role))
})

// ✅ 选中状态逻辑
const activeKey = computed(() => {
  const name = route.name as string
  const role = userStore.role
  
  if (name === 'AdminDashboard') return 'dashboard'
  if (name === 'UserManagement') return 'user-manage'
  if (name === 'CodeManager') return 'code-manage'
  
  // 🔥 新增：邮件营销高亮判断
  if (name === 'AdminMailCenter') return 'mail-center'
  
  if (name === 'UserAuthManager') return 'auth-manage'
  if (name === 'AuditLogs') return 'audit-log'
  if (name === 'ProductManager') return 'product-manage'
  if (name === 'NoteManagement') return 'note-manage'
  if (name === 'ResourceManager') return 'resource-manage'
  if (name === 'AdminForum') return 'forum-manage' 
  if (name === 'FeedbackManager') return 'feedback-manage'
  if (name === 'PlatformFeedbackManager') return 'platform-feedback-manage'
  
  if (name === 'DiscountSettings') {
    return role === 'agent' ? 'my-discount' : 'agent-discount-monitor'
  }
  
  return null
})

// ✅ 菜单点击跳转逻辑
const handleMenuUpdate = (key: string) => {
  switch (key) {
    case 'dashboard': router.push('/admin'); break;
    case 'sys-config': router.push('/admin/configs'); break;
    case 'user-manage': router.push('/admin/users'); break;
    case 'code-manage': router.push('/admin/codes'); break;
    
    // 🔥 新增跳转：前往邮件营销页面
    case 'mail-center': router.push('/admin/mail-center'); break;
    
    case 'auth-manage': router.push('/admin/user-auths'); break;
    case 'audit-log': router.push('/admin/audit-logs'); break;
    case 'product-manage': router.push('/admin/products'); break;
    case 'note-manage': router.push('/admin/notes'); break;
    case 'resource-manage': router.push('/admin/resources'); break;
    case 'forum-manage': router.push('/admin/forum'); break; 
    case 'feedback-manage': router.push('/admin/feedbacks'); break;
    case 'platform-feedback-manage': router.push('/admin/platform-feedbacks'); break;
    
    case 'my-discount': 
    case 'agent-discount-monitor':
      router.push('/admin/discount-settings'); 
      break;
      
    case 'back-home': router.push('/'); break;
  }
}

const userOptions = [{ label: '退出登录', key: 'logout', icon: renderIcon(LogOutOutline) }]

const handleUserSelect = (key: string) => {
  if (key === 'logout') {
    userStore.logout()
    router.push('/login')
    message.success('已退出')
  }
}
</script>

<template>
  <div class="admin-layout">
    <n-layout has-sider position="absolute">
      <n-layout-sider
        bordered 
        collapse-mode="width" 
        :collapsed-width="64" 
        :width="220"
        show-trigger 
        inverted 
        v-model:collapsed="collapsed"
        style="background-color: #001529;"
      >
        <div class="logo">
          <n-icon size="28" color="#18a058"><SettingsOutline /></n-icon>
          <span v-show="!collapsed" class="logo-title">
             {{ userStore.role === 'agent' ? '代理控制台' : '系统管理' }}
          </span>
        </div>

        <n-menu
          :collapsed-width="64" 
          :icon-size="20"
          :options="menuOptions" 
          :value="activeKey"
          @update:value="handleMenuUpdate"
          inverted
        />
      </n-layout-sider>

      <n-layout>
        <n-layout-header bordered style="height: 60px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between;">
          <div style="font-size: 16px; font-weight: bold; color: #333;">
            {{ route.meta.title || '系统管理' }}
          </div>
          
          <n-dropdown :options="userOptions" @select="handleUserSelect">
            <div style="display: flex; align-items: center; cursor: pointer;">
              <n-avatar 
                round 
                size="small" 
                :src="adminAvatar"
                fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
                style="margin-right: 8px; border: 1px solid #eee;"
              />
              <span>{{ userStore.nickname || userStore.username }}</span>
              <n-tag v-if="userStore.role === 'agent'" type="info" size="tiny" style="margin-left: 6px">代理</n-tag>
            </div>
          </n-dropdown>
        </n-layout-header>

        <n-layout-content content-style="padding: 20px; background-color: #f0f2f5; min-height: calc(100vh - 60px);">
          <router-view v-slot="{ Component }">
             <transition name="fade" mode="out-in">
               <component :is="Component" />
             </transition>
          </router-view>
        </n-layout-content>
      </n-layout>
    </n-layout>
  </div>
</template>

<style scoped>
.logo { 
  height: 60px; 
  display: flex; 
  align-items: center; 
  justify-content: center; 
  color: #fff; 
  border-bottom: 1px solid rgba(255,255,255,0.1); 
  overflow: hidden; 
  transition: all 0.3s;
}

.logo-title {
  margin-left: 10px; 
  font-weight: bold; 
  font-size: 16px;
  white-space: nowrap; 
  opacity: 1;
  transition: opacity 0.3s;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>