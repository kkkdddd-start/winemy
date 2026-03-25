<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>进程管理</h2>
        <p class="description">M2 - 进程列表/树、查杀、Dump</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('process-list')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程列表</div>
              <div class="card-desc">查看所有运行中的进程</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('process-tree')">
            <div class="card-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程树</div>
              <div class="card-desc">父子进程关系视图</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('process-kill')">
            <div class="card-icon danger">
              <el-icon><Close /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程查杀</div>
              <div class="card-desc">终止可疑进程</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('process-dump')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程Dump</div>
              <div class="card-desc">导出进程内存数据</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>进程列表</span>
            <el-input v-model="searchKeyword" placeholder="搜索进程名称或PID" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredProcessList" v-loading="loading" stripe>
          <el-table-column prop="pid" label="PID" width="100" />
          <el-table-column prop="name" label="进程名称" min-width="150" />
          <el-table-column prop="username" label="用户名" width="120" />
          <el-table-column prop="cpu" label="CPU%" width="80" />
          <el-table-column prop="memory" label="内存%" width="80" />
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="handleKill(row)">查杀</el-button>
              <el-button size="small" @click="handleDump(row)">Dump</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Connection, Close, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ProcessInfo {
  pid: number
  name: string
  username: string
  cpu: string
  memory: string
  path: string
}

const loading = ref(false)
const searchKeyword = ref('')
const processList = ref<ProcessInfo[]>([])

const filteredProcessList = computed(() => {
  if (!searchKeyword.value) return processList.value
  const keyword = searchKeyword.value.toLowerCase()
  return processList.value.filter(p =>
    p.name.toLowerCase().includes(keyword) ||
    String(p.pid).includes(keyword)
  )
})

async function loadProcessList() {
  loading.value = true
  try {
    const data = await Go.GetProcessList()
    if (data) {
      processList.value = data as ProcessInfo[]
    }
  } catch (error) {
    console.error('Failed to load process list:', error)
    ElMessage.error('加载进程列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadProcessList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleKill(row: ProcessInfo) {
  ElMessage.info(`查杀进程: ${row.name} (PID: ${row.pid})`)
}

function handleDump(row: ProcessInfo) {
  ElMessage.info(`Dump进程: ${row.name} (PID: ${row.pid})`)
}

onMounted(() => {
  loadProcessList()
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

.feature-cards {
  margin-bottom: 20px;
}

.feature-card {
  background: #16213e;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 15px;
}

.feature-card:hover {
  background: #1a2a4a;
  transform: translateY(-2px);
}

.card-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 5px;
}

.card-desc {
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
