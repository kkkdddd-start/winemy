<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>系统补丁</h2>
        <p class="description">M8 - 已安装补丁、缺失补丁</p>
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
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('installed')">
            <div class="card-icon">
              <el-icon><SuccessFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">已安装补丁</div>
              <div class="card-desc">系统已安装补丁列表</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('missing')">
            <div class="card-icon warning">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">缺失补丁</div>
              <div class="card-desc">建议安装的补丁</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('export')">
            <div class="card-icon">
              <el-icon><Download /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">导出报告</div>
              <div class="card-desc">导出补丁分析报告</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>已安装补丁列表</span>
            <el-input v-model="searchKeyword" placeholder="搜索KB编号" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredPatchList" v-loading="loading" stripe>
          <el-table-column prop="kb" label="KB编号" width="120" />
          <el-table-column prop="name" label="补丁名称" min-width="200" />
          <el-table-column prop="installed_on" label="安装日期" width="120" />
          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
          <el-table-column prop="installed_by" label="安装者" width="120" />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, SuccessFilled, WarningFilled, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface PatchInfo {
  kb: string
  name: string
  installed_on: string
  description: string
  installed_by: string
}

const loading = ref(false)
const searchKeyword = ref('')
const patchList = ref<PatchInfo[]>([])

const filteredPatchList = computed(() => {
  if (!searchKeyword.value) return patchList.value
  const keyword = searchKeyword.value.toLowerCase()
  return patchList.value.filter(p => p.kb.toLowerCase().includes(keyword))
})

async function loadPatchList() {
  loading.value = true
  try {
    const data = await Go.GetPatchList()
    if (data) {
      patchList.value = data as PatchInfo[]
    }
  } catch (error) {
    console.error('Failed to load patch list:', error)
    ElMessage.error('加载补丁列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadPatchList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

onMounted(() => {
  loadPatchList()
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
