<template>
  <div class="layout-container">
    <div class="sidebar">
      <div class="logo">
        <h2>ERT</h2>
        <span class="version">v13.0</span>
      </div>
      <el-menu
        :default-active="currentRoute"
        class="sidebar-menu"
        background-color="#1a1a2e"
        text-color="#a0aec0"
        active-text-color="#409eff"
        router
      >
        <el-menu-item index="/m1">
          <el-icon><Monitor /></el-icon>
          <span>系统概览</span>
        </el-menu-item>
        <el-menu-item index="/m2">
          <el-icon><Document /></el-icon>
          <span>进程管理</span>
        </el-menu-item>
        <el-menu-item index="/m3">
          <el-icon><Connection /></el-icon>
          <span>网络分析</span>
        </el-menu-item>
        <el-menu-item index="/m4">
          <el-icon><Tickets /></el-icon>
          <span>注册表分析</span>
        </el-menu-item>
        <el-menu-item index="/m5">
          <el-icon><Setting /></el-icon>
          <span>服务管理</span>
        </el-menu-item>
        <el-menu-item index="/m6">
          <el-icon><Clock /></el-icon>
          <span>计划任务</span>
        </el-menu-item>
        <el-menu-item index="/m7">
          <el-icon><DataLine /></el-icon>
          <span>系统监控</span>
        </el-menu-item>
        <el-menu-item index="/m13">
          <el-icon><DocumentCopy /></el-icon>
          <span>日志分析</span>
        </el-menu-item>
        <el-menu-item index="/m14">
          <el-icon><User /></el-icon>
          <span>账户分析</span>
        </el-menu-item>
        <el-menu-item index="/m17">
          <el-icon><Warning /></el-icon>
          <span>应急处置</span>
        </el-menu-item>
        <el-menu-item index="/m18">
          <el-icon><Lightning /></el-icon>
          <span>自启动项目</span>
        </el-menu-item>
        <el-menu-item index="/m22">
          <el-icon><Download /></el-icon>
          <span>报告导出</span>
        </el-menu-item>
        <el-menu-item index="/m25">
          <el-icon><Grid /></el-icon>
          <span>编解码工具</span>
        </el-menu-item>
      </el-menu>
    </div>
    <div class="main-content">
      <div class="header">
        <div class="header-title">{{ currentTitle }}</div>
        <div class="header-actions">
          <el-button @click="handleRefresh" :icon="Refresh" circle />
          <el-button @click="handleSearch" :icon="Search" circle />
          <el-button @click="handleExport" :icon="Download" circle />
        </div>
      </div>
      <div class="content-area">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Refresh, Search, Download, Monitor, Document, Connection, Tickets, Setting, Clock, DataLine, DocumentCopy, User, Warning, Lightning, Grid } from '@element-plus/icons-vue'

const route = useRoute()

const currentRoute = computed(() => route.path)
const currentTitle = computed(() => (route.meta?.title as string) || 'ERT')
</script>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  background-color: #1a1a2e;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #2d2d44;
}

.logo {
  padding: 20px;
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.logo h2 {
  color: #409eff;
  font-size: 24px;
  margin: 0;
}

.version {
  color: #606266;
  font-size: 12px;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  height: 60px;
  background-color: #16213e;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid #2d2d44;
}

.header-title {
  font-size: 18px;
  font-weight: 500;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.content-area {
  flex: 1;
  overflow: auto;
  padding: 20px;
}
</style>
