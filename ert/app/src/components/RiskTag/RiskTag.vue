<template>
  <el-tag :type="tagType" :effect="effect" :size="size" :closable="closable" :disable-transitions="disableTransitions" @close="handleClose">
    <slot>{{ displayText }}</slot>
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { TagProps } from 'element-plus'

type RiskLevel = 0 | 1 | 2 | 3 | 'low' | 'medium' | 'high' | 'critical'

interface Props {
  riskLevel: RiskLevel
  showText?: boolean
  textMap?: Record<string, string>
  type?: TagProps['type']
  effect?: 'light' | 'dark' | 'plain'
  size?: 'large' | 'default' | 'small'
  closable?: boolean
  disableTransitions?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showText: true,
  textMap: () => ({
    '0': '低风险',
    'low': '低风险',
    '1': '中风险',
    'medium': '中风险',
    '2': '高风险',
    'high': '高风险',
    '3': '严重',
    'critical': '严重'
  }),
  effect: 'light',
  size: 'default',
  closable: false,
  disableTransitions: false
})

const emit = defineEmits<{
  close: []
}>()

const tagType = computed(() => {
  const level = props.riskLevel
  if (props.type) return props.type

  switch (String(level)) {
    case '0':
    case 'low':
      return 'success'
    case '1':
    case 'medium':
      return 'warning'
    case '2':
    case 'high':
      return 'danger'
    case '3':
    case 'critical':
      return 'danger'
    default:
      return 'info'
  }
})

const displayText = computed(() => {
  if (!props.showText) return ''
  const text = props.textMap[String(props.riskLevel)]
  return text || String(props.riskLevel)
})

function handleClose() {
  emit('close')
}
</script>

<style scoped>
.el-tag {
  font-weight: 500;
}
</style>
