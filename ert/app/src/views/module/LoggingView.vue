<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>日志分析</h2>
        <p class="description">M13 - 事件日志、EVTX解析、全文搜索</p>
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
          <div class="feature-card" @click="handleFeature('event-log')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">事件日志</div>
              <div class="card-desc">Windows事件日志</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('evtx')">
            <div class="card-icon">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">EVTX解析</div>
              <div class="card-desc">解析EVTX文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('search')">
            <div class="card-icon">
              <el-icon><Search /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">全文搜索</div>
              <div class="card-desc">日志内容搜索</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('export')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出日志</div>
              <div class="card-desc">导出分析结果</div>
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
          <el-table-column prop="level" label="级别" width="80">
            <template #default="{ row }">
              <el-tag :type="getLevelType(row.level)">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="source" label="来源" width="150" />
          <el-table-column prop="event_id" label="事件ID" width="100" />
          <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
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
import { Refresh, Document, FolderOpened, Search, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface LogInfo {
  time: string
  level: string
  source: string
  event_id: number
  message: string
}

const loading = ref(false)
const searchKeyword = ref('')
const logList = ref<LogInfo[]>([])

const filteredLogList = computed(() => {
  if (!searchKeyword.value) return logList.value
  const keyword = searchKeyword.value.toLowerCase()
  return logList.value.filter(l =>
    l.message.toLowerCase().includes(keyword) ||
    l.source.toLowerCase().includes(keyword)
  )
})

function getLevelType(level: string): string {
  const typeMap: Record<string, string> = {
    'Error': 'danger',
    'Warning': 'warning',
    'Information': 'success',
    'Critical': 'danger'
  }
  return typeMap[level] || 'info'
}

async function loadLogList() {
  loading.value = true
  try {
    const data = await Go.GetLogList()
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
  ElMessage.info(`查看日志: ${row.event_id}`)
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
