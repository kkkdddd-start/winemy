<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>报告导出</h2>
        <p class="description">M22 - HTML/PDF/JSON 导出</p>
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
          <div class="feature-card" @click="handleExport('html')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出HTML</div>
              <div class="card-desc">HTML报告格式</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExport('pdf')">
            <div class="card-icon">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出PDF</div>
              <div class="card-desc">PDF报告格式</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExport('json')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出JSON</div>
              <div class="card-desc">JSON数据格式</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExport('csv')">
            <div class="card-icon">
              <el-icon><Grid /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出CSV</div>
              <div class="card-desc">CSV表格格式</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>导出历史</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="HTML" value="html" />
              <el-option label="PDF" value="pdf" />
              <el-option label="JSON" value="json" />
              <el-option label="CSV" value="csv" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredHistoryList" v-loading="loading" stripe>
          <el-table-column prop="time" label="时间" width="180" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ row.type.toUpperCase() }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="报告名称" min-width="200" />
          <el-table-column prop="size" label="大小" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === '完成' ? 'success' : 'warning'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleDownload(row)">下载</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Document, FolderOpened, List, Grid } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ReportHistory {
  time: string
  type: string
  name: string
  size: string
  status: string
}

const loading = ref(false)
const filterType = ref('')
const historyList = ref<ReportHistory[]>([])

const filteredHistoryList = computed(() => {
  if (!filterType.value) return historyList.value
  return historyList.value.filter(h => h.type === filterType.value)
})

async function loadHistoryList() {
  loading.value = true
  try {
    const data = await Go.GetReportHistory()
    if (data) {
      historyList.value = data as ReportHistory[]
    }
  } catch (error) {
    console.error('Failed to load report history:', error)
    ElMessage.error('加载导出历史失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadHistoryList()
}

function handleExport(type: string) {
  ElMessage.success(`正在导出 ${type.toUpperCase()} 报告...`)
}

function handleDownload(row: ReportHistory) {
  ElMessage.success(`下载报告: ${row.name}`)
}

function handleDelete(row: ReportHistory) {
  ElMessage.warning(`删除报告: ${row.name}`)
}

onMounted(() => {
  loadHistoryList()
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
