<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>WMIC检测</h2>
        <p class="description">M21 - WMIC 命令历史检测</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索命令内容" style="width: 200px" clearable @keyup.enter="handleSearch">
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
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ historyList.length }}</div>
              <div class="card-label">命令总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.success }}</div>
              <div class="card-label">成功</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.suspicious }}</div>
              <div class="card-label">可疑命令</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.failed }}</div>
              <div class="card-label">失败</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('command-history')">
            <div class="card-icon info">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">命令历史</div>
              <div class="card-desc">WMIC历史命令</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('persistence-check')">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">持久化检测</div>
              <div class="card-desc">WMIC持久化检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleExport">
            <div class="card-icon success">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出报告</div>
              <div class="card-desc">导出WMIC分析</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>WMIC 历史记录</span>
            <div class="header-operations">
              <el-select v-model="filterResult" placeholder="筛选结果" style="width: 120px" clearable>
                <el-option label="全部" value="" />
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
                <el-option label="可疑" value="suspicious" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="time" label="时间" width="180" sortable />
          <el-table-column prop="user" label="用户" width="120" sortable show-overflow-tooltip />
          <el-table-column prop="command" label="命令" min-width="350" show-overflow-tooltip sortable />
          <el-table-column prop="result" label="结果" width="100" sortable>
            <template #default="{ row }">
              <RiskTag :risk-level="getRiskLevel(row.result)" :show-text="false" />
              <span style="margin-left: 6px;">{{ row.result }}</span>
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
            :total="filteredHistoryList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="WMIC 命令详情" width="700px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="时间">{{ selectedItem.time }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ selectedItem.user }}</el-descriptions-item>
        <el-descriptions-item label="执行结果" :span="2">
          <RiskTag :risk-level="getRiskLevel(selectedItem.result)" />
        </el-descriptions-item>
        <el-descriptions-item label="命令" :span="2">
          <code class="command-code">{{ selectedItem.command }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="完整输出" :span="2" v-if="selectedItem.output">
          <el-input type="textarea" :rows="4" :value="selectedItem.output" readonly />
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleAnalyze">分析</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Clock, Warning, Download, Search, SuccessFilled, WarningFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import RiskTag from '@/components/RiskTag/RiskTag.vue'
import { Go } from '@wailsjs/go/main/App'

interface WMICHistory {
  time: string
  user: string
  command: string
  result: string
  output?: string
}

const loading = ref(false)
const searchKeyword = ref('')
const filterResult = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const historyList = ref<WMICHistory[]>([])
const selectedItem = ref<WMICHistory | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<WMICHistory[]>([])

const stats = computed(() => {
  const list = historyList.value
  return {
    success: list.filter(h => h.result === '成功').length,
    failed: list.filter(h => h.result === '失败').length,
    suspicious: list.filter(h => h.result === '可疑').length
  }
})

const filteredHistoryList = computed(() => {
  let result = historyList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(h =>
      h.command.toLowerCase().includes(keyword) ||
      h.user.toLowerCase().includes(keyword)
    )
  }
  if (filterResult.value) {
    result = result.filter(h => {
      if (filterResult.value === 'success') return h.result === '成功'
      if (filterResult.value === 'failed') return h.result === '失败'
      if (filterResult.value === 'suspicious') return h.result === '可疑'
      return true
    })
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredHistoryList.value.slice(start, end)
})

function getRiskLevel(result: string): number {
  switch (result) {
    case '成功': return 0
    case '可疑': return 1
    case '失败': return 2
    default: return 0
  }
}

async function loadHistoryList() {
  loading.value = true
  try {
    const data = await Go.GetWMICHistory()
    if (data) {
      historyList.value = data as WMICHistory[]
    }
  } catch (error) {
    console.error('Failed to load WMIC history:', error)
    ElMessage.error('加载WMIC历史失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadHistoryList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: WMICHistory[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  if (feature === 'persistence-check') {
    filterResult.value = 'suspicious'
  } else {
    filterResult.value = ''
  }
}

function handleView(row: WMICHistory) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

function handleAnalyze() {
  ElMessage.info('正在分析命令...')
  detailDialogVisible.value = false
}

function handleExport() {
  ElMessage.info('正在导出数据...')
}

onMounted(() => {
  loadHistoryList()
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
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }

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

.command-code {
  display: block;
  padding: 8px;
  background: #1a1a2e;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #e6a23c;
  word-break: break-all;
}
</style>
