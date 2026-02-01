<script setup lang="ts">
import { NMessageProvider, NConfigProvider, type GlobalThemeOverrides } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#2563eb',
    primaryColorHover: '#3b82f6',
    primaryColorPressed: '#1d4ed8',
    primaryColorSuppl: '#3b82f6',
    borderRadius: '12px'
  },
  Button: {
    borderRadiusMedium: '16px',
    borderRadiusSmall: '12px',
    borderRadiusLarge: '16px'
  },
  Input: {
    borderRadius: '16px'
  },
  Card: {
    borderRadius: '16px'
  }
}
</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider>
      <router-view />
    </n-message-provider>
  </n-config-provider>
</template>

<style>
/* 全局样式重置 */
body {
  margin: 0;
  font-family: "Inter", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Helvetica Neue", Helvetica, Arial, sans-serif;
  background-color: #f8fafc;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
  color: #334155;
}

/* 🔥 Naive UI 组件全局美化 (圆角风格 - 统一 16px) */
.n-input, .n-input-wrapper, .n-base-selection .n-base-selection-label {
    border-radius: 16px !important;
}

/* 让输入框在 hover/focus 时更有质感 */
.n-input:hover, .n-base-selection:hover {
    box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.1);
}

.n-button {
    border-radius: 16px !important; 
}

/* 卡片稍大一点圆角 */
.n-card {
    border-radius: 16px !important;
}
</style>