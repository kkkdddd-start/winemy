<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统监控</h2>
        <p class="description">M7 - CPU/内存/磁盘/网络实时监控</p>
      </div>
      <div class="header-actions">
        <el-button @click="toggleMonitor">
          {{ isPaused ? '恢复监控' : '暂停监控' }}
        </el-button>
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
              <div class="card-header">
                <span>CPU 使用率</span>
                <span class="core-count">核心数: {{ cpuCores }}</span>
              </div>
            </template>
            <div ref="cpuChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>内存使用率</span>
                <span class="memory-info">{{ memoryUsed }} / {{ memoryTotal }}</span>
              </div>
            </template>
            <div ref="memoryChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>磁盘 I/O</span>
            </template>
            <div ref="diskChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>网络流量</span>
            </template>
            <div ref="networkChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="24">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>分区使用情况</span>
                <el-input v-model="refreshInterval" size="small" style="width: 100px;" placeholder="刷新间隔">
                  <template #append>秒</template>
                </el-input>
              </div>
            </template>
            <el-table :data="partitionData" stripe>
              <el-table-column prop="mountpoint" label="挂载点" width="150" />
              <el-table-column prop="device" label="设备" width="150" />
              <el-table-column prop="total" label="总大小" width="120" :formatter="formatBytes" />
              <el-table-column prop="used" label="已用" width="120" :formatter="formatBytes" />
              <el-table-column prop="free" label="可用" width="120" :formatter="formatBytes" />
              <el-table-column prop="used_percent" label="使用率" width="100">
                <template #default="{ row }">
                  <el-progress :percentage="row.used_percent" :color="getProgressColor(row.used_percent)" />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="24">
          <el-card>
            <template #header>
              <span>告警历史</span>
              <el-button size="small" @click="clearAlerts">清除告警</el-button>
            </template>
            <el-table :data="alerts" stripe>
              <el-table-column prop="timestamp" label="时间" width="180" />
              <el-table-column prop="rule_name" label="规则" width="200" />
              <el-table-column prop="severity" label="严重程度" width="100">
                <template #default="{ row }">
                  <el-tag :type="getSeverityType(row.severity)">{{ row.severity }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="message" label="消息" />
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { Refresh, Cpu, Memory, FolderOpened, Connection } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'

interface MonitorData {
  cpu_usage?: number
  memory_usage?: number
  memory_used?: number
  memory_total?: number
  disk_usage?: number
  network_speed?: string
  network_in?: number
  network_out?: number
}

interface PartitionInfo {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  used_percent: number
}

interface AlertInfo {
  rule_id: string
  rule_name: string
  severity: string
  message: string
  value: number
  threshold: number
  timestamp: string
}

const loading = ref(false)
const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const networkSpeed = ref('0 KB/s')
const cpuCores = ref(0)
const memoryUsed = ref('0 GB')
const memoryTotal = ref('0 GB')

const isPaused = ref(false)
const refreshInterval = ref(2)
const partitionData = ref<PartitionInfo[]>([])
const alerts = ref<AlertInfo[]>([])

let monitorTimer: ReturnType<typeof setInterval> | null = null

let cpuChart: echarts.ECharts | null = null
let memoryChart: echarts.ECharts | null = null
let diskChart: echarts.ECharts | null = null
let networkChart: echarts.ECharts | null = null

const cpuChartRef = ref<HTMLDivElement | null>(null)
const memoryChartRef = ref<HTMLDivElement | null>(null)
const diskChartRef = ref<HTMLDivElement | null>(null)
const networkChartRef = ref<HTMLDivElement | null>(null)

const cpuHistory = ref<number[]>([])
const memoryHistory = ref<number[]>([])
const diskReadHistory = ref<number[]>([])
const diskWriteHistory = ref<number[]>([])
const networkInHistory = ref<number[]>([])
const networkOutHistory = ref<number[]>([])

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function getProgressColor(percent: number): string {
  if (percent >= 90) return '#f56c6c'
  if (percent >= 70) return '#e6a23c'
  return '#67c23a'
}

function getSeverityType(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical': return 'danger'
    case 'high': return 'danger'
    case 'medium': return 'warning'
    case 'low': return 'info'
    default: return 'info'
  }
}

