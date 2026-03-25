<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统监控</h2>
        <p class="description">M7 - CPU/内存/磁盘/网络实时监控</p>
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
              <div class="card-value">{{ cpuUsage }}%</div>
              <div class="card-label">CPU 使用率</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon memory">
              <el-icon><Memory /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ memoryUsage }}%</div>
              <div class="card-label">内存使用率</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon disk">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ diskUsage }}%</div>
              <div class="card-label">磁盘使用率</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon network">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ networkSpeed }}</div>
              <div class="card-label">网络速度</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>CPU 使用率</span>
            </template>
            <div class="chart-placeholder">
              <el-progress type="dashboard" :percentage="cpuUsage" :color="getCpuColor(cpuUsage)" />
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>内存使用率</span>
            </template>
            <div class="chart-placeholder">
              <el-progress type="dashboard" :percentage="memoryUsage" :color="getMemoryColor(memoryUsage)" />
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh, Cpu, Memory, FolderOpened, Connection } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

const loading = ref(false)
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const networkSpeed = ref('0 KB/s')

let monitorTimer: ReturnType<typeof setInterval> | null = null

function getCpuColor(usage: number): string {
  if (usage >= 90) return '#f56c6c'
  if (usage >= 70) return '#e6a23c'
  return '#67c23a'
}

function getMemoryColor(usage: number): string {
  if (usage >= 90) return '#f56c6c'
  if (usage >= 70) return '#e6a23c'
  return '#67c23a'
}

async function loadMonitorData() {
  loading.value = true
  try {
    const data = await Go.GetMonitorData()
    if (data) {
      cpuUsage.value = data.cpu_usage || 0
      memoryUsage.value = data.memory_usage || 0
      diskUsage.value = data.disk_usage || 0
      networkSpeed.value = data.network_speed || '0 KB/s'
    }
  } catch (error) {
    console.error('Failed to load monitor data:', error)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadMonitorData()
}

function startMonitor() {
  monitorTimer = setInterval(() => {
    loadMonitorData()
  }, 2000)
}

onMounted(() => {
  loadMonitorData()
  startMonitor()
})

onUnmounted(() => {
  if (monitorTimer) {
    clearInterval(monitorTimer)
  }
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

.content-area {
  margin-top: 20px;
}

.chart-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}
</style>
