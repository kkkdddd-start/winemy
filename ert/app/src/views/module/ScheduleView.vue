<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>计划任务</h2>
        <p class="description">M6 - 任务列表、异常检测</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索任务" style="width: 200px" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterStatus" placeholder="状态" style="width: 110px; margin-left: 8px;" clearable>
          <el-option label="全部" value="" />
          <el-option label="就绪" value="Ready" />
          <el-option label="运行中" value="Running" />
          <el-option label="已禁用" value="Disabled" />
        </el-select>
        <el-button type="primary" @click="handleRefresh" :loading="loading" style="margin-left: 8px;">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon total"><el-icon><List /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ taskStats.total }}</div>
              <div class="card-label">任务总数</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon anomaly"><el-icon><Warning /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ taskStats.anomaly }}</div>
              <div class="card-label">异常任务</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon running"><el-icon><VideoPlay /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ taskStats.running }}</div>
              <div class="card-label">运行中</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon disabled"><el-icon><CircleClose /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ taskStats.disabled }}</div>
              <div class="card-label">已禁用</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('task-list')">
            <div class="card-icon"><el-icon><List /></el-icon></div>
            <div class="card-content">
              <div class="card-title">任务列表</div>
              <div class="card-desc">所有计划任务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('anomaly')">
            <div class="card-icon warning"><el-icon><Warning /></el-icon></div>
            <div class="card-content">
              <div class="card-title">异常检测</div>
              <div class="card-desc">可疑任务检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExportXML">
            <div class="card-icon info"><el-icon><Document /></el-icon></div>
            <div class="card-content">
              <div class="card-title">导出 XML</div>
              <div class="card-desc">导出任务配置</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('task-delete')">
            <div class="card-icon danger"><el-icon><Delete /></el-icon></div>
            <div class="card-content">
              <div class="card-title">删除任务</div>
              <div class="card-desc">删除计划任务</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <el-table :data="filteredTaskList" v-loading="loading" stripe :row-class-name="getRowClassName" @row-click="handleRowClick">
          <el-table-column prop="name" label="任务名称" min-width="180" sortable show-overflow-tooltip />
          <el-table-column prop="state" label="状态" width="100" sortable>
            <template #default="{ row }">
              <el-tag v-if="row.state === 'Ready'" type="success" size="small">就绪</el-tag>
              <el-tag v-else-if="row.state === 'Running'" type="warning" size="small">运行中</el-tag>
              <el-tag v-else-if="row.state === 'Disabled'" type="info" size="small">已禁用</el-tag>
              <el-tag v-else size="small">{{ row.state }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="last_run_time" label="上次运行" width="160">
            <template #default="{ row }">
              <span :class="{ 'risk-text': isAnomalyTime(row.last_run_time) }">{{ row.last_run_time || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="next_run_time" label="下次运行" width="160" />
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column prop="risk_level" label="风险" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.risk_level === 3" type="danger" size="small">严重</el-tag>
              <el-tag v-else-if="row.risk_level === 2" type="danger" size="small">高</el-tag>
              <el-tag v-else-if="row.risk_level === 1" type="warning" size="small">中</el-tag>
              <el-tag v-else type="success" size="small">低</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
              <el-button type="success" size="small" @click="handleRun(row)">运行</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="任务详情" width="600px">
      <el-descriptions v-if="selectedTask" :column="2" border>
        <el-descriptions-item label="任务名称" :span="2">{{ selectedTask.name }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="selectedTask.state === 'Ready'" type="success">就绪</el-tag>
          <el-tag v-else-if="selectedTask.state === 'Running'" type="warning">运行中</el-tag>
          <el-tag v-else type="info">{{ selectedTask.state }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="风险等级">
          <el-tag v-if="selectedTask.risk_level === 3" type="danger">严重</el-tag>
          <el-tag v-else-if="selectedTask.risk_level === 2" type="danger">高</el-tag>
          <el-tag v-else-if="selectedTask.risk_level === 1" type="warning">中</el-tag>
          <el-tag v-else type="success">低</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="路径" :span="2">{{ selectedTask.path }}</el-descriptions-item>
        <el-descriptions-item label="命令">{{ selectedTask.command || '-' }}</el-descriptions-item>
        <el-descriptions-item label="触发器">{{ selectedTask.trigger || '-' }}</el-descriptions-item>
        <el-descriptions-item label="上次运行">{{ selectedTask.last_run_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="下次运行">{{ selectedTask.next_run_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ selectedTask.created_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="修改时间">{{ selectedTask.modified_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ selectedTask.description || '-' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="handleExportXMLForTask">导出 XML</el-button>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Search, List, Warning, Document, Delete, VideoPlay, CircleClose } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface ScheduledTask {
  name: string
  path: string
  command?: string
  trigger?: string
  state: string
  last_run_time?: string
  next_run_time?: string
  created_time?: string
  modified_time?: string
  description?: string
  risk_level: number
}

const loading = ref(false)
const searchKeyword = ref('')
const filterStatus = ref('')
const detailDialogVisible = ref(false)
const selectedTask = ref<ScheduledTask | null>(null)

const mockTaskList = ref<ScheduledTask[]>([
  { name: 'WindowsUpdate', path: '\\Microsoft\\Windows\\WindowsUpdate', state: 'Ready', last_run_time: '2024-03-20 03:00:00', next_run_time: '2024-03-21 03:00:00', command: 'wuauclt.exe /detectnow', trigger: '每日 03:00', risk_level: 0, description: 'Windows 自动更新检测' },
  { name: 'OfficeAutomaticUpdates', path: '\\Microsoft\\Office\\OfficeAutomaticUpdates', state: 'Ready', last_run_time: '2024-03-19 08:00:00', next_run_time: '2024-03-20 08:00:00', trigger: '每日 08:00', risk_level: 0, description: 'Office 自动更新' },
  { name: 'SuspiciousTask', path: '\\Microsoft\\Windows\\Temp', state: 'Ready', last_run_time: '2024-03-15 02:30:00', next_run_time: '2024-03-22 02:30:00', command: 'powershell -enc JABhAHIA...', trigger: '每日 02:30', risk_level: 3, description: '可疑的 PowerShell 任务 - 建议立即检查' },
  { name: 'UserLogon', path: '\\Microsoft\\Windows\\Test', state: 'Running', last_run_time: '2024-03-21 09:15:00', next_run_time: '-', command: 'C:\\temp\\logon.bat', trigger: '用户登录时', risk_level: 2, description: '用户登录时执行 - 路径可疑' },
  { name: 'SystemCleanup', path: '\\Microsoft\\Windows\\DiskCleanup', state: 'Ready', last_run_time: '2024-03-18 01:00:00', next_run_time: '2024-03-25 01:00:00', trigger: '每周一 01:00', risk_level: 0, description: '磁盘清理任务' },
  { name: 'SecurityHealth', path: '\\Microsoft\\Windows\\SecurityHealth', state: 'Ready', last_run_time: '2024-03-20 10:00:00', next_run_time: '2024-03-21 10:00:00', trigger: '每日 10:00', risk_level: 0, description: '安全中心健康检查' },
  { name: 'TempFilesCleanup', path: '\\Microsoft\\Windows\\Temp\\User_Files', state: 'Disabled', last_run_time: '2024-02-01 00:00:00', next_run_time: '-', trigger: '手动', risk_level: 1, description: '临时文件清理（已禁用）' },
])

const taskStats = computed(() => ({
  total: mockTaskList.value.length,
  anomaly: mockTaskList.value.filter(t => t.risk_level >= 2).length,
  running: mockTaskList.value.filter(t => t.state === 'Running').length,
  disabled: mockTaskList.value.filter(t => t.state === 'Disabled').length
}))

const filteredTaskList = computed(() => {
  let result = mockTaskList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(t =>
      t.name.toLowerCase().includes(keyword) ||
      t.path.toLowerCase().includes(keyword)
    )
  }
  if (filterStatus.value) {
    result = result.filter(t => t.state === filterStatus.value)
  }
  return result
})

function isAnomalyTime(time?: string): boolean {
  if (!time) return false
  const hour = parseInt(time.split(' ')[1]?.split(':')[0] || '0')
  return hour >= 0 && hour <= 5
}

function getRowClassName({ row }: { row: ScheduledTask }): string {
  if (row.risk_level === 3) return 'risk-critical-row'
  if (row.risk_level === 2) return 'risk-high-row'
  return ''
}

function handleRowClick(row: ScheduledTask) {
  selectedTask.value = row
}

function handleView(row: ScheduledTask) {
  selectedTask.value = row
  detailDialogVisible.value = true
}

async function handleRun(row: ScheduledTask) {
  try {
    await ElMessageBox.confirm(`确定要立即运行任务 "${row.name}" 吗？`, '运行任务', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' })
    ElMessage.success(`任务 ${row.name} 已触发执行`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

async function handleDelete(row: ScheduledTask) {
  try {
    await ElMessageBox.confirm(`确定要删除任务 "${row.name}" 吗？此操作不可恢复！`, '删除任务', { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'warning' })
    ElMessage.warning(`任务 ${row.name} 已删除`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleExportXML() {
  ElMessage.success('任务配置已导出为 XML')
}

function handleExportXMLForTask() {
  if (selectedTask.value) {
    ElMessage.success(`任务 ${selectedTask.value.name} 已导出为 XML`)
  }
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

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
.module-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-info h2 { margin: 0 0 5px 0; font-size: 20px; }
.description { margin: 0; color: #909399; font-size: 14px; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.info-cards, .feature-cards { margin-bottom: 20px; }
.info-card { background: #16213e; border-radius: 8px; padding: 16px; display: flex; align-items: center; gap: 12px; }
.feature-card { background: #16213e; border-radius: 8px; padding: 16px; cursor: pointer; transition: all 0.3s; display: flex; align-items: center; gap: 12px; }
.feature-card:hover { background: #1a2a4a; transform: translateY(-2px); }
.card-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 20px; background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.total { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.anomaly { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.running { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.disabled { background: rgba(144, 147, 153, 0.2); color: #909399; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-value { font-size: 24px; font-weight: 600; color: #fff; }
.card-label { font-size: 12px; color: #909399; }
.card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 4px; }
.card-desc { font-size: 12px; color: #909399; }
.content-area { margin-top: 20px; }
.risk-text { color: #f56c6c; }
:deep(.risk-critical-row) { background-color: rgba(245, 108, 108, 0.1) !important; }
:deep(.risk-high-row) { background-color: rgba(230, 162, 60, 0.1) !important; }
</style>
