<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>域控检测</h2>
        <p class="description">M19 - 域用户/组/OU/GPO、离线降级</p>
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
          <div class="feature-card" @click="handleFeature('domain-user')">
            <div class="card-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">域用户</div>
              <div class="card-desc">域用户账户</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('domain-group')">
            <div class="card-icon">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">域组</div>
              <div class="card-desc">域安全组</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('ou')">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">OU结构</div>
              <div class="card-desc">组织单位</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('gpo')">
            <div class="card-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">GPO策略</div>
              <div class="card-desc">组策略对象</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>域控信息</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="用户" value="user" />
              <el-option label="组" value="group" />
              <el-option label="计算机" value="computer" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredDomainList" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="dn" label="DN路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
          <el-table-column prop="when_changed" label="修改时间" width="160" />
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
import { Refresh, User, UserFilled, Folder, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface DomainInfo {
  type: string
  name: string
  dn: string
  description: string
  when_changed: string
}

const loading = ref(false)
const filterType = ref('')
const domainList = ref<DomainInfo[]>([])

const filteredDomainList = computed(() => {
  if (!filterType.value) return domainList.value
  return domainList.value.filter(d => d.type.toLowerCase() === filterType.value)
})

async function loadDomainList() {
  loading.value = true
  try {
    const data = await Go.GetDomainList()
    if (data) {
      domainList.value = data as DomainInfo[]
    }
  } catch (error) {
    console.error('Failed to load domain list:', error)
    ElMessage.error('加载域控列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadDomainList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: DomainInfo) {
  ElMessage.info(`查看: ${row.name}`)
}

onMounted(() => {
  loadDomainList()
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
