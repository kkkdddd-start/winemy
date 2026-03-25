<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统概览</h2>
        <p class="description">主机信息、资源监控、实时图表</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon cpu">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ systemInfo.cpu_count || 0 }}</div>
              <div class="card-label">CPU 核心数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon memory">
              <el-icon><Memory /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ formatBytes(systemInfo.memory_total) }}</div>
              <div class="card-label">总内存</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon disk">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ formatBytes(systemInfo.disk_total) }}</div>
              <div class="card-label">总磁盘</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon network">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ systemInfo.is_domain ? '已加域' : '工作组' }}</div>
              <div class="card-label">网络状态</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="system-details">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>系统信息</span>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="主机名">{{ systemInfo.hostname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="当前用户">{{ systemInfo.current_user || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作系统">{{ systemInfo.os_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="系统版本">{{ systemInfo.os_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="架构">{{ systemInfo.architecture || '-' }}</el-descriptions-item>
          <el-descriptions-item label="启动时间">{{ systemInfo.boot_time || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh, Cpu, Memory, FolderOpened, Connection } from '@element-plus/icons-vue'
import { Go } from '@wailsjs/go/main/App'

interface SystemInfo {
  hostname?: string
  os_name?: string
  os_version?: string
  architecture?: string
  boot_time?: string
  current_user?: string
  cpu_count?: number
  memory_total?: number
  disk_total?: number
  is_domain?: boolean
  domain_name?: string
}

const loading = ref(false)
const systemInfo = ref<SystemInfo>({})

function formatBytes(bytes: number | undefined): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  while (bytes >= 1024 && index < units.length - 1) {
    bytes /= 1024
    index++
  }
  return `${bytes.toFixed(2)} ${units[index]}`
}

async function loadSystemInfo() {
  loading.value = true
  try {
    const data = await Go.GetSystemInfo()
    if (data && data.length > 0) {
      systemInfo.value = data[0] as SystemInfo
    }
  } catch (error) {
    console.error('Failed to load system info:', error)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadSystemInfo()
}

onMounted(() => {
  loadSystemInfo()
})
</script>

<style scoped>
.module-view {
  height: 100%;
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-info h2 {
  margin: 0 0 5px 0;
  font-size: 20px;
}

.description {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.info-cards {
  margin-bottom: 20px;
}

.info-card {
  background: #16213e;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 15px;
}

.card-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.card-icon.cpu {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-icon.memory {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
}

.card-icon.disk {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
}

.card-icon.network {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
}

.card-value {
  font-size: 20px;
  font-weight: 600;
  color: #fff;
}

.card-label {
  font-size: 12px;
  color: #909399;
}

.system-details {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
