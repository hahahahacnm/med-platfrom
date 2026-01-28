<script setup lang="ts">
import { NMessageProvider } from 'naive-ui'
</script>

<template>
  <n-message-provider>
    <router-view />
  </n-message-provider>
</template>

<style>
/* 全局样式重置 */
body {
  margin: 0;
  font-family: v-sans, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
</style>