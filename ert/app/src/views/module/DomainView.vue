<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>域控检测</h2>
        <p class="description">M19 - 域用户/组/OU/GPO、离线降级</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索名称/DN路径" style="width: 200px" clearable @keyup.enter="handleSearch">
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
            <div class="card-icon info">
              <el-icon><User /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.users }}</div>
              <div class="card-label">域用户</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon success">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.groups }}</div>
              <div class="card-label">域组</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon warning">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ stats.computers }}</div>
              <div class="card-label">计算机</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="info-card">
            <div class="card-icon">
              <el-icon><InfoFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-value">{{ domainList.length }}</div>
              <div class="card-label">总计</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('user')">
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
          <div class="feature-card" @click="handleFeature('group')">
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
            <span>域控信息列表</span>
            <div class="header-operations">
              <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
                <el-option label="全部" value="" />
                <el-option label="用户" value="user" />
                <el-option label="组" value="group" />
                <el-option label="计算机" value="computer" />
              </el-select>
            </div>
          </div>
        </template>
        <el-table :data="paginatedData" v-loading="loading" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column prop="type" label="类型" width="100" sortable>
            <template #default="{ row }">
              <el-tag>{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip sortable />
          <el-table-column prop="dn" label="DN路径" min-width="250" show-overflow-tooltip sortable />
          <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip sortable />
          <el-table-column prop="when_changed" label="修改时间" width="160" sortable />
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
            :total="filteredDomainList.length"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailDialogVisible" title="域控对象详情" width="600px">
      <el-descriptions :column="2" border v-if="selectedItem">
        <el-descriptions-item label="名称">{{ selectedItem.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ selectedItem.type }}</el-descriptions-item>
        <el-descriptions-item label="DN路径" :span="2">{{ selectedItem.dn }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ selectedItem.description || '无' }}</el-descriptions-item>
        <el-descriptions-item label="最后修改" :span="2">{{ selectedItem.when_changed }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, User, UserFilled, Folder, Setting, Search, Download, InfoFilled } from '@element-plus/icons-vue'
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
const searchKeyword = ref('')
const filterType = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const domainList = ref<DomainInfo[]>([])
const selectedItem = ref<DomainInfo | null>(null)
const detailDialogVisible = ref(false)
const selectedItems = ref<DomainInfo[]>([])

const stats = computed(() => {
  const list = domainList.value
  return {
    users: list.filter(d => d.type === '用户').length,
    groups: list.filter(d => d.type === '组').length,
    computers: list.filter(d => d.type === '计算机').length
  }
})

const filteredDomainList = computed(() => {
  let result = domainList.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(d =>
      d.name.toLowerCase().includes(keyword) ||
      d.dn.toLowerCase().includes(keyword) ||
      d.description.toLowerCase().includes(keyword)
    )
  }
  if (filterType.value) {
    result = result.filter(d => d.type.toLowerCase() === filterType.value)
  }
  return result
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredDomainList.value.slice(start, end)
})

async function loadDomainList() {
  loading.value = true
  try {
    const data = await Go.GetDomainInfo()
    if (data) {
      domainList.value = Array.isArray(data) ? data as DomainInfo[] : [data as DomainInfo]
    }
  } catch (error) {
    console.error('Failed to load domain list:', error)
    ElMessage.error('加载域控列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
}

function handleRefresh() {
  loadDomainList()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
}

function handlePageChange(page: number) {
  currentPage.value = page
}

function handleSelectionChange(selection: DomainInfo[]) {
  selectedItems.value = selection
}

function handleFeature(feature: string) {
  filterType.value = feature
}

function handleView(row: DomainInfo) {
  selectedItem.value = row
  detailDialogVisible.value = true
}

function handleExport() {
  ElMessage.info('正在导出数据...')
}

onMounted(() => {
  loadDomainList()
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
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }

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
</style>
