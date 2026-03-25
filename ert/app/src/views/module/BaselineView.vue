<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>安全基线检查</h2>
        <p class="description">M23 - 密码/账户/审核/网络安全</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索检查项" style="width: 200px" clearable @keyup.enter="handleSearch">
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
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.failed }}</div>
              <div class="card-label">不合规项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.passed }}</div>
              <div class="card-label">合规项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.warning }}</div>
              <div class="card-label">警告项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineScore }}%</div>
              <div class="card-label">合规评分</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('password')">
            <div class="card-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">密码策略</div>
              <div class="card-desc">密码复杂度检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('account')">
            <div class="card-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">账户策略</div>
              <div class="card-desc">账户安全检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('audit')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">审核策略</div>
              <div class="card-desc">安全审核检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('network')">
            <div class="card-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">网络安全</div>
              <div class="card-desc">网络配置检查</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>基线检查结果</span>
            <div class="header-operations">
              <el-select v-model="filterStatus" placeholder="筛选状态" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="不合规" value="failed" />
                <el-option label="合规" value="passed" />
                <el-option label="警告" value="warning" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="category" label="类别" width="120" sortable show-overflow-tooltip />
          <el-table-column prop="item" label="检查项" min-width="200" show-overflow-tooltip sortable />
          <el-table-column prop="status" label="状态" width="100" sortable>
            <template #default="{ row }">
              <RiskTag :risk-level="getRiskLevel(row.status)" />
            </template>
          </el-table-column>
          <el-table-column prop="current" label="当前值" width="150" sortable show-overflow-tooltip />
          <el-table-column prop="expected" label="期望值" width="150" sortable show-overflow-tooltip />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
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
            :total="filteredBaselineList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="基线检查详情" width="650px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="检查项">{{ selectedItem.item }}</el-descriptions-item>
        <el-descriptions-item label="类别">{{ selectedItem.category }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <RiskTag :risk-level="getRiskLevel(selectedItem.status)" />
        </el-descriptions-item>
        <el-descriptions-item label="合规标准" :span="2">{{ selectedItem.expected }}</el-descriptions-item>
        <el-descriptions-item label="当前配置" :span="2">
          <code class="config-code">{{ selectedItem.current }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="详细说明" :span="2">{{ selectedItem.description || '无' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="warning" @click="handleFix(selectedItem!)" :disabled="selectedItem?.status === '合规'">修复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Warning, SuccessFilled, WarningFilled, InfoFilled, Lock, User, Document, Connection, Search, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RiskTag from '@/components/RiskTag/RiskTag.vue'
import { Go } from '@wailsjs/go/main/App'

interface BaselineItem {
  category: string
  item: string
  status: string
  current: string
  expected: string
  description: string
}

const loading = ref(false)
const searchKeyword = ref('')
const filterStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const baselineList = ref<BaselineItem[]>([])
const selectedItem = ref<BaselineItem | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<BaselineItem[]>([])
const baselineStats = ref({
  failed: 0,
  passed: 0,
  warning: 0
})
const baselineScore = ref(0)

const filteredBaselineList = computed(() => {
  let result = baselineList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(b =>
      b.item.toLowerCase().includes(keyword) ||
      b.category.toLowerCase().includes(keyword) ||
      b.description.toLowerCase().includes(keyword)
    )
  }
  if (filterStatus.value) {
    result = result.filter(b => b.status === (filterStatus.value === 'passed' ? '合规' : filterStatus.value === 'failed' ? '不合规' : '警告'))
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredBaselineList.value.slice(start, end)
})

function getRiskLevel(status: string): number {
  switch (status) {
    case '不合规': return 2
    case '警告': return 1
    case '合规': return 0
    default: return 0
  }
}

function updateStats() {
  baselineStats.value = {
    failed: baselineList.value.filter(b => b.status === '不合规').length,
    passed: baselineList.value.filter(b => b.status === '合规').length,
    warning: baselineList.value.filter(b => b.status === '警告').length
  }
  const total = baselineList.value.length
  baselineScore.value = total > 0 ? Math.round((baselineStats.value.passed / total) * 100) : 0
}

async function loadBaselineList() {
  loading.value = true
  try {
    const data = await Go.GetBaselineResults()
    if (data) {
      baselineList.value = data as BaselineItem[]
      updateStats()
    }
  } catch (error) {
    console.error('Failed to load baseline list:', error)
    ElMessage.error('加载基线检查失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadBaselineList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: BaselineItem[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  filterStatus.value = ''
}

function handleView(row: BaselineItem) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

async function handleFix(row: BaselineItem) {
  if (!row) return
  try {
    await ElMessageBox.confirm(
      `确定要自动修复 "${row.item}" 吗？`,
      '修复确认',
      { confirmButtonText: '确认修复', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success('正在修复...')
    detailDialogVisible.value = false
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleExport() {
  ElMessage.info('正在导出基线检查报告...')
}

onMounted(() => {
  loadBaselineList()
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
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }

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

.config-code {
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
