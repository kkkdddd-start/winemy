<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>IIS日志</h2>
        <p class="description">M24 - IIS/Apache/SQL Server 日志</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索日志内容" style="width: 200px" clearable @keyup.enter="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ logList.length }}</div>
              <div class="card-label">日志总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
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
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ uniqueIPs }}</div>
              <div class="card-label">独立IP</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('iis')">
            <div class="card-icon info">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">IIS日志</div>
              <div class="card-desc">IIS Web日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('apache')">
            <div class="card-icon warning">
              <el-icon><Grid /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Apache日志</div>
              <div class="card-desc">Apache访问日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('sql')">
            <div class="card-icon">
              <el-icon><Database /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">SQL Server日志</div>
              <div class="card-desc">SQL Server审计</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExport">
            <div class="card-icon success">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">日志分析</div>
              <div class="card-desc">日志统计报表</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>日志列表</span>
            <div class="header-operations">
              <el-select v-model="filterSource" placeholder="筛选来源" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="IIS" value="iis" />
                <el-option label="Apache" value="apache" />
                <el-option label="SQL Server" value="sql" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="time" label="时间" width="180" sortable />
          <el-table-column prop="source" label="来源" width="120" sortable>
            <template #default="{ row }">
              <el-tag>{{ row.source }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="client_ip" label="客户端IP" width="140" sortable show-overflow-tooltip />
          <el-table-column prop="method" label="方法" width="80" sortable>
            <template #default="{ row }">
              <el-tag :type="getMethodType(row.method)">{{ row.method }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="uri" label="URI" min-width="250" show-overflow-tooltip sortable />
          <el-table-column prop="status" label="状态码" width="80" sortable>
            <template #default="{ row }">
              <RiskTag :risk-level="getStatusRiskLevel(row.status)" :show-text="false" />
              <span style="margin-left: 4px;">{{ row.status }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-area">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredLogList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="日志详情" width="700px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="时间">{{ selectedItem.time }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ selectedItem.source }}</el-descriptions-item>
        <el-descriptions-item label="客户端IP">{{ selectedItem.client_ip }}</el-descriptions-item>
        <el-descriptions-item label="状态码">
          <RiskTag :risk-level="getStatusRiskLevel(selectedItem.status)" />
        </el-descriptions-item>
        <el-descriptions-item label="请求方法" :span="2">
          <el-tag :type="getMethodType(selectedItem.method)">{{ selectedItem.method }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="URI" :span="2">
          <code class="uri-code">{{ selectedItem.uri }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="User-Agent" :span="2" v-if="selectedItem.user_agent">
          {{ selectedItem.user_agent }}
        </el-descriptions-item>
        <el-descriptions-item label="响应时间" :span="2" v-if="selectedItem.response_time">
          {{ selectedItem.response_time }}ms
        </el-descriptions-item>
        <el-descriptions-item label="字节数" :span="2" v-if="selectedItem.bytes">
          {{ selectedItem.bytes }}
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleBlockIP">封禁IP</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Monitor, Grid, Database, DataAnalysis, Search, Download, Warning, WarningFilled, SuccessFilled, Document } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RiskTag from '@/components/RiskTag/RiskTag.vue'
import { Go } from '@wailsjs/go/main/App'

interface LogInfo {
  time: string
  source: string
  client_ip: string
  method: string
  uri: string
  status: number
  user_agent?: string
  response_time?: number
  bytes?: string
}

const loading = ref(false)
const searchKeyword = ref('')
const filterSource = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const logList = ref<LogInfo[]>([])
const selectedItem = ref<LogInfo | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<LogInfo[]>([])

const stats = computed(() => {
  const list = logList.value
  return {
    errors: list.filter(l => l.status >= 500).length,
    warnings: list.filter(l => l.status >= 400 && l.status < 500).length
  }
})

const uniqueIPs = computed(() => {
  return new Set(logList.value.map(l => l.client_ip)).size
})

const filteredLogList = computed(() => {
  let result = logList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(l =>
      l.uri.toLowerCase().includes(keyword) ||
      l.client_ip.includes(keyword) ||
      (l.user_agent && l.user_agent.toLowerCase().includes(keyword))
    )
  }
  if (filterSource.value) {
    result = result.filter(l => l.source.toLowerCase() === filterSource.value)
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredLogList.value.slice(start, end)
})

function getMethodType(method: string): string {
  switch (method.toUpperCase()) {
    case 'GET': return 'success'
    case 'POST': return 'primary'
    case 'PUT': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}

function getStatusRiskLevel(status: number): number {
  if (status >= 500) return 2
  if (status >= 400) return 1
  return 0
}

async function loadLogList() {
  loading.value = true
  try {
    const data = await Go.GetIISLogs('')
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

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadLogList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: LogInfo[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  filterSource.value = feature
}

function handleView(row: LogInfo) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

async function handleBlockIP() {
  if (!selectedItem.value) return
  try {
    await ElMessageBox.confirm(
      `确定要封禁 IP "${selectedItem.value.client_ip}" 吗？`,
      '封禁确认',
      { confirmButtonText: '确认封禁', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success('IP已封禁')
    detailDialogVisible.value = false
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleExport() {
  ElMessage.info('正在导出日志数据...')
}

onMounted(() => {
  loadLogList()
})
</script>

<style scoped>
.module-view { height: 100%; }

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-info h2 { margin: 0 0 5px 0; font-size: 20px; }
.description { margin: 0; color: #909399; font-size: 14px; }

.header-actions { display: flex; gap: 10px; align-items: center; }

.info-cards, .feature-cards { margin-bottom: 20px; }

.info-card {
  background: #16213e;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 15px;
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

.feature-card:hover { background: #1a2a4a; transform: translateY(-2px); }

.card-icon {
  width: 50px; height: 50px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 24px;
}
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }

.card-value { font-size: 24px; font-weight: 600; color: #fff; }
.card-label { font-size: 12px; color: #909399; }

.card-title { font-size: 16px; font-weight: 600; color: #fff; margin-bottom: 5px; }
.card-desc { font-size: 12px; color: #909399; }

.content-area { margin-top: 20px; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-operations { display: flex; gap: 10px; }

.pagination-area { margin-top: 16px; display: flex; justify-content: flex-end; }

.uri-code {
  display: block;
  padding: 8px;
  background: #1a1a2e;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #409eff;
  word-break: break-all;
}
</style>
