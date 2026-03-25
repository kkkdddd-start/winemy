<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>安全基线检查</h2>
        <p class="description">M23 - 密码/账户/审核/网络安全</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="info-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon danger">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.failed }}</div>
              <div class="card-label">不合规项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.passed }}</div>
              <div class="card-label">合规项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineStats.warning }}</div>
              <div class="card-label">警告项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon info">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ baselineScore }}%</div>
              <div class="card-label">合规评分</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('password')">
            <div class="card-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">密码策略</div>
              <div class="card-desc">密码复杂度检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('account')">
            <div class="card-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">账户策略</div>
              <div class="card-desc">账户安全检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('audit')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">审核策略</div>
              <div class="card-desc">安全审核检查</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('network')">
            <div class="card-icon">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">网络安全</div>
              <div class="card-desc">网络配置检查</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>基线检查结果</span>
            <el-select v-model="filterStatus" placeholder="筛选状态" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="不合规" value="failed" />
              <el-option label="合规" value="passed" />
              <el-option label="警告" value="warning" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredBaselineList" v-loading="loading" stripe>
          <el-table-column prop="category" label="类别" width="120" />
          <el-table-column prop="item" label="检查项" min-width="200" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="current" label="当前值" width="150" />
          <el-table-column prop="expected" label="期望值" width="150" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleFix(row)">修复</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Warning, SuccessFilled, WarningFilled, InfoFilled, Lock, User, Document, Connection } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface BaselineItem {
  category: string
  item: string
  status: string
  current: string
  expected: string
  description: string
}

const loading = ref(false)
const filterStatus = ref('')
const baselineList = ref<BaselineItem[]>([])
const baselineStats = ref({
  failed: 0,
  passed: 0,
  warning: 0
})
const baselineScore = ref(0)

const filteredBaselineList = computed(() => {
  if (!filterStatus.value) return baselineList.value
  return baselineList.value.filter(b => b.status === filterStatus.value)
})

function getStatusType(status: string): string {
  const typeMap: Record<string, string> = {
    '不合规': 'danger',
    '合规': 'success',
    '警告': 'warning'
  }
  return typeMap[status] || 'info'
}

async function loadBaselineList() {
  loading.value = true
  try {
    const data = await Go.GetBaselineList()
    if (data) {
      baselineList.value = data as BaselineItem[]
      baselineStats.value = {
        failed: baselineList.value.filter(b => b.status === '不合规').length,
        passed: baselineList.value.filter(b => b.status === '合规').length,
        warning: baselineList.value.filter(b => b.status === '警告').length
      }
      const total = baselineList.value.length
      baselineScore.value = total > 0 ? Math.round((baselineStats.value.passed / total) * 100) : 0
    }
  } catch (error) {
    console.error('Failed to load baseline list:', error)
    ElMessage.error('加载基线检查失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadBaselineList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleFix(row: BaselineItem) {
  ElMessage.success(`修复检查项: ${row.item}`)
}

onMounted(() => {
  loadBaselineList()
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

.info-cards {
  margin-bottom: 20px;
}

.info-card {
  background: #16213e;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 15px;
}

.card-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
}

.card-icon.success {
  background: rgba(103, 194, 58, 0.2);
  color: #67c23a;
}

.card-icon.warning {
  background: rgba(230, 162, 60, 0.2);
  color: #e6a23c;
}

.card-icon.info {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-value {
  font-size: 20px;
  font-weight: 600;
  color: #fff;
}

.card-label {
  font-size: 12px;
  color: #909399;
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
