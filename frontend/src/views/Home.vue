<script setup lang="ts">
import Dashboard from '../components/Dashboard.vue'
import { NLayout, NLayoutContent } from 'naive-ui'
</script>

<template>
  <div class="home-container">
    <div class="overview-header">
      <h2 class="page-title">总览面板</h2>
    </div>
    <n-layout style="height: 100%; flex: 1; background-color: #f8fafc;">
      <n-layout-content content-style="padding: 24px;" :native-scrollbar="true">
        <Dashboard />
      </n-layout-content>
    </n-layout>
  </div>
</template>

<style scoped>
.home-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fff;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  border: 1px solid #e2e8f0;
}

.overview-header {
  padding: 16px 24px;
  border-bottom: 1px solid #f0f0f0;
  background-color: #fff;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #1e293b;
}
</style>