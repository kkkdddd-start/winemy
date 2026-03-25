<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>注册表分析</h2>
        <p class="description">M4 - 关键项检测、持久化、自启动</p>
      </div>
      <div class="header-actions">
        <el-input v-model="searchKeyword" placeholder="搜索注册表项" style="width: 200px" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterType" placeholder="筛选类型" style="width: 130px; margin-left: 8px;" clearable>
          <el-option label="全部" value="" />
          <el-option label="自启动" value="autostart" />
          <el-option label="持久化" value="persistence" />
          <el-option label="可疑" value="suspicious" />
        </el-select>
        <el-button type="primary" @click="handleRefresh" :loading="loading" style="margin-left: 8px;">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="feature-card" :class="{ active: activeFeature === 'key-detection' }" @click="setActiveFeature('key-detection')">
            <div class="card-icon"><el-icon><Key /></el-icon></div>
            <div class="card-content">
              <div class="card-title">关键项检测</div>
              <div class="card-desc">检测可疑注册表项</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: activeFeature === 'persistence' }" @click="setActiveFeature('persistence')">
            <div class="card-icon warning"><el-icon><Lock /></el-icon></div>
            <div class="card-content">
              <div class="card-title">持久化检测</div>
              <div class="card-desc">自启动项检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: activeFeature === 'autostart' }" @click="setActiveFeature('autostart')">
            <div class="card-icon success"><el-icon><Switch /></el-icon></div>
            <div class="card-content">
              <div class="card-title">自启动项</div>
              <div class="card-desc">Run键自启动</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" :class="{ active: activeFeature === 'reg-diff' }" @click="setActiveFeature('reg-diff')">
            <div class="card-icon info"><el-icon><Document /></el-icon></div>
            <div class="card-content">
              <div class="card-title">注册表对比</div>
              <div class="card-desc">基线对比分析</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-row :gutter="20">
        <el-col :span="activeView === 'tree' ? 8 : 24">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>注册表项 ({{ filteredRegistryList.length }})</span>
                <div class="header-actions">
                  <el-button-group>
                    <el-button :type="activeView === 'tree' ? 'primary' : 'default'" size="small" @click="activeView = 'tree'">
                      <el-icon><Operation /></el-icon>
                    </el-button>
                    <el-button :type="activeView === 'table' ? 'primary' : 'default'" size="small" @click="activeView = 'table'">
                      <el-icon><List /></el-icon>
                    </el-button>
                  </el-button-group>
                  <el-button type="success" size="small" @click="handleExport" style="margin-left: 8px;">
                    <el-icon><Download /></el-icon>
                    导出
                  </el-button>
                </div>
              </div>
            </template>

            <div v-if="activeView === 'tree'">
              <el-tree :data="treeData" :props="treeProps" node-key="path" default-expand-all @node-click="handleNodeClick">
                <template #default="{ node, data }">
                  <span class="tree-node">
                    <span class="node-label">{{ node.label }}</span>
                    <el-tag v-if="data.risk === '高危'" type="danger" size="small">{{ data.risk }}</el-tag>
                    <el-tag v-else-if="data.risk === '中危'" type="warning" size="small">{{ data.risk }}</el-tag>
                  </span>
                </template>
              </el-tree>
            </div>

            <el-table v-else :data="filteredRegistryList" v-loading="loading" stripe @row-click="handleRowClick" :row-class-name="getRowClassName">
              <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
              <el-table-column prop="name" label="名称" min-width="120" show-overflow-tooltip />
              <el-table-column prop="value_type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.value_type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="value" label="值" min-width="150" show-overflow-tooltip />
              <el-table-column prop="risk_level" label="风险" width="80">
                <template #default="{ row }">
                  <el-tag v-if="row.risk_level === 3" type="danger" size="small">严重</el-tag>
                  <el-tag v-else-if="row.risk_level === 2" type="danger" size="small">高</el-tag>
                  <el-tag v-else-if="row.risk_level === 1" type="warning" size="small">中</el-tag>
                  <el-tag v-else type="success" size="small">低</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>

        <el-col v-if="activeView === 'tree'" :span="16">
          <el-card>
            <template #header>
              <span>键值详情</span>
            </template>
            <el-descriptions v-if="selectedItem" :column="2" border>
              <el-descriptions-item label="路径">{{ selectedItem.path }}</el-descriptions-item>
              <el-descriptions-item label="名称">{{ selectedItem.name }}</el-descriptions-item>
              <el-descriptions-item label="值类型">{{ selectedItem.value_type }}</el-descriptions-item>
              <el-descriptions-item label="风险等级">
                <el-tag v-if="selectedItem.risk_level === 3" type="danger">严重</el-tag>
                <el-tag v-else-if="selectedItem.risk_level === 2" type="danger">高</el-tag>
                <el-tag v-else-if="selectedItem.risk_level === 1" type="warning">中</el-tag>
                <el-tag v-else type="success">低</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="值" :span="2">
                <el-input type="textarea" :value="selectedItem.value" readonly :rows="3" />
              </el-descriptions-item>
              <el-descriptions-item label="修改时间">{{ selectedItem.modified || '-' }}</el-descriptions-item>
              <el-descriptions-item label="描述">{{ selectedItem.description || '-' }}</el-descriptions-item>
            </el-descriptions>
            <el-empty v-else description="请选择一个注册表项查看详情" />
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Refresh, Search, Key, Lock, Switch, Document, Operation, List, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface RegistryItem {
  path: string
  name: string
  value: string
  value_type: string
  risk_level: number
  risk?: string
  modified?: string
  description?: string
  children?: RegistryItem[]
}

