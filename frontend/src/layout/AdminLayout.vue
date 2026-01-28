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
  ShieldCheckmarkOutline 
} from '@vicons/ionicons5'
// 修正引用路径
import { useUserStore } from '../stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const message = useMessage()

// 控制侧边栏折叠状态
const collapsed = ref(false)

// 计算头像地址（仅用于右上角）
const adminAvatar = computed(() => {
  if (userStore.avatar) return `http://localhost:8080${userStore.avatar}`
  return undefined
})

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 📋 菜单配置
const menuOptions = [
  { 
    label: '用户权限管理', 
    key: 'user-manage', 
    icon: renderIcon(PersonOutline) 
  },
  { 
    label: '业务授权管理', 
    key: 'auth-manage', 
    icon: renderIcon(CardOutline) 
  },
  { 
    label: '授权审计', 
    key: 'audit-log', 
    icon: renderIcon(ShieldCheckmarkOutline) 
  },
  { 
    label: '商品配置', 
    key: 'product-manage', 
    icon: renderIcon(SettingsOutline) 
  },
  { 
    label: '资源管理', 
    key: 'resource-manage', 
    icon: renderIcon(FolderOpenOutline) 
  },
  { 
    label: '返回前台刷题', 
    key: 'back-home', 
    icon: renderIcon(HomeOutline) 
  }
]

// 选中状态逻辑
const activeKey = computed(() => {
  const name = route.name as string
  if (name === 'UserManagement') return 'user-manage'
  if (name === 'UserAuthManager') return 'auth-manage'
  if (name === 'AuditLogs') return 'audit-log'
  if (name === 'ProductManager') return 'product-manage'
  if (name === 'ResourceManager') return 'resource-manage' 
  return null
})

// 菜单点击跳转逻辑
const handleMenuUpdate = (key: string) => {
  switch (key) {
    case 'user-manage': router.push('/admin/users'); break;
    case 'auth-manage': router.push('/admin/user-auths'); break;
    case 'audit-log': router.push('/admin/audit-logs'); break;
    case 'product-manage': router.push('/admin/products'); break;
    case 'resource-manage': router.push('/admin/resources'); break;
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
          <span v-show="!collapsed" class="logo-title">管理控制台</span>
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

/* 简单的淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>