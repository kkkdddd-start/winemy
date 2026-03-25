<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>IIS日志</h2>
        <p class="description">M24 - IIS/Apache/SQL Server 日志</p>
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
          <div class="feature-card" @click="handleFeature('iis')">
            <div class="card-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">IIS日志</div>
              <div class="card-desc">IIS Web日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('apache')">
            <div class="card-icon">
              <el-icon><Grid /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Apache日志</div>
              <div class="card-desc">Apache访问日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('sql')">
            <div class="card-icon">
              <el-icon><Database /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">SQL Server日志</div>
              <div class="card-desc">SQL Server审计</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('analysis')">
            <div class="card-icon">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">日志分析</div>
              <div class="card-desc">日志统计报表</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>日志列表</span>
            <el-input v-model="searchKeyword" placeholder="搜索日志内容" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredLogList" v-loading="loading" stripe>
          <el-table-column prop="time" label="时间" width="180" />
          <el-table-column prop="source" label="来源" width="120">
            <template #default="{ row }">
              <el-tag>{{ row.source }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="client_ip" label="客户端IP" width="140" />
          <el-table-column prop="method" label="方法" width="80" />
          <el-table-column prop="uri" label="URI" min-width="250" show-overflow-tooltip />
          <el-table-column prop="status" label="状态码" width="80">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
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
import { Refresh, Monitor, Grid, Database, DataAnalysis } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface LogInfo {
  time: string
  source: string
  client_ip: string
  method: string
  uri: string
  status: number
}

const loading = ref(false)
const searchKeyword = ref('')
const logList = ref<LogInfo[]>([])

const filteredLogList = computed(() => {
  if (!searchKeyword.value) return logList.value
  const keyword = searchKeyword.value.toLowerCase()
  return logList.value.filter(l =>
    l.uri.toLowerCase().includes(keyword) ||
    l.client_ip.includes(keyword)
  )
})

function getStatusType(status: number): string {
  if (status >= 500) return 'danger'
  if (status >= 400) return 'warning'
  if (status >= 300) return 'info'
  return 'success'
}

async function loadLogList() {
  loading.value = true
  try {
    const data = await Go.GetIISLogList()
    if (data) {
      logList.value = data as LogInfo[]
    }
  } catch (error) {
    console.error('Failed to load log list:', error)
    ElMessage.error('加载日志列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadLogList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: LogInfo) {
  ElMessage.info(`查看日志: ${row.uri}`)
}

onMounted(() => {
  loadLogList()
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
