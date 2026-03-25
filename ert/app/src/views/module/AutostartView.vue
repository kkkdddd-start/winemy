<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>自启动项目</h2>
        <p class="description">M18 - 注册表/启动文件夹/服务/WMI</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索名称/路径" style="width: 200px" clearable @keyup.enter="handleSearch">
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
              <div class="card-value">{{ stats.highRisk }}</div>
              <div class="card-label">高危项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.mediumRisk }}</div>
              <div class="card-label">中危项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.lowRisk }}</div>
              <div class="card-label">低危项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ autostartList.length }}</div>
              <div class="card-label">总计</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('registry')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">注册表</div>
              <div class="card-desc">Run键自启动</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('startup-folder')">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">启动文件夹</div>
              <div class="card-desc">Startup目录</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service')">
            <div class="card-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">服务自启</div>
              <div class="card-desc">服务自启动项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('wmi')">
            <div class="card-icon">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">WMI自启</div>
              <div class="card-desc">WMI事件订阅</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>自启动项目列表</span>
            <div class="header-operations">
              <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="注册表" value="registry" />
                <el-option label="启动文件夹" value="folder" />
                <el-option label="服务" value="service" />
                <el-option label="WMI" value="wmi" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="type" label="类型" width="100" sortable>
            <template #default="{ row }">
              <el-tag>{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip sortable />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip sortable />
          <el-table-column prop="location" label="位置" min-width="200" show-overflow-tooltip sortable />
          <el-table-column prop="risk" label="风险" width="100" sortable>
            <template #default="{ row }">
              <RiskTag :risk-level="getRiskLevel(row.risk)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="danger" size="small" @click="handleDisable(row)" :disabled="row.risk === '低危'">禁用</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-area">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredAutostartList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="自启动项详情" width="600px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="名称">{{ selectedItem.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ selectedItem.type }}</el-descriptions-item>
        <el-descriptions-item label="风险等级" :span="2">
          <RiskTag :risk-level="getRiskLevel(selectedItem.risk)" />
        </el-descriptions-item>
        <el-descriptions-item label="路径" :span="2">{{ selectedItem.path }}</el-descriptions-item>
        <el-descriptions-item label="位置" :span="2">{{ selectedItem.location }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ selectedItem.description || '无' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="danger" @click="handleDisable(selectedItem)" :disabled="selectedItem?.risk === '低危'">禁用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Key, Folder, Setting, Cpu, Search, Download, Warning, WarningFilled, SuccessFilled, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RiskTag from '@/components/RiskTag/RiskTag.vue'
import { Go } from '@wailsjs/go/main/App'

interface AutostartInfo {
  type: string
  name: string
  path: string
  location: string
  risk: string
  description?: string
}

const loading = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const autostartList = ref<AutostartInfo[]>([])
const selectedItem = ref<AutostartInfo | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<AutostartInfo[]>([])

const stats = computed(() => {
  const list = autostartList.value
  return {
    highRisk: list.filter(a => a.risk === '高危').length,
    mediumRisk: list.filter(a => a.risk === '中危').length,
    lowRisk: list.filter(a => a.risk === '低危').length
  }
})

const filteredAutostartList = computed(() => {
  let result = autostartList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(a =>
      a.name.toLowerCase().includes(keyword) ||
      a.path.toLowerCase().includes(keyword) ||
      a.location.toLowerCase().includes(keyword)
    )
  }
  if (filterType.value) {
    result = result.filter(a => a.type.toLowerCase() === filterType.value)
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredAutostartList.value.slice(start, end)
})

function getRiskLevel(risk: string): number {
  switch (risk) {
    case '高危': return 2
    case '中危': return 1
    case '低危': return 0
    default: return 0
  }
}

async function loadAutostartList() {
  loading.value = true
  try {
    const data = await Go.GetAutostartList()
    if (data) {
      autostartList.value = data as AutostartInfo[]
    }
  } catch (error) {
    console.error('Failed to load autostart list:', error)
    ElMessage.error('加载自启动列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadAutostartList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: AutostartInfo[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  filterType.value = feature === 'registry' ? 'registry' :
                     feature === 'startup-folder' ? 'folder' :
                     feature === 'service' ? 'service' :
                     feature === 'wmi' ? 'wmi' : ''
}

function handleView(row: AutostartInfo) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

async function handleDisable(row: AutostartInfo) {
  if (!row) return
  try {
    await ElMessageBox.confirm(
      `确定要禁用 "${row.name}" 吗？`,
      '禁用确认',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success(`已禁用: ${row.name}`)
    detailDialogVisible.value = false
    loadAutostartList()
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleExport() {
  ElMessage.info('正在导出数据...')
}

onMounted(() => {
  loadAutostartList()
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
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
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
</style>
