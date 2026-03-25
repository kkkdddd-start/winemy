<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>计划任务</h2>
        <p class="description">M6 - 任务列表、异常检测</p>
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
          <div class="feature-card" @click="handleFeature('task-list')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">任务列表</div>
              <div class="card-desc">所有计划任务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('anomaly')">
            <div class="card-icon warning">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">异常检测</div>
              <div class="card-desc">可疑任务检测</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('task-create')">
            <div class="card-icon">
              <el-icon><Plus /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">创建任务</div>
              <div class="card-desc">新建计划任务</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('task-delete')">
            <div class="card-icon danger">
              <el-icon><Delete /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">删除任务</div>
              <div class="card-desc">删除计划任务</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>计划任务列表</span>
            <el-select v-model="filterStatus" placeholder="筛选状态" style="width: 150px" clearable>
              <el-option label="全部" value="" />
              <el-option label="正常" value="normal" />
              <el-option label="异常" value="anomaly" />
            </el-select>
          </div>
        </template>
        <el-table :data="filteredTaskList" v-loading="loading" stripe>
          <el-table-column prop="name" label="任务名称" min-width="150" />
          <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
          <el-table-column prop="trigger" label="触发器" width="120" />
          <el-table-column prop="next_run" label="下次运行" width="160" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'anomaly' ? 'danger' : 'success'">{{ row.status === 'anomaly' ? '异常' : '正常' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, List, Warning, Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface ScheduledTask {
  name: string
  path: string
  trigger: string
  next_run: string
  status: string
}

const loading = ref(false)
const filterStatus = ref('')
const taskList = ref<ScheduledTask[]>([])

const filteredTaskList = computed(() => {
  if (!filterStatus.value) return taskList.value
  return taskList.value.filter(t => t.status === filterStatus.value)
})

async function loadTaskList() {
  loading.value = true
  try {
    const data = await Go.GetScheduledTaskList()
    if (data) {
      taskList.value = data as ScheduledTask[]
    }
  } catch (error) {
    console.error('Failed to load scheduled task list:', error)
    ElMessage.error('加载计划任务列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadTaskList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: ScheduledTask) {
  ElMessage.info(`查看任务: ${row.name}`)
}

function handleDelete(row: ScheduledTask) {
  ElMessage.warning(`删除任务: ${row.name}`)
}

onMounted(() => {
  loadTaskList()
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

.card-icon.danger {
  background: rgba(245, 108, 108, 0.2);
  color: #f56c6c;
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
