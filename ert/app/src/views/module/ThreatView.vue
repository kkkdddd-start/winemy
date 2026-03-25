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
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ threatStats.solved }}</div>
              <div class="card-label">已处置</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'all' }" @click="handleFeature('all')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">全部威胁</div>
              <div class="card-desc">所有检测到的威胁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'process' }" @click="handleFeature('process')">
            <div class="card-icon danger">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">恶意进程</div>
              <div class="card-desc">可疑进程检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'network' }" @click="handleFeature('network')">
            <div class="card-icon warning">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">可疑网络</div>
              <div class="card-desc">异常网络连接</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'behavior' }" @click="handleFeature('behavior')">
            <div class="card-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">敏感行为</div>
              <div class="card-desc">敏感操作监控</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索威胁名称" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterLevel" placeholder="威胁等级" style="width: 150px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="高危" value="high" />
                <el-option label="中危" value="medium" />
                <el-option label="低危" value="low" />
              </el-select>
              <el-select v-model="filterStatus" placeholder="处置状态" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="待处置" value="pending" />
                <el-option label="已处置" value="solved" />
                <el-option label="已忽略" value="ignored" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'time', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="level" label="等级" width="80" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.level" />
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="120" sortable="custom" />
          <el-table-column prop="name" label="威胁名称" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="location" label="位置" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="time" label="发现时间" width="160" sortable="custom" />
          <el-table-column prop="status" label="状态" width="100" sortable="custom">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="handleKill(row)" :disabled="row.status !== '待处置'">查杀</el-button>
              <el-button type="warning" size="small" @click="handleIsolate(row)" :disabled="row.status !== '待处置'">隔离</el-button>
              <el-button type="info" size="small" @click="handleIgnore(row)">忽略</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredThreatList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="威胁详情" width="700px" destroy-on-close>
      <div class="detail-content" v-if="currentThreat">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="威胁名称" :span="2">{{ currentThreat.name }}</el-descriptions-item>
          <el-descriptions-item label="等级">
            <RiskTag :level="currentThreat.level" />
          </el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentThreat.type }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentThreat.status)">{{ currentThreat.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发现时间">{{ currentThreat.time }}</el-descriptions-item>
          <el-descriptions-item label="位置" :span="2">{{ currentThreat.location }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentThreat.description || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div class="related-section" v-if="currentThreat.related">
          <h4>关联信息</h4>
          <el-tag v-for="item in currentThreat.related" :key="item" type="info" style="margin-right: 5px; margin-bottom: 5px;">{{ item }}</el-tag>
        </div>
        <div class="timeline-section" v-if="currentThreat.timeline">
          <h4>威胁活动时间线</h4>
          <Timeline :items="currentThreat.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="danger" @click="handleKill(currentThreat)" :disabled="currentThreat?.status !== '待处置'">查杀</el-button>
        <el-button type="warning" @click="handleIsolate(currentThreat)" :disabled="currentThreat?.status !== '待处置'">隔离</el-button>
        <el-button type="info" @click="handleIgnore(currentThreat)">忽略</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Warning, WarningFilled, InfoFilled, SuccessFilled, Download, List, Cpu, Connection, Monitor, CircleCheck } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'
import Timeline from '@/components/common/Timeline.vue'

interface ThreatInfo {
  level: string
  type: string
  name: string
  location: string
  time: string
  status: string
  description?: string
  related?: string[]
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterLevel = ref('')
const filterStatus = ref('')
const currentFeature = ref('all')
const threatList = ref<ThreatInfo[]>([])
const detailDialogVisible = ref(false)
const currentThreat = ref<ThreatInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const threatStats = computed(() => ({
  high: threatList.value.filter(t => t.level === '高危' || t.level === 'high').length,
  medium: threatList.value.filter(t => t.level === '中危' || t.level === 'medium').length,
  low: threatList.value.filter(t => t.level === '低危' || t.level === 'low').length,
  solved: threatList.value.filter(t => t.status === '已处置' || t.status === 'solved').length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'all': '全部威胁',
    'process': '恶意进程',
    'network': '可疑网络',
    'behavior': '敏感行为'
  }
  return titles[currentFeature.value] || '威胁列表'
})

const filteredThreatList = computed(() => {
  let result = threatList.value
  
  if (currentFeature.value === 'process') {
    result = result.filter(t => t.type === '恶意进程' || t.type === 'process')
  } else if (currentFeature.value === 'network') {
    result = result.filter(t => t.type === '可疑网络' || t.type === 'network')
  } else if (currentFeature.value === 'behavior') {
    result = result.filter(t => t.type === '敏感行为' || t.type === 'behavior')
  }
  
  if (filterLevel.value) {
    result = result.filter(t => t.level.toLowerCase() === filterLevel.value)
  }
  
  if (filterStatus.value) {
    result = result.filter(t => t.status.toLowerCase() === filterStatus.value)
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(t =>
      t.name.toLowerCase().includes(keyword) ||
      t.location.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredThreatList.value.slice(start, end)
})

function getStatusType(status: string): string {
  const typeMap: Record<string, string> = {
    '待处置': 'warning',
    '已处置': 'success',
    '已忽略': 'info',
    'pending': 'warning',
    'solved': 'success',
    'ignored': 'info'
  }
  return typeMap[status] || 'info'
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

async function loadThreatList() {
  loading.value = true
  try {
    const data = await Go.GetThreats()
    if (data) {
      threatList.value = data as ThreatInfo[]
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

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleKill(row: ThreatInfo | null) {
  const target = row || currentThreat.value
  if (!target) return
  ElMessageBox.confirm(`确定要查杀威胁 ${target.name} 吗？`, '查杀威胁', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'danger'
  }).then(() => {
    ElMessage.success(`已查杀威胁: ${target.name}`)
    detailDialogVisible.value = false
    loadThreatList()
  }).catch(() => {})
}

function handleIsolate(row: ThreatInfo | null) {
  const target = row || currentThreat.value
  if (!target) return
  ElMessageBox.confirm(`确定要隔离威胁 ${target.name} 吗？`, '隔离威胁', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`已隔离威胁: ${target.name}`)
    detailDialogVisible.value = false
    loadThreatList()
  }).catch(() => {})
}

function handleIgnore(row: ThreatInfo | null) {
  const target = row || currentThreat.value
  if (!target) return
  ElMessageBox.confirm(`确定要忽略威胁 ${target.name} 吗？`, '忽略威胁', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(() => {
    ElMessage.success(`已忽略威胁: ${target.name}`)
    detailDialogVisible.value = false
    loadThreatList()
  }).catch(() => {})
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `threat_report_${timestamp}.csv`
    let csv = '等级,类型,威胁名称,位置,发现时间,状态\n'
    filteredThreatList.value.forEach(t => {
      csv += `"${t.level}","${t.type}","${t.name}","${t.location}","${t.time}","${t.status}"\n`
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

.card-icon.success {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
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

.related-section {
  margin-top: 20px;
  padding: 15px;
  background: #1a1a2e;
  border-radius: 8px;
}

.related-section h4 {
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
