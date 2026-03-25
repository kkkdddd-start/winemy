<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>文件系统</h2>
        <p class="description">M11 - 文件枚举、哈希、大文件处理</p>
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
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">文件总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.normal }}</div>
              <div class="card-label">正常文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.recent }}</div>
              <div class="card-label">最近修改</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.suspicious }}</div>
              <div class="card-label">可疑文件</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'file-list' }" @click="handleFeature('file-list')">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件枚举</div>
              <div class="card-desc">目录文件浏览</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'hash' }" @click="handleFeature('hash')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件哈希</div>
              <div class="card-desc">MD5/SHA1/SHA256</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'large-file' }" @click="handleFeature('large-file')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">大文件</div>
              <div class="card-desc">查找大文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'recent' }" @click="handleFeature('recent')">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">最近文件</div>
              <div class="card-desc">最近修改的文件</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索文件名称" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterRisk" placeholder="风险等级" style="width: 120px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="正常" value="normal" />
                <el-option label="可疑" value="suspicious" />
                <el-option label="危险" value="danger" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'modified', order: 'descending' }" @sort-change="handleSort">
          <el-table-column prop="name" label="文件名" min-width="200" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="path" label="路径" min-width="300" show-overflow-tooltip />
          <el-table-column prop="size" label="大小" width="120" sortable="custom">
            <template #default="{ row }">
              {{ formatSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="modified" label="修改时间" width="160" sortable="custom" />
          <el-table-column prop="risk" label="风险等级" width="100" sortable="custom">
            <template #default="{ row }">
              <RiskTag :level="row.risk" />
            </template>
          </el-table-column>
          <el-table-column prop="hash" label="哈希" width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="info" size="small" @click="handleHash(row)">哈希</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredFileList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="文件详情" width="700px" destroy-on-close>
      <div class="detail-content" v-if="currentFile">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="文件名" :span="2">{{ currentFile.name }}</el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentFile.path }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ formatSize(currentFile.size) }}</el-descriptions-item>
          <el-descriptions-item label="风险等级">
            <RiskTag :level="currentFile.risk" />
          </el-descriptions-item>
          <el-descriptions-item label="修改时间">{{ currentFile.modified }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentFile.created || '未知' }}</el-descriptions-item>
          <el-descriptions-item label="MD5" :span="2">{{ currentFile.md5 || '计算中...' }}</el-descriptions-item>
          <el-descriptions-item label="SHA256" :span="2">{{ currentFile.sha256 || '计算中...' }}</el-descriptions-item>
        </el-descriptions>
        <div class="timeline-section" v-if="currentFile.timeline">
          <h4>文件活动时间线</h4>
          <Timeline :items="currentFile.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="info" @click="handleHash(currentFile)">计算哈希</el-button>
        <el-button type="danger" @click="handleDelete(currentFile)">删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Folder, Key, Document, Clock, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import RiskTag from '@/components/common/RiskTag.vue'
import Timeline from '@/components/common/Timeline.vue'

interface FileInfo {
  name: string
  path: string
  size: number
  modified: string
  hash: string
  risk: string
  created?: string
  md5?: string
  sha256?: string
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterRisk = ref('')
const currentFeature = ref('file-list')
const fileList = ref<FileInfo[]>([])
const detailDialogVisible = ref(false)
const currentFile = ref<FileInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => {
  const oneWeekAgo = new Date()
  oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)
  return {
    total: fileList.value.length,
    normal: fileList.value.filter(f => f.risk === 'normal').length,
    recent: fileList.value.filter(f => new Date(f.modified) > oneWeekAgo).length,
    suspicious: fileList.value.filter(f => f.risk === 'suspicious' || f.risk === 'danger').length
  }
})

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'file-list': '文件列表',
    'hash': '文件哈希',
    'large-file': '大文件',
    'recent': '最近文件'
  }
  return titles[currentFeature.value] || '文件列表'
})

const filteredFileList = computed(() => {
  let result = fileList.value
  const oneWeekAgo = new Date()
  oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)
  
  if (currentFeature.value === 'recent') {
    result = result.filter(f => new Date(f.modified) > oneWeekAgo)
  } else if (currentFeature.value === 'large-file') {
    result = result.filter(f => f.size > 100 * 1024 * 1024)
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(f => f.name.toLowerCase().includes(keyword))
  }
  
  if (filterRisk.value) {
    result = result.filter(f => f.risk === filterRisk.value)
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredFileList.value.slice(start, end)
})

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  while (bytes >= 1024 && index < units.length - 1) {
    bytes /= 1024
    index++
  }
  return `${bytes.toFixed(2)} ${units[index]}`
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

async function loadFileList() {
  loading.value = true
  try {
    const data = await Go.GetFileList()
    if (data) {
      fileList.value = data as FileInfo[]
    }
  } catch (error) {
    console.error('Failed to load file list:', error)
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadFileList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleView(row: FileInfo) {
  currentFile.value = row
  detailDialogVisible.value = true
}

function handleHash(row: FileInfo | null) {
  const target = row || currentFile.value
  if (!target) return
  ElMessage.info(`正在计算文件哈希: ${target.name}`)
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `file_list_${timestamp}.csv`
    let csv = '文件名,路径,大小,修改时间,风险等级,哈希\n'
    filteredFileList.value.forEach(f => {
      csv += `"${f.name}","${f.path}","${formatSize(f.size)}","${f.modified}","${f.risk}","${f.hash}"\n`
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

function handleDelete(row: FileInfo | null) {
  const target = row || currentFile.value
  if (!target) return
  ElMessageBox.confirm(`确定要删除文件 ${target.name} 吗？`, '删除文件', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`已删除文件: ${target.name}`)
    detailDialogVisible.value = false
    loadFileList()
  }).catch(() => {})
}

onMounted(() => {
  loadFileList()
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
