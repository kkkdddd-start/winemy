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
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('local-account')">
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
          <div class="feature-card" @click="handleFeature('domain-account')">
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
          <div class="feature-card" @click="handleFeature('group')">
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
          <div class="feature-card" @click="handleFeature('privilege')">
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
            <span>账户列表</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="本地账户" value="local" />
              <el-option label="域账户" value="domain" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredAccountList" v-loading="loading" stripe>
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="row.type === 'Local' ? 'primary' : 'success'">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="full_name" label="全名" width="150" />
          <el-table-column prop="groups" label="所属组" min-width="150" show-overflow-tooltip />
          <el-table-column prop="last_login" label="最后登录" width="160" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Enabled' ? 'success' : 'danger'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="warning" size="small" @click="handleDisable(row)">{{ row.status === 'Enabled' ? '禁用' : '启用' }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, User, Avatar, UserFilled, Key } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface AccountInfo {
  username: string
  type: string
  full_name: string
  groups: string
  last_login: string
  status: string
}

const loading = ref(false)
const filterType = ref('')
const accountList = ref<AccountInfo[]>([])

const filteredAccountList = computed(() => {
  if (!filterType.value) return accountList.value
  return accountList.value.filter(a =>
    (filterType.value === 'local' && a.type === 'Local') ||
    (filterType.value === 'domain' && a.type === 'Domain')
  )
})

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
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: AccountInfo) {
  ElMessage.info(`查看账户: ${row.username}`)
}

function handleDisable(row: AccountInfo) {
  ElMessage.warning(`${row.status === 'Enabled' ? '禁用' : '启用'}账户: ${row.username}`)
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
}

.feature-card:hover {
  background: #1a2a4a;
  transform: translateY(-2px);
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

.card-icon.warning {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
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
</style>
