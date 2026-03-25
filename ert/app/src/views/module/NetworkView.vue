<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>网络分析</h2>
        <p class="description">网络连接、监听端口、IP地理位置</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索 IP/端口/进程" style="width: 200px" clearable @keyup.enter="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="protocolFilter" placeholder="协议" clearable style="width: 100px; margin-left: 8px;">
          <el-option label="TCP" value="tcp" />
          <el-option label="UDP" value="udp" />
        </el-select>
        <el-select v-model="stateFilter" placeholder="状态" clearable style="width: 120px; margin-left: 8px;">
          <el-option label="ESTABLISHED" value="ESTABLISHED" />
          <el-option label="LISTENING" value="LISTENING" />
          <el-option label="TIME_WAIT" value="TIME_WAIT" />
          <el-option label="CLOSE_WAIT" value="CLOSE_WAIT" />
        </el-select>
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon total"><el-icon><Connection /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ connectionStats.total }}</div>
              <div class="card-label">总连接数</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon danger"><el-icon><Warning /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ connectionStats.risk }}</div>
              <div class="card-label">可疑连接</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon listening"><el-icon><Monitor /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ connectionStats.listening }}</div>
              <div class="card-label">监听端口</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon foreign"><el-icon><Location /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ connectionStats.foreign }}</div>
              <div class="card-label">境外连接</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="连接列表" name="connections">
          <el-card>
            <el-table :data="filteredConnections" v-loading="loading" stripe @row-click="handleRowClick">
              <el-table-column prop="protocol" label="协议" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.protocol === 'TCP' ? 'success' : 'warning'" size="small">{{ row.protocol }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="local_addr" label="本地地址" width="150" show-overflow-tooltip />
              <el-table-column prop="local_port" label="本地端口" width="100" />
              <el-table-column prop="remote_addr" label="远程地址" width="150" show-overflow-tooltip>
                <template #default="{ row }">
                  <span :class="{ 'risk-ip': row.is_risk }">{{ row.remote_addr }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="remote_port" label="远程端口" width="100" />
              <el-table-column prop="state" label="状态" width="120">
                <template #default="{ row }">
                  <el-tag :type="getStateTagType(row.state)" size="small">{{ row.state }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="pid" label="PID" width="80" />
              <el-table-column prop="process_name" label="进程" width="150" show-overflow-tooltip />
              <el-table-column prop="location" label="地理位置" width="120">
                <template #default="{ row }">
                  <span v-if="row.location">{{ row.location }}</span>
                  <span v-else-if="row.is_foreign" class="risk-text">境外</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column prop="risk_level" label="风险" width="80">
                <template #default="{ row }">
                  <el-tag v-if="row.risk_level === 3" type="danger" size="small">严重</el-tag>
                  <el-tag v-else-if="row.risk_level === 2" type="danger" size="small">高</el-tag>
                  <el-tag v-else-if="row.risk_level === 1" type="warning" size="small">中</el-tag>
                  <el-tag v-else type="success" size="small">低</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="监听端口" name="listening">
          <el-card>
            <el-table :data="listeningPorts" v-loading="loading" stripe>
              <el-table-column prop="port" label="端口" width="100" sortable />
              <el-table-column prop="protocol" label="协议" width="80">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.protocol }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="process_name" label="进程" min-width="150" show-overflow-tooltip />
              <el-table-column prop="pid" label="PID" width="80" />
              <el-table-column prop="risk_level" label="风险" width="100">
                <template #default="{ row }">
                  <el-tag v-if="row.is_suspicious" type="danger" size="small">可疑</el-tag>
                  <el-tag v-else type="success" size="small">正常</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="说明" min-width="200">
                <template #default="{ row }">
                  <span v-if="row.is_suspicious" class="risk-text">{{ row.description }}</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog v-model="detailDialogVisible" title="连接详情" width="600px">
      <el-descriptions v-if="selectedConnection" :column="2" border>
        <el-descriptions-item label="协议">{{ selectedConnection.protocol }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStateTagType(selectedConnection.state)" size="small">{{ selectedConnection.state }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="本地地址">{{ selectedConnection.local_addr }}</el-descriptions-item>
        <el-descriptions-item label="本地端口">{{ selectedConnection.local_port }}</el-descriptions-item>
        <el-descriptions-item label="远程地址">{{ selectedConnection.remote_addr }}</el-descriptions-item>
        <el-descriptions-item label="远程端口">{{ selectedConnection.remote_port }}</el-descriptions-item>
        <el-descriptions-item label="PID">{{ selectedConnection.pid }}</el-descriptions-item>
        <el-descriptions-item label="进程名">{{ selectedConnection.process_name }}</el-descriptions-item>
        <el-descriptions-item label="地理位置" :span="2">{{ selectedConnection.location || '未知' }}</el-descriptions-item>
        <el-descriptions-item label="风险等级" :span="2">
          <el-tag v-if="selectedConnection.risk_level === 3" type="danger">严重</el-tag>
          <el-tag v-else-if="selectedConnection.risk_level === 2" type="danger">高</el-tag>
          <el-tag v-else-if="selectedConnection.risk_level === 1" type="warning">中</el-tag>
          <el-tag v-else type="success">低</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="handleBlockIP" type="danger">封禁 IP</el-button>
        <el-button @click="handleKillProcess" type="warning">结束进程</el-button>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Search, Connection, Warning, Monitor, Location } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface NetworkConnection {
  id: number
  protocol: string
  local_addr: string
  local_port: number
  remote_addr: string
  remote_port: number
  state: string
  pid: number
  process_name: string
  location?: string
  is_foreign?: boolean
  is_risk?: boolean
  risk_level: number
}

const loading = ref(false)
const searchKeyword = ref('')
const protocolFilter = ref('')
const stateFilter = ref('')
const activeTab = ref('connections')
const detailDialogVisible = ref(false)
const selectedConnection = ref<NetworkConnection | null>(null)

const mockConnections = ref<NetworkConnection[]>([
  { id: 1, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 50685, remote_addr: '142.250.185.4', remote_port: 443, state: 'ESTABLISHED', pid: 1024, process_name: 'chrome.exe', location: '美国', is_foreign: true, is_risk: false, risk_level: 1 },
  { id: 2, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 443, remote_addr: '0.0.0.0', remote_port: 0, state: 'LISTENING', pid: 4, process_name: 'System', risk_level: 0 },
  { id: 3, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 3389, remote_addr: '0.0.0.0', remote_port: 0, state: 'LISTENING', pid: 980, process_name: 'svchost.exe', risk_level: 2 },
  { id: 4, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 49678, remote_addr: '192.168.1.1', remote_port: 445, state: 'ESTABLISHED', pid: 4, process_name: 'System', risk_level: 0 },
  { id: 5, protocol: 'UDP', local_addr: '192.168.1.100', local_port: 53, remote_addr: '0.0.0.0', remote_port: 0, state: '-', pid: 1300, process_name: 'svchost.exe', risk_level: 0 },
  { id: 6, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 49789, remote_addr: '23.105.185.42', remote_port: 4444, state: 'ESTABLISHED', pid: 5120, process_name: 'trojan.exe', location: '境外', is_foreign: true, is_risk: true, risk_level: 3 },
  { id: 7, protocol: 'TCP', local_addr: '192.168.1.100', local_port: 49790, remote_addr: '10.0.0.1', remote_port: 5555, state: 'ESTABLISHED', pid: 6144, process_name: 'meterpreter.exe', location: '内网', is_risk: true, risk_level: 3 },
])

const connectionStats = computed(() => ({
  total: mockConnections.value.length,
  risk: mockConnections.value.filter(c => c.risk_level >= 2).length,
  listening: mockConnections.value.filter(c => c.state === 'LISTENING').length,
  foreign: mockConnections.value.filter(c => c.is_foreign).length
}))

const filteredConnections = computed(() => {
  let result = mockConnections.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(c => 
      c.remote_addr.toLowerCase().includes(keyword) ||
      String(c.remote_port).includes(keyword) ||
      c.process_name.toLowerCase().includes(keyword) ||
      String(c.pid).includes(keyword)
    )
  }
  if (protocolFilter.value) {
    result = result.filter(c => c.protocol.toLowerCase() === protocolFilter.value.toLowerCase())
  }
  if (stateFilter.value) {
    result = result.filter(c => c.state === stateFilter.value)
  }
  return result
})

