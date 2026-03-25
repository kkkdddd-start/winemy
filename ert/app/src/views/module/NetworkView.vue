<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>网络分析</h2>
        <p class="description">M3 - 连接列表、端口监听、IP 地理</p>
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
          <div class="feature-card" @click="handleFeature('connection-list')">
            <div class="card-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">连接列表</div>
              <div class="card-desc">TCP/UDP 连接状态</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('port-listener')">
            <div class="card-icon">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">端口监听</div>
              <div class="card-desc">监听端口服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('ip-geo')">
            <div class="card-icon">
              <el-icon><Location /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">IP 地理</div>
              <div class="card-desc">IP 地理位置查询</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('net-stat')">
            <div class="card-icon">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">网络统计</div>
              <div class="card-desc">流量统计视图</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>网络连接</span>
            <el-input v-model="searchKeyword" placeholder="搜索IP或端口" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredConnectionList" v-loading="loading" stripe>
          <el-table-column prop="protocol" label="协议" width="80" />
          <el-table-column prop="local_addr" label="本地地址" min-width="150" />
          <el-table-column prop="remote_addr" label="远程地址" min-width="150" />
          <el-table-column prop="state" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStateType(row.state)">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="pid" label="PID" width="100" />
          <el-table-column prop="program" label="程序" min-width="150" show-overflow-tooltip />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Connection, Monitor, Location, DataAnalysis } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ConnectionInfo {
  protocol: string
  local_addr: string
  remote_addr: string
  state: string
  pid: number
  program: string
}

const loading = ref(false)
const searchKeyword = ref('')
const connectionList = ref<ConnectionInfo[]>([])

const filteredConnectionList = computed(() => {
  if (!searchKeyword.value) return connectionList.value
  const keyword = searchKeyword.value.toLowerCase()
  return connectionList.value.filter(c =>
    c.local_addr.toLowerCase().includes(keyword) ||
    c.remote_addr.toLowerCase().includes(keyword) ||
    String(c.pid).includes(keyword)
  )
})

function getStateType(state: string): string {
  const stateMap: Record<string, string> = {
    'ESTABLISHED': 'success',
    'LISTEN': 'primary',
    'TIME_WAIT': 'warning',
    'CLOSE_WAIT': 'warning',
    'SYN_SENT': 'info'
  }
  return stateMap[state] || 'info'
}

async function loadConnectionList() {
  loading.value = true
  try {
    const data = await Go.GetNetworkList()
    if (data) {
      connectionList.value = data as ConnectionInfo[]
    }
  } catch (error) {
    console.error('Failed to load network list:', error)
    ElMessage.error('加载网络连接失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadConnectionList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

onMounted(() => {
  loadConnectionList()
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
