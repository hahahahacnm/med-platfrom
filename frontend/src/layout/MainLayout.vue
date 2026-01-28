<script setup lang="ts">
import { h, ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../stores/user'
import { 
  HomeOutline, 
  BookOutline, 
  StarOutline, 
  JournalOutline, 
  PersonOutline, 
  SettingsOutline,
  LogOutOutline,
  MenuOutline,
  CloseOutline,
  SchoolOutline,
  LibraryOutline
} from '@vicons/ionicons5'
import { NIcon, NAvatar, NDrawer, NDrawerContent } from 'naive-ui'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isMobileMenuOpen = ref(false)

const userAvatar = computed(() => {
  if (userStore.avatar) return userStore.avatar.startsWith('http') ? userStore.avatar : `http://localhost:8080${userStore.avatar}`
  return undefined
})

const menuItems = computed(() => [
  { label: '总览', key: 'Home', icon: HomeOutline, path: '/' },
  { label: '题库', key: 'QuizBank', icon: LibraryOutline, path: '/quiz' },
  { label: '错题集', key: 'Mistakes', icon: BookOutline, path: '/mistakes' },
  { label: '收藏夹', key: 'Favorites', icon: StarOutline, path: '/favorites' },
  { label: '笔记本', key: 'MyNotes', icon: JournalOutline, path: '/my-notes' },
  ...(userStore.role === 'admin' ? [{ label: '管理员后台', key: 'Admin', icon: SettingsOutline, path: '/admin' }] : [])
])

const handleNavigate = (path: string) => {
  router.push(path)
  isMobileMenuOpen.value = false
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="app-layout">
    <!-- Desktop Sidebar -->
    <aside class="sidebar desktop-sidebar">
      <div class="logo-area">
        <div class="logo-icon">
          <n-icon size="24" color="#fff"><SchoolOutline /></n-icon>
        </div>
        <span class="logo-text">题酷</span>
      </div>

      <nav class="nav-menu">
        <div 
          v-for="item in menuItems" 
          :key="item.key"
          class="nav-item"
          :class="{ active: route.name === item.key || route.path === item.path }"
          @click="handleNavigate(item.path)"
        >
          <n-icon size="20" class="nav-icon"><component :is="item.icon" /></n-icon>
          <span class="nav-label">{{ item.label }}</span>
        </div>
      </nav>

      <div class="user-profile-area" @click="handleNavigate('/profile')">
        <div class="avatar-wrapper">
          <n-avatar round size="small" :src="userAvatar" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
        </div>
        <div class="user-info">
          <div class="user-name">{{ userStore.username || '同学' }}</div>
          <div class="user-role">{{ userStore.role === 'admin' ? '管理员' : '普通用户' }}</div>
        </div>
      </div>
    </aside>

    <!-- Mobile Header -->
    <header class="mobile-header">
      <div class="logo-area-mobile">
        <div class="logo-icon-mobile">
          <n-icon size="20" color="#fff"><SchoolOutline /></n-icon>
        </div>
        <span class="logo-text-mobile">题酷</span>
      </div>
      <button class="menu-btn" @click="isMobileMenuOpen = true">
        <n-icon size="24"><MenuOutline /></n-icon>
      </button>
    </header>

    <!-- Mobile Drawer -->
    <n-drawer v-model:show="isMobileMenuOpen" placement="right" width="260">
      <n-drawer-content title="菜单" closable>
        <nav class="mobile-nav-menu">
           <div 
            v-for="item in menuItems" 
            :key="item.key"
            class="nav-item mobile-item"
            :class="{ active: route.name === item.key || route.path === item.path }"
            @click="handleNavigate(item.path)"
          >
            <n-icon size="20" class="nav-icon"><component :is="item.icon" /></n-icon>
            <span class="nav-label">{{ item.label }}</span>
          </div>
          <div class="nav-item mobile-item logout-item" @click="handleLogout">
             <n-icon size="20" class="nav-icon"><LogOutOutline /></n-icon>
             <span class="nav-label">退出登录</span>
          </div>
        </nav>
      </n-drawer-content>
    </n-drawer>

    <!-- Main Content -->
    <main class="main-content">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background-color: #f8fafc; /* slate-50 */
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
}

/* Sidebar Styles */
.sidebar {
  background-color: #fff;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  z-index: 20;
}

.desktop-sidebar {
  width: 250px;
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
}

.logo-area {
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid #f1f5f9;
}

.logo-icon {
  background-color: #2563eb; /* blue-600 */
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 20px;
  font-weight: 800;
  color: #2563eb;
}

.nav-menu {
  flex: 1;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #64748b; /* slate-500 */
}

.nav-item:hover {
  background-color: #f1f5f9; /* slate-100 */
  color: #0f172a; /* slate-900 */
}

.nav-item.active {
  background-color: #2563eb;
  color: #fff;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.nav-label {
  font-size: 14px;
  font-weight: 600;
}

.user-profile-area {
  padding: 16px;
  margin: 16px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  border-radius: 12px;
  transition: background-color 0.2s;
}

.user-profile-area:hover {
  background-color: #f8fafc;
}

.user-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.user-name {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 12px;
  color: #94a3b8;
}

/* Mobile Header */
.mobile-header {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid #e2e8f0;
  z-index: 50;
  padding: 0 16px;
  align-items: center;
  justify-content: space-between;
}

.logo-area-mobile {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-icon-mobile {
  background-color: #2563eb;
  padding: 6px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text-mobile {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
}

.menu-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #64748b;
  padding: 4px;
}

.mobile-nav-menu {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mobile-item {
  padding: 12px;
}

.logout-item {
  margin-top: 20px;
  color: #ef4444; /* red-500 */
}
.logout-item:hover {
  background-color: #fef2f2;
}

/* Main Content */
.main-content {
  flex: 1;
  margin-left: 250px;
  width: calc(100% - 250px);
  padding: 24px;
  height: 100vh;
  overflow-y: auto; 
  position: relative;
  box-sizing: border-box;
}

:global(body) {
  overflow: hidden;
}

/* Responsive */
@media (max-width: 768px) {
  .desktop-sidebar {
    display: none;
  }
  
  .mobile-header {
    display: flex;
  }

  .main-content {
    margin-left: 0;
    width: 100%;
    padding: 80px 16px 24px 16px; /* Reduced side padding from 24px to 16px */
  }
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
