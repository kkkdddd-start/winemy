<template>
  <div class="progress-component">
    <div v-if="showHeader" class="progress-header">
      <span class="progress-title">{{ title }}</span>
      <span class="progress-percentage">{{ percentage }}%</span>
    </div>
    <el-progress
      :percentage="percentage"
      :type="type"
      :stroke-width="strokeWidth"
      :color="progressColor"
      :show-text="showText"
      :status="progressStatus"
      :indeterminate="indeterminate"
    >
      <slot>{{ displayText }}</slot>
    </el-progress>
    <div v-if="showInfo" class="progress-info">
      <span v-if="message">{{ message }}</span>
      <span v-if="eta" class="eta">预计剩余: {{ eta }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ProgressType } from 'element-plus'

interface Props {
  percentage?: number
  type?: ProgressType
  strokeWidth?: number
  color?: string | Record<string, string> | ((percentage: number) => string)
  showText?: boolean
  status?: 'success' | 'warning' | 'exception' | ''
  indeterminate?: boolean
  title?: string
  showHeader?: boolean
  showInfo?: boolean
  message?: string
  eta?: string
  severity?: 'low' | 'medium' | 'high' | 'critical'
}

const props = withDefaults(defineProps<Props>(), {
  percentage: 0,
  type: 'line',
  strokeWidth: 6,
  color: undefined,
  showText: true,
  status: '',
  indeterminate: false,
  showHeader: false,
  showInfo: false
})

const displayText = computed(() => {
  if (props.message) return props.message
  return ''
})

const progressColor = computed(() => {
  if (props.color) return props.color

  const pct = props.percentage || 0
  if (pct >= 100) return '#67c23a'
  if (pct >= 80) return '#f56c6c'
  if (pct >= 60) return '#e6a23c'
  return '#409eff'
})

const progressStatus = computed(() => {
  if (props.status) return props.status
  if ((props.percentage || 0) >= 100) return 'success'
  return ''
})
</script>

<style scoped>
.progress-component {
  width: 100%;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.progress-percentage {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.eta {
  color: var(--el-text-color-secondary);
}
</style>
