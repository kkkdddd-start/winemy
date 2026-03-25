<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统补丁</h2>
        <p class="description">M8 - 已安装补丁、缺失补丁</p>
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
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">补丁总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.installed }}</div>
              <div class="card-label">已安装</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.missing }}</div>
              <div class="card-label">缺失补丁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.critical }}</div>
              <div class="card-label">严重补丁</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'installed' }" @click="handleFeature('installed')">
            <div class="card-icon">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">已安装补丁</div>
              <div class="card-desc">系统已安装补丁列表</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'missing' }" @click="handleFeature('missing')">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">缺失补丁</div>
              <div class="card-desc">建议安装的补丁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'export' }" @click="handleFeature('export')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出报告</div>
              <div class="card-desc">导出补丁分析报告</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>{{ currentFeature === 'missing' ? '缺失补丁列表' : '已安装补丁列表' }}</span>
            <div class="header-operations">
              <el-input v-model="searchKeyword" placeholder="搜索KB编号/名称" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterRisk" placeholder="风险等级" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="严重" value="critical" />
                <el-option label="重要" value="important" />
                <el-option label="中等" value="moderate" />
                <el-option label="低级" value="low" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'installed_on', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="kb" label="KB编号" width="130" sortable="custom" />
          <el-table-column prop="name" label="补丁名称" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="risk" label="风险等级" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.risk" />
            </template>
          </el-table-column>
          <el-table-column prop="installed_on" label="安装日期" width="120" sortable="custom" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column prop="installed_by" label="安装者" width="120" />
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
            :total="filteredPatchList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="补丁详情" width="600px" destroy-on-close>
      <div class="detail-content" v-if="currentPatch">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="KB编号">{{ currentPatch.kb }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <RiskTag :level="currentPatch.risk" />
          </el-descriptions-item>
          <el-descriptions-item label="补丁名称" :span="2">{{ currentPatch.name }}</el-descriptions-item>
          <el-descriptions-item label="安装日期">{{ currentPatch.installed_on }}</el-descriptions-item>
          <el-descriptions-item label="安装者">{{ currentPatch.installed_by }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentPatch.description }}</el-descriptions-item>
          <el-descriptions-item label="相关漏洞" :span="2">{{ currentPatch.cve || '无' }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleInstallPatch(currentPatch)" v-if="currentPatch?.status === 'missing'">安装补丁</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, SuccessFilled, WarningFilled, Download, CircleCheck, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'

interface PatchInfo {
  kb: string
  name: string
  installed_on: string
  description: string
  installed_by: string
  risk: string
  status: string
  cve?: string
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterRisk = ref('')
const currentFeature = ref('installed')
const patchList = ref<PatchInfo[]>([])
const detailDialogVisible = ref(false)
const currentPatch = ref<PatchInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)
const sortProp = ref('installed_on')
const sortOrder = ref<'ascending' | 'descending'>('descending')

const stats = computed(() => ({
  total: patchList.value.length,
  installed: patchList.value.filter(p => p.status === 'installed').length,
  missing: patchList.value.filter(p => p.status === 'missing').length,
  critical: patchList.value.filter(p => p.risk === 'critical').length
}))

const filteredPatchList = computed(() => {
  let result = patchList.value
  
  if (currentFeature.value === 'installed') {
    result = result.filter(p => p.status === 'installed')
  } else if (currentFeature.value === 'missing') {
    result = result.filter(p => p.status === 'missing')
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(p => 
      p.kb.toLowerCase().includes(keyword) ||
      p.name.toLowerCase().includes(keyword)
    )
  }
  
  if (filterRisk.value) {
    result = result.filter(p => p.risk === filterRisk.value)
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredPatchList.value.slice(start, end)
})

function handleSort({ prop, order }: { prop: string; order: 'ascending' | 'descending' }) {
  sortProp.value = prop
  sortOrder.value = order
}

function handleSearch() {
  currentPage.value = 1
}

function handleFilter() {
  currentPage.value = 1
}

function handleSizeChange(val: number) {
  pageSize.value = val
  currentPage.value = 1
}

function handleCurrentChange() {
}

async function loadPatchList() {
  loading.value = true
  try {
    const data = await Go.GetPatchList()
    if (data) {
      patchList.value = data as PatchInfo[]
    }
  } catch (error) {
    console.error('Failed to load patch list:', error)
    ElMessage.error('加载补丁列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadPatchList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
  if (feature === 'export') {
    handleExport()
  }
}

function handleView(row: PatchInfo) {
  currentPatch.value = row
  detailDialogVisible.value = true
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `patch_report_${timestamp}.csv`
    let csv = 'KB编号,补丁名称,风险等级,安装日期,安装者,描述\n'
    filteredPatchList.value.forEach(p => {
      csv += `"${p.kb}","${p.name}","${p.risk}","${p.installed_on}","${p.installed_by}","${p.description}"\n`
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

function handleInstallPatch(patch: PatchInfo | null) {
  if (!patch) return
  ElMessageBox.confirm(`确定要安装补丁 ${patch.kb} 吗？`, '安装补丁', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`开始安装补丁: ${patch.kb}`)
  }).catch(() => {})
}

onMounted(() => {
  loadPatchList()
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
</style>
