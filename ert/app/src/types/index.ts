export interface SystemInfo {
  hostname: string
  os_name: string
  os_version: string
  architecture: string
  boot_time: string
  current_user: string
  cpu_count: number
  memory_total: number
  disk_total: number
  is_domain: boolean
  domain_name: string
}

export interface ProcessDTO {
  pid: number
  name: string
  path: string
  command_line: string
  user: string
  cpu: number
  memory: number
  start_time: string
  risk_level: RiskLevel
}

export interface NetworkConnDTO {
  pid: number
  protocol: string
  local_addr: string
  local_port: number
  remote_addr: string
  remote_port: number
  state: string
  risk_level: RiskLevel
}

export interface RegistryKeyDTO {
  path: string
  name: string
  value_type: string
  value: string
  modified: string
  risk_level: RiskLevel
}

export interface ServiceDTO {
  name: string
  display_name: string
  status: string
  start_type: string
  path: string
  risk_level: RiskLevel
}

export interface ScheduledTaskDTO {
  name: string
  path: string
  state: string
  last_run_time: string
  next_run_time: string
  risk_level: RiskLevel
}

export interface EventLogDTO {
  event_id: number
  event_type: string
  level: string
  source: string
  channel: string
  computer: string
  time_created: string
  raw_xml: string
  message: string
}

export interface AccountDTO {
  name: string
  full_name: string
  sid: string
  domain: string
  status: string
  last_logon: string
  risk_level: RiskLevel
}

export interface DriverDTO {
  name: string
  path: string
  base_address: string
  size: number
  is_signed: boolean
  signature: string
  risk_level: RiskLevel
}

export interface MemoryDumpDTO {
  pid: number
  process_name: string
  dump_type: string
  file_path: string
  file_size: number
  sha256: string
  created_at: string
}

export interface Progress {
  module_id: number
  current: number
  total: number
  percentage: number
  message: string
  eta: string
}

export interface AlertEvent {
  id: string
  rule_id: string
  rule_name: string
  severity: RiskLevel
  message: string
  value: number
  threshold: number
  timestamp: string
  module_id: number
}

export interface CompareResult {
  session1: string
  session2: string
  added_processes: ProcessDTO[]
  removed_processes: ProcessDTO[]
  added_network: NetworkConnDTO[]
  removed_network: NetworkConnDTO[]
  added_registry: RegistryKeyDTO[]
  removed_registry: RegistryKeyDTO[]
  added_services: ServiceDTO[]
  removed_services: ServiceDTO[]
}

export type RiskLevel = 0 | 1 | 2 | 3

export const RiskLevelLabels: Record<RiskLevel, string> = {
  0: '低风险',
  1: '中风险',
  2: '高风险',
  3: '严重'
}

export const RiskLevelColors: Record<RiskLevel, string> = {
  0: '#67c23a',
  1: '#e6a23c',
  2: '#f56c6c',
  3: '#909399'
}

export interface Config {
  app: {
    name: string
    version: string
    debug: boolean
  }
  server: {
    host: string
    port: number
  }
  database: {
    main: {
      path: string
      wal_mode: boolean
      busy_timeout: number
      max_open_conns: number
    }
    cache: {
      max_cost: number
      buffer_size: number
      ttl: number
    }
  }
  storage: {
    data_dir: string
    dump_dir: string
    report_dir: string
    max_storage: number
  }
  ui: {
    theme: string
    language: string
    refresh_rate: string
    table: {
      page_size: number
      virtual_scroll: boolean
    }
    chart: {
      animation: boolean
      theme: string
    }
    shortcuts: {
      enabled: boolean
    }
  }
  modules: {
    [key: string]: any
  }
}