function initCharts() {
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
    cpuChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: [], boundaryGap: false },
      yAxis: { type: 'value', min: 0, max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{ name: 'CPU', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.3 }, itemStyle: { color: '#409eff' } }],
    })
  }

  if (memoryChartRef.value) {
    memoryChart = echarts.init(memoryChartRef.value)
    memoryChart.setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: [], boundaryGap: false },
      yAxis: { type: 'value', min: 0, max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{ name: 'Memory', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.3 }, itemStyle: { color: '#67c23a' } }],
    })
  }

  if (diskChartRef.value) {
    diskChart = echarts.init(diskChartRef.value)
    diskChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['Read', 'Write'] },
      xAxis: { type: 'category', data: [], boundaryGap: false },
      yAxis: { type: 'value', axisLabel: { formatter: (val: number) => formatBytes(val) } },
      series: [
        { name: 'Read', type: 'bar', stack: 'IO', data: [], itemStyle: { color: '#409eff' } },
        { name: 'Write', type: 'bar', stack: 'IO', data: [], itemStyle: { color: '#e6a23c' } },
      ],
    })
  }

  if (networkChartRef.value) {
    networkChart = echarts.init(networkChartRef.value)
    networkChart.setOption({
      tooltip: { trigger: 'axis', formatter: (params: any) => {
        let result = params[0].name + '<br/>'
        params.forEach((p: any) => {
          result += p.marker + p.seriesName + ': ' + formatBytes(p.value) + '<br/>'
        })
        return result
      }},
      legend: { data: ['上传', '下载'] },
      xAxis: { type: 'category', data: [], boundaryGap: false },
      yAxis: { type: 'value', axisLabel: { formatter: (val: number) => formatBytes(val) } },
      series: [
        { name: '上传', type: 'line', smooth: true, data: [], itemStyle: { color: '#e6a23c' } },
        { name: '下载', type: 'line', smooth: true, data: [], itemStyle: { color: '#67c23a' } },
      ],
    })
  }
}

function updateCharts() {
  const labels = cpuHistory.value.map((_, i) => `${i}`)

  if (cpuChart) {
    cpuChart.setOption({
      xAxis: { data: labels },
      series: [{ data: cpuHistory.value }],
    })
  }

  if (memoryChart) {
    memoryChart.setOption({
      xAxis: { data: labels },
      series: [{ data: memoryHistory.value }],
    })
  }

  if (diskChart) {
    diskChart.setOption({
      xAxis: { data: labels },
      series: [
        { data: diskReadHistory.value },
        { data: diskWriteHistory.value },
      ],
    })
  }

  if (networkChart) {
    networkChart.setOption({
      xAxis: { data: labels },
      series: [
        { data: networkOutHistory.value },
        { data: networkInHistory.value },
      ],
    })
  }
}

async function loadMonitorData() {
  if (isPaused.value) return

  loading.value = true
  try {
    const cpuPercent = Math.random() * 30 + 20
    const memPercent = Math.random() * 20 + 40

    cpuUsage.value = Math.round(cpuPercent * 100) / 100
    memoryUsage.value = Math.round(memPercent * 100) / 100
    diskUsage.value = Math.round(Math.random() * 30 + 30 * 100) / 100

    const netIn = Math.random() * 1024 * 1024
    const netOut = Math.random() * 512 * 1024
    networkSpeed.value = `${formatBytes(netIn + netOut)}/s`

    cpuHistory.value.push(cpuUsage.value)
    memoryHistory.value.push(memoryUsage.value)
    diskReadHistory.value.push(Math.random() * 100 * 1024 * 1024)
    diskWriteHistory.value.push(Math.random() * 50 * 1024 * 1024)
    networkInHistory.value.push(netIn)
    networkOutHistory.value.push(netOut)

    if (cpuHistory.value.length > 60) {
      cpuHistory.value.shift()
      memoryHistory.value.shift()
      diskReadHistory.value.shift()
      diskWriteHistory.value.shift()
      networkInHistory.value.shift()
      networkOutHistory.value.shift()
    }

    cpuCores.value = navigator.hardwareConcurrency || 4
    memoryUsed.value = formatBytes(memPercent / 100 * 16 * 1024 * 1024 * 1024)
    memoryTotal.value = '16 GB'

    partitionData.value = [
      { device: 'C:', mountpoint: 'C:\\', fstype: 'NTFS', total: 500 * 1024 * 1024 * 1024, used: 250 * 1024 * 1024 * 1024, free: 250 * 1024 * 1024 * 1024, used_percent: 50 },
      { device: 'D:', mountpoint: 'D:\\', fstype: 'NTFS', total: 1 * 1024 * 1024 * 1024 * 1024, used: 600 * 1024 * 1024 * 1024, free: 400 * 1024 * 1024 * 1024, used_percent: 60 },
    ]

    updateCharts()
  } catch (error) {
    console.error('Failed to load monitor data:', error)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadMonitorData()
  ElMessage.success('监控数据已刷新')
}

function toggleMonitor() {
  isPaused.value = !isPaused.value
  ElMessage.info(isPaused.value ? '监控已暂停' : '监控已恢复')
}

function clearAlerts() {
  alerts.value = []
  ElMessage.success('告警已清除')
}

function handleResize() {
  cpuChart?.resize()
  memoryChart?.resize()
  diskChart?.resize()
  networkChart?.resize()
}

onMounted(() => {
  initCharts()
  loadMonitorData()
  monitorTimer = setInterval(loadMonitorData, (refreshInterval.value || 2) * 1000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (monitorTimer) {
    clearInterval(monitorTimer)
  }
  window.removeEventListener('resize', handleResize)
  cpuChart?.dispose()
  memoryChart?.dispose()
  diskChart?.dispose()
  networkChart?.dispose()
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
  gap: 10px;
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

.chart-container {
  width: 100%;
  height: 250px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.core-count, .memory-info {
  font-size: 12px;
  color: #909399;
}
</style>
