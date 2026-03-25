import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

export function useShortcuts() {
  const router = useRouter()
  const shortcuts = new Map<string, () => void>()

  shortcuts.set('ctrl+shift+t', () => {
    window.runtime.EventsEmit('ert:triage')
  })

  shortcuts.set('ctrl+e', () => {
    window.runtime.EventsEmit('ert:global-search')
  })

  shortcuts.set('ctrl+d', () => {
    router.push('/m25')
  })

  for (let i = 1; i <= 9; i++) {
    const moduleId = i
    shortcuts.set(`ctrl+${i}`, () => {
      router.push(`/m${moduleId}`)
    })
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
      return
    }

    const key = [
      e.ctrlKey ? 'ctrl' : '',
      e.shiftKey ? 'shift' : '',
      e.key.toLowerCase()
    ].filter(Boolean).join('+')

    const handler = shortcuts.get(key)
    if (handler) {
      e.preventDefault()
      handler()
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })
}
