<template>
  <div class="compare-component">
    <div class="compare-header">
      <div class="compare-session">
        <span class="session-label">{{ session1Label }}</span>
        <el-select v-model="session1Value" placeholder="选择会话1" @change="handleSession1Change">
          <el-option
            v-for="session in sessions"
            :key="session.id"
            :label="session.label"
            :value="session.id"
          />
        </el-select>
      </div>
      <div class="compare-action">
        <el-button type="primary" @click="handleCompare" :loading="loading" :disabled="!canCompare">
          <el-icon><Connection /></el-icon>
          对比
        </el-button>
      </div>
      <div class="compare-session">
        <span class="session-label">{{ session2Label }}</span>
        <el-select v-model="session2Value" placeholder="选择会话2" @change="handleSession2Change">
          <el-option
            v-for="session in sessions"
            :key="session.id"
            :label="session.label"
            :value="session.id"
          />
        </el-select>
      </div>
    </div>

    <div v-if="result" class="compare-result">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="新增项" name="added">
          <div class="result-section">
            <div v-if="result.added && result.added.length > 0" class="result-list">
              <div v-for="(item, index) in result.added" :key="index" class="result-item added">
                <el-icon class="item-icon"><Plus /></el-icon>
                <span class="item-text">{{ formatItem(item) }}</span>
              </div>
            </div>
            <el-empty v-else description="无新增项" />
          </div>
        </el-tab-pane>
        <el-tab-pane label="删除项" name="removed">
          <div class="result-section">
            <div v-if="result.removed && result.removed.length > 0" class="result-list">
              <div v-for="(item, index) in result.removed" :key="index" class="result-item removed">
                <el-icon class="item-icon"><Minus /></el-icon>
                <span class="item-text">{{ formatItem(item) }}</span>
              </div>
            </div>
            <el-empty v-else description="无删除项" />
          </div>
        </el-tab-pane>
        <el-tab-pane label="共同项" name="common">
          <div class="result-section">
            <div v-if="result.common && result.common.length > 0" class="result-list">
              <div v-for="(item, index) in result.common" :key="index" class="result-item common">
                <el-icon class="item-icon"><Check /></el-icon>
                <span class="item-text">{{ formatItem(item) }}</span>
              </div>
            </div>
            <el-empty v-else description="无共同项" />
          </div>
        </el-tab-pane>
      </el-tabs>

      <div class="compare-summary">
        <el-statistic title="新增" :value="result.added?.length || 0" />
        <el-statistic title="删除" :value="result.removed?.length || 0" />
        <el-statistic title="共同" :value="result.common?.length || 0" />
      </div>
    </div>

    <div v-else class="compare-placeholder">
      <el-empty description="请选择两个会话进行对比" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Connection, Plus, Minus, Check } from '@element-plus/icons-vue'

interface Session {
  id: string
  label: string
  timestamp?: string
}

interface CompareResult {
  added?: any[]
  removed?: any[]
  common?: any[]
}

interface Props {
  sessions?: Session[]
  session1Label?: string
  session2Label?: string
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  sessions: () => [],
  session1Label: '会话 1',
  session2Label: '会话 2',
  loading: false
})

const emit = defineEmits<{
  compare: [session1: string, session2: string]
}>()

const session1Value = ref<string>('')
const session2Value = ref<string>('')
const activeTab = ref('added')
const result = ref<CompareResult | null>(null)

const canCompare = computed(() => {
  return session1Value.value && session2Value.value && session1Value.value !== session2Value.value
})

function handleSession1Change(val: string) {
  session1Value.value = val
}

function handleSession2Change(val: string) {
  session2Value.value = val
}

function handleCompare() {
  if (canCompare.value) {
    emit('compare', session1Value.value, session2Value.value)
  }
}

function formatItem(item: any): string {
  if (typeof item === 'string') return item
  if (item.name) return item.name
  if (item.path) return item.path
  if (item.title) return item.title
  return JSON.stringify(item)
}

function setResult(newResult: CompareResult) {
  result.value = newResult
}

defineExpose({
  setResult
})
</script>

<style scoped>
.compare-component {
  width: 100%;
}

.compare-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 20px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.compare-session {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.session-label {
  font-weight: 500;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

.compare-action {
  flex-shrink: 0;
}

.compare-result {
  margin-top: 16px;
}

.result-section {
  min-height: 200px;
}

.result-list {
  max-height: 400px;
  overflow-y: auto;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 4px;
  border-radius: 4px;
  font-size: 14px;
}

.result-item.added {
  background: rgba(103, 194, 58, 0.1);
  color: #67c23a;
}

.result-item.removed {
  background: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.result-item.common {
  background: rgba(64, 120, 192, 0.1);
  color: #409eff;
}

.item-icon {
  flex-shrink: 0;
}

.item-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.compare-summary {
  display: flex;
  justify-content: space-around;
  margin-top: 20px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.compare-placeholder {
  padding: 60px 0;
}
</style>
