<template>
  <div id="app" class="ert-app" :class="{ 'dark-theme': isDarkTheme, 'light-theme': !isDarkTheme }">
    <router-view />
    <el-backtop :right="20" :bottom="20" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useShortcuts } from './shortcuts/useShortcuts'

const isDarkTheme = ref(true)

const theme = computed(() => isDarkTheme.value ? 'dark' : 'light')

function toggleTheme() {
  isDarkTheme.value = !isDarkTheme.value
  localStorage.setItem('ert-theme', isDarkTheme.value ? 'dark' : 'light')
  document.documentElement.setAttribute('data-theme', isDarkTheme.value ? 'dark' : 'light')
}

function setTheme(newTheme: 'dark' | 'light') {
  isDarkTheme.value = newTheme === 'dark'
  localStorage.setItem('ert-theme', newTheme)
  document.documentElement.setAttribute('data-theme', newTheme)
}

defineExpose({
  toggleTheme,
  setTheme,
  isDarkTheme
})

onMounted(() => {
  const savedTheme = localStorage.getItem('ert-theme')
  if (savedTheme) {
    isDarkTheme.value = savedTheme === 'dark'
    document.documentElement.setAttribute('data-theme', savedTheme)
  }
  
  useShortcuts()
})
</script>

<style>
:root {
  --ert-bg-primary: #0a0e27;
  --ert-bg-secondary: #16213e;
  --ert-bg-tertiary: #1a2a4a;
  --ert-text-primary: #ffffff;
  --ert-text-secondary: #909399;
  --ert-border-color: #2d3a5a;
  --ert-accent: #409eff;
}

[data-theme="light"] {
  --ert-bg-primary: #f5f7fa;
  --ert-bg-secondary: #ffffff;
  --ert-bg-tertiary: #ebeef5;
  --ert-text-primary: #303133;
  --ert-text-secondary: #909399;
  --ert-border-color: #dcdfe6;
  --ert-accent: #409eff;
}

.ert-app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: var(--ert-bg-primary);
  color: var(--ert-text-primary);
  transition: background-color 0.3s, color 0.3s;
}

.dark-theme {
  --el-bg-color: #16213e;
  --el-bg-color-page: #0a0e27;
  --el-text-color-primary: #ffffff;
  --el-text-color-regular: #b4b4b4;
  --el-text-color-secondary: #909399;
  --el-border-color: #2d3a5a;
  --el-fill-color-blank: #16213e;
}

.light-theme {
  --el-bg-color: #ffffff;
  --el-bg-color-page: #f5f7fa;
  --el-text-color-primary: #303133;
  --el-text-color-regular: #606266;
  --el-text-color-secondary: #909399;
  --el-border-color: #dcdfe6;
  --el-fill-color-blank: #ffffff;
}
</style>