const listeningPorts = computed(() => {
  return mockConnections.value
    .filter(c => c.state === 'LISTENING')
    .map(c => ({
      ...c,
      is_suspicious: [4444, 5555, 6666, 7777].includes(c.local_port),
      description: [4444, 5555].includes(c.local_port) ? '高危端口，可能为后门' : ''
    }))
})

function getStateTagType(state: string): string {
  switch (state) {
    case 'ESTABLISHED': return 'success'
    case 'LISTENING': return 'warning'
    case 'TIME_WAIT':
    case 'CLOSE_WAIT': return 'info'
    default: return 'info'
  }
}

function handleRowClick(row: NetworkConnection) {
  selectedConnection.value = row
  detailDialogVisible.value = true
}

function handleBlockIP() {
  if (selectedConnection.value) {
    ElMessage.success(`IP ${selectedConnection.value.remote_addr} 已封禁`)
    detailDialogVisible.value = false
  }
}

function handleKillProcess() {
  if (selectedConnection.value) {
    ElMessage.warning(`结束进程 PID: ${selectedConnection.value.pid}`)
    detailDialogVisible.value = false
  }
}

function handleSearch() {}

function handleRefresh() {
  loading.value = true
  setTimeout(() => {
    loading.value = false
    ElMessage.success('刷新成功')
  }, 500)
}
</script>

<style scoped>
.module-view { height: 100%; }

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-info h2 { margin: 0 0 5px 0; font-size: 20px; }
.description { margin: 0; color: #909399; font-size: 14px; }

.header-actions { display: flex; gap: 10px; align-items: center; }

.info-cards, .feature-cards { margin-bottom: 20px; }

.info-card {
  background: #16213e;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.card-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;
}
.card-icon.total { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.listening { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.foreign { background: rgba(103, 194, 58, 0.2); color: #67c23a; }

.card-value { font-size: 24px; font-weight: 600; color: #fff; }
.card-label { font-size: 12px; color: #909399; }

.content-area { margin-top: 20px; }

.risk-ip { color: #f56c6c; font-weight: 600; }
.risk-text { color: #f56c6c; }
</style>
