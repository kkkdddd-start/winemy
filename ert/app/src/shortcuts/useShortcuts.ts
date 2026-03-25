import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

export function useShortcuts() {
  const router = useRouter()
  const shortcuts = new Map<string, () => void>()
  const isFullscreen = ref(false)

  shortcuts.set('ctrl+shift+t', () => {
    window.runtime?.EventsEmit('ert:triage')
    ElMessage.success('开始 Triage 采集...')
  })

  shortcuts.set('ctrl+e', () => {
    window.runtime?.EventsEmit('ert:global-search')
  })

  shortcuts.set('ctrl+d', () => {
    router.push('/m25')
  })

  shortcuts.set('ctrl+s', () => {
    window.runtime?.EventsEmit('ert:export')
    ElMessage.success('触发数据导出...')
  })

  shortcuts.set('ctrl+r', () => {
    window.runtime?.EventsEmit('ert:refresh')
    ElMessage.success('刷新当前模块...')
  })

  shortcuts.set('ctrl+f', () => {
    window.runtime?.EventsEmit('ert:page-search')
  })

  shortcuts.set('f5', () => {
    window.runtime?.EventsEmit('ert:refresh')
    ElMessage.success('刷新当前模块...')
  })

  shortcuts.set('f11', () => {
    toggleFullscreen()
  })

  shortcuts.set('escape', () => {
    window.runtime?.EventsEmit('ert:close-dialog')
  })

  for (let i = 1; i <= 9; i++) {
    const moduleId = i
    shortcuts.set(`ctrl+${i}`, () => {
      router.push(`/m${moduleId}`)
    })
  }

  shortcuts.set('ctrl+0', () => {
    router.push('/m10')
  })

  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().then(() => {
        isFullscreen.value = true
        ElMessage.success('已切换到全屏模式')
      }).catch(() => {
        ElMessage.error('全屏模式不可用')
      })
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().then(() => {
          isFullscreen.value = false
          ElMessage.info('已退出全屏模式')
        }).catch(() => {})
      }
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement || e.target instanceof HTMLSelectElement) {
      if (e.key === 'Escape') {
        const handler = shortcuts.get('escape')
        if (handler) {
          e.preventDefault()
          handler()
        }
      }
      return
    }

    const keyParts: string[] = []
    if (e.ctrlKey) keyParts.push('ctrl')
    if (e.shiftKey) keyParts.push('shift')
    if (e.altKey) keyParts.push('alt')
    
    let key = e.key.toLowerCase()
    if (key === ' ') key = 'space'
    if (key === '+') key = 'plus'
    if (key === '-') key = 'minus'
    
    keyParts.push(key)
    const shortcutKey = keyParts.filter(Boolean).join('+')

    const handler = shortcuts.get(shortcutKey)
    if (handler) {
      e.preventDefault()
      handler()
    }
  }

  function handleFullscreenChange() {
    isFullscreen.value = !!document.fullscreenElement
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
    document.addEventListener('fullscreenchange', handleFullscreenChange)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
  })

  return {
    isFullscreen,
    toggleFullscreen
  }
}

export function useGlobalSearch() {
  function openGlobalSearch() {
    window.runtime?.EventsEmit('ert:global-search')
  }

  function closeGlobalSearch() {
    window.runtime?.EventsEmit('ert:close-search')
  }

  return {
    openGlobalSearch,
    closeGlobalSearch
  }
}

export function useTriage() {
  async function startTriage() {
    window.runtime?.EventsEmit('ert:triage')
    ElMessage.info('开始 Triage 采集...')
  }

  return {
    startTriage
  }
}
