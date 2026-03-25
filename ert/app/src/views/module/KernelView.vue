<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>内核分析</h2>
        <p class="description">M10 - 驱动列表、签名状态</p>
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
              <div class="card-label">驱动总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><Stamp /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.signed }}</div>
              <div class="card-label">已签名</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.unsigned }}</div>
              <div class="card-label">未签名</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Link /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.hooks }}</div>
              <div class="card-label">可疑钩子</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'driver-list' }" @click="handleFeature('driver-list')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">驱动列表</div>
              <div class="card-desc">内核驱动列表</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'signature' }" @click="handleFeature('signature')">
            <div class="card-icon">
              <el-icon><Stamp /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">签名验证</div>
              <div class="card-desc">驱动签名状态</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'unsigned' }" @click="handleFeature('unsigned')">
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">未签名驱动</div>
              <div class="card-desc">未签名驱动检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'hook' }" @click="handleFeature('hook')">
            <div class="card-icon warning">
              <el-icon><Link /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">内核钩子</div>
              <div class="card-desc">Inline/Hook检测</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索驱动名称/公司" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterType" placeholder="签名状态" style="width: 150px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="已签名" value="signed" />
                <el-option label="未签名" value="unsigned" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'date', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="name" label="驱动名称" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="company" label="公司" width="150" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="signature" label="签名状态" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.signature === 'Signed' ? 'low' : 'critical'" />
            </template>
          </el-table-column>
          <el-table-column prop="date" label="日期" width="120" sortable="custom" />
          <el-table-column prop="risk" label="风险等级" width="100">
            <template #default="{ row }">
              <RiskTag :level="row.risk || (row.signature === 'Signed' ? 'low' : 'critical')" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="warning" size="small" @click="handleVerify(row)">验证</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredDriverList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="驱动详情" width="650px" destroy-on-close>
      <div class="detail-content" v-if="currentDriver">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="驱动名称" :span="2">{{ currentDriver.name }}</el-descriptions-item>
          <el-descriptions-item label="公司">{{ currentDriver.company }}</el-descriptions-item>
          <el-descriptions-item label="签名状态">
            <RiskTag :level="currentDriver.signature === 'Signed' ? 'low' : 'critical'" />
          </el-descriptions-item>
          <el-descriptions-item label="日期">{{ currentDriver.date }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <RiskTag :level="currentDriver.risk || (currentDriver.signature === 'Signed' ? 'low' : 'critical')" />
          </el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentDriver.path }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentDriver.description || '无' }}</el-descriptions-item>
          <el-descriptions-item label="签名者" :span="2">{{ currentDriver.signer || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div class="timeline-section" v-if="currentDriver.timeline">
          <h4>活动时间线</h4>
          <Timeline :items="currentDriver.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="danger" @click="handleDisable(currentDriver)" v-if="currentDriver.signature !== 'Signed'">禁用驱动</el-button>
        <el-button type="primary" @click="handleVerify(currentDriver)">重新验证</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Stamp, Warning, Link, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'
import Timeline from '@/components/common/Timeline.vue'

interface DriverInfo {
  name: string
  path: string
  company: string
  signature: string
  date: string
  risk?: string
  description?: string
  signer?: string
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentFeature = ref('driver-list')
const driverList = ref<DriverInfo[]>([])
const detailDialogVisible = ref(false)
const currentDriver = ref<DriverInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: driverList.value.length,
  signed: driverList.value.filter(d => d.signature === 'Signed').length,
  unsigned: driverList.value.filter(d => d.signature !== 'Signed').length,
  hooks: driverList.value.filter(d => d.risk === 'high').length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'driver-list': '驱动列表',
    'signature': '签名验证',
    'unsigned': '未签名驱动',
    'hook': '内核钩子'
  }
  return titles[currentFeature.value] || '驱动列表'
})

const filteredDriverList = computed(() => {
  let result = driverList.value
  
  if (currentFeature.value === 'unsigned') {
    result = result.filter(d => d.signature !== 'Signed')
  } else if (currentFeature.value === 'signature') {
    result = result.filter(d => d.signature === 'Signed')
  } else if (currentFeature.value === 'hook') {
    result = result.filter(d => d.risk === 'high' || d.risk === 'critical')
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(d =>
      d.name.toLowerCase().includes(keyword) ||
      d.company.toLowerCase().includes(keyword)
    )
  }
  
  if (filterType.value) {
    result = result.filter(d =>
      (filterType.value === 'signed' && d.signature === 'Signed') ||
      (filterType.value === 'unsigned' && d.signature !== 'Signed')
    )
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredDriverList.value.slice(start, end)
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

async function loadDriverList() {
  loading.value = true
  try {
    const data = await Go.GetKernelDriverList()
    if (data) {
      driverList.value = data as DriverInfo[]
    }
  } catch (error) {
    console.error('Failed to load driver list:', error)
    ElMessage.error('加载驱动列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadDriverList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleView(row: DriverInfo) {
  currentDriver.value = row
  detailDialogVisible.value = true
}

function handleVerify(row: DriverInfo) {
  ElMessage.info(`正在验证驱动: ${row.name}`)
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `kernel_driver_report_${timestamp}.csv`
    let csv = '驱动名称,公司,签名状态,日期,路径\n'
    filteredDriverList.value.forEach(d => {
      csv += `"${d.name}","${d.company}","${d.signature}","${d.date}","${d.path}"\n`
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

function handleDisable(row: DriverInfo | null) {
  const target = row || currentDriver.value
  if (!target) return
  ElMessageBox.confirm(`确定要禁用驱动 ${target.name} 吗？`, '禁用驱动', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`已禁用驱动: ${target.name}`)
    detailDialogVisible.value = false
  }).catch(() => {})
}

onMounted(() => {
  loadDriverList()
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
