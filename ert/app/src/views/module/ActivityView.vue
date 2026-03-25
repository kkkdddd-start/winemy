<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>活动痕迹</h2>
        <p class="description">M12 - 最近打开、USB使用、浏览器历史</p>
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
          <div class="feature-card" @click="handleFeature('recent-files')">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">最近文件</div>
              <div class="card-desc">最近打开的文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('usb')">
            <div class="card-icon">
              <el-icon><Usb /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">USB使用</div>
              <div class="card-desc">USB设备使用记录</div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="feature-card" @click="handleFeature('browser')">
            <div class="card-icon">
              <el-icon><Browser /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">浏览器历史</div>
              <div class="card-desc">浏览器访问记录</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>活动痕迹</span>
            <el-select v-model="filterType" placeholder="筛选类型" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="最近文件" value="recent" />
              <el-option label="USB" value="usb" />
              <el-option label="浏览器" value="browser" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredActivityList" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="path" label="路径" min-width="250" show-overflow-tooltip />
          <el-table-column prop="time" label="时间" width="160" />
          <el-table-column prop="detail" label="详情" min-width="150" show-overflow-tooltip />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, Clock, Usb, Browser } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ActivityInfo {
  type: string
  name: string
  path: string
  time: string
  detail: string
}

const loading = ref(false)
const filterType = ref('')
const activityList = ref<ActivityInfo[]>([])

const filteredActivityList = computed(() => {
  if (!filterType.value) return activityList.value
  return activityList.value.filter(a => a.type.toLowerCase() === filterType.value)
})

async function loadActivityList() {
  loading.value = true
  try {
    const data = await Go.GetActivityList()
    if (data) {
      activityList.value = data as ActivityInfo[]
    }
  } catch (error) {
    console.error('Failed to load activity list:', error)
    ElMessage.error('加载活动痕迹失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadActivityList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

onMounted(() => {
  loadActivityList()
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
