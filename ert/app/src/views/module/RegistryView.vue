<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>注册表分析</h2>
        <p class="description">M4 - 关键项检测、持久化、自启动</p>
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
          <div class="feature-card" @click="handleFeature('key-detection')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">关键项检测</div>
              <div class="card-desc">检测可疑注册表项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('persistence')">
            <div class="card-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">持久化检测</div>
              <div class="card-desc">自启动项检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('autostart')">
            <div class="card-icon">
              <el-icon><Switch /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">自启动项</div>
              <div class="card-desc">Run键自启动</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('reg-diff')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">注册表对比</div>
              <div class="card-desc">基线对比分析</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>注册表项</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="自启动" value="autostart" />
              <el-option label="持久化" value="persistence" />
              <el-option label="可疑" value="suspicious" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredRegistryList" v-loading="loading" stripe>
          <el-table-column prop="key" label="注册表键" min-width="250" show-overflow-tooltip />
          <el-table-column prop="value" label="值" min-width="200" show-overflow-tooltip />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="getTypeColor(row.type)">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="risk" label="风险等级" width="100">
            <template #default="{ row }">
              <el-tag :type="getRiskType(row.risk)">{{ row.risk }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Key, Lock, Switch, Document } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface RegistryItem {
  key: string
  value: string
  type: string
  risk: string
}

const loading = ref(false)
const filterType = ref('')
const registryList = ref<RegistryItem[]>([])

const filteredRegistryList = computed(() => {
  if (!filterType.value) return registryList.value
  return registryList.value.filter(r => r.type === filterType.value)
})

function getTypeColor(type: string): string {
  const colorMap: Record<string, string> = {
    'autostart': 'primary',
    'persistence': 'warning',
    'suspicious': 'danger'
  }
  return colorMap[type] || 'info'
}

function getRiskType(risk: string): string {
  const riskMap: Record<string, string> = {
    '高危': 'danger',
    '中危': 'warning',
    '低危': 'info'
  }
  return riskMap[risk] || 'info'
}

async function loadRegistryList() {
  loading.value = true
  try {
    const data = await Go.GetRegistryList()
    if (data) {
      registryList.value = data as RegistryItem[]
    }
  } catch (error) {
    console.error('Failed to load registry list:', error)
    ElMessage.error('加载注册表列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadRegistryList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: RegistryItem) {
  ElMessage.info(`查看: ${row.key}`)
}

onMounted(() => {
  loadRegistryList()
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
