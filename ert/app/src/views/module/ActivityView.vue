<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>活动痕迹</h2>
        <p class="description">M12 - 最近打开、USB使用、浏览器历史</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>
          导出记录
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">活动总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.recentFiles }}</div>
              <div class="card-label">最近文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Usb /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.usb }}</div>
              <div class="card-label">USB使用</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><Browser /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.browser }}</div>
              <div class="card-label">浏览器历史</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'recent-files' }" @click="handleFeature('recent-files')">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">最近文件</div>
              <div class="card-desc">最近打开的文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'usb' }" @click="handleFeature('usb')">
            <div class="card-icon">
              <el-icon><Usb /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">USB使用</div>
              <div class="card-desc">USB设备使用记录</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" :class="{ active: currentFeature === 'browser' }" @click="handleFeature('browser')">
            <div class="card-icon">
              <el-icon><Browser /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">浏览器历史</div>
              <div class="card-desc">浏览器访问记录</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索名称/路径" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="最近文件" value="recent" />
                <el-option label="USB" value="usb" />
                <el-option label="浏览器" value="browser" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'time', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="type" label="类型" width="100" sortable="custom">
            <template #default="{ row }">
              <el-tag :type="getTypeTagType(row.type)">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="time" label="时间" width="160" sortable="custom" />
          <el-table-column prop="detail" label="详情" min-width="150" show-overflow-tooltip />
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
            :total="filteredActivityList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="活动详情" width="650px" destroy-on-close>
      <div class="detail-content" v-if="currentActivity">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="类型">
            <el-tag :type="getTypeTagType(currentActivity.type)">{{ currentActivity.type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="时间">{{ currentActivity.time }}</el-descriptions-item>
          <el-descriptions-item label="名称" :span="2">{{ currentActivity.name }}</el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentActivity.path }}</el-descriptions-item>
          <el-descriptions-item label="详情" :span="2">{{ currentActivity.detail || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div class="timeline-section" v-if="currentActivity.timeline">
          <h4>活动时间线</h4>
          <Timeline :items="currentActivity.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleOpenLocation(currentActivity)" v-if="currentActivity.path">打开位置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Clock, Usb, Browser, Download, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import Timeline from '@/components/common/Timeline.vue'

interface ActivityInfo {
  type: string
  name: string
  path: string
  time: string
  detail: string
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentFeature = ref('recent-files')
const activityList = ref<ActivityInfo[]>([])
const detailDialogVisible = ref(false)
const currentActivity = ref<ActivityInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: activityList.value.length,
  recentFiles: activityList.value.filter(a => a.type === 'recent').length,
  usb: activityList.value.filter(a => a.type === 'usb').length,
  browser: activityList.value.filter(a => a.type === 'browser').length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'recent-files': '最近文件',
    'usb': 'USB使用记录',
    'browser': '浏览器历史'
  }
  return titles[currentFeature.value] || '活动痕迹'
})

const filteredActivityList = computed(() => {
  let result = activityList.value
  
  if (currentFeature.value === 'recent-files') {
    result = result.filter(a => a.type === 'recent')
  } else if (currentFeature.value === 'usb') {
    result = result.filter(a => a.type === 'usb')
  } else if (currentFeature.value === 'browser') {
    result = result.filter(a => a.type === 'browser')
  }
  
  if (filterType.value) {
    result = result.filter(a => a.type.toLowerCase() === filterType.value)
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(a =>
      a.name.toLowerCase().includes(keyword) ||
      a.path.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredActivityList.value.slice(start, end)
})

function getTypeTagType(type: string): string {
  const typeMap: Record<string, string> = {
    'recent': 'primary',
    'usb': 'warning',
    'browser': 'success'
  }
  return typeMap[type] || 'info'
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

async function loadActivityList() {
  loading.value = true
  try {
    const data = await Go.GetActivityList()
    if (data) {
      activityList.value = data as ActivityInfo[]
    }
  } catch (error) {
    console.error('Failed to load activity list:', error)
    ElMessage.error('加载活动痕迹失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadActivityList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleView(row: ActivityInfo) {
  currentActivity.value = row
  detailDialogVisible.value = true
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `activity_report_${timestamp}.csv`
    let csv = '类型,名称,路径,时间,详情\n'
    filteredActivityList.value.forEach(a => {
      csv += `"${a.type}","${a.name}","${a.path}","${a.time}","${a.detail}"\n`
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

function handleOpenLocation(row: ActivityInfo | null) {
  const target = row || currentActivity.value
  if (!target) return
  ElMessage.info(`打开位置: ${target.path}`)
}

onMounted(() => {
  loadActivityList()
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
