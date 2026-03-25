import { ref, computed, defineStore } from 'vue'
import { Go } from '@wailsjs/go/main/App'

export function useBaseStore(moduleId: number) {
  const data = ref<any[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdate = ref<string | null>(null)

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const filteredData = computed(() => data.value)
  const hasData = computed(() => data.value.length > 0)

  async function fetchData(query = '') {
    loading.value = true
    error.value = null
    try {
      const result = await Go.GetModuleData(moduleId, query)
      if (result) {
        data.value = result
        pagination.value.total = data.value.length
        lastUpdate.value = new Date().toISOString()
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function collect() {
    loading.value = true
    try {
      await Go.CollectModule(moduleId)
      await fetchData()
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  function setPage(page: number) { pagination.value.page = page }
  function setPageSize(pageSize: number) { pagination.value.pageSize = pageSize; pagination.value.page = 1 }
  function reset() { data.value = []; error.value = null; pagination.value = { page: 1, pageSize: 20, total: 0 } }

  return { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset }
}

export const useSystemStore = defineStore('system', () => {
  const { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset } = useBaseStore(1)
  const systemInfo = ref<any>(null)
  async function fetchSystemInfo() {
    loading.value = true
    try {
      const result = await Go.GetModuleData(1, '')
      if (result?.length) systemInfo.value = result[0]
    } catch (e) { error.value = e instanceof Error ? e.message : String(e) }
    finally { loading.value = false }
  }
  return { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset, systemInfo, fetchSystemInfo }
})

export const useProcessStore = defineStore('process', () => {
  return useBaseStore(2)
})

export const useNetworkStore = defineStore('network', () => {
  return useBaseStore(3)
})

export const useRegistryStore = defineStore('registry', () => {
  return useBaseStore(4)
})

export const useServiceStore = defineStore('service', () => {
  return useBaseStore(5)
})

export const useScheduleStore = defineStore('schedule', () => {
  return useBaseStore(6)
})

export const useMonitorStore = defineStore('monitor', () => {
  const { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset } = useBaseStore(7)
  const realtimeHistory = ref<any>(null)
  async function fetchRealtimeHistory() {
    try { const result = await Go.GetModuleData(7, 'realtime'); if (result) realtimeHistory.value = result[0] }
    catch (e) { error.value = e instanceof Error ? e.message : String(e) }
  }
  return { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset, realtimeHistory, fetchRealtimeHistory }
})

export const usePatchStore = defineStore('patch', () => {
  return useBaseStore(8)
})

export const useSoftwareStore = defineStore('software', () => {
  return useBaseStore(9)
})

export const useKernelStore = defineStore('kernel', () => {
  return useBaseStore(10)
})

export const useFilesystemStore = defineStore('filesystem', () => {
  return useBaseStore(11)
})

export const useActivityStore = defineStore('activity', () => {
  return useBaseStore(12)
})

export const useLoggingStore = defineStore('logging', () => {
  return useBaseStore(13)
})

export const useAccountStore = defineStore('account', () => {
  return useBaseStore(14)
})

export const useMemoryStore = defineStore('memory', () => {
  return useBaseStore(15)
})

export const useThreatStore = defineStore('threat', () => {
  return useBaseStore(16)
})

export const useResponseStore = defineStore('response', () => {
  return useBaseStore(17)
})

export const useAutostartStore = defineStore('autostart', () => {
  return useBaseStore(18)
})

export const useDomainStore = defineStore('domain', () => {
  return useBaseStore(19)
})

export const useDomainHackStore = defineStore('domainHack', () => {
  return useBaseStore(20)
})

export const useWMICStore = defineStore('wmic', () => {
  return useBaseStore(21)
})

export const useReportStore = defineStore('report', () => {
  return useBaseStore(22)
})

export const useBaselineStore = defineStore('baseline', () => {
  return useBaseStore(23)
})

export const useIISStore = defineStore('iis', () => {
  return useBaseStore(24)
})

export const useCodecStore = defineStore('codec', () => {
  const { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset } = useBaseStore(25)
  const input = ref('')
  const output = ref('')
  const encodingType = ref('base64')
  function setInput(val: string) { input.value = val }
  function setOutput(val: string) { output.value = val }
  function setEncodingType(val: string) { encodingType.value = val }
  return { data, loading, error, lastUpdate, pagination, filteredData, hasData, fetchData, collect, setPage, setPageSize, reset, input, output, encodingType, setInput, setOutput, setEncodingType }
})
