<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>WMIC检测</h2>
        <p class="description">M21 - WMIC 命令历史检测</p>
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
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('command-history')">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">命令历史</div>
              <div class="card-desc">WMIC历史命令</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('persistence-check')">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">持久化检测</div>
              <div class="card-desc">WMIC持久化检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('export')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出报告</div>
              <div class="card-desc">导出WMIC分析</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>WMIC 历史</span>
            <el-input v-model="searchKeyword" placeholder="搜索命令" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredHistoryList" v-loading="loading" stripe>
          <el-table-column prop="time" label="时间" width="180" />
          <el-table-column prop="user" label="用户" width="120" />
          <el-table-column prop="command" label="命令" min-width="400" show-overflow-tooltip />
          <el-table-column prop="result" label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="row.result === '成功' ? 'success' : 'danger'">{{ row.result }}</el-tag>
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
import { Refresh, Clock, Warning, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface WMICHistory {
  time: string
  user: string
  command: string
  result: string
}

const loading = ref(false)
const searchKeyword = ref('')
const historyList = ref<WMICHistory[]>([])

const filteredHistoryList = computed(() => {
  if (!searchKeyword.value) return historyList.value
  const keyword = searchKeyword.value.toLowerCase()
  return historyList.value.filter(h => h.command.toLowerCase().includes(keyword))
})

async function loadHistoryList() {
  loading.value = true
  try {
    const data = await Go.GetWMICHistory()
    if (data) {
      historyList.value = data as WMICHistory[]
    }
  } catch (error) {
    console.error('Failed to load WMIC history:', error)
    ElMessage.error('加载WMIC历史失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadHistoryList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: WMICHistory) {
  ElMessage.info(`查看命令: ${row.command}`)
}

onMounted(() => {
  loadHistoryList()
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
