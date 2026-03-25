<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>自启动项目</h2>
        <p class="description">M18 - 注册表/启动文件夹/服务/WMI</p>
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
          <div class="feature-card" @click="handleFeature('registry')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">注册表</div>
              <div class="card-desc">Run键自启动</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('startup-folder')">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">启动文件夹</div>
              <div class="card-desc">Startup目录</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service')">
            <div class="card-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">服务自启</div>
              <div class="card-desc">服务自启动项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('wmi')">
            <div class="card-icon">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">WMI自启</div>
              <div class="card-desc">WMI事件订阅</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>自启动项目</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="注册表" value="registry" />
              <el-option label="启动文件夹" value="folder" />
              <el-option label="服务" value="service" />
              <el-option label="WMI" value="wmi" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredAutostartList" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="location" label="位置" min-width="200" show-overflow-tooltip />
          <el-table-column prop="risk" label="风险" width="100">
            <template #default="{ row }">
              <el-tag :type="row.risk === '高危' ? 'danger' : row.risk === '中危' ? 'warning' : 'success'">{{ row.risk }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="danger" size="small" @click="handleDisable(row)">禁用</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Key, Folder, Setting, Cpu } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface AutostartInfo {
  type: string
  name: string
  path: string
  location: string
  risk: string
}

const loading = ref(false)
const filterType = ref('')
const autostartList = ref<AutostartInfo[]>([])

const filteredAutostartList = computed(() => {
  if (!filterType.value) return autostartList.value
  return autostartList.value.filter(a => a.type.toLowerCase() === filterType.value)
})

async function loadAutostartList() {
  loading.value = true
  try {
    const data = await Go.GetAutostartList()
    if (data) {
      autostartList.value = data as AutostartInfo[]
    }
  } catch (error) {
    console.error('Failed to load autostart list:', error)
    ElMessage.error('加载自启动列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadAutostartList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: AutostartInfo) {
  ElMessage.info(`查看: ${row.name}`)
}

function handleDisable(row: AutostartInfo) {
  ElMessage.warning(`禁用: ${row.name}`)
}

onMounted(() => {
  loadAutostartList()
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
