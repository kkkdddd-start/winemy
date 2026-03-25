<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>文件系统</h2>
        <p class="description">M11 - 文件枚举、哈希、大文件处理</p>
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
          <div class="feature-card" @click="handleFeature('file-list')">
            <div class="card-icon">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件枚举</div>
              <div class="card-desc">目录文件浏览</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('hash')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">文件哈希</div>
              <div class="card-desc">MD5/SHA1/SHA256</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('large-file')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">大文件</div>
              <div class="card-desc">查找大文件</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="feature-card" @click="handleFeature('recent')">
            <div class="card-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">最近文件</div>
              <div class="card-desc">最近修改的文件</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>文件列表</span>
            <el-input v-model="searchKeyword" placeholder="搜索文件名称" style="width: 200px" clearable />
          </div>
        </template>
        <el-table :data="filteredFileList" v-loading="loading" stripe>
          <el-table-column prop="name" label="文件名" min-width="200" />
          <el-table-column prop="path" label="路径" min-width="300" show-overflow-tooltip />
          <el-table-column prop="size" label="大小" width="100">
            <template #default="{ row }">
              {{ formatSize(row.size) }}
            </template>
          </el-table-column>
          <el-table-column prop="modified" label="修改时间" width="160" />
          <el-table-column prop="hash" label="哈希" width="120" show-overflow-tooltip />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleView(row)">查看</el-button>
              <el-button type="info" size="small" @click="handleHash(row)">哈希</el-button>
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
import { Refresh, Folder, Key, Document, Clock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

interface FileInfo {
  name: string
  path: string
  size: number
  modified: string
  hash: string
}

const loading = ref(false)
const searchKeyword = ref('')
const fileList = ref<FileInfo[]>([])

const filteredFileList = computed(() => {
  if (!searchKeyword.value) return fileList.value
  const keyword = searchKeyword.value.toLowerCase()
  return fileList.value.filter(f => f.name.toLowerCase().includes(keyword))
})

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  while (bytes >= 1024 && index < units.length - 1) {
    bytes /= 1024
    index++
  }
  return `${bytes.toFixed(2)} ${units[index]}`
}

async function loadFileList() {
  loading.value = true
  try {
    const data = await Go.GetFileList()
    if (data) {
      fileList.value = data as FileInfo[]
    }
  } catch (error) {
    console.error('Failed to load file list:', error)
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadFileList()
}

function handleFeature(feature: string) {
  ElMessage.info(`功能: ${feature}`)
}

function handleView(row: FileInfo) {
  ElMessage.info(`查看文件: ${row.name}`)
}

function handleHash(row: FileInfo) {
  ElMessage.info(`计算哈希: ${row.name}`)
}

function handleDelete(row: FileInfo) {
  ElMessage.warning(`删除文件: ${row.name}`)
}

onMounted(() => {
  loadFileList()
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
