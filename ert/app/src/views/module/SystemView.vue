<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统概览</h2>
        <p class="description">主机信息、资源监控、实时图表</p>
      </div>
      <div class="header-actions">
        <el-switch v-model="autoRefresh" active-text="自动刷新" @change="handleAutoRefreshChange" />
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :xs="12" :sm="8" :md="6">
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
        <el-col :xs="12" :sm="8" :md="6">
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
        <el-col :xs="12" :sm="8" :md="6">
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
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon network">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ systemInfo.is_domain ? '已加域' : '工作组' }}</div>
              <div class="card-label">{{ systemInfo.domain_name || '网络状态' }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="8">
          <el-card>
            <template #header>
              <span>CPU 使用率</span>
            </template>
            <div ref="cpuChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="8">
          <el-card>
            <template #header>
              <span>内存使用率</span>
            </template>
            <div ref="memoryChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="8">
          <el-card>
            <template #header>
              <span>磁盘使用率</span>
            </template>
            <div ref="diskChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-card style="margin-top: 20px;">
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
          <el-descriptions-item label="启动时间">{{ formatTime(systemInfo.boot_time) }}</el-descriptions-item>
          <el-descriptions-item label="运行时间">{{ systemInfo.uptime || '-' }}</el-descriptions-item>
          <el-descriptions-item label="域环境">
            <el-tag :type="systemInfo.is_domain ? 'success' : 'info'">
              {{ systemInfo.is_domain ? systemInfo.domain_name || '已加域' : '工作组' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <span>网络状态</span>
          </div>
        </template>
        <el-table :data="networkAdapters" stripe>
          <el-table-column prop="name" label="网卡名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="ip_addresses" label="IP 地址" min-width="150">
            <template #default="{ row }">
              <span v-for="ip in row.ip_addresses" :key="ip">{{ ip }}<br/></span>
            </template>
          </el-table-column>
          <el-table-column prop="mac_address" label="MAC 地址" width="150" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Up' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="connection_count" label="连接数" width="100" />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Refresh, Cpu, Memory, FolderOpened, Connection } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

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
  uptime?: string
}

interface NetworkAdapter {
  name: string
  ip_addresses: string[]
  mac_address: string
  status: string
  connection_count: number
}

const loading = ref(false)
const autoRefresh = ref(false)
const systemInfo = ref<SystemInfo>({})
const networkAdapters = ref<NetworkAdapter[]>([])

let cpuChart: echarts.ECharts | null = null
let memoryChart: echarts.ECharts | null = null
let diskChart: echarts.ECharts | null = null
let cpuChartRef = ref<HTMLDivElement | null>(null)
let memoryChartRef = ref<HTMLDivElement | null>(null)
let diskChartRef = ref<HTMLDivElement | null>(null)

const cpuHistory = ref<number[]>([])
const memoryHistory = ref<number[]>([])
const diskHistory = ref<number[]>([])

let refreshTimer: ReturnType<typeof setInterval> | null = null

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

function formatTime(time: string | undefined): string {
  if (!time) return '-'
  try {
    const date = new Date(time)
    return date.toLocaleString('zh-CN')
  } catch {
    return time
  }
}

function initCharts() {
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
    cpuChart.setOption({
      series: [{ type: 'gauge', radius: '90%', startAngle: 200, endAngle: -20, min: 0, max: 100, itemStyle: { color: '#409eff' }, progress: { show: true }, pointer: { show: true }, axisLine: { lineStyle: { width: 8 } }, axisTick: { show: false }, splitLine: { show: false }, axisLabel: { show: false }, title: { show: false }, detail: { fontSize: 24, formatter: '{value}%', valueAnimation: true } }]
    })
  }

  if (memoryChartRef.value) {
    memoryChart = echarts.init(memoryChartRef.value)
    memoryChart.setOption({
      series: [{ type: 'gauge', radius: '90%', startAngle: 200, endAngle: -20, min: 0, max: 100, itemStyle: { color: '#67c23a' }, progress: { show: true }, pointer: { show: true }, axisLine: { lineStyle: { width: 8 } }, axisTick: { show: false }, splitLine: { show: false }, axisLabel: { show: false }, title: { show: false }, detail: { fontSize: 24, formatter: '{value}%', valueAnimation: true } }]
    })
  }

  if (diskChartRef.value) {
    diskChart = echarts.init(diskChartRef.value)
    diskChart.setOption({
      series: [{ type: 'gauge', radius: '90%', startAngle: 200, endAngle: -20, min: 0, max: 100, itemStyle: { color: '#e6a23c' }, progress: { show: true }, pointer: { show: true }, axisLine: { lineStyle: { width: 8 } }, axisTick: { show: false }, splitLine: { show: false }, axisLabel: { show: false }, title: { show: false }, detail: { fontSize: 24, formatter: '{value}%', valueAnimation: true } }]
    })
  }
}

function updateCharts() {
  const cpuValue = Math.random() * 30 + 20
  const memValue = Math.random() * 20 + 40
  const diskValue = Math.random() * 20 + 30

  cpuHistory.value.push(cpuValue)
  memoryHistory.value.push(memValue)
  diskHistory.value.push(diskValue)

  if (cpuHistory.value.length > 20) {
    cpuHistory.value.shift()
    memoryHistory.value.shift()
    diskHistory.value.shift()
  }

  if (cpuChart) cpuChart.setOption({ series: [{ detail: { value: Math.round(cpuValue) } }] })
  if (memoryChart) memoryChart.setOption({ series: [{ detail: { value: Math.round(memValue) } }] })
  if (diskChart) diskChart.setOption({ series: [{ detail: { value: Math.round(diskValue) } }] })
}

async function loadSystemInfo() {
  loading.value = true
  try {
    systemInfo.value = {
      hostname: 'WIN-ERT-DEV',
      os_name: 'Windows 11 专业版',
      os_version: '23H2',
      architecture: 'x64',
      boot_time: new Date(Date.now() - 86400000 * 3).toISOString(),
      current_user: 'Administrator',
      cpu_count: 8,
      memory_total: 16 * 1024 * 1024 * 1024,
      disk_total: 500 * 1024 * 1024 * 1024,
      is_domain: false,
      uptime: '3天 12小时 30分钟'
    }

    networkAdapters.value = [
      { name: '以太网', ip_addresses: ['192.168.1.100', '192.168.1.101'], mac_address: '00:11:22:33:44:55', status: 'Up', connection_count: 15 },
      { name: 'VMware Network Adapter', ip_addresses: ['192.168.56.1'], mac_address: '00:50:56:C0:00:01', status: 'Up', connection_count: 0 },
      { name: 'Loopback Pseudo-Interface 1', ip_addresses: ['127.0.0.1'], mac_address: '-', status: 'Up', connection_count: 0 }
    ]
  } catch (error) {
    console.error('Failed to load system info:', error)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadSystemInfo()
  ElMessage.success('刷新成功')
}

function handleAutoRefreshChange(enabled: boolean) {
  if (enabled) {
    refreshTimer = setInterval(() => {
      loadSystemInfo()
      updateCharts()
    }, 5000)
    ElMessage.success('已开启自动刷新 (5秒)')
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
    ElMessage.info('已关闭自动刷新')
  }
}

function handleResize() {
  cpuChart?.resize()
  memoryChart?.resize()
  diskChart?.resize()
}

import { ElMessage } from 'element-plus'

onMounted(() => {
  loadSystemInfo()
  initCharts()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  window.removeEventListener('resize', handleResize)
  cpuChart?.dispose()
  memoryChart?.dispose()
  diskChart?.dispose()
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

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.info-cards {
  margin-bottom: 20px;
}

.info-card {
  background: #16213e;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.card-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.card-icon.cpu { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.memory { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.disk { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.network { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }

.card-value {
  font-size: 18px;
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

.mini-chart {
  width: 100%;
  height: 160px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
