import { defineStore } from 'pinia'
import { useBaseStore } from '../base'

export const useSystemStore = defineStore('system', () => {
  const base = useBaseStore(1)
  const systemInfo = ref<any>(null)

  async function fetchSystemInfo() {
    base.loading.value = true
    try {
      const result = await Go.GetModuleData(1, '')
      if (result && result.length > 0) {
        systemInfo.value = result[0]
      }
    } catch (e) {
      base.error.value = e instanceof Error ? e.message : String(e)
    } finally {
      base.loading.value = false
    }
  }

  return {
    ...base,
    systemInfo,
    fetchSystemInfo
  }
})
