<template>
  <div class="search-component">
    <el-input
      v-model="searchValue"
      :placeholder="placeholder"
      :clearable="clearable"
      :disabled="disabled"
      :size="size"
      @input="handleInput"
      @change="handleChange"
      @clear="handleClear"
      @keyup.enter="handleSearch"
    >
      <template #prefix>
        <el-icon><Search /></el-icon>
      </template>
      <template #append v-if="showButton">
        <el-button @click="handleSearch" :disabled="disabled">
          <el-icon><Search /></el-icon>
        </el-button>
      </template>
    </el-input>
    <div v-if="showSuggestion && suggestions.length > 0" class="suggestions">
      <div
        v-for="(item, index) in suggestions"
        :key="index"
        class="suggestion-item"
        @click="handleSuggestionClick(item)"
      >
        {{ item }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { InputSize } from 'element-plus'

interface Props {
  modelValue?: string
  placeholder?: string
  clearable?: boolean
  disabled?: boolean
  size?: InputSize
  debounce?: number
  showButton?: boolean
  showSuggestion?: boolean
  suggestions?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '搜索...',
  clearable: true,
  disabled: false,
  size: 'default',
  debounce: 300,
  showButton: false,
  showSuggestion: false,
  suggestions: () => []
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'search': [value: string]
  'change': [value: string]
  'clear': []
  'suggestion-click': [item: string]
}>()

const searchValue = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  searchValue.value = val
})

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function handleInput(value: string) {
  emit('update:modelValue', value)

  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    emit('change', value)
  }, props.debounce)
}

function handleChange(value: string) {
  emit('change', value)
}

function handleClear() {
  emit('update:modelValue', '')
  emit('clear')
}

function handleSearch() {
  emit('search', searchValue.value)
}

function handleSuggestionClick(item: string) {
  emit('suggestion-click', item)
}
</script>

<style scoped>
.search-component {
  position: relative;
}

.suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-top: none;
  border-radius: 0 0 var(--el-border-radius-base) var(--el-border-radius-base);
  box-shadow: var(--el-box-shadow-light);
  z-index: 1000;
  max-height: 200px;
  overflow-y: auto;
}

.suggestion-item {
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
}

.suggestion-item:hover {
  background: var(--el-fill-color-light);
}
</style>
