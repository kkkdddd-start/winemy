<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>威胁检测</h2>
        <p class="description">M16 - 恶意进程、可疑网络、敏感行为</p>
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
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ threatStats.high }}</div>
              <div class="card-label">高危威胁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ threatStats.medium }}</div>
              <div class="card-label">中危威胁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ threatStats.low }}</div>
              <div class="card-label">低危威胁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ threatStats.solved }}</div>
              <div class="card-label">已处置</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>威胁列表</span>
            <el-select v-model="filterLevel" placeholder="筛选等级" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="高危" value="high" />
              <el-option label="中危" value="medium" />
              <el-option label="低危" value="low" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredThreatList" v-loading="loading" stripe>
          <el-table-column prop="level" label="等级" width="80">
            <template #default="{ row }">
              <el-tag :type="getLevelType(row.level)">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="name" label="威胁名称" min-width="200" />
          <el-table-column prop="location" label="位置" min-width="200" show-overflow-tooltip />
          <el-table-column prop="time" label="发现时间" width="160" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === '待处置' ? 'warning' : 'success'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="handleKill(row)">查杀</el-button>
              <el-button type="warning" size="small" @click="handleIsolate(row)">隔离</el-button>
              <el-button type="info" size="small" @click="handleIgnore(row)">忽略</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Warning, WarningFilled, InfoFilled, SuccessFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ThreatInfo {
  level: string
  type: string
  name: string
  location: string
  time: string
  status: string
}

const loading = ref(false)
const filterLevel = ref('')
const threatList = ref<ThreatInfo[]>([])
const threatStats = ref({
  high: 0,
  medium: 0,
  low: 0,
  solved: 0
})

const filteredThreatList = computed(() => {
  if (!filterLevel.value) return threatList.value
  return threatList.value.filter(t => t.level.toLowerCase() === filterLevel.value)
})

function getLevelType(level: string): string {
  const typeMap: Record<string, string> = {
    '高危': 'danger',
    '中危': 'warning',
    '低危': 'info'
  }
  return typeMap[level] || 'info'
}

async function loadThreatList() {
  loading.value = true
  try {
    const data = await Go.GetThreatList()
    if (data) {
      threatList.value = data as ThreatInfo[]
      threatStats.value = {
        high: threatList.value.filter(t => t.level === '高危').length,
        medium: threatList.value.filter(t => t.level === '中危').length,
        low: threatList.value.filter(t => t.level === '低危').length,
        solved: threatList.value.filter(t => t.status === '已处置').length
      }
    }
  } catch (error) {
    console.error('Failed to load threat list:', error)
    ElMessage.error('加载威胁列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadThreatList()
}

function handleKill(row: ThreatInfo) {
  ElMessage.success(`查杀威胁: ${row.name}`)
}

function handleIsolate(row: ThreatInfo) {
  ElMessage.success(`隔离文件: ${row.name}`)
}

function handleIgnore(row: ThreatInfo) {
  ElMessage.info(`忽略威胁: ${row.name}`)
}

onMounted(() => {
  loadThreatList()
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

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
}

.card-icon.warning {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
}

.card-icon.info {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-icon.success {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
