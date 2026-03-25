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
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('driver-list')">
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
          <div class="feature-card" @click="handleFeature('signature')">
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
          <div class="feature-card" @click="handleFeature('unsigned')">
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
          <div class="feature-card" @click="handleFeature('hook')">
            <div class="card-icon">
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
            <span>驱动列表</span>
            <el-select v-model="filterType" placeholder="筛选状态" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="已签名" value="signed" />
              <el-option label="未签名" value="unsigned" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredDriverList" v-loading="loading" stripe>
          <el-table-column prop="name" label="驱动名称" min-width="200" />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="company" label="公司" width="150" show-overflow-tooltip />
          <el-table-column prop="signature" label="签名状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.signature === 'Signed' ? 'success' : 'danger'">{{ row.signature }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="date" label="日期" width="120" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Stamp, Warning, Link } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface DriverInfo {
  name: string
  path: string
  company: string
  signature: string
  date: string
}

const loading = ref(false)
const filterType = ref('')
const driverList = ref<DriverInfo[]>([])

const filteredDriverList = computed(() => {
  if (!filterType.value) return driverList.value
  return driverList.value.filter(d =>
    (filterType.value === 'signed' && d.signature === 'Signed') ||
    (filterType.value === 'unsigned' && d.signature !== 'Signed')
  )
})

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
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: DriverInfo) {
  ElMessage.info(`查看驱动: ${row.name}`)
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

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
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
