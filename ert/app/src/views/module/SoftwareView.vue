<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>软件列表</h2>
        <p class="description">M9 - 已安装软件、异常检测</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>
          导出列表
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
              <div class="card-label">软件总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.normal }}</div>
              <div class="card-label">正常软件</div>
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
              <div class="card-label">可疑软件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Delete /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.outdated }}</div>
              <div class="card-label">版本过时</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'software-list' }" @click="handleFeature('software-list')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">软件列表</div>
              <div class="card-desc">已安装软件列表</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'anomaly' }" @click="handleFeature('anomaly')">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">异常检测</div>
              <div class="card-desc">可疑软件检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'version' }" @click="handleFeature('version')">
            <div class="card-icon">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">版本信息</div>
              <div class="card-desc">软件版本详情</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'uninstall' }" @click="handleFeature('uninstall')">
            <div class="card-icon danger">
              <el-icon><Delete /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">卸载软件</div>
              <div class="card-desc">卸载选中软件</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索软件名称/供应商" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterRisk" placeholder="风险等级" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="正常" value="normal" />
                <el-option label="可疑" value="suspicious" />
                <el-option label="危险" value="danger" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'name', order: 'ascending' }" @sort-change="handleSort">
          <el-table-column prop="name" label="软件名称" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="version" label="版本" width="120" sortable="custom" />
          <el-table-column prop="vendor" label="供应商" min-width="150" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="install_date" label="安装日期" width="120" sortable="custom" />
          <el-table-column prop="risk" label="风险" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.risk" />
            </template>
          </el-table-column>
          <el-table-column prop="install_path" label="安装路径" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="danger" size="small" @click="handleUninstall(row)" v-if="currentFeature === 'uninstall'">卸载</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredSoftwareList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="软件详情" width="650px" destroy-on-close>
      <div class="detail-content" v-if="currentSoftware">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="软件名称" :span="2">{{ currentSoftware.name }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ currentSoftware.version }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <RiskTag :level="currentSoftware.risk" />
          </el-descriptions-item>
          <el-descriptions-item label="供应商">{{ currentSoftware.vendor }}</el-descriptions-item>
          <el-descriptions-item label="安装日期">{{ currentSoftware.install_date }}</el-descriptions-item>
          <el-descriptions-item label="安装路径" :span="2">{{ currentSoftware.install_path }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentSoftware.description || '无' }}</el-descriptions-item>
          <el-descriptions-item label="文件大小">{{ currentSoftware.size || '未知' }}</el-descriptions-item>
          <el-descriptions-item label="最后更新">{{ currentSoftware.last_update || '未知' }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="danger" @click="handleUninstall(currentSoftware)">卸载</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Warning, InfoFilled, Delete, Download, CircleCheck } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'

interface SoftwareInfo {
  name: string
  version: string
  vendor: string
  install_date: string
  install_path: string
  risk: string
  description?: string
  size?: string
  last_update?: string
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterRisk = ref('')
const currentFeature = ref('software-list')
const softwareList = ref<SoftwareInfo[]>([])
const detailDialogVisible = ref(false)
const currentSoftware = ref<SoftwareInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: softwareList.value.length,
  normal: softwareList.value.filter(s => s.risk === 'normal').length,
  suspicious: softwareList.value.filter(s => s.risk === 'suspicious').length,
  outdated: softwareList.value.filter(s => s.risk === 'danger').length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'software-list': '已安装软件',
    'anomaly': '可疑软件',
    'version': '版本信息',
    'uninstall': '卸载管理'
  }
  return titles[currentFeature.value] || '软件列表'
})

const filteredSoftwareList = computed(() => {
  let result = softwareList.value
  
  if (currentFeature.value === 'anomaly') {
    result = result.filter(s => s.risk === 'suspicious' || s.risk === 'danger')
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(keyword) ||
      s.vendor.toLowerCase().includes(keyword)
    )
  }
  
  if (filterRisk.value) {
    result = result.filter(s => s.risk === filterRisk.value)
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredSoftwareList.value.slice(start, end)
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

async function loadSoftwareList() {
  loading.value = true
  try {
    const data = await Go.GetSoftwareList()
    if (data) {
      softwareList.value = data as SoftwareInfo[]
    }
  } catch (error) {
    console.error('Failed to load software list:', error)
    ElMessage.error('加载软件列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadSoftwareList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleView(row: SoftwareInfo) {
  currentSoftware.value = row
  detailDialogVisible.value = true
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `software_list_${timestamp}.csv`
    let csv = '软件名称,版本,供应商,安装日期,风险等级,安装路径\n'
    filteredSoftwareList.value.forEach(s => {
      csv += `"${s.name}","${s.version}","${s.vendor}","${s.install_date}","${s.risk}","${s.install_path}"\n`
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

function handleUninstall(row: SoftwareInfo | null) {
  const target = row || currentSoftware.value
  if (!target) return
  ElMessageBox.confirm(`确定要卸载软件 ${target.name} 吗？`, '卸载软件', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`开始卸载: ${target.name}`)
    detailDialogVisible.value = false
  }).catch(() => {})
}

onMounted(() => {
  loadSoftwareList()
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
</style>
