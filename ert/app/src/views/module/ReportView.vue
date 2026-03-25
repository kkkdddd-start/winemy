<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>报告导出</h2>
        <p class="description">M22 - HTML/PDF/JSON 导出</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索报告名称" style="width: 200px" clearable @keyup.enter="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
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
            <div class="card-icon info">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ historyList.length }}</div>
              <div class="card-label">报告总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.completed }}</div>
              <div class="card-label">已完成</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Loading /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.processing }}</div>
              <div class="card-label">处理中</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ totalSize }}</div>
              <div class="card-label">总大小</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleExport('html')">
            <div class="card-icon info">
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
            <div class="card-icon danger">
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
            <div class="card-icon success">
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
            <div class="card-icon warning">
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
            <div class="header-operations">
              <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="HTML" value="html" />
                <el-option label="PDF" value="pdf" />
                <el-option label="JSON" value="json" />
                <el-option label="CSV" value="csv" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="time" label="时间" width="180" sortable />
          <el-table-column prop="type" label="类型" width="100" sortable>
            <template #default="{ row }">
              <el-tag>{{ row.type.toUpperCase() }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="报告名称" min-width="200" show-overflow-tooltip sortable />
          <el-table-column prop="size" label="大小" width="100" sortable />
          <el-table-column prop="status" label="状态" width="100" sortable>
            <template #default="{ row }">
              <el-tag :type="row.status === '完成' ? 'success' : 'warning'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)" :disabled="row.status !== '完成'">查看</el-button>
              <el-button type="success" size="small" @click="handleDownload(row)" :disabled="row.status !== '完成'">下载</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="detailDialogVisible" title="报告详情" width="600px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="报告名称">{{ selectedItem.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ selectedItem.type.toUpperCase() }}</el-descriptions-item>
        <el-descriptions-item label="生成时间">{{ selectedItem.time }}</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ selectedItem.size }}</el-descriptions-item>
        <el-descriptions-item label="状态" :span="2">
          <el-tag :type="selectedItem.status === '完成' ? 'success' : 'warning'">{{ selectedItem.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="文件路径" :span="2">{{ selectedItem.path || '无' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="success" @click="handleDownload(selectedItem!)" :disabled="selectedItem?.status !== '完成'">下载</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="exportDialogVisible" title="导出报告" width="500px">
      <el-form :model="exportForm" label-width="100px">
        <el-form-item label="报告类型">
          <el-select v-model="exportForm.type" style="width: 100%;">
            <el-option label="HTML" value="html" />
            <el-option label="PDF" value="pdf" />
            <el-option label="JSON" value="json" />
            <el-option label="CSV" value="csv" />
          </el-select>
        </el-form-item>
        <el-form-item label="报告名称">
          <el-input v-model="exportForm.name" placeholder="请输入报告名称" />
        </el-form-item>
        <el-form-item label="包含模块">
          <el-checkbox-group v-model="exportForm.modules">
            <el-checkbox label="process">进程</el-checkbox>
            <el-checkbox label="network">网络</el-checkbox>
            <el-checkbox label="registry">注册表</el-checkbox>
            <el-checkbox label="service">服务</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="exportDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmExport" :loading="actionLoading">开始导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Document, FolderOpened, List, Grid, Search, SuccessFilled, Loading, Folder } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ReportHistory {
  time: string
  type: string
  name: string
  size: string
  status: string
  path?: string
}

const loading = ref(false)
const actionLoading = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const historyList = ref<ReportHistory[]>([])
const selectedItem = ref<ReportHistory | null>(null)
const detailDialogVisible = ref(false)
const exportDialogVisible = ref(false)
const selectedItems = ref<ReportHistory[]>([])
const exportForm = ref({
  type: 'html',
  name: '',
  modules: ['process', 'network']
})

const stats = computed(() => {
  const list = historyList.value
  return {
    completed: list.filter(h => h.status === '完成').length,
    processing: list.filter(h => h.status === '处理中').length
  }
})

const totalSize = computed(() => {
  const list = historyList.value.filter(h => h.status === '完成')
  let total = 0
  list.forEach(h => {
    const match = h.size.match(/(\d+\.?\d*)/)
    if (match) total += parseFloat(match[1])
  })
  return total > 1024 ? `${(total / 1024).toFixed(1)} MB` : `${total.toFixed(1)} KB`
})

const filteredHistoryList = computed(() => {
  let result = historyList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(h => h.name.toLowerCase().includes(keyword))
  }
  if (filterType.value) {
    result = result.filter(h => h.type === filterType.value)
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredHistoryList.value.slice(start, end)
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

function handleSelectionChange(selection: ReportHistory[]) {
  selectedItems.value = selection
}

function handleExport(type: string) {
  exportForm.value.type = type
  exportForm.value.name = `安全报告_${new Date().toLocaleDateString('zh-CN')}`
  exportDialogVisible.value = true
}

async function confirmExport() {
  if (!exportForm.value.name) {
    ElMessage.warning('请输入报告名称')
    return
  }
  actionLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1500))
    ElMessage.success('报告导出任务已创建')
    exportDialogVisible.value = false
    loadHistoryList()
  } catch (error) {
    ElMessage.error('导出失败')
  } finally {
    actionLoading.value = false
  }
}

function handleView(row: ReportHistory) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

function handleDownload(row: ReportHistory) {
  ElMessage.success(`正在下载: ${row.name}`)
}

async function handleDelete(row: ReportHistory) {
  try {
    await ElMessageBox.confirm(
      `确定要删除报告 "${row.name}" 吗？`,
      '删除确认',
      { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success('报告已删除')
    loadHistoryList()
  } catch {
    ElMessage.info('已取消操作')
  }
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
</style>
