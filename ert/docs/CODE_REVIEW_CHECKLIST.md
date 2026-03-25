# ERT 代码审查检查清单

本文档定义了 ERT (应急响应工具箱) 项目的代码审查标准和检查流程。

---

## 一、功能与业务逻辑检查 (Functionality & Business Logic)

### 1.1 核心功能闭环

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 模块接口完整性 | 25个模块是否都实现了 Module 接口 | 🔴 高 | ☐ |
| 前后端API匹配 | 前端调用的 Go.XXX() 方法是否都在 app.go 中暴露 | 🔴 高 | ☐ |
| 数据流闭环 | handleRefresh → Go.XXX() → 后端 GetData() → 前端展示 | 🔴 高 | ☐ |
| Registry注册 | 所有模块是否都已注册 | 🔴 高 | ☐ |

### 1.2 边界条件处理

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 空数据处理 | GetData() 返回空数组时前端是否有友好展示 | 🟡 中 | ☐ |
| 错误提示 | API调用失败是否有 ElMessage.error() | 🔴 高 | ☐ |
| 加载状态 | 异步操作是否有 v-loading | 🔴 高 | ☐ |
| 超时处理 | Wails 调用是否有超时机制 | 🟡 中 | ☐ |
| 二次确认 | 危险操作是否有确认对话框 | 🔴 高 | ☐ |

### 1.3 Windows 特性兼容

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 路径分隔符 | 路径是否正确处理 | 🔴 高 | ☐ |
| 系统权限 | 需要管理员权限的操作是否有提示 | 🔴 高 | ☐ |
| 中文路径 | 路径包含中文时是否正确处理 | 🟡 中 | ☐ |
| UAC提升 | 需要提权的操作是否能正确触发UAC | 🟡 中 | ☐ |

---

## 二、Go 后端检查 (Backend)

### 2.1 代码安全

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 敏感信息硬编码 | 检查是否有密码、密钥硬编码 | 🔴 高 | ☐ |
| 命令注入防护 | exec 执行命令是否校验输入 | 🔴 高 | ☐ |
| 路径遍历防护 | 文件操作是否防止 ../ 越界 | 🔴 高 | ☐ |
| PowerShell安全 | PowerShell 命令是否使用 -ErrorAction | 🟡 中 | ☐ |
| 进程保护 | 关键系统进程是否有保护 | 🔴 高 | ☐ |

### 2.2 Go 代码规范

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 错误处理 | 所有 error 返回值是否被正确处理 | 🔴 高 | ☐ |
| 资源释放 | 文件操作是否使用 defer 关闭 | 🔴 高 | ☐ |
| 上下文传递 | 长操作是否支持 context.Context | 🟡 中 | ☐ |
| 日志脱敏 | 敏感数据是否脱敏 | 🔴 高 | ☐ |
| go:build 标签 | Windows 专用代码是否标注 | 🔴 高 | ☐ |

### 2.3 Wails 特定检查

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 结构体导出 | 暴露给前端结构体字段是否大写 | 🔴 高 | ☐ |
| 方法绑定 | wails.Bind() 注册的方法是否 public | 🔴 高 | ☐ |
| 参数类型 | Wails 绑定方法参数类型是否支持 | 🔴 高 | ☐ |
| 窗口配置 | wails.Run() 窗口配置是否合理 | 🟢 低 | ☐ |

---

## 三、Vue 前端检查 (Frontend)

### 3.1 Vue 3 + Element Plus

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 组件销毁 | onUnmounted 是否清理定时器/监听器 | 🔴 高 | ☐ |
| 响应式数据 | ref/reactive 使用是否正确 | 🔴 高 | ☐ |
| 异步加载 | 大数据量是否使用分页 | 🟡 中 | ☐ |
| 表单验证 | el-form 验证规则是否完整 | 🟡 中 | ☐ |
| 错误边界 | 是否有全局错误捕获 | 🟡 中 | ☐ |

### 3.2 TypeScript 类型定义

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 接口定义 | 每个模块是否有完整 TypeScript 接口 | 🟡 中 | ☐ |
| 类型安全 | 是否避免 any 类型乱用 | 🟡 中 | ☐ |
| 前后端DTO匹配 | TypeScript 接口是否与 Go DTO 匹配 | 🔴 高 | ☐ |

### 3.3 UI/UX 体验

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 加载状态 | 异步操作是否有 v-loading | 🔴 高 | ☐ |
| 操作反馈 | 成功/失败是否有 ElMessage 通知 | 🔴 高 | ☐ |
| 确认对话框 | 危险操作是否有 ElMessageBox.confirm | 🔴 高 | ☐ |
| 快捷键支持 | 常用操作是否支持快捷键 | 🟡 中 | ☐ |

---

## 四、架构设计检查

### 4.1 模块化设计

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| Registry模式 | 模块注册机制是否正确 | 🔴 高 | ☐ |
| 接口一致性 | 25个模块是否遵循统一接口 | 🔴 高 | ☐ |
| 优先级设计 | 模块优先级是否合理设置 | 🟡 中 | ☐ |

