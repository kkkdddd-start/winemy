<template>
  <div class="timeline-component">
    <div v-if="title" class="timeline-header">
      <span>{{ title }}</span>
    </div>
    <div class="timeline">
      <div
        v-for="(item, index) in items"
        :key="index"
        class="timeline-item"
        :class="{ 'timeline-item-last': index === items.length - 1 }"
      >
        <div class="timeline-marker" :style="{ backgroundColor: getMarkerColor(item.severity) }">
          <el-icon v-if="item.icon" :size="12"><component :is="item.icon" /></el-icon>
        </div>
        <div class="timeline-content">
          <div class="timeline-header">
            <span class="timeline-title">{{ item.title }}</span>
            <span class="timeline-time">{{ formatTime(item.timestamp) }}</span>
          </div>
          <div v-if="item.description" class="timeline-description">
            {{ item.description }}
          </div>
          <div v-if="item.tags && item.tags.length > 0" class="timeline-tags">
            <el-tag
              v-for="(tag, tagIndex) in item.tags"
              :key="tagIndex"
              size="small"
              :type="getTagType(item.severity)"
            >
              {{ tag }}
            </el-tag>
          </div>
          <div v-if="item.expandable && expandedItems.includes(index)" class="timeline-expanded">
            <slot :name="`item-${index}`" :item="item">{{ item.description }}</slot>
          </div>
          <div v-if="item.expandable" class="timeline-expand" @click="toggleExpand(index)">
            <el-link type="primary" :underline="false">
              {{ expandedItems.includes(index) ? '收起' : '展开' }}
              <el-icon class="el-icon--right">
                <ArrowUp v-if="expandedItems.includes(index)" />
                <ArrowDown v-else />
              </el-icon>
            </el-link>
          </div>
        </div>
      </div>
    </div>
    <div v-if="loading" class="timeline-loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>加载中...</span>
    </div>
    <div v-if="items.length === 0 && !loading" class="timeline-empty">
      <el-empty description="暂无时间线数据" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowUp, ArrowDown, Loading } from '@element-plus/icons-vue'

interface TimelineItem {
  id?: string | number
  title: string
  description?: string
  timestamp: string | Date | number
  severity?: 'low' | 'medium' | 'high' | 'critical' | 'info'
  icon?: string
  tags?: string[]
  expandable?: boolean
  data?: any
}

interface Props {
  items: TimelineItem[]
  title?: string
  loading?: boolean
  sortable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  loading: false,
  sortable: false
})

const expandedItems = ref<number[]>([])

function getMarkerColor(severity?: string): string {
  switch (severity) {
    case 'critical':
    case 'high':
      return '#f56c6c'
    case 'medium':
      return '#e6a23c'
    case 'low':
      return '#67c23a'
    case 'info':
    default:
      return '#409eff'
  }
}

function getTagType(severity?: string): string {
  switch (severity) {
    case 'critical':
    case 'high':
      return 'danger'
    case 'medium':
      return 'warning'
    case 'low':
      return 'success'
    case 'info':
    default:
      return 'info'
  }
}

function formatTime(timestamp: string | Date | number): string {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  if (isNaN(date.getTime())) return String(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

function toggleExpand(index: number) {
  const idx = expandedItems.value.indexOf(index)
  if (idx >= 0) {
    expandedItems.value.splice(idx, 1)
  } else {
    expandedItems.value.push(index)
  }
}
</script>

<style scoped>
.timeline-component {
  width: 100%;
}

.timeline-header {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--el-text-color-primary);
}

.timeline {
  position: relative;
  padding-left: 28px;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 7px;
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--el-border-color);
}

.timeline-item {
  position: relative;
  padding-bottom: 24px;
}

.timeline-item-last {
  padding-bottom: 0;
}

.timeline-marker {
  position: absolute;
  left: -24px;
  top: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  z-index: 1;
}

.timeline-content {
  background: var(--el-fill-color-light);
  border-radius: 4px;
  padding: 12px;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.timeline-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.timeline-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.timeline-description {
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-bottom: 8px;
  line-height: 1.5;
}

.timeline-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.timeline-expanded {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--el-border-color);
}

.timeline-expand {
  margin-top: 8px;
}

.timeline-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  padding: 20px;
  color: var(--el-text-color-secondary);
}

.timeline-empty {
  padding: 40px 0;
}
</style>
