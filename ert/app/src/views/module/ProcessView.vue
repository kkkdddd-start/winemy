<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>进程管理</h2>
        <p class="description">进程列表、进程树、进程查杀与Dump</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索进程名称/PID" style="width: 200px" clearable @keyup.enter="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="viewMode" style="width: 120px; margin-left: 8px;">
          <el-option label="列表视图" value="list" />
          <el-option label="树形视图" value="tree" />
        </el-select>
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="showKillProcessDialog">
            <div class="card-icon danger">
              <el-icon><Close /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程查杀</div>
              <div class="card-desc">强制终止进程</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="showDumpProcessDialog">
            <div class="card-icon warning">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">进程 Dump</div>
              <div class="card-desc">导出进程内存</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card">
            <div class="card-icon info">
              <el-icon><Upload /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出列表</div>
              <div class="card-desc">导出为 CSV/JSON</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card-mini">
            <div class="stat-value">{{ processCount }}</div>
            <div class="stat-label">进程总数</div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <el-table 
          :data="filteredProcesses" 
          v-loading="loading" 
          stripe 
          @selection-change="handleSelectionChange"
          :row-class-name="getRowClassName"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column prop="pid" label="PID" width="80" sortable />
          <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <span :class="{ 'risk-process': row.risk_level > 1 }">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="user" label="用户名" width="120" show-overflow-tooltip />
          <el-table-column prop="cpu" label="CPU%" width="80" sortable>
            <template #default="{ row }">
              <span :class="{ 'high-usage': row.cpu > 50 }">{{ row.cpu }}%</span>
            </template>
          </el-table-column>
          <el-table-column prop="memory" label="内存%" width="80" sortable>
            <template #default="{ row }">
              <span :class="{ 'high-usage': row.memory > 50 }">{{ row.memory }}%</span>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column prop="risk_level" label="风险" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.risk_level === 3" type="danger" size="small">严重</el-tag>
              <el-tag v-else-if="row.risk_level === 2" type="danger" size="small">高</el-tag>
              <el-tag v-else-if="row.risk_level === 1" type="warning" size="small">中</el-tag>
              <el-tag v-else type="success" size="small">低</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="handleKill(row)" :disabled="isProtectedProcess(row.name)">
                查杀
              </el-button>
              <el-button type="warning" size="small" @click="handleDump(row)">
                Dump
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-area">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredProcesses.length"
            :page-sizes="[20, 50, 100, 200]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="killDialogVisible" title="进程查杀确认" width="450px">
      <el-alert type="error" :closable="false" show-icon>
        <template #title>
          <strong>警告：此操作不可逆！</strong>
        </template>
      </el-alert>
      <div style="margin-top: 16px;">
        <p>即将终止以下进程：</p>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="PID">{{ selectedProcess?.pid }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ selectedProcess?.name }}</el-descriptions-item>
          <el-descriptions-item label="路径">{{ selectedProcess?.path || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="killDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmKill" :loading="actionLoading">确认查杀</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dumpDialogVisible" title="进程 Dump" width="450px">
      <el-form :model="dumpForm" label-width="100px">
        <el-form-item label="进程">
          <el-input :value="`${selectedProcess?.pid} - ${selectedProcess?.name}`" disabled />
        </el-form-item>
        <el-form-item label="Dump 类型">
          <el-select v-model="dumpForm.type" style="width: 100%;">
            <el-option label="Mini Dump" value="mini" />
            <el-option label="Full Dump" value="full" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dumpDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmDump" :loading="actionLoading">开始 Dump</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dumpProgressVisible" title="Dump 进度" width="500px" :close-on-click-modal="false">
      <el-progress :percentage="dumpProgress" :status="dumpProgressStatus" :stroke-width="20" />
      <p style="margin-top: 16px; text-align: center;">{{ dumpProgressMessage }}</p>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Close, Download, Upload, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface ProcessInfo {
  pid: number
  name: string
  path?: string
  user?: string
  cpu: number
  memory: number
  ppid?: number
  risk_level: number
}

const loading = ref(false)
const actionLoading = ref(false)
const searchKeyword = ref('')
const viewMode = ref('list')
const currentPage = ref(1)
const pageSize = ref(20)

const killDialogVisible = ref(false)
const dumpDialogVisible = ref(false)
const dumpProgressVisible = ref(false)
const dumpProgress = ref(0)
const dumpProgressMessage = ref('')
const dumpProgressStatus = ref('')

const selectedProcess = ref<ProcessInfo | null>(null)
const dumpForm = ref({ type: 'mini' })

const selectedProcesses = ref<ProcessInfo[]>([])

const processCount = computed(() => processes.value.length)

const processes = ref<ProcessInfo[]>([])

const filteredProcesses = computed(() => {
  let result = processes.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(p => 
      p.name.toLowerCase().includes(keyword) || 
      String(p.pid).includes(keyword) ||
      (p.path && p.path.toLowerCase().includes(keyword))
    )
  }
  return result
})

const protectedProcesses = ['System', 'lsass.exe', 'winlogon.exe', 'csrss.exe', 'smss.exe', 'services.exe', 'wininit.exe']

function isProtectedProcess(name: string): boolean {
  return protectedProcesses.some(p => p.toLowerCase() === name.toLowerCase())
}

function getRowClassName({ row }: { row: ProcessInfo }): string {
  if (row.risk_level === 3) return 'risk-critical-row'
  if (row.risk_level === 2) return 'risk-high-row'
  return ''
}

function showKillProcessDialog() {
  if (selectedProcesses.value.length === 1) {
    selectedProcess.value = selectedProcesses.value[0]
    killDialogVisible.value = true
  } else {
    ElMessage.warning('请先选择一个进程')
  }
}

function showDumpProcessDialog() {
  if (selectedProcesses.value.length === 1) {
    selectedProcess.value = selectedProcesses.value[0]
    dumpDialogVisible.value = true
  } else {
    ElMessage.warning('请先选择一个进程')
  }
}

async function handleKill(process: ProcessInfo) {
  if (isProtectedProcess(process.name)) {
    ElMessage.error(`禁止终止关键系统进程: ${process.name}`)
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要终止进程 "${process.name}" (PID: ${process.pid}) 吗？此操作不可恢复。`,
      '进程查杀确认',
      { confirmButtonText: '确认查杀', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success(`进程 ${process.name} 已终止`)
  } catch {
    ElMessage.info('已取消操作')
  }
}

async function confirmKill() {
  if (!selectedProcess.value) return
  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.KillProcess(selectedProcess.value.pid)
    ElMessage.success(`进程 ${selectedProcess.value.name} 已终止`)
    killDialogVisible.value = false
    handleRefresh()
  } catch (error) {
    ElMessage.error('进程查杀失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleDump(process: ProcessInfo) {
  selectedProcess.value = process
  dumpDialogVisible.value = true
}

async function confirmDump() {
  if (!selectedProcess.value) return
  actionLoading.value = true
  dumpProgressVisible.value = true
  dumpProgress.value = 0
  dumpProgressMessage.value = '正在创建 Dump 文件...'
  dumpProgressStatus.value = ''

  try {
    const { Go } = await import('@wailsjs/go/main/App')
    const result = await Go.DumpProcess(selectedProcess.value.pid)
    for (let i = 0; i <= 100; i += 20) {
      await new Promise(resolve => setTimeout(resolve, 100))
      dumpProgress.value = i
      dumpProgressMessage.value = `Dump 进度: ${i}%`
    }
    dumpProgressStatus.value = 'success'
    dumpProgressMessage.value = `Dump 完成！文件: ${result || 'unknown'}`
    ElMessage.success('进程 Dump 已完成')
    dumpDialogVisible.value = false
  } catch (error) {
    dumpProgressStatus.value = 'exception'
    dumpProgressMessage.value = 'Dump 失败: ' + (error instanceof Error ? error.message : String(error))
    ElMessage.error('进程 Dump 失败')
  } finally {
    actionLoading.value = false
    setTimeout(() => { dumpProgressVisible.value = false }, 1500)
  }
}

function handleSelectionChange(selection: ProcessInfo[]) {
  selectedProcesses.value = selection
}

function handleSearch() {
  currentPage.value = 1
}

async function handleRefresh() {
  loading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    const result = await Go.GetProcessList()
    if (result && Array.isArray(result)) {
      processes.value = result.map((p: any) => ({
        pid: p.pid,
        name: p.name || '',
        path: p.path || '',
        user: p.user || '',
        cpu: p.cpu || 0,
        memory: p.memory || 0,
        ppid: p.ppid || 0,
        risk_level: p.risk_level || 0
      }))
    }
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to load processes:', error)
    ElMessage.error('刷新失败')
  } finally {
    loading.value = false
  }
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
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

.feature-card, .info-card-mini {
  background: #16213e;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.feature-card:hover { background: #1a2a4a; transform: translateY(-2px); }

.card-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;
}
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }

.card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 4px; }
.card-desc { font-size: 12px; color: #909399; }

.info-card-mini { text-align: center; }
.stat-value { font-size: 28px; font-weight: bold; color: #409eff; }
.stat-label { font-size: 12px; color: #909399; margin-top: 4px; }

.content-area { margin-top: 20px; }

.pagination-area { margin-top: 16px; display: flex; justify-content: flex-end; }

.risk-process { color: #f56c6c; font-weight: 600; }
.high-usage { color: #e6a23c; }

:deep(.risk-critical-row) { background-color: rgba(245, 108, 108, 0.1) !important; }
:deep(.risk-high-row) { background-color: rgba(230, 162, 60, 0.1) !important; }
</style>
