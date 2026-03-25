<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>服务管理</h2>
        <p class="description">M5 - 服务列表、启停操作</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索服务" style="width: 200px" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterStatus" placeholder="状态" style="width: 110px; margin-left: 8px;" clearable>
          <el-option label="全部" value="" />
          <el-option label="运行中" value="Running" />
          <el-option label="已停止" value="Stopped" />
        </el-select>
        <el-select v-model="filterStartType" placeholder="启动类型" style="width: 120px; margin-left: 8px;" clearable>
          <el-option label="全部" value="" />
          <el-option label="自动" value="Auto" />
          <el-option label="手动" value="Manual" />
          <el-option label="禁用" value="Disabled" />
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
            <div class="card-icon total"><el-icon><Monitor /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ serviceStats.total }}</div>
              <div class="card-label">服务总数</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon running"><el-icon><VideoPlay /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ serviceStats.running }}</div>
              <div class="card-label">运行中</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon stopped"><el-icon><VideoPause /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ serviceStats.stopped }}</div>
              <div class="card-label">已停止</div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="8" :md="6">
          <div class="info-card">
            <div class="card-icon risk"><el-icon><Warning /></el-icon></div>
            <div class="card-content">
              <div class="card-value">{{ serviceStats.risk }}</div>
              <div class="card-label">风险服务</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-list')">
            <div class="card-icon"><el-icon><List /></el-icon></div>
            <div class="card-content">
              <div class="card-title">服务列表</div>
              <div class="card-desc">所有系统服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-start')">
            <div class="card-icon success"><el-icon><VideoPlay /></el-icon></div>
            <div class="card-content">
              <div class="card-title">启动服务</div>
              <div class="card-desc">启动停止的服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-stop')">
            <div class="card-icon danger"><el-icon><VideoPause /></el-icon></div>
            <div class="card-content">
              <div class="card-title">停止服务</div>
              <div class="card-desc">停止运行中的服务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('service-config')">
            <div class="card-icon"><el-icon><Setting /></el-icon></div>
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
        <el-table :data="filteredServiceList" v-loading="loading" stripe :row-class-name="getRowClassName" @row-click="handleRowClick">
          <el-table-column prop="name" label="服务名" min-width="150" sortable />
          <el-table-column prop="display_name" label="显示名称" min-width="180" sortable show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100" sortable>
            <template #default="{ row }">
              <el-tag :type="row.status === 'Running' ? 'success' : 'info'" size="small">
                {{ row.status === 'Running' ? '运行中' : '已停止' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="start_type" label="启动类型" width="100" sortable>
            <template #default="{ row }">
              <el-tag v-if="row.start_type === 'Auto'" type="primary" size="small">自动</el-tag>
              <el-tag v-else-if="row.start_type === 'Manual'" type="info" size="small">手动</el-tag>
              <el-tag v-else-if="row.start_type === 'Disabled'" type="danger" size="small">禁用</el-tag>
              <el-tag v-else size="small">{{ row.start_type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="risk_level" label="风险" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.risk_level === 3" type="danger" size="small">严重</el-tag>
              <el-tag v-else-if="row.risk_level === 2" type="danger" size="small">高</el-tag>
              <el-tag v-else-if="row.risk_level === 1" type="warning" size="small">中</el-tag>
              <el-tag v-else type="success" size="small">低</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status !== 'Running'" type="success" size="small" @click.stop="handleStart(row)">启动</el-button>
              <el-button v-if="row.status === 'Running'" type="warning" size="small" @click.stop="handleStop(row)">停止</el-button>
              <el-button type="primary" size="small" @click.stop="handleRestart(row)" :disabled="row.status !== 'Running'">重启</el-button>
              <el-button type="info" size="small" @click.stop="handleConfig(row)">配置</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="configDialogVisible" title="服务配置" width="500px">
      <el-descriptions v-if="selectedService" :column="1" border>
        <el-descriptions-item label="服务名">{{ selectedService.name }}</el-descriptions-item>
        <el-descriptions-item label="显示名称">{{ selectedService.display_name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ selectedService.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="路径">{{ selectedService.path }}</el-descriptions-item>
      </el-descriptions>
      <el-form style="margin-top: 20px;">
        <el-form-item label="启动类型">
          <el-select v-model="configForm.startType" style="width: 100%;">
            <el-option label="自动" value="Auto" />
            <el-option label="手动" value="Manual" />
            <el-option label="禁用" value="Disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmConfig">保存配置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Search, List, VideoPlay, VideoPause, Setting, Monitor, Warning } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface ServiceInfo {
  name: string
  display_name: string
  status: string
  start_type: string
  path: string
  description?: string
  risk_level: number
}

const loading = ref(false)
const searchKeyword = ref('')
const filterStatus = ref('')
const filterStartType = ref('')
const configDialogVisible = ref(false)
const selectedService = ref<ServiceInfo | null>(null)
const configForm = ref({ startType: 'Auto' })

const mockServiceList = ref<ServiceInfo[]>([])

const serviceStats = computed(() => ({
  total: services.value.length,
  running: services.value.filter(s => s.status === 'Running').length,
  stopped: services.value.filter(s => s.status !== 'Running').length,
  risk: services.value.filter(s => s.risk_level >= 2).length
}))

const services = ref<ServiceInfo[]>([])

const filteredServiceList = computed(() => {
  let result = services.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(s =>
      s.name.toLowerCase().includes(keyword) ||
      s.display_name.toLowerCase().includes(keyword)
    )
  }
  if (filterStatus.value) {
    result = result.filter(s => s.status === filterStatus.value)
  }
  if (filterStartType.value) {
    result = result.filter(s => s.start_type === filterStartType.value)
  }
  return result
})

function getRowClassName({ row }: { row: ServiceInfo }): string {
  if (row.risk_level === 3) return 'risk-critical-row'
  if (row.risk_level === 2) return 'risk-high-row'
  return ''
}

function handleRowClick(row: ServiceInfo) {
  selectedService.value = row
}

async function handleStart(row: ServiceInfo) {
  try {
    await ElMessageBox.confirm(`确定要启动服务 "${row.display_name}" 吗？`, '启动服务', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' })
    ElMessage.success(`服务 ${row.display_name} 已启动`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

async function handleStop(row: ServiceInfo) {
  try {
    await ElMessageBox.confirm(`确定要停止服务 "${row.display_name}" 吗？`, '停止服务', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    ElMessage.warning(`服务 ${row.display_name} 已停止`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

async function handleRestart(row: ServiceInfo) {
  try {
    await ElMessageBox.confirm(`确定要重启服务 "${row.display_name}" 吗？`, '重启服务', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    ElMessage.success(`服务 ${row.display_name} 正在重启...`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleConfig(row: ServiceInfo) {
  selectedService.value = row
  configForm.value.startType = row.start_type
  configDialogVisible.value = true
}

function confirmConfig() {
  if (selectedService.value) {
    ElMessage.success(`服务 ${selectedService.value.display_name} 配置已保存`)
    configDialogVisible.value = false
  }
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleRefresh() {
  loading.value = true
  const { Go } = await import('@wailsjs/go/main/App')
  Go.GetServices().then((result: any) => {
    if (result && Array.isArray(result)) {
      services.value = result.map((s: any) => ({
        name: s.name || '',
        display_name: s.display_name || '',
        status: s.status || 'Unknown',
        start_type: s.start_type || 'Unknown',
        path: s.path || '',
        description: s.description || '',
        risk_level: s.risk_level || 0
      }))
    }
    ElMessage.success('刷新成功')
  }).catch((error: any) => {
    console.error('Failed to load services:', error)
    ElMessage.error('刷新失败')
  }).finally(() => {
    loading.value = false
  })
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
.card-icon.running { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.stopped { background: rgba(144, 147, 153, 0.2); color: #909399; }
.card-icon.risk { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-value { font-size: 24px; font-weight: 600; color: #fff; }
.card-label { font-size: 12px; color: #909399; }
.card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 4px; }
.card-desc { font-size: 12px; color: #909399; }
.content-area { margin-top: 20px; }
:deep(.risk-critical-row) { background-color: rgba(245, 108, 108, 0.1) !important; }
:deep(.risk-high-row) { background-color: rgba(230, 162, 60, 0.1) !important; }
</style>
