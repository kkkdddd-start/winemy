<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>服务管理</h2>
        <p class="description">M5 - 服务列表、启停操作</p>
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
          <div class="feature-card" @click="handleFeature('service-list')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">服务列表</div>
              <div class="card-desc">所有系统服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-start')">
            <div class="card-icon success">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">启动服务</div>
              <div class="card-desc">启动停止的服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-stop')">
            <div class="card-icon danger">
              <el-icon><VideoPause /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">停止服务</div>
              <div class="card-desc">停止运行中的服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-config')">
            <div class="card-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">服务配置</div>
              <div class="card-desc">服务属性配置</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>服务列表</span>
            <el-input v-model="searchKeyword" placeholder="搜索服务名称" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredServiceList" v-loading="loading" stripe>
          <el-table-column prop="name" label="服务名" min-width="150" />
          <el-table-column prop="display_name" label="显示名称" min-width="150" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Running' ? 'success' : 'info'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="start_type" label="启动类型" width="120" />
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status !== 'Running'" type="success" size="small" @click="handleStart(row)">启动</el-button>
              <el-button v-if="row.status === 'Running'" type="warning" size="small" @click="handleStop(row)">停止</el-button>
              <el-button type="primary" size="small" @click="handleConfig(row)">配置</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, VideoPlay, VideoPause, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ServiceInfo {
  name: string
  display_name: string
  status: string
  start_type: string
  path: string
}

const loading = ref(false)
const searchKeyword = ref('')
const serviceList = ref<ServiceInfo[]>([])

const filteredServiceList = computed(() => {
  if (!searchKeyword.value) return serviceList.value
  const keyword = searchKeyword.value.toLowerCase()
  return serviceList.value.filter(s =>
    s.name.toLowerCase().includes(keyword) ||
    s.display_name.toLowerCase().includes(keyword)
  )
})

async function loadServiceList() {
  loading.value = true
  try {
    const data = await Go.GetServiceList()
    if (data) {
      serviceList.value = data as ServiceInfo[]
    }
  } catch (error) {
    console.error('Failed to load service list:', error)
    ElMessage.error('加载服务列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadServiceList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleStart(row: ServiceInfo) {
  ElMessage.success(`启动服务: ${row.display_name}`)
}

function handleStop(row: ServiceInfo) {
  ElMessage.warning(`停止服务: ${row.display_name}`)
}

function handleConfig(row: ServiceInfo) {
  ElMessage.info(`配置服务: ${row.display_name}`)
}

onMounted(() => {
  loadServiceList()
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

.card-icon.success {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
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
