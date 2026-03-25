<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>应急处置</h2>
        <p class="description">M17 - 进程查杀、文件隔离、审计日志</p>
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
          <div class="feature-card" @click="showIsolateFileDialog">
            <div class="card-icon warning">
              <el-icon><FolderDelete /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件隔离</div>
              <div class="card-desc">隔离可疑文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="showBlockIPDialog">
            <div class="card-icon">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">IP封禁</div>
              <div class="card-desc">封禁恶意IP</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="showDisableServiceDialog">
            <div class="card-icon">
              <el-icon><VideoPause /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">服务禁用</div>
              <div class="card-desc">停止并禁用服务</div>
            </div>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="6">
          <div class="feature-card" @click="showRestoreRegistryDialog">
            <div class="card-icon">
              <el-icon><Edit /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">注册表修复</div>
              <div class="card-desc">恢复被篡改的注册表</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="showBackupDialog">
            <div class="card-icon">
              <el-icon><DocumentCopy /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件备份</div>
              <div class="card-desc">备份可疑文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="showRestoreFileDialog">
            <div class="card-icon">
              <el-icon><RefreshLeft /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件恢复</div>
              <div class="card-desc">从备份恢复文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleExportLog">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出审计日志</div>
              <div class="card-desc">导出操作记录</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>处置历史</span>
            <div class="header-actions">
              <el-input v-model="searchKeyword" placeholder="搜索处置记录" style="width: 200px" clearable />
              <el-button type="danger" size="small" @click="handleClearHistory">清空历史</el-button>
            </div>
          </div>
        </template>
        <el-table :data="filteredHistoryList" v-loading="loading" stripe>
          <el-table-column prop="timestamp" label="时间" width="180" />
          <el-table-column prop="type" label="操作" width="150">
            <template #default="{ row }">
              <el-tag :type="getActionType(row.type)">{{ getActionName(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="target" label="目标" min-width="200">
            <template #default="{ row }">
              <span v-if="row.pid">PID: {{ row.pid }} - {{ row.name || '未知进程' }}</span>
              <span v-else-if="row.ip">IP: {{ row.ip }}</span>
              <span v-else-if="row.service">服务: {{ row.service }}</span>
              <span v-else-if="row.path || row.original_path">{{ row.path || row.original_path }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : 'danger'">{{ row.status === 'success' ? '成功' : '失败' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="operator" label="操作人" width="120">
            <template #default>
              <span>当前用户</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="killProcessDialogVisible" title="进程查杀确认" width="500px">
      <div class="dialog-content">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            <strong>危险操作</strong> - 此操作将强制终止进程，可能导致数据丢失
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="进程 PID">
            <el-input-number v-model="killProcessForm.pid" :min="1" style="width: 100%;" />
          </el-form-item>
          <el-form-item label="进程名称">
            <el-input v-model="killProcessForm.name" placeholder="可选" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="killProcessDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="handleKillProcess" :loading="actionLoading">确认查杀</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="isolateFileDialogVisible" title="文件隔离确认" width="500px">
      <div class="dialog-content">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            <strong>文件将被移动到隔离区</strong> - 可以在之后恢复
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="文件路径" required>
            <el-input v-model="isolateFileForm.path" placeholder="请输入完整文件路径" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="isolateFileDialogVisible = false">取消</el-button>
        <el-button type="warning" @click="handleIsolateFile" :loading="actionLoading">确认隔离</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="blockIPDialogVisible" title="IP 封禁确认" width="500px">
      <div class="dialog-content">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            <strong>IP 封禁</strong> - 将在防火墙上添加规则阻止该 IP
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="IP 地址" required>
            <el-input v-model="blockIPForm.ip" placeholder="例如: 192.168.1.100" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="blockIPDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBlockIP" :loading="actionLoading">确认封禁</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="disableServiceDialogVisible" title="服务禁用确认" width="500px">
      <div class="dialog-content">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            <strong>危险操作</strong> - 禁用服务可能导致系统功能异常
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="服务名称" required>
            <el-input v-model="disableServiceForm.serviceName" placeholder="例如: wuauserv" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="disableServiceDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="handleDisableService" :loading="actionLoading">确认禁用</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="restoreRegistryDialogVisible" title="注册表修复确认" width="500px">
      <div class="dialog-content">
        <el-alert type="warning" :closable="false" show-icon>
          <template #title>
            <strong>注册表修改</strong> - 此操作将删除指定的注册表值
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="注册表路径" required>
            <el-input v-model="restoreRegistryForm.path" placeholder="例如: HKLM\Software\..." />
          </el-form-item>
          <el-form-item label="值名称" required>
            <el-input v-model="restoreRegistryForm.valueName" placeholder="要删除的值名称" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="restoreRegistryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRestoreRegistry" :loading="actionLoading">确认修复</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="backupDialogVisible" title="文件备份确认" width="500px">
      <div class="dialog-content">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            <strong>文件备份</strong> - 将创建文件的安全副本
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="文件路径" required>
            <el-input v-model="backupForm.path" placeholder="请输入完整文件路径" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="backupDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBackup" :loading="actionLoading">确认备份</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="restoreFileDialogVisible" title="文件恢复确认" width="500px">
      <div class="dialog-content">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            <strong>文件恢复</strong> - 从备份副本恢复文件
          </template>
        </el-alert>
        <el-form style="margin-top: 20px;">
          <el-form-item label="备份文件路径" required>
            <el-input v-model="restoreFileForm.backupPath" placeholder="请输入备份文件路径" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="restoreFileDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRestoreFile" :loading="actionLoading">确认恢复</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="操作详情" width="600px">
      <el-descriptions :column="2" border v-if="selectedAction">
        <el-descriptions-item label="时间">{{ selectedAction.timestamp }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">{{ getActionName(selectedAction.type) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="selectedAction.status === 'success' ? 'success' : 'danger'">
            {{ selectedAction.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作人">当前用户</el-descriptions-item>
        <el-descriptions-item label="详细信息" :span="2">
          <pre style="margin: 0; white-space: pre-wrap;">{{ formatDetails(selectedAction) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Close, FolderDelete, CircleClose, VideoPause, Edit, DocumentCopy, RefreshLeft, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface ActionInfo {
  timestamp: string
  type: string
  pid?: number
  name?: string
  ip?: string
  service?: string
  serviceName?: string
  path?: string
  original_path?: string
  quarantine_path?: string
  backup_path?: string
  restore_path?: string
  status: string
  message?: string
  error?: string
  reason?: string
}

const loading = ref(false)
const actionLoading = ref(false)
const searchKeyword = ref('')
const historyList = ref<ActionInfo[]>([])

const killProcessDialogVisible = ref(false)
const isolateFileDialogVisible = ref(false)
const blockIPDialogVisible = ref(false)
const disableServiceDialogVisible = ref(false)
const restoreRegistryDialogVisible = ref(false)
const backupDialogVisible = ref(false)
const restoreFileDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const selectedAction = ref<ActionInfo | null>(null)

const killProcessForm = ref({ pid: 0, name: '' })
const isolateFileForm = ref({ path: '' })
const blockIPForm = ref({ ip: '' })
const disableServiceForm = ref({ serviceName: '' })
const restoreRegistryForm = ref({ path: '', valueName: '' })
const backupForm = ref({ path: '' })
const restoreFileForm = ref({ backupPath: '' })

const filteredHistoryList = computed(() => {
  if (!searchKeyword.value) return historyList.value
  const keyword = searchKeyword.value.toLowerCase()
  return historyList.value.filter(h =>
    (h.path && h.path.toLowerCase().includes(keyword)) ||
    (h.name && h.name.toLowerCase().includes(keyword)) ||
    (h.ip && h.ip.includes(keyword)) ||
    (h.service && h.service.toLowerCase().includes(keyword)) ||
    (h.type && h.type.toLowerCase().includes(keyword))
  )
})

function getActionName(typeName: string): string {
  const names: Record<string, string> = {
    'kill_process': '进程查杀',
    'isolate_file': '文件隔离',
    'block_ip': 'IP封禁',
    'unblock_ip': 'IP解封',
    'disable_service': '服务禁用',
    'restore_registry': '注册表修复',
    'backup_file': '文件备份',
    'restore_file': '文件恢复',
    'disconnect_network': '网络断开',
  }
  return names[typeName] || typeName
}

function getActionType(typeName: string): string {
  const types: Record<string, string> = {
    'kill_process': 'danger',
    'isolate_file': 'warning',
    'block_ip': 'info',
    'unblock_ip': 'success',
    'disable_service': 'danger',
    'restore_registry': 'warning',
    'backup_file': 'info',
    'restore_file': 'success',
    'disconnect_network': 'danger',
  }
  return types[typeName] || 'info'
}

function formatDetails(action: ActionInfo): string {
  const lines: string[] = []
  if (action.pid) lines.push(`PID: ${action.pid}`)
  if (action.name) lines.push(`进程名: ${action.name}`)
  if (action.ip) lines.push(`IP: ${action.ip}`)
  if (action.service || action.serviceName) lines.push(`服务: ${action.service || action.serviceName}`)
  if (action.path) lines.push(`路径: ${action.path}`)
  if (action.original_path) lines.push(`原始路径: ${action.original_path}`)
  if (action.quarantine_path) lines.push(`隔离路径: ${action.quarantine_path}`)
  if (action.backup_path) lines.push(`备份路径: ${action.backup_path}`)
  if (action.restore_path) lines.push(`恢复路径: ${action.restore_path}`)
  if (action.message) lines.push(`消息: ${action.message}`)
  if (action.error) lines.push(`错误: ${action.error}`)
  if (action.reason) lines.push(`原因: ${action.reason}`)
  return lines.join('\n') || '无详细信息'
}

function showKillProcessDialog() {
  killProcessForm.value = { pid: 0, name: '' }
  killProcessDialogVisible.value = true
}

function showIsolateFileDialog() {
  isolateFileForm.value = { path: '' }
  isolateFileDialogVisible.value = true
}

function showBlockIPDialog() {
  blockIPForm.value = { ip: '' }
  blockIPDialogVisible.value = true
}

function showDisableServiceDialog() {
  disableServiceForm.value = { serviceName: '' }
  disableServiceDialogVisible.value = true
}

function showRestoreRegistryDialog() {
  restoreRegistryForm.value = { path: '', valueName: '' }
  restoreRegistryDialogVisible.value = true
}

function showBackupDialog() {
  backupForm.value = { path: '' }
  backupDialogVisible.value = true
}

function showRestoreFileDialog() {
  restoreFileForm.value = { backupPath: '' }
  restoreFileDialogVisible.value = true
}

async function handleKillProcess() {
  if (killProcessForm.value.pid <= 0) {
    ElMessage.warning('请输入有效的 PID')
    return
  }

  actionLoading.value = true
  try {
    const criticalProcs = ['system', 'lsass.exe', 'winlogon.exe', 'csrss.exe', 'smss.exe', 'services.exe', 'wininit.exe']
    if (killProcessForm.value.name && criticalProcs.includes(killProcessForm.value.name.toLowerCase())) {
      ElMessage.error('禁止终止关键系统进程')
      return
    }

    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('kill_process', String(killProcessForm.value.pid))

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'kill_process',
      pid: killProcessForm.value.pid,
      name: killProcessForm.value.name,
      status: 'success',
      message: '进程已终止'
    })

    killProcessDialogVisible.value = false
    ElMessage.success('进程已终止')
  } catch (error) {
    ElMessage.error('进程查杀失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleIsolateFile() {
  if (!isolateFileForm.value.path) {
    ElMessage.warning('请输入文件路径')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('isolate_file', isolateFileForm.value.path)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'isolate_file',
      path: isolateFileForm.value.path,
      status: 'success',
      message: '文件已隔离'
    })

    isolateFileDialogVisible.value = false
    ElMessage.success('文件已隔离到隔离区')
  } catch (error) {
    ElMessage.error('文件隔离失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleBlockIP() {
  if (!blockIPForm.value.ip) {
    ElMessage.warning('请输入 IP 地址')
    return
  }

  const ipPattern = /^(\d{1,3}\.){3}\d{1,3}$/
  if (!ipPattern.test(blockIPForm.value.ip)) {
    ElMessage.warning('请输入有效的 IP 地址')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('block_ip', blockIPForm.value.ip)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'block_ip',
      ip: blockIPForm.value.ip,
      status: 'success',
      message: 'IP 已封禁'
    })

    blockIPDialogVisible.value = false
    ElMessage.success('IP 已封禁')
  } catch (error) {
    ElMessage.error('IP 封禁失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleDisableService() {
  if (!disableServiceForm.value.serviceName) {
    ElMessage.warning('请输入服务名称')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('disable_service', disableServiceForm.value.serviceName)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'disable_service',
      service: disableServiceForm.value.serviceName,
      status: 'success',
      message: '服务已禁用'
    })

    disableServiceDialogVisible.value = false
    ElMessage.success('服务已禁用')
  } catch (error) {
    ElMessage.error('服务禁用失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleRestoreRegistry() {
  if (!restoreRegistryForm.value.path || !restoreRegistryForm.value.valueName) {
    ElMessage.warning('请输入完整的注册表路径和值名称')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('restore_registry', `${restoreRegistryForm.value.path}|${restoreRegistryForm.value.valueName}`)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'restore_registry',
      path: `${restoreRegistryForm.value.path}\\${restoreRegistryForm.value.valueName}`,
      status: 'success',
      message: '注册表已恢复'
    })

    restoreRegistryDialogVisible.value = false
    ElMessage.success('注册表已恢复')
  } catch (error) {
    ElMessage.error('注册表修复失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleBackup() {
  if (!backupForm.value.path) {
    ElMessage.warning('请输入文件路径')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('backup_file', backupForm.value.path)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'backup_file',
      path: backupForm.value.path,
      status: 'success',
      message: '文件已备份'
    })

    backupDialogVisible.value = false
    ElMessage.success('文件已备份')
  } catch (error) {
    ElMessage.error('文件备份失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

async function handleRestoreFile() {
  if (!restoreFileForm.value.backupPath) {
    ElMessage.warning('请输入备份文件路径')
    return
  }

  actionLoading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    await Go.ResponseAction('restore_file', restoreFileForm.value.backupPath)

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'restore_file',
      backup_path: restoreFileForm.value.backupPath,
      status: 'success',
      message: '文件已恢复'
    })

    restoreFileDialogVisible.value = false
    ElMessage.success('文件已恢复')
  } catch (error) {
    ElMessage.error('文件恢复失败: ' + (error instanceof Error ? error.message : String(error)))
  } finally {
    actionLoading.value = false
  }
}

  actionLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'restore_registry',
      path: `${restoreRegistryForm.value.path}\\${restoreRegistryForm.value.valueName}`,
      status: 'success',
      message: '注册表已恢复'
    })

    restoreRegistryDialogVisible.value = false
    ElMessage.success('注册表已恢复')
  } catch (error) {
    ElMessage.error('注册表修复失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleBackup() {
  if (!backupForm.value.path) {
    ElMessage.warning('请输入文件路径')
    return
  }

  actionLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'backup_file',
      path: backupForm.value.path,
      status: 'success',
      message: '文件已备份'
    })

    backupDialogVisible.value = false
    ElMessage.success('文件已备份')
  } catch (error) {
    ElMessage.error('文件备份失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleRestoreFile() {
  if (!restoreFileForm.value.backupPath) {
    ElMessage.warning('请输入备份文件路径')
    return
  }

  actionLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    historyList.value.unshift({
      timestamp: new Date().toISOString(),
      type: 'restore_file',
      backup_path: restoreFileForm.value.backupPath,
      status: 'success',
      message: '文件已恢复'
    })

    restoreFileDialogVisible = false
    ElMessage.success('文件已恢复')
  } catch (error) {
    ElMessage.error('文件恢复失败')
  } finally {
    actionLoading.value = false
  }
}

async function handleExportLog() {
  try {
    const logContent = JSON.stringify(historyList.value, null, 2)
    const blob = new Blob([logContent], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit_log_${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('审计日志已导出')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

function handleClearHistory() {
  ElMessageBox.confirm('确定要清空所有处置历史吗？此操作不可恢复。', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    historyList.value = []
    ElMessage.success('历史已清空')
  }).catch(() => {})
}

function handleView(row: ActionInfo) {
  selectedAction.value = row
  detailDialogVisible.value = true
}

function handleRefresh() {
  loadHistoryList()
}

function loadHistoryList() {
  loading.value = true
  setTimeout(() => {
    loading.value = false
  }, 500)
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

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
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

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
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

.dialog-content {
  padding: 10px 0;
}
</style>
