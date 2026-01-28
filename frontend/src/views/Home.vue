<script setup lang="ts">
import Dashboard from '../components/Dashboard.vue'
import { NLayout, NLayoutContent } from 'naive-ui'
</script>

<template>
  <div class="home-container">
    <!-- Removed Overview Header -->
    <Dashboard />
  </div>
</template>

.home-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Removed card-like container styles (border, radius, shadow) */

.overview-header {
  padding: 16px 24px;
/* border-bottom removed to blend in, or keep for separation? Let's remove to be cleaner */
  padding-bottom: 0; /* Just spacing */
}

.page-title {
  margin: 0;
  font-size: 24px; /* Larger title for full page feel */
  font-weight: 800;
  color: #1e293b;
}