<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>域内渗透检测</h2>
        <p class="description">M20 - Kerberoasting、Golden Ticket</p>
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
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="高危" value="high" />
              <el-option label="中危" value="medium" />
              <el-option label="低危" value="low" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredResultList" v-loading="loading" stripe>
          <el-table-column prop="technique" label="攻击技术" width="150" />
          <el-table-column prop="target" label="目标" min-width="150" />
          <el-table-column prop="user" label="涉及用户" width="120" />
          <el-table-column prop="risk" label="风险等级" width="100">
            <template #default="{ row }">
              <el-tag :type="row.risk === '高危' ? 'danger' : row.risk === '中危' ? 'warning' : 'info'">{{ row.risk }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="time" label="检测时间" width="160" />
          <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
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
import { Refresh, Key, Ticket, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface HackResult {
  technique: string
  target: string
  user: string
  risk: string
  time: string
  detail: string
}

const loading = ref(false)
const filterType = ref('')
const resultList = ref<HackResult[]>([])

const filteredResultList = computed(() => {
  if (!filterType.value) return resultList.value
  return resultList.value.filter(r => r.risk.toLowerCase() === filterType.value)
})

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

function handleRefresh() {
  loadResultList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: HackResult) {
  ElMessage.info(`查看: ${row.technique}`)
}

onMounted(() => {
  loadResultList()
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
</style>
