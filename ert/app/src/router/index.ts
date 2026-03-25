import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/m1' },
  { path: '/m1', component: () => import('@/views/module/SystemView.vue'), meta: { title: '系统概览', moduleId: 1 } },
  { path: '/m2', component: () => import('@/views/module/ProcessView.vue'), meta: { title: '进程管理', moduleId: 2 } },
  { path: '/m3', component: () => import('@/views/module/NetworkView.vue'), meta: { title: '网络分析', moduleId: 3 } },
  { path: '/m4', component: () => import('@/views/module/RegistryView.vue'), meta: { title: '注册表分析', moduleId: 4 } },
  { path: '/m5', component: () => import('@/views/module/ServiceView.vue'), meta: { title: '服务管理', moduleId: 5 } },
  { path: '/m6', component: () => import('@/views/module/ScheduleView.vue'), meta: { title: '计划任务', moduleId: 6 } },
  { path: '/m7', component: () => import('@/views/module/MonitorView.vue'), meta: { title: '系统监控', moduleId: 7 } },
  { path: '/m8', component: () => import('@/views/module/PatchView.vue'), meta: { title: '系统补丁', moduleId: 8 } },
  { path: '/m9', component: () => import('@/views/module/SoftwareView.vue'), meta: { title: '软件列表', moduleId: 9 } },
  { path: '/m10', component: () => import('@/views/module/KernelView.vue'), meta: { title: '内核分析', moduleId: 10 } },
  { path: '/m11', component: () => import('@/views/module/FilesystemView.vue'), meta: { title: '文件系统', moduleId: 11 } },
  { path: '/m12', component: () => import('@/views/module/ActivityView.vue'), meta: { title: '活动痕迹', moduleId: 12 } },
  { path: '/m13', component: () => import('@/views/module/LoggingView.vue'), meta: { title: '日志分析', moduleId: 13 } },
  { path: '/m14', component: () => import('@/views/module/AccountView.vue'), meta: { title: '账户分析', moduleId: 14 } },
  { path: '/m15', component: () => import('@/views/module/MemoryView.vue'), meta: { title: '内存取证', moduleId: 15 } },
  { path: '/m16', component: () => import('@/views/module/ThreatView.vue'), meta: { title: '威胁检测', moduleId: 16 } },
  { path: '/m17', component: () => import('@/views/module/ResponseView.vue'), meta: { title: '应急处置', moduleId: 17 } },
  { path: '/m18', component: () => import('@/views/module/AutostartView.vue'), meta: { title: '自启动项目', moduleId: 18 } },
  { path: '/m19', component: () => import('@/views/module/DomainView.vue'), meta: { title: '域控检测', moduleId: 19 } },
  { path: '/m20', component: () => import('@/views/module/DomainHackView.vue'), meta: { title: '域内渗透', moduleId: 20 } },
  { path: '/m21', component: () => import('@/views/module/WMICView.vue'), meta: { title: 'WMIC 检测', moduleId: 21 } },
  { path: '/m22', component: () => import('@/views/module/ReportView.vue'), meta: { title: '报告导出', moduleId: 22 } },
  { path: '/m23', component: () => import('@/views/module/BaselineView.vue'), meta: { title: '安全基线', moduleId: 23 } },
  { path: '/m24', component: () => import('@/views/module/IISView.vue'), meta: { title: 'IIS 日志', moduleId: 24 } },
  { path: '/m25', component: () => import('@/views/module/CodecView.vue'), meta: { title: '编解码工具', moduleId: 25 } },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export function useModuleRoutes() {
  return routes.filter(r => r.path !== '/').map(r => ({
    path: r.path,
    title: r.meta?.title as string,
    moduleId: r.meta?.moduleId as number
  }))
}
