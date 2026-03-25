<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>日志分析</h2>
        <p class="description">M13 - 事件日志、EVTX解析、全文搜索</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>
          导出日志
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">日志总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.errors }}</div>
              <div class="card-label">错误</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.warnings }}</div>
              <div class="card-label">警告</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.security }}</div>
              <div class="card-label">安全事件</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'event-log' }" @click="handleFeature('event-log')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">事件日志</div>
              <div class="card-desc">Windows事件日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'evtx' }" @click="handleFeature('evtx')">
            <div class="card-icon">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">EVTX解析</div>
              <div class="card-desc">解析EVTX文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'search' }" @click="handleFeature('search')">
            <div class="card-icon">
              <el-icon><Search /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">全文搜索</div>
              <div class="card-desc">日志内容搜索</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'export' }" @click="handleFeature('export')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出日志</div>
              <div class="card-desc">导出分析结果</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>{{ featureTitle }}</span>
            <div class="header-operations">
              <el-input v-model="searchKeyword" placeholder="搜索日志内容" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterLevel" placeholder="日志级别" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="Error" value="Error" />
                <el-option label="Warning" value="Warning" />
                <el-option label="Information" value="Information" />
                <el-option label="Critical" value="Critical" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'time', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="time" label="时间" width="180" sortable="custom" />
          <el-table-column prop="level" label="级别" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="getLevelRisk(row.level)" :text="row.level" />
            </template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="150" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="event_id" label="事件ID" width="100" sortable="custom" />
          <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredLogList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="日志详情" width="700px" destroy-on-close>
      <div class="detail-content" v-if="currentLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="时间">{{ currentLog.time }}</el-descriptions-item>
          <el-descriptions-item label="级别">
            <RiskTag :level="getLevelRisk(currentLog.level)" :text="currentLog.level" />
          </el-descriptions-item>
          <el-descriptions-item label="来源">{{ currentLog.source }}</el-descriptions-item>
          <el-descriptions-item label="事件ID">{{ currentLog.event_id }}</el-descriptions-item>
          <el-descriptions-item label="消息" :span="2">{{ currentLog.message }}</el-descriptions-item>
        </el-descriptions>
        <div class="raw-data" v-if="currentLog.raw_data">
          <h4>原始数据</h4>
          <pre>{{ currentLog.raw_data }}</pre>
        </div>
        <div class="timeline-section" v-if="currentLog.timeline">
          <h4>关联事件时间线</h4>
          <Timeline :items="currentLog.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleSearchRelated(currentLog)" v-if="currentLog.event_id">搜索相关事件</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Document, FolderOpened, Search, Download, CircleClose, WarningFilled, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'
import Timeline from '@/components/common/Timeline.vue'

interface LogInfo {
  time: string
  level: string
  source: string
  event_id: number
  message: string
  raw_data?: string
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterLevel = ref('')
const currentFeature = ref('event-log')
const logList = ref<LogInfo[]>([])
const detailDialogVisible = ref(false)
const currentLog = ref<LogInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: logList.value.length,
  errors: logList.value.filter(l => l.level === 'Error' || l.level === 'Critical').length,
  warnings: logList.value.filter(l => l.level === 'Warning').length,
  security: logList.value.filter(l => l.source.includes('Security')).length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'event-log': '事件日志',
    'evtx': 'EVTX解析',
    'search': '全文搜索',
    'export': '导出日志'
  }
  return titles[currentFeature.value] || '日志列表'
})

const filteredLogList = computed(() => {
  let result = logList.value
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(l =>
      l.message.toLowerCase().includes(keyword) ||
      l.source.toLowerCase().includes(keyword)
    )
  }
  
  if (filterLevel.value) {
    result = result.filter(l => l.level === filterLevel.value)
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredLogList.value.slice(start, end)
})

function getLevelRisk(level: string): string {
  const levelMap: Record<string, string> = {
    'Critical': 'critical',
    'Error': 'high',
    'Warning': 'medium',
    'Information': 'low'
  }
  return levelMap[level] || 'low'
}

function handleSearch() {
  currentPage.value = 1
}

function handleFilter() {
  currentPage.value = 1
}

function handleSort() {
  currentPage.value = 1
}

function handleSizeChange(val: number) {
  pageSize.value = val
  currentPage.value = 1
}

function handleCurrentChange() {}

async function loadLogList() {
  loading.value = true
  try {
    const data = await Go.GetEventLogs('', '', 0)
    if (data) {
      logList.value = data as LogInfo[]
    }
  } catch (error) {
    console.error('Failed to load log list:', error)
    ElMessage.error('加载日志列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadLogList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
  if (feature === 'export') {
    handleExport()
  }
}

function handleView(row: LogInfo) {
  currentLog.value = row
  detailDialogVisible.value = true
}

function handleSearchRelated(row: LogInfo | null) {
  const target = row || currentLog.value
  if (!target) return
  searchKeyword.value = target.event_id.toString()
  currentFeature.value = 'search'
  handleSearch()
  detailDialogVisible.value = false
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `log_report_${timestamp}.csv`
    let csv = '时间,级别,来源,事件ID,消息\n'
    filteredLogList.value.forEach(l => {
      csv += `"${l.time}","${l.level}","${l.source}","${l.event_id}","${l.message}"\n`
    })
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('Export failed:', error)
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  loadLogList()
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
  border: 2px solid transparent;
}

.feature-card:hover {
  background: #1a2a4a;
  transform: translateY(-2px);
}

.feature-card.active {
  border-color: #409eff;
  background: #1a2a4a;
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

.card-icon.warning {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
}

.card-icon.info {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-value {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.card-label {
  font-size: 12px;
  color: #909399;
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

.header-operations {
  display: flex;
  gap: 10px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.detail-content {
  padding: 10px 0;
}

.raw-data {
  margin-top: 20px;
  padding: 15px;
  background: #1a1a2e;
  border-radius: 8px;
}

.raw-data h4 {
  margin: 0 0 10px 0;
  color: #fff;
}

.raw-data pre {
  margin: 0;
  color: #a0a0a0;
  font-size: 12px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.timeline-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #303030;
}

.timeline-section h4 {
  margin: 0 0 15px 0;
  color: #fff;
}
</style>
