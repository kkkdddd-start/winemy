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
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('software-list')">
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
          <div class="feature-card" @click="handleFeature('anomaly')">
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
          <div class="feature-card" @click="handleFeature('version')">
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
          <div class="feature-card" @click="handleFeature('uninstall')">
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
            <span>已安装软件</span>
            <el-input v-model="searchKeyword" placeholder="搜索软件名称" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredSoftwareList" v-loading="loading" stripe>
          <el-table-column prop="name" label="软件名称" min-width="200" />
          <el-table-column prop="version" label="版本" width="120" />
          <el-table-column prop="vendor" label="供应商" min-width="150" show-overflow-tooltip />
          <el-table-column prop="install_date" label="安装日期" width="120" />
          <el-table-column prop="install_path" label="安装路径" min-width="200" show-overflow-tooltip />
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
import { Refresh, List, Warning, InfoFilled, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface SoftwareInfo {
  name: string
  version: string
  vendor: string
  install_date: string
  install_path: string
}

const loading = ref(false)
const searchKeyword = ref('')
const softwareList = ref<SoftwareInfo[]>([])

const filteredSoftwareList = computed(() => {
  if (!searchKeyword.value) return softwareList.value
  const keyword = searchKeyword.value.toLowerCase()
  return softwareList.value.filter(s =>
    s.name.toLowerCase().includes(keyword) ||
    s.vendor.toLowerCase().includes(keyword)
  )
})

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
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: SoftwareInfo) {
  ElMessage.info(`查看: ${row.name}`)
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