const loading = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const activeFeature = ref('key-detection')
const activeView = ref('table')
const selectedItem = ref<RegistryItem | null>(null)

const mockRegistryData = ref<RegistryItem[]>([])

const registryData = ref<RegistryItem[]>([])

const treeData = computed(() => {
  const root: Record<string, any> = {}
  registryData.value.forEach(item => {
    const parts = item.path.split('\\')
    let current = root
    parts.forEach((part, index) => {
      if (!current[part]) {
        current[part] = { children: {} }
      }
      if (index === parts.length - 1) {
        current[part].data = item
      }
      current = current[part]
    })
  })

  function buildTree(data: Record<string, any>, prefix = ''): any[] {
    return Object.keys(data).map(key => {
      const node = data[key]
      const fullPath = prefix ? `${prefix}\\${key}` : key
      return {
        label: key,
        path: fullPath,
        risk: node.data?.risk_level >= 2 ? '高危' : node.data?.risk_level >= 1 ? '中危' : '低危',
        risk_level: node.data?.risk_level || 0,
        children: buildTree(node.children || {}, fullPath)
      }
    })
  }

  return buildTree(root)
})

const treeProps = {
  children: 'children',
  label: 'label'
}

const filteredRegistryList = computed(() => {
  let result = registryData.value
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(r =>
      r.path.toLowerCase().includes(keyword) ||
      r.name.toLowerCase().includes(keyword) ||
      r.value.toLowerCase().includes(keyword)
    )
  }
  if (filterType.value) {
    result = result.filter(r => r.description?.toLowerCase().includes(filterType.value))
  }
  return result
})

function getRowClassName({ row }: { row: RegistryItem }): string {
  if (row.risk_level === 3) return 'risk-critical-row'
  if (row.risk_level === 2) return 'risk-high-row'
  return ''
}

function setActiveFeature(feature: string) {
  activeFeature.value = feature
}

function handleRowClick(row: RegistryItem) {
  selectedItem.value = row
}

function handleNodeClick(data: RegistryItem) {
  if (data.data) {
    selectedItem.value = data.data
  }
}

async function handleRefresh() {
  loading.value = true
  try {
    const { Go } = await import('@wailsjs/go/main/App')
    const result = await Go.GetRegistryKeys()
    if (result && Array.isArray(result)) {
      registryData.value = result.map((r: any) => ({
        path: r.path || '',
        name: r.name || '',
        value: r.value || '',
        value_type: r.value_type || 'REG_SZ',
        risk_level: r.risk_level || 0,
        modified: r.modified || '',
        description: r.description || ''
      }))
    }
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to load registry data:', error)
    ElMessage.error('刷新失败')
  } finally {
    loading.value = false
  }
}

function handleExport() {
  ElMessage.success('导出成功')
}
</script>

<style scoped>
.module-view { height: 100%; }
.module-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-info h2 { margin: 0 0 5px 0; font-size: 20px; }
.description { margin: 0; color: #909399; font-size: 14px; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.feature-cards, .feature-cards :deep(.el-row) .el-col { margin-bottom: 12px; }
.feature-card { background: #16213e; border-radius: 8px; padding: 16px; cursor: pointer; transition: all 0.3s; display: flex; align-items: center; gap: 12px; }
.feature-card:hover, .feature-card.active { background: #1a2a4a; transform: translateY(-2px); }
.card-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 20px; background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-icon.warning { background: rgba(230, 162, 60, 0.2); color: #e6a23c; }
.card-icon.success { background: rgba(103, 194, 58, 0.2); color: #67c23a; }
.card-icon.info { background: rgba(64, 158, 255, 0.2); color: #409eff; }
.card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 4px; }
.card-desc { font-size: 12px; color: #909399; }
.content-area { margin-top: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.tree-node { display: flex; align-items: center; gap: 8px; width: 100%; }
.node-label { flex: 1; }
:deep(.risk-critical-row) { background-color: rgba(245, 108, 108, 0.1) !important; }
:deep(.risk-high-row) { background-color: rgba(230, 162, 60, 0.1) !important; }
</style>
