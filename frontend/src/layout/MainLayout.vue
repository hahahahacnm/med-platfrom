<script setup lang="ts">
import { h, ref, computed, onMounted } from 'vue'
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
  CheckmarkCircle,
  LibraryOutline,
  CartOutline,
  ChatboxEllipsesOutline // 🔥 社区图标
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

// 🔥 菜单配置
const menuItems = computed(() => [
  { label: '总览', key: 'Home', icon: HomeOutline, path: '/' },
  { label: '题库', key: 'QuizBank', icon: LibraryOutline, path: '/quiz' },
  { label: '错题集', key: 'Mistakes', icon: BookOutline, path: '/mistakes' },
  { label: '收藏夹', key: 'Favorites', icon: StarOutline, path: '/favorites' },
  { label: '笔记本', key: 'MyNotes', icon: JournalOutline, path: '/my-notes' },
  { label: '订阅商城', key: 'PaymentTest', icon: CartOutline, path: '/payment-test' }, 
  // 🔥🔥🔥 修改处：将原来的意见反馈改为社区交流
  { label: '社区交流', key: 'ForumHome', icon: ChatboxEllipsesOutline, path: '/forum' },
  
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

// 🛡️ 每次加载布局时，静默刷新一次用户信息（防缓存过期）
onMounted(() => {
    if (userStore.token) {
        userStore.fetchProfile()
    }
})
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar desktop-sidebar">
      <div class="logo-area">
        <div class="logo-icon">
          <n-icon size="24" color="#fff"><CheckmarkCircle /></n-icon>
        </div>
        <span class="logo-text">题酷</span>
      </div>

      <nav class="nav-menu">
        <div 
          v-for="item in menuItems" 
          :key="item.key"
          class="nav-item"
          :class="{ active: route.name === item.key || route.path.startsWith(item.path) && item.path !== '/' }"
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

    <header class="mobile-header">
      <div class="logo-area-mobile">
        <div class="logo-icon-mobile">
          <n-icon size="20" color="#fff"><CheckmarkCircle /></n-icon>
        </div>
        <span class="logo-text-mobile">题酷</span>
      </div>
      <button class="menu-btn" @click="isMobileMenuOpen = true">
        <n-icon size="24"><MenuOutline /></n-icon>
      </button>
    </header>

    <n-drawer v-model:show="isMobileMenuOpen" placement="right" width="280">
      <n-drawer-content title="菜单" closable body-content-style="padding: 0;">
        <nav class="mobile-nav-menu">
           <div 
            v-for="item in menuItems" 
            :key="item.key"
            class="nav-item mobile-item"
            :class="{ active: route.name === item.key || route.path.startsWith(item.path) && item.path !== '/' }"
            @click="handleNavigate(item.path)"
          >
            <n-icon size="20" class="nav-icon"><component :is="item.icon" /></n-icon>
            <span class="nav-label">{{ item.label }}</span>
            </div>
        </nav>
        
        <template #footer>
            <div class="mobile-drawer-footer">
                <div class="user-profile-area mobile-profile-card" @click="handleNavigate('/profile')">
                    <div class="avatar-wrapper">
                    <n-avatar round size="medium" :src="userAvatar" fallback-src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg" />
                    </div>
                    <div class="user-info">
                    <div class="user-name">{{ userStore.username || '同学' }}</div>
                    <div class="user-role">点击查看个人中心</div>
                    </div>
                </div>
                 <div class="nav-item mobile-item logout-item" @click="handleLogout">
                    <n-icon size="20" class="nav-icon"><LogOutOutline /></n-icon>
                    <span class="nav-label">退出登录</span>
                </div>
            </div>
        </template>
      </n-drawer-content>
    </n-drawer>

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

  display: flex;
  flex-direction: column;
  z-index: 100;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.desktop-sidebar {
  width: 68px; /* Collapsed width */
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
}

.desktop-sidebar:hover {
  width: 210px; /* Expanded width on hover */
  box-shadow: 4px 0 24px rgba(0,0,0,0.08); /* Add shadow when expanded */
}

/* Ensure content pushes appropriately or stays put. 
   Since it's a hover expansion, content staying at collapsed margin is usually better UX 
   to avoid content jumping, allowing menu to overlay. */

.logo-area {
  padding: 20px 0; /* Adjusted padding */
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 0; /* Gap handled by logic/padding */
  border-bottom: 1px solid #f1f5f9;
  height: 73px; /* Fixed height to match design */
  box-sizing: border-box;
  white-space: nowrap;
}

.logo-icon {
  background-color: #2563eb;
  min-width: 68px; /* Center in collapsed width */
  height: 32px;
  background: none; /* Remove bg to look cleaner in collapsed or keep it? Original had blue bg box. Let's keep icon centered. */
  /* Actually original had a box. Let's adjust to be centered properly */
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Re-styling logo icon wrapper for the transition */
.logo-icon {
    min-width: 68px;
    display: flex;
    justify-content: center;
    align-items: center;
}
.logo-icon .n-icon {
  background-color: #2563eb;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.logo-text {
  font-size: 18px;
  font-weight: 500;
  color: #2563eb;
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
  pointer-events: none;
}

.desktop-sidebar:hover .logo-text {
  opacity: 1;
  transform: translateX(0);
  pointer-events: auto;
}

.nav-menu {
  flex: 1;
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0; /* Gap reset */
  padding: 8px 0; /* Vertical padding only, horizontal handled by min-width of icon */
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #64748b;
  white-space: nowrap;
}

.nav-item:hover {
  background-color: #f1f5f9;
  color: #0f172a;
}

.nav-item.active {
  background-color: #eff6ff;
  color: #2563eb;
  font-weight: 500;
}

/* Wrapper for icon to force it to be centered in the collapsed strip */
.nav-icon {
  min-width: 52px; /* 68px sidebar - padding? let's standardise */
  /* The sidebar is 68px. The menu has padding 8px. So available width is 52px. */
  min-width: 52px; 
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-label {
  font-size: 14px;
  font-weight: 400;
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
}

.desktop-sidebar:hover .nav-label {
  opacity: 1;
  transform: translateX(0);
}

.user-profile-area {
  padding: 12px 0;
  margin: 12px 8px; /* Keep margin but ensure fit */
  border-top: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  gap: 0;
  cursor: pointer;
  border-radius: 8px;
  transition: background-color 0.2s;
  overflow: hidden;
}

.user-profile-area:hover {
  background-color: #f8fafc;
}

.avatar-wrapper {
  min-width: 52px; /* Center within the available 52px width (68 - 16 margin/padding) */
  display: flex;
  justify-content: center;
}

.user-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
  white-space: nowrap;
}

.desktop-sidebar:hover .user-info {
  opacity: 1;
  transform: translateX(0);
}

.user-name {
  font-size: 13px;
  font-weight: 500;
  color: #0f172a;
  text-overflow: ellipsis;
  overflow: hidden;
}

.user-role {
  font-size: 12px;
  color: #94a3b8;
}

/* Mobile Header - Unchanged */
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
  padding: 16px;
}

.mobile-drawer-footer {
    display: flex;
    flex-direction: column;
    gap: 4px;
    width: 100%;
}

.mobile-profile-card {
    margin: 0 0 12px 0;
    border: 1px solid #f1f5f9;
    background: #f8fafc;
    /* Reset mobile profile card styles that might inherit from global classes */
    padding: 12px;
    display: flex;
    align-items: center;
    gap: 12px;
}
/* Fix mobile profile card internals having opacity 0 due to shared class names */
.mobile-profile-card .user-info {
    opacity: 1 !important;
    transform: none !important;
    gap: 4px;
}
.mobile-profile-card .avatar-wrapper {
    min-width: auto;
}

.mobile-item {
  padding: 12px;
  /* Reset gap */
  gap: 12px;
}
.mobile-item .nav-label {
    opacity: 1 !important;
    transform: none !important;
}
.mobile-item .nav-icon {
    min-width: auto;
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
  margin-left: 68px; /* Adjusted to match collapsed sidebar */
  width: calc(100% - 68px); /* Full width based on new margin */
  padding: 0; 
  height: 100vh;
  overflow-y: auto; 
  position: relative;
  box-sizing: border-box;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
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