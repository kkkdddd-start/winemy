<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>账户分析</h2>
        <p class="description">M14 - 本地/域账户、组、权限</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>
          导出账户
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.total }}</div>
              <div class="card-label">账户总数</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.enabled }}</div>
              <div class="card-label">已启用</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.disabled }}</div>
              <div class="card-label">已禁用</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.admin }}</div>
              <div class="card-label">管理员</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'local-account' }" @click="handleFeature('local-account')">
            <div class="card-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">本地账户</div>
              <div class="card-desc">本地用户账户</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'domain-account' }" @click="handleFeature('domain-account')">
            <div class="card-icon">
              <el-icon><Avatar /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">域账户</div>
              <div class="card-desc">域用户账户</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'group' }" @click="handleFeature('group')">
            <div class="card-icon">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">用户组</div>
              <div class="card-desc">用户组管理</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: currentFeature === 'privilege' }" @click="handleFeature('privilege')">
            <div class="card-icon warning">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">权限分析</div>
              <div class="card-desc">账户权限分析</div>
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
              <el-input v-model="searchKeyword" placeholder="搜索用户名" style="width: 200px" clearable @input="handleSearch" />
              <el-select v-model="filterType" placeholder="账户类型" style="width: 150px" clearable @change="handleFilter">
                <el-option label="全部" value="" />
                <el-option label="本地账户" value="local" />
                <el-option label="域账户" value="domain" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe :default-sort="{ prop: 'username', order: 'ascending' }" @sort-change="handleSort">
          <el-table-column prop="username" label="用户名" width="150" sortable="custom" />
          <el-table-column prop="type" label="类型" width="100" sortable="custom">
            <template #default="{ row }">
              <el-tag :type="row.type === 'Local' ? 'primary' : 'success'">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="full_name" label="全名" width="150" sortable="custom" />
          <el-table-column prop="groups" label="所属组" min-width="150" sortable="custom" show-overflow-tooltip />
          <el-table-column prop="last_login" label="最后登录" width="160" sortable="custom" />
          <el-table-column prop="status" label="状态" width="100" sortable="custom">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Enabled' ? 'success' : 'danger'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="warning" size="small" @click="handleToggleStatus(row)">
                {{ row.status === 'Enabled' ? '禁用' : '启用' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="filteredAccountList.length"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="账户详情" width="650px" destroy-on-close>
      <div class="detail-content" v-if="currentAccount">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户名">{{ currentAccount.username }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="currentAccount.type === 'Local' ? 'primary' : 'success'">{{ currentAccount.type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="全名">{{ currentAccount.full_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentAccount.status === 'Enabled' ? 'success' : 'danger'">{{ currentAccount.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="所属组" :span="2">{{ currentAccount.groups }}</el-descriptions-item>
          <el-descriptions-item label="最后登录" :span="2">{{ currentAccount.last_login }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentAccount.description || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div class="privilege-section" v-if="currentAccount.privileges">
          <h4>权限信息</h4>
          <el-tag v-for="priv in currentAccount.privileges" :key="priv" type="warning" style="margin-right: 5px; margin-bottom: 5px;">{{ priv }}</el-tag>
        </div>
        <div class="timeline-section" v-if="currentAccount.timeline">
          <h4>账户活动时间线</h4>
          <Timeline :items="currentAccount.timeline" />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="warning" @click="handleToggleStatus(currentAccount)">
          {{ currentAccount?.status === 'Enabled' ? '禁用账户' : '启用账户' }}
        </el-button>
        <el-button type="danger" @click="handleDeleteAccount(currentAccount)" v-if="currentAccount?.type === 'Local'">删除账户</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, User, Avatar, UserFilled, Key, Download, CircleCheck, Lock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'
import Timeline from '@/components/common/Timeline.vue'

interface AccountInfo {
  username: string
  type: string
  full_name: string
  groups: string
  last_login: string
  status: string
  description?: string
  privileges?: string[]
  timeline?: Array<{ time: string; event: string; type: string }>
}

const loading = ref(false)
const exporting = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentFeature = ref('local-account')
const accountList = ref<AccountInfo[]>([])
const detailDialogVisible = ref(false)
const currentAccount = ref<AccountInfo | null>(null)

const currentPage = ref(1)
const pageSize = ref(20)

const stats = computed(() => ({
  total: accountList.value.length,
  enabled: accountList.value.filter(a => a.status === 'Enabled').length,
  disabled: accountList.value.filter(a => a.status === 'Disabled').length,
  admin: accountList.value.filter(a => a.groups.includes('Administrators') || a.groups.includes('Admin')).length
}))

const featureTitle = computed(() => {
  const titles: Record<string, string> = {
    'local-account': '本地账户',
    'domain-account': '域账户',
    'group': '用户组',
    'privilege': '权限分析'
  }
  return titles[currentFeature.value] || '账户列表'
})

const filteredAccountList = computed(() => {
  let result = accountList.value
  
  if (currentFeature.value === 'local-account') {
    result = result.filter(a => a.type === 'Local')
  } else if (currentFeature.value === 'domain-account') {
    result = result.filter(a => a.type === 'Domain')
  }
  
  if (filterType.value) {
    result = result.filter(a =>
      (filterType.value === 'local' && a.type === 'Local') ||
      (filterType.value === 'domain' && a.type === 'Domain')
    )
  }
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(a => a.username.toLowerCase().includes(keyword))
  }
  
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredAccountList.value.slice(start, end)
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

async function loadAccountList() {
  loading.value = true
  try {
    const data = await Go.GetAccountList()
    if (data) {
      accountList.value = data as AccountInfo[]
    }
  } catch (error) {
    console.error('Failed to load account list:', error)
    ElMessage.error('加载账户列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadAccountList()
}

function handleFeature(feature: string) {
  currentFeature.value = feature
  currentPage.value = 1
}

function handleView(row: AccountInfo) {
  currentAccount.value = row
  detailDialogVisible.value = true
}

function handleToggleStatus(row: AccountInfo | null) {
  const target = row || currentAccount.value
  if (!target) return
  const action = target.status === 'Enabled' ? '禁用' : '启用'
  ElMessageBox.confirm(`确定要${action}账户 ${target.username} 吗？`, `${action}账户`, {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success(`${action}账户: ${target.username}`)
    detailDialogVisible.value = false
    loadAccountList()
  }).catch(() => {})
}

function handleDeleteAccount(row: AccountInfo | null) {
  const target = row || currentAccount.value
  if (!target) return
  ElMessageBox.confirm(`确定要删除账户 ${target.username} 吗？此操作不可恢复！`, '删除账户', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'danger'
  }).then(() => {
    ElMessage.success(`已删除账户: ${target.username}`)
    detailDialogVisible.value = false
    loadAccountList()
  }).catch(() => {})
}

async function handleExport() {
  exporting.value = true
  try {
    const timestamp = new Date().toISOString().slice(0, 10)
    const filename = `account_report_${timestamp}.csv`
    let csv = '用户名,类型,全名,所属组,最后登录,状态\n'
    filteredAccountList.value.forEach(a => {
      csv += `"${a.username}","${a.type}","${a.full_name}","${a.groups}","${a.last_login}","${a.status}"\n`
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
  loadAccountList()
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

.privilege-section {
  margin-top: 20px;
  padding: 15px;
  background: #1a1a2e;
  border-radius: 8px;
}

.privilege-section h4 {
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
