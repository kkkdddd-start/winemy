<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>内存取证</h2>
        <p class="description">M15 - 进程/系统内存 Dump</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>
          导出报告
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">进程总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.suspicious }}</div>
              <div class="card-label">可疑进程</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.hidden }}</div>
              <div class="card-label">隐藏进程</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.dumped }}</div>
              <div class="card-label">已Dump</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'process-dump' }" @click="handleFeature('process-dump')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程Dump</div>
              <div class="card-desc">导出进程内存</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'system-dump' }" @click="handleFeature('system-dump')">
            <div class="card-icon">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">系统Dump</div>
              <div class="card-desc">完整内存镜像</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'analysis' }" @click="handleFeature('analysis')">
            <div class="card-icon">
              <el-icon><Search /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">内存分析</div>
              <div class="card-desc">字符串提取分析</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'yara' }" @click="handleFeature('yara')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Yara扫描</div>
              <div class="card-desc">规则匹配检测</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索进程名称" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterRisk" placeholder="风险等级" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="正常" value="normal" />
                <el-option label="可疑" value="suspicious" />
                <el-option label="危险" value="danger" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'memory', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="pid" label="PID" width="100" sortable="custom" />
          <el-table-column prop="name" label="进程名称" min-width="150" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="username" label="用户名" width="120" sortable="custom" />
          <el-table-column prop="memory" label="内存大小" width="120" sortable="custom" />
          <el-table-column prop="risk" label="风险" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.risk" />
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleDump(row)">Dump</el-button>
              <el-button type="warning" size="small" @click="handleScan(row)">扫描</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredProcessList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="进程详情" width="700px" destroy-on-close>
      <div class="detail-content" v-if="currentProcess">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="PID">{{ currentProcess.pid }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <RiskTag :level="currentProcess.risk" />
          </el-descriptions-item>
          <el-descriptions-item label="进程名称" :span="2">{{ currentProcess.name }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ currentProcess.username }}</el-descriptions-item>
          <el-descriptions-item label="内存大小">{{ currentProcess.memory }}</el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentProcess.path }}</el-descriptions-item>
          <el-descriptions-item label="命令行" :span="2">{{ currentProcess.command_line || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div class="module-list" v-if="currentProcess.modules">
          <h4>加载的模块</h4>
          <el-tag v-for="mod in currentProcess.modules" :key="mod" type="info" style="margin-right: 5px; margin-bottom: 5px;">{{ mod }}</el-tag>
        </div>
        <div class="timeline-section" v-if="currentProcess.timeline">
          <h4>进程活动时间线</h4>
          <Timeline :items="currentProcess.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleDump(currentProcess)">Dump进程</el-button>
        <el-button type="warning" @click="handleScan(currentProcess)">Yara扫描</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Cpu, Search, Document, Download, Warning, CircleClose } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'
import Timeline from '@/components/common/Timeline.vue'

interface ProcessInfo {
  pid: number
  name: string
  username: string
  memory: string
  path: string
  risk: string
  command_line?: string
  modules?: string[]
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterRisk = ref('')
const currentFeature = ref('process-dump')
const processList = ref<ProcessInfo[]>([])
const detailDialogVisible = ref(false)
const currentProcess = ref<ProcessInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: processList.value.length,
  suspicious: processList.value.filter(p => p.risk === 'suspicious').length,
  hidden: processList.value.filter(p => p.risk === 'danger').length,
  dumped: 0
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'process-dump': '进程列表',
    'system-dump': '系统Dump',
    'analysis': '内存分析',
    'yara': 'Yara扫描'
  }
  return titles[currentFeature.value] || '进程列表'
})

const filteredProcessList = computed(() => {
  let result = processList.value
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(p => p.name.toLowerCase().includes(keyword))
  }
  
  if (filterRisk.value) {
    result = result.filter(p => p.risk === filterRisk.value)
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredProcessList.value.slice(start, end)
})

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
  currentFeature.value = feature
  currentPage.value = 1
}

function handleDump(row: ProcessInfo | null) {
  const target = row || currentProcess.value
  if (!target) return
  ElMessageBox.confirm(`确定要Dump进程 ${target.name} (PID: ${target.pid}) 吗？`, 'Dump进程', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`正在Dump进程: ${target.name}`)
    detailDialogVisible.value = false
  }).catch(() => {})
}

function handleScan(row: ProcessInfo | null) {
  const target = row || currentProcess.value
  if (!target) return
  ElMessage.info(`正在扫描进程: ${target.name}`)
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `memory_report_${timestamp}.csv`
    let csv = 'PID,进程名称,用户名,内存大小,风险等级,路径\n'
    filteredProcessList.value.forEach(p => {
      csv += `"${p.pid}","${p.name}","${p.username}","${p.memory}","${p.risk}","${p.path}"\n`
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

.card-icon.success {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
}

.card-icon.warning {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
}

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
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

.module-list {
  margin-top: 20px;
  padding: 15px;
  background: #1a1a2e;
  border-radius: 8px;
}

.module-list h4 {
  margin: 0 0 10px 0;
  color: #fff;
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