### 4.2 前后端交互

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| API命名规范 | 前端调用名与后端方法是否一致 | 🔴 高 | ☐ |
| 数据格式 | []map[string]interface{} 是否正确序列化 | 🔴 高 | ☐ |

---

## 五、安全检查

### 5.1 敏感操作保护

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 进程保护 | 关键系统进程禁止终止 | 🔴 高 | ☐ |
| 服务保护 | 关键服务禁止禁用 | 🔴 高 | ☐ |
| 注册表保护 | 关键注册表键禁止修改 | 🔴 高 | ☐ |
| 文件保护 | 系统关键文件禁止删除 | 🔴 高 | ☐ |

### 5.2 权限控制

| 检查项 | 说明 | 优先级 | 检查状态 |
|--------|------|--------|----------|
| 管理员检查 | 需要管理员权限的操作是否检查 | 🔴 高 | ☐ |
| 操作审计 | 敏感操作是否记录审计日志 | 🟡 中 | ☐ |

---

## 六、ERT 项目特定检查

### 6.1 前端视图与后端API对应

| 前端文件 | 调用API | 检查状态 |
|----------|---------|----------|
| SystemView.vue | Go.GetSystemInfo() | ☐ |
| ProcessView.vue | Go.GetProcessList(), Go.GetProcessTree(), Go.KillProcess(), Go.DumpProcess() | ☐ |
| NetworkView.vue | Go.GetNetworkList() | ☐ |
| ServiceView.vue | Go.GetServices(), Go.StartService(), Go.StopService(), Go.RestartService() | ☐ |
| RegistryView.vue | Go.GetRegistryKeys() | ☐ |
| ScheduleView.vue | Go.GetScheduledTasks(), Go.ExportScheduledTaskToXML() | ☐ |
| MonitorView.vue | Go.GetMonitorData() | ☐ |
| PatchView.vue | Go.GetPatches() | ☐ |
| SoftwareView.vue | Go.GetSoftwareList() | ☐ |
| KernelView.vue | Go.GetDrivers() | ☐ |
| FilesystemView.vue | Go.GetFiles() | ☐ |
| ActivityView.vue | Go.GetActivity() | ☐ |
| LoggingView.vue | Go.GetEventLogs() | ☐ |
| AccountView.vue | Go.GetAccounts() | ☐ |
| MemoryView.vue | Go.GetMemoryDumps(), Go.DumpProcess() | ☐ |
| ThreatView.vue | Go.GetThreats() | ☐ |
| ResponseView.vue | Go.ResponseAction() | ☐ |
| AutostartView.vue | Go.GetAutostartItems() | ☐ |
| DomainView.vue | Go.GetDomainInfo() | ☐ |
| DomainHackView.vue | Go.GetDomainHackDetections() | ☐ |
| WMICView.vue | Go.GetWMICHistory() | ☐ |
| ReportView.vue | Go.GetReportHistory(), Go.ExportReport() | ☐ |
| BaselineView.vue | Go.GetBaselineResults() | ☐ |
| IISView.vue | Go.GetIISLogs() | ☐ |
| CodecView.vue | Go.CodecEncode(), Go.CodecDecode(), Go.CodecAutoDetect() | ☐ |

### 6.2 模块功能完整性

| 模块 | 检查项 | 优先级 | 检查状态 |
|------|--------|--------|----------|
| M2 进程 | GetProcessTree() 是否正确构建 | 🔴 高 | ☐ |
| M2 进程 | 进程 Dump 是否使用真实 API | 🔴 高 | ☐ |
| M3 网络 | IP 地理位置库是否可用 | 🟡 中 | ☐ |
| M7 监控 | 实时数据采集是否真实 | 🔴 高 | ☐ |
| M15 内存 | Dump 文件 SHA256 是否计算 | 🟡 中 | ☐ |
| M17 应急 | IP 封锁是否添加防火墙规则 | 🔴 高 | ☐ |

---

## 七、检查执行记录

### 检查时间: 2026-03-25

### 检查人员: AI Assistant

### 检查结果汇总

| 类别 | 通过 | 失败 | 未检查 |
|------|------|------|--------|
| 功能与业务逻辑 | - | - | - |
| Go 后端 | - | - | - |
| Vue 前端 | - | - | - |
| 架构设计 | - | - | - |
| 安全检查 | - | - | - |
| 项目特定 | - | - | - |

### 发现的问题

1. [问题编号] - [问题描述]
   - 位置: [文件:行号]
   - 优先级: [高/中/低]
   - 状态: [待修复/已修复]

---

## 八、快速检查命令

```bash
# Go 语法检查
cd /workspace/ert && go vet ./...

# 依赖检查
cd /workspace/ert && go mod tidy

# 前端 lint
cd /workspace/ert/app && npm run lint

# 构建测试
cd /workspace/ert && wails build
```
