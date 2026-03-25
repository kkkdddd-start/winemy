<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>域内渗透检测</h2>
        <p class="description">M20 - Kerberoasting、Golden Ticket</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索攻击技术/目标" style="width: 200px" clearable @keyup.enter="handleSearch">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="success" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.high }}</div>
              <div class="card-label">高危</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.medium }}</div>
              <div class="card-label">中危</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.low }}</div>
              <div class="card-label">低危</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ resultList.length }}</div>
              <div class="card-label">总计</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('kerberoasting')">
            <div class="card-icon danger">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Kerberoasting</div>
              <div class="card-desc">SPN账户检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('golden-ticket')">
            <div class="card-icon danger">
              <el-icon><Ticket /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Golden Ticket</div>
              <div class="card-desc">票据伪造检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('silver-ticket')">
            <div class="card-icon warning">
              <el-icon><Ticket /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Silver Ticket</div>
              <div class="card-desc">服务票据检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('pass-the-hash')">
            <div class="card-icon warning">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Pass The Hash</div>
              <div class="card-desc">哈希传递检测</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>渗透检测结果</span>
            <div class="header-operations">
              <el-select v-model="filterType" placeholder="筛选风险等级" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="高危" value="high" />
                <el-option label="中危" value="medium" />
                <el-option label="低危" value="low" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="technique" label="攻击技术" width="150" sortable show-overflow-tooltip />
          <el-table-column prop="target" label="目标" min-width="150" sortable show-overflow-tooltip />
          <el-table-column prop="user" label="涉及用户" width="120" sortable show-overflow-tooltip />
          <el-table-column prop="risk" label="风险等级" width="100" sortable>
            <template #default="{ row }">
              <RiskTag :risk-level="getRiskLevel(row.risk)" />
            </template>
          </el-table-column>
          <el-table-column prop="time" label="检测时间" width="160" sortable />
          <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-area">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="filteredResultList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="攻击详情" width="650px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="攻击技术">{{ selectedItem.technique }}</el-descriptions-item>
        <el-descriptions-item label="风险等级">
          <RiskTag :risk-level="getRiskLevel(selectedItem.risk)" />
        </el-descriptions-item>
        <el-descriptions-item label="目标" :span="2">{{ selectedItem.target }}</el-descriptions-item>
        <el-descriptions-item label="涉及用户" :span="2">{{ selectedItem.user }}</el-descriptions-item>
        <el-descriptions-item label="检测时间" :span="2">{{ selectedItem.time }}</el-descriptions-item>
        <el-descriptions-item label="详细描述" :span="2">{{ selectedItem.detail || '无' }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="selectedItem?.evidence" class="evidence-section">
        <h4>证据信息</h4>
        <el-input type="textarea" :rows="4" :value="selectedItem.evidence" readonly />
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="warning" @click="handleBlock">阻断</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Key, Ticket, Lock, Search, Download, Warning, SuccessFilled, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import RiskTag from '@/components/RiskTag/RiskTag.vue'
import { Go } from '@wailsjs/go/main/App'

interface HackResult {
  technique: string
  target: string
  user: string
  risk: string
  time: string
  detail: string
  evidence?: string
}

const loading = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const resultList = ref<HackResult[]>([])
const selectedItem = ref<HackResult | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<HackResult[]>([])

const stats = computed(() => {
  const list = resultList.value
  return {
    high: list.filter(r => r.risk === '高危').length,
    medium: list.filter(r => r.risk === '中危').length,
    low: list.filter(r => r.risk === '低危').length
  }
})

const filteredResultList = computed(() => {
  let result = resultList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(r =>
      r.technique.toLowerCase().includes(keyword) ||
      r.target.toLowerCase().includes(keyword) ||
      r.user.toLowerCase().includes(keyword)
    )
  }
  if (filterType.value) {
    result = result.filter(r => r.risk.toLowerCase() === filterType.value)
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredResultList.value.slice(start, end)
})

function getRiskLevel(risk: string): number {
  switch (risk) {
    case '高危': return 2
    case '中危': return 1
    case '低危': return 0
    default: return 0
  }
}

async function loadResultList() {
  loading.value = true
  try {
    const data = await Go.GetDomainHackList()
    if (data) {
      resultList.value = data as HackResult[]
    }
  } catch (error) {
    console.error('Failed to load hack result list:', error)
    ElMessage.error('加载检测结果失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadResultList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: HackResult[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  filterType.value = feature === 'kerberoasting' ? 'high' :
                     feature === 'golden-ticket' ? 'high' :
                     feature === 'silver-ticket' ? 'medium' :
                     feature === 'pass-the-hash' ? 'medium' : ''
}

function handleView(row: HackResult) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

async function handleBlock() {
  if (!selectedItem.value) return
  try {
    await ElMessageBox.confirm(
      `确定要阻断此攻击吗？`,
      '阻断确认',
      { confirmButtonText: '确认阻断', cancelButtonText: '取消', type: 'warning' }
    )
    ElMessage.success('已发起阻断请求')
    detailDialogVisible.value = false
  } catch {
    ElMessage.info('已取消操作')
  }
}

function handleExport() {
  ElMessage.info('正在导出数据...')
}

onMounted(() => {
  loadResultList()
})
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
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 15px;
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

.feature-card:hover { background: #1a2a4a; transform: translateY(-2px); }

.card-icon {
  width: 50px; height: 50px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-size: 24px;
}
.card-icon.danger { background: rgba(245, 108, 108, 0.2); color: #f56c6c; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }

.card-value { font-size: 24px; font-weight: 600; color: #fff; }
.card-label { font-size: 12px; color: #909399; }

.card-title { font-size: 16px; font-weight: 600; color: #fff; margin-bottom: 5px; }
.card-desc { font-size: 12px; color: #909399; }

.content-area { margin-top: 20px; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-operations { display: flex; gap: 10px; }

.pagination-area { margin-top: 16px; display: flex; justify-content: flex-end; }

.evidence-section { margin-top: 16px; }
.evidence-section h4 { margin-bottom: 8px; color: #fff; }
</style>
