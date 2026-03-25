<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>编解码工具</h2>
        <p class="description">M25 - Base64/Hex/Unicode/URL/HTML</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="showHistory">
          <el-icon><Clock /></el-icon>
          历史记录
        </el-button>
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'base64' }" @click="handleEncodeSelect('base64')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Base64</div>
              <div class="card-desc">Base64编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'hex' }" @click="handleEncodeSelect('hex')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Hex</div>
              <div class="card-desc">十六进制编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'unicode' }" @click="handleEncodeSelect('unicode')">
            <div class="card-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Unicode</div>
              <div class="card-desc">Unicode转换</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'url' }" @click="handleEncodeSelect('url')">
            <div class="card-icon">
              <el-icon><Link /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">URL</div>
              <div class="card-desc">URL编码解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'html' }" @click="handleEncodeSelect('html')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">HTML</div>
              <div class="card-desc">HTML实体编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" :class="{ active: currentCodec === 'hash' }" @click="handleEncodeSelect('hash')">
            <div class="card-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Hash</div>
              <div class="card-desc">哈希计算</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>编解码工具</span>
            <div class="header-operations">
              <el-select v-model="currentCodec" placeholder="选择编解码类型" style="width: 200px" @change="handleCodecChange">
                <el-option label="Base64" value="base64" />
                <el-option label="Hex" value="hex" />
                <el-option label="Unicode" value="unicode" />
                <el-option label="URL" value="url" />
                <el-option label="HTML" value="html" />
                <el-option label="Hash (MD5/SHA1/SHA256)" value="hash" />
              </el-select>
            </div>
          </div>
        </template>
        <div class="codec-tool">
          <div class="codec-section">
            <div class="codec-label">输入</div>
            <el-input v-model="inputText" type="textarea" :rows="6" :placeholder="getInputPlaceholder()" />
          </div>
          <div class="codec-actions">
            <el-button type="primary" @click="handleEncode" :loading="actionLoading">
              <el-icon><Key /></el-icon>
              编码/加密
            </el-button>
            <el-button @click="handleDecode" :loading="actionLoading">
              <el-icon><Lock /></el-icon>
              解码/解密
            </el-button>
            <el-button type="info" @click="handleClear">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
            <el-button type="success" @click="handleCopy" :disabled="!outputText">
              <el-icon><DocumentCopy /></el-icon>
              复制结果
            </el-button>
          </div>
          <div class="codec-section">
            <div class="codec-label">
              输出
              <el-tag size="small" style="margin-left: 8px;">{{ currentCodec.toUpperCase() }}</el-tag>
            </div>
            <el-input v-model="outputText" type="textarea" :rows="6" readonly placeholder="结果将显示在这里" />
          </div>
        </div>
      </el-card>
    </div>

    <el-dialog v-model="historyDialogVisible" title="历史记录" width="800px">
      <el-table :data="paginatedHistory" v-loading="loading" stripe>
        <el-table-column prop="time" label="时间" width="160" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operation" label="操作" width="80">
          <template #default="{ row }">
            <el-tag :type="row.operation === 'encode' ? 'primary' : 'success'">
              {{ row.operation === 'encode' ? '编码' : '解码' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="input" label="输入" min-width="150" show-overflow-tooltip />
        <el-table-column prop="output" label="输出" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleRestore(row)">恢复</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-area">
        <el-pagination
          v-model:current-page="historyPage"
          v-model:page-size="historyPageSize"
          :total="historyList.length"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleHistoryPageSizeChange"
          @current-change="handleHistoryPageChange"
        />
      </div>
      <template #footer>
        <el-button @click="handleClearHistory">清空历史</el-button>
        <el-button @click="historyDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Key, List, ChatDotRound, Link, Document, Lock, Clock, Delete, DocumentCopy } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface HistoryItem {
  time: string
  type: string
  operation: string
  input: string
  output: string
}

const loading = ref(false)
const actionLoading = ref(false)
const currentCodec = ref('base64')
const inputText = ref('')
const outputText = ref('')
const historyDialogVisible = ref(false)
const historyPage = ref(1)
const historyPageSize = ref(10)
const historyList = ref<HistoryItem[]>([])

const paginatedHistory = computed(() => {
  const start = (historyPage.value - 1) * historyPageSize.value
  const end = start + historyPageSize.value
  return historyList.value.slice(start, end)
})

function getInputPlaceholder(): string {
  switch (currentCodec.value) {
    case 'base64': return '请输入需要 Base64 编码或解码的文本'
    case 'hex': return '请输入需要 Hex 编码或解码的文本'
    case 'unicode': return '请输入需要 Unicode 转换的文本'
    case 'url': return '请输入需要 URL 编码或解码的文本'
    case 'html': return '请输入需要 HTML 实体编码或解码的文本'
    case 'hash': return '请输入需要计算哈希值的文本'
    default: return '请输入文本'
  }
}

function handleCodecChange() {
  inputText.value = ''
  outputText.value = ''
}

function handleEncodeSelect(codec: string) {
  currentCodec.value = codec
  inputText.value = ''
  outputText.value = ''
}

function handleEncode() {
  if (!inputText.value) {
    ElMessage.warning('请输入需要编码的文本')
    return
  }
  actionLoading.value = true
  try {
    let result = ''
    switch (currentCodec.value) {
      case 'base64':
        result = btoa(unescape(encodeURIComponent(inputText.value)))
        break
      case 'hex':
        result = inputText.value.split('').map(c => c.charCodeAt(0).toString(16).padStart(2, '0')).join('')
        break
      case 'unicode':
        result = inputText.value.split('').map(c => '\\u' + c.charCodeAt(0).toString(16).padStart(4, '0')).join('')
        break
      case 'url':
        result = encodeURIComponent(inputText.value)
        break
      case 'html':
        result = inputText.value.replace(/[<>&"']/g, c => ({ '<': '&lt;', '>': '&gt;', '&': '&amp;', '"': '&quot;', "'": '&#39;' }[c]))
        break
      case 'hash':
        result = 'hash_' + inputText.value.length
        ElMessage.info('Hash 功能需要后端支持')
        break
    }
    outputText.value = result
    addHistory('encode', result)
    ElMessage.success('编码成功')
  } catch (error) {
    ElMessage.error('编码失败')
  } finally {
    actionLoading.value = false
  }
}

function handleDecode() {
  if (!inputText.value) {
    ElMessage.warning('请输入需要解码的文本')
    return
  }
  actionLoading.value = true
  try {
    let result = ''
    switch (currentCodec.value) {
      case 'base64':
        result = decodeURIComponent(escape(atob(inputText.value)))
        break
      case 'hex':
        result = inputText.value.replace(/../g, d => String.fromCharCode(parseInt(d, 16)))
        break
      case 'unicode':
        result = inputText.value.replace(/\\u([0-9a-fA-F]{4})/g, (_, p) => String.fromCharCode(parseInt(p, 16)))
        break
      case 'url':
        result = decodeURIComponent(inputText.value)
        break
      case 'html':
        result = inputText.value.replace(/&lt;|&gt;|&amp;|&quot;|&#39;/g, c => ({ '&lt;': '<', '&gt;': '>', '&amp;': '&', '&quot;': '"', '&#39;': "'" }[c]))
        break
      default:
        ElMessage.warning('此类型不支持解码')
        return
    }
    outputText.value = result
    addHistory('decode', result)
    ElMessage.success('解码成功')
  } catch (error) {
    ElMessage.error('解码失败，请检查输入格式')
  } finally {
    actionLoading.value = false
  }
}

function handleClear() {
  inputText.value = ''
  outputText.value = ''
}

async function handleCopy() {
  if (!outputText.value) {
    ElMessage.warning('没有可复制的内容')
    return
  }
  try {
    await navigator.clipboard.writeText(outputText.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

function addHistory(operation: string, output: string) {
  historyList.value.unshift({
    time: new Date().toLocaleString('zh-CN'),
    type: currentCodec.value,
    operation,
    input: inputText.value,
    output
  })
}

function showHistory() {
  historyDialogVisible.value = true
}

function handleRestore(row: HistoryItem) {
  currentCodec.value = row.type
  inputText.value = row.input
  outputText.value = row.output
  historyDialogVisible.value = false
  ElMessage.success('已恢复记录')
}

function handleHistoryPageSizeChange(size: number) {
  historyPageSize.value = size
  historyPage.value = 1
}

function handleHistoryPageChange(page: number) {
  historyPage.value = page
}

async function handleClearHistory() {
  try {
    await ElMessageBox.confirm('确定要清空所有历史记录吗？', '清空确认', { type: 'warning' })
    historyList.value = []
    ElMessage.success('历史记录已清空')
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleRefresh() {
  ElMessage.success('工具已刷新')
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

.feature-cards { margin-bottom: 20px; }

.feature-card {
  background: #16213e;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 12px;
}

.feature-card:hover { background: #1a2a4a; transform: translateY(-2px); }
.feature-card.active { background: #1a2a4a; border: 2px solid #409eff; }

.card-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px;
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 4px; }
.card-desc { font-size: 12px; color: #909399; }

.content-area { margin-top: 20px; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-operations { display: flex; gap: 10px; }

.codec-tool { padding: 10px 0; }

.codec-section { margin-bottom: 20px; }

.codec-label {
  margin-bottom: 10px;
  font-size: 14px;
  color: #909399;
  display: flex;
  align-items: center;
}

.codec-actions { display: flex; gap: 10px; margin-bottom: 20px; }

.pagination-area { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
