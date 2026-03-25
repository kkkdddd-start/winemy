# ERT 代码功能实现检查列表

本文档根据设计文档 `docs/USER_MANUAL.md` 编制，提供最详细的粒度来检查每个模块的功能实现状态。

---

## M1 系统概览

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M1.1 | 主机基本信息：计算机名、操作系统、当前用户、启动时间 | ☐ |
| M1.2 | 系统资源：CPU 使用率 | ☐ |
| M1.3 | 系统资源：内存使用率 | ☐ |
| M1.4 | 系统资源：磁盘使用率 | ☐ |
| M1.5 | 网络状态：网卡信息 | ☐ |
| M1.6 | 网络状态：IP 地址 | ☐ |
| M1.7 | 网络状态：网络连接数 | ☐ |
| M1.8 | 实时监控图表：CPU 曲线 | ☐ |
| M1.9 | 实时监控图表：内存曲线 | ☐ |
| M1.10 | 实时监控图表：磁盘曲线 | ☐ |
| M1.11 | 实时监控图表：网络流量曲线 | ☐ |
| M1.12 | 域环境检测（IsDomain） | ☐ |
| M1.13 | 域名称获取（DomainName） | ☐ |
| M1.14 | Uptime 运行时间 | ☐ |
| M1.15 | 引导时间（BootTime） | ☐ |

### 实现文件
- `internal/modules/m1_system/system.go`

---

## M2 进程管理

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M2.1 | 进程列表：PID | ☐ |
| M2.2 | 进程列表：PPID（父进程ID） | ☐ |
| M2.3 | 进程列表：Name（名称） | ☐ |
| M2.4 | 进程列表：Path（路径） | ☐ |
| M2.5 | 进程列表：User（用户名）- 通过 Windows API 获取 | ☐ |
| M2.6 | 进程列表：CommandLine（命令行） | ☐ |
| M2.7 | 进程列表：CPU使用率 | ☐ |
| M2.8 | 进程列表：Memory（内存使用） | ☐ |
| M2.9 | 进程列表：StartTime（启动时间） | ☐ |
| M2.10 | 进程树：父子进程关系树形结构 | ☐ |
| M2.11 | 进程搜索：按名称搜索 | ☐ |
| M2.12 | 进程搜索：按 PID 搜索 | ☐ |
| M2.13 | 进程搜索：按路径搜索 | ☐ |
| M2.14 | 风险标记：无签名进程标黄 | ☐ |
| M2.15 | 风险标记：可疑进程标红（mimikatz, meterpreter 等） | ☐ |
| M2.16 | 进程查杀：KillProcess 方法 | ☐ |
| M2.17 | 进程查杀：关键系统进程保护（lsass, winlogon, csrss 等） | ☐ |
| M2.18 | 进程 Dump：DumpProcess 方法生成 .dmp 文件 | ☐ |
| M2.19 | 进程 Dump：使用 MiniDumpWriteDump API | ☐ |
| M2.20 | 进程 Dump：Dump 文件 SHA256 校验 | ☐ |
| M2.21 | GetProcessTree() 返回树形结构 | ☐ |
| M2.22 | GetProcess(pid) 根据 PID 获取详情 | ☐ |
| M2.23 | Search(keyword) 模糊搜索 | ☐ |

### 实现文件
- `internal/modules/m2_process/process.go`

---

## M3 网络分析

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M3.1 | 连接列表：Protocol（TCP/UDP） | ☐ |
| M3.2 | 连接列表：LocalAddr（本地地址） | ☐ |
| M3.3 | 连接列表：LocalPort（本地端口） | ☐ |
| M3.4 | 连接列表：RemoteAddr（远程地址） | ☐ |
| M3.5 | 连接列表：RemotePort（远程端口） | ☐ |
| M3.6 | 连接列表：State（状态：ESTABLISHED, LISTENING 等） | ☐ |
| M3.7 | 连接列表：PID（关联进程） | ☐ |
| M3.8 | 监听端口：所有 LISTENING 状态端口 | ☐ |
| M3.9 | 监听端口：显示对应进程名称 | ☐ |
| M3.10 | IP 地理位置：显示远程 IP 的国家/城市 | ☐ |
| M3.11 | 风险检测：可疑端口（4444, 5555 等） | ☐ |
| M3.12 | 风险检测：可疑外部连接（境外 IP） | ☐ |
| M3.13 | 风险检测：内网端口映射检测 | ☐ |
| M3.14 | 网络统计：各协议连接数 | ☐ |
| M3.15 | 网络统计：各状态连接数 | ☐ |

### 实现文件
- `internal/modules/m3_network/network.go`

---

## M4 注册表分析

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M4.1 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run | ☐ |
| M4.2 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce | ☐ |
| M4.3 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnceEx | ☐ |
| M4.4 | HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run | ☐ |
| M4.5 | HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce | ☐ |
| M4.6 | HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Run | ☐ |
| M4.7 | HKLM\SYSTEM\CurrentControlSet\Services（服务列表） | ☐ |
| M4.8 | HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon | ☐ |
| M4.9 | HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\RunMRU | ☐ |
| M4.10 | HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run | ☐ |
| M4.11 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run | ☐ |
| M4.12 | HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options | ☐ |
| M4.13 | HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Userinit | ☐ |
| M4.14 | HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell | ☐ |
| M4.15 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths | ☐ |
| M4.16 | HKLM\SOFTWARE\Classes\*\shell\open\command | ☐ |
| M4.17 | HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\FileExts | ☐ |
| M4.18 | AppInit_DLLs 检测 | ☐ |
| M4.19 | KnownDLLs 检测 | ☐ |
| M4.20 | LSA 保护设置检测（RunAsPPL） | ☐ |
| M4.21 | 注册表值类型解析：REG_SZ | ☐ |
| M4.22 | 注册表值类型解析：REG_DWORD | ☐ |
| M4.23 | 注册表值类型解析：REG_BINARY | ☐ |
| M4.24 | 注册表值类型解析：REG_MULTI_SZ | ☐ |
| M4.25 | 注册表值类型解析：REG_EXPAND_SZ | ☐ |
| M4.26 | 服务路径解析（ImagePath） | ☐ |
| M4.27 | 服务启动类型解析（Start 类型） | ☐ |
| M4.28 | 注册表最后修改时间 | ☐ |
| M4.29 | Search(keyword) 搜索功能 | ☐ |
| M4.30 | 风险等级评估 | ☐ |

### 实现文件
- `internal/modules/m4_registry/registry.go`

---

## M5 服务管理

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M5.1 | 服务列表：ServiceName | ☐ |
| M5.2 | 服务列表：DisplayName | ☐ |
| M5.3 | 服务列表：Status（Running/Stopped） | ☐ |
| M5.4 | 服务列表：StartType（自动/手动/禁用） | ☐ |
| M5.5 | 服务列表：Path（可执行文件路径） | ☐ |
| M5.6 | 服务列表：Description | ☐ |
| M5.7 | 服务依赖关系 | ☐ |
| M5.8 | 风险检测：异常服务 | ☐ |
| M5.9 | 风险检测：禁用安全服务（如 Windows Defender） | ☐ |
| M5.10 | 服务操作：启动服务 | ☐ |
| M5.11 | 服务操作：停止服务 | ☐ |
| M5.12 | 服务操作：重启服务 | ☐ |
| M5.13 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m5_service/service.go`

---

## M6 计划任务

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M6.1 | 任务列表：TaskName | ☐ |
| M6.2 | 任务列表：State（状态：Ready, Running, Disabled） | ☐ |
| M6.3 | 任务列表：LastRunTime | ☐ |
| M6.4 | 任务列表：NextRunTime | ☐ |
| M6.5 | 任务路径 | ☐ |
| M6.6 | 任务操作（命令行参数） | ☐ |
| M6.7 | 异常检测：隐藏任务 | ☐ |
| M6.8 | 异常检测：异常路径任务（temp, appdata） | ☐ |
| M6.9 | XML 导出任务配置 | ☐ |
| M6.10 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m6_schedule/schedule.go`

---

## M7 系统监控

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M7.1 | CPU 监控：实时使用率 | ☐ |
| M7.2 | CPU 监控：历史峰值 | ☐ |
| M7.3 | CPU 监控：每核心使用率 | ☐ |
| M7.4 | 内存监控：使用量/总量 | ☐ |
| M7.5 | 内存监控：换页率 | ☐ |
| M7.6 | 磁盘监控：读写速度 | ☐ |
| M7.7 | 磁盘监控：空间使用率 | ☐ |
| M7.8 | 磁盘监控：每个分区的使用情况 | ☐ |
| M7.9 | 网络监控：流量实时曲线 | ☐ |
| M7.10 | 网络监控：各网卡流量 | ☐ |
| M7.11 | 告警规则：CPU 阈值 | ☐ |
| M7.12 | 告警规则：内存阈值 | ☐ |
| M7.13 | 告警规则：磁盘空间阈值 | ☐ |
| M7.14 | 告警规则：网络流量阈值 | ☐ |
| M7.15 | 数据导出：JSON 格式 | ☐ |
| M7.16 | 数据导出：CSV 格式 | ☐ |

### 实现文件
- `internal/modules/m7_monitor/monitor.go`

---

## M8 系统补丁

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M8.1 | 已安装补丁列表：HotFixID（KB编号） | ☐ |
| M8.2 | 已安装补丁列表：Description（描述） | ☐ |
| M8.3 | 已安装补丁列表：InstalledOn（安装日期） | ☐ |
| M8.4 | 安全补丁识别 | ☐ |
| M8.5 | 缺失补丁检测：DetectMissingPatches(kbList) | ☐ |
| M8.6 | 缺失补丁检测：DetectMissingPatchesFromMicrosoft() | ☐ |
| M8.7 | 漏洞关联：CVE 编号 | ☐ |
| M8.8 | 补丁来源信息 | ☐ |
| M8.9 | 补丁大小 | ☐ |
| M8.10 | 更新回滚检测 | ☐ |
| M8.11 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m8_patch/patch.go`

---

## M9 软件列表

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M9.1 | 软件列表：DisplayName | ☐ |
| M9.2 | 软件列表：Version | ☐ |
| M9.3 | 软件列表：Publisher | ☐ |
| M9.4 | 软件列表：InstallDate | ☐ |
| M9.5 | 软件安装路径 | ☐ |
| M9.6 | 异常检测：无可信签名的软件 | ☐ |
| M9.7 | 异常检测：可疑软件（potentially unwanted） | ☐ |
| M9.8 | 异常检测：无版本信息的软件 | ☐ |
| M9.9 | Search(keyword) 搜索 | ☐ |
| M9.10 | 导出功能 | ☐ |

### 实现文件
- `internal/modules/m9_software/software.go`

---

## M10 内核分析

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M10.1 | 驱动列表：DriverName | ☐ |
| M10.2 | 驱动列表：Path（驱动文件路径） | ☐ |
| M10.3 | 驱动列表：BaseAddress（基地址） | ☐ |
| M10.4 | 驱动列表：Size（内存大小） | ☐ |
| M10.5 | 驱动签名验证：Authenticode 签名状态 | ☐ |
| M10.6 | 驱动签名验证：签名者信息 | ☐ |
| M10.7 | 异常驱动检测：无签名驱动 | ☐ |
| M10.8 | 异常驱动检测：签名无效驱动 | ☐ |
| M10.9 | 异常驱动检测：可疑名称（rootkit, keylog 等） | ☐ |
| M10.10 | 驱动文件版本信息 | ☐ |
| M10.11 | 驱动公司名称 | ☐ |
| M10.12 | 过滤驱动检测（fltmc） | ☐ |
| M10.13 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m10_kernel/kernel.go`

---

## M11 文件系统

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M11.1 | 目录扫描：指定目录递归扫描 | ☐ |
| M11.2 | 目录扫描：深度限制（maxDepth） | ☐ |
| M11.3 | 文件枚举：FileName | ☐ |
| M11.4 | 文件枚举：FilePath | ☐ |
| M11.5 | 文件枚举：FileSize | ☐ |
| M11.6 | 文件时间：创建时间（CreationTime） | ☐ |
| M11.7 | 文件时间：修改时间（ModifyTime） | ☐ |
| M11.8 | 文件时间：访问时间（AccessTime） | ☐ |
| M11.9 | 文件属性：Hidden | ☐ |
| M11.10 | 文件属性：System | ☐ |
| M11.11 | 文件属性：ReadOnly | ☐ |
| M11.12 | 文件哈希：MD5 | ☐ |
| M11.13 | 文件哈希：SHA1 | ☐ |
| M11.14 | 文件哈希：SHA256 | ☐ |
| M11.15 | Authenticode 签名验证 | ☐ |
| M11.16 | 签名者信息提取 | ☐ |
| M11.17 | 替代数据流（ADS）检测 | ☐ |
| M11.18 | 文件权限检查（ACL） | ☐ |
| M11.19 | 快捷方式（LNK）解析 | ☐ |
| M11.20 | 敏感文件监控：SAM | ☐ |
| M11.21 | 敏感文件监控：SYSTEM | ☐ |
| M11.22 | 敏感文件监控：SECURITY | ☐ |
| M11.23 | 大文件标记（>100MB） | ☐ |
| M11.24 | 大文件流式读取（GB级） | ☐ |
| M11.25 | 文件搜索：按名称 | ☐ |
| M11.26 | 文件搜索：按大小范围 | ☐ |
| M11.27 | 文件搜索：按日期范围 | ☐ |
| M11.28 | 风险检测：可疑扩展名（.vbs, .js, .scr） | ☐ |
| M11.29 | 风险检测：temp/tmp 目录文件 | ☐ |
| M11.30 | GetFileHash(path) 获取单个文件哈希 | ☐ |
| M11.31 | ScanPath(path, recursive) 路径扫描 | ☐ |
| M11.32 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m11_filesystem/filesystem.go`

---

## M12 活动痕迹

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M12.1 | 最近打开文件：Recent 文件夹解析 | ☐ |
| M12.2 | 最近打开文件：文件名 | ☐ |
| M12.3 | 最近打开文件：访问时间 | ☐ |
| M12.4 | LNK 快捷方式解析：实际路径 | ☐ |
| M12.5 | LNK 快捷方式解析：创建时间 | ☐ |
| M12.6 | USB 使用：设备名称 | ☐ |
| M12.7 | USB 使用：DeviceID | ☐ |
| M12.8 | USB 使用：最后插入时间 | ☐ |
| M12.9 | USB 使用：序列号 | ☐ |
| M12.10 | 浏览器历史：Chrome 历史读取 | ☐ |
| M12.11 | 浏览器历史：Edge 历史读取 | ☐ |
| M12.12 | 浏览器历史：Firefox 历史读取 | ☐ |
| M12.13 | 浏览器历史：URL | ☐ |
| M12.14 | 浏览器历史：Title | ☐ |
| M12.15 | 浏览器历史：访问时间 | ☐ |
| M12.16 | 浏览器历史：访问次数 | ☐ |
| M12.17 | 文件操作记录：创建 | ☐ |
| M12.18 | 文件操作记录：修改 | ☐ |
| M12.19 | 文件操作记录：删除 | ☐ |
| M12.20 | RDP 连接历史 | ☐ |
| M12.21 | 网络共享访问历史 | ☐ |
| M12.22 | 打印历史 | ☐ |
| M12.23 | 风险检测：可疑路径 | ☐ |
| M12.24 | Search(keyword) 搜索 | ☐ |
| M12.25 | 数据导出 | ☐ |

### 实现文件
- `internal/modules/m12_activity/activity.go`

---

## M13 日志分析

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M13.1 | 安全日志 Security | ☐ |
| M13.2 | 系统日志 System | ☐ |
| M13.3 | 应用程序日志 Application | ☐ |
| M13.4 | DNS Server 日志 | ☐ |
| M13.5 | 日志筛选：按 Level（Error, Warning, Info） | ☐ |
| M13.6 | 日志筛选：按 Source | ☐ |
| M13.7 | 日志筛选：按 Time | ☐ |
| M13.8 | 日志筛选：按 EventID | ☐ |
| M13.9 | 关键词搜索：全文检索 | ☐ |
| M13.10 | 日志解析：EventID | ☐ |
| M13.11 | 日志解析：TimeCreated | ☐ |
| M13.12 | 日志解析：Message | ☐ |
| M13.13 | 日志解析：RawXML | ☐ |
| M13.14 | EVTX 文件解析 | ☐ |
| M13.15 | 一键分析：4624（登录成功） | ☐ |
| M13.16 | 一键分析：4625（登录失败） | ☐ |
| M13.17 | 一键分析：4634（注销） | ☐ |
| M13.18 | 一键分析：4672（特殊登录） | ☐ |
| M13.19 | 一键分析：4688（进程创建） | ☐ |
| M13.20 | 一键分析：4698（计划任务创建） | ☐ |
| M13.21 | 导出报告：HTML | ☐ |
| M13.22 | 导出报告：JSON | ☐ |

### 实现文件
- `internal/modules/m13_logging/logging.go`

---

## M14 账户分析

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M14.1 | 本地账户：AccountName | ☐ |
| M14.2 | 本地账户：SID | ☐ |
| M14.3 | 本地账户：Status（Enabled/Disabled） | ☐ |
| M14.4 | 本地账户：LastLogon | ☐ |
| M14.5 | 本地账户：PasswordLastSet | ☐ |
| M14.6 | 本地账户：PasswordExpired | ☐ |
| M14.7 | 账户组：GroupName | ☐ |
| M14.8 | 账户组：Members（成员列表） | ☐ |
| M14.9 | 特殊账户：Guest | ☐ |
| M14.10 | 特殊账户：Administrator | ☐ |
| M14.11 | 特殊账户：隐藏账户检测 | ☐ |
| M14.12 | SID 解析：SID -> AccountName | ☐ |
| M14.13 | SID 解析：账户类型（User, Group, Domain） | ☐ |
| M14.14 | 风险检测：空密码账户 | ☐ |
| M14.15 | 风险检测：永不过期密码 | ☐ |
| M14.16 | 风险检测：Guest 启用 | ☐ |
| M14.17 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m14_account/account.go`

---

## M15 内存取证

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M15.1 | 进程内存 Dump：指定 PID | ☐ |
| M15.2 | 进程内存 Dump：MiniDump 格式 | ☐ |
| M15.3 | 进程内存 Dump：FullDump 格式 | ☐ |
| M15.4 | 系统内存 Dump：完整物理内存 | ☐ |
| M15.5 | 系统内存 Dump：休眠文件解析 | ☐ |
| M15.6 | Dump 列表：历史记录 | ☐ |
| M15.7 | Dump 文件：SHA256 校验 | ☐ |
| M15.8 | Dump 文件：大小 | ☐ |
| M15.9 | Dump 文件：创建时间 | ☐ |
| M15.10 | Dump 操作：KillProcess 后自动 Dump | ☐ |
| M15.11 | 内存分析：strings 提取 | ☐ |
| M15.12 | 内存分析：YARA 规则匹配 | ☐ |
| M15.13 | Dump 管理：删除 Dump | ☐ |
| M15.14 | Dump 管理：导出 Dump | ☐ |

### 实现文件
- `internal/modules/m15_memory/memory.go`

---

## M16 威胁检测

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M16.1 | 恶意进程检测：哈希黑名单匹配 | ☐ |
| M16.2 | 恶意进程检测：进程名称黑名单 | ☐ |
| M16.3 | 恶意进程检测：路径黑名单 | ☐ |
| M16.4 | 恶意进程检测：命令行特征 | ☐ |
| M16.5 | 可疑网络检测：恶意 IP 黑名单 | ☐ |
| M16.6 | 可疑网络检测：恶意域名黑名单 | ☐ |
| M16.7 | 敏感行为：注册表修改（Run键） | ☐ |
| M16.8 | 敏感行为：服务创建 | ☐ |
| M16.9 | 敏感行为：计划任务创建 | ☐ |
| M16.10 | 敏感行为：远程线程创建 | ☐ |
| M16.11 | 威胁情报：离线哈希库 | ☐ |
| M16.12 | 威胁情报：YARA 规则 | ☐ |
| M16.13 | 告警生成：AlertEvent | ☐ |
| M16.14 | 告警去重 | ☐ |
| M16.15 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m16_threat/threat.go`

---

## M17 应急处置

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M17.1 | 进程查杀：KillProcess(pid) | ☐ |
| M17.2 | 进程查杀：强制终止 | ☐ |
| M17.3 | 进程查杀：带确认对话框 | ☐ |
| M17.4 | 文件隔离：移动到隔离区 | ☐ |
| M17.5 | 文件隔离：备份原文件 | ☐ |
| M17.6 | 网络断开：断开连接 | ☐ |
| M17.7 | 服务禁用：Stop + 禁用启动类型 | ☐ |
| M17.8 | 注册表修复：恢复被篡改的键值 | ☐ |
| M17.9 | 二次确认：操作确认对话框 | ☐ |
| M17.10 | 审计日志：操作人 | ☐ |
| M17.11 | 审计日志：操作时间 | ☐ |
| M17.12 | 审计日志：操作对象 | ☐ |
| M17.13 | 审计日志：操作结果 | ☐ |
| M17.14 | 关键进程保护：lsass.exe | ☐ |
| M17.15 | 关键进程保护：winlogon.exe | ☐ |
| M17.16 | 关键进程保护：csrss.exe | ☐ |
| M17.17 | 关键进程保护：smss.exe | ☐ |
| M17.18 | 关键进程保护：services.exe | ☐ |
| M17.19 | 关键进程保护：system | ☐ |

### 实现文件
- `internal/modules/m17_response/response.go`

---

## M18 自启动项目

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M18.1 | 注册表自启动：Run 键 | ☐ |
| M18.2 | 注册表自启动：RunOnce 键 | ☐ |
| M18.3 | 注册表自启动：RunOnceEx 键 | ☐ |
| M18.4 | 启动文件夹：当前用户 Startup | ☐ |
| M18.5 | 启动文件夹：All Users Startup | ☐ |
| M18.6 | 计划任务自启动： | ☐ |
| M18.7 | 服务自启动：Auto Start 服务 | ☐ |
| M18.8 | WMI 自启动：Event Subscriber | ☐ |
| M18.9 | Winlogon 自启动 | ☐ |
| M18.10 | AppInit_DLLs 自启动 | ☐ |
| M18.11 | IE 浏览器扩展 | ☐ |
| M18.12 | 快捷方式劫持 | ☐ |
| M18.13 | 风险评估 | ☐ |
| M18.14 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m18_autostart/autostart.go`

---

## M19 域控检测专项

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M19.1 | 域信息：DomainName | ☐ |
| M19.2 | 域信息：DomainController | ☐ |
| M19.3 | 域用户：UserPrincipalName | ☐ |
| M19.4 | 域用户：sAMAccountName | ☐ |
| M19.5 | 域用户：SID | ☐ |
| M19.6 | 域用户：MemberOf（所属组） | ☐ |
| M19.7 | 域用户：LastLogon | ☐ |
| M19.8 | 域用户：pwdLastSet | ☐ |
| M19.9 | 域组：GroupName | ☐ |
| M19.10 | 域组：Description | ☐ |
| M19.11 | 域组：Members | ☐ |
| M19.12 | OU 结构：OrganizationalUnit | ☐ |
| M19.13 | OU 结构：GPO 链接 | ☐ |
| M19.14 | GPO：Group Policy Objects | ☐ |
| M19.15 | GPO：安全设置 | ☐ |
| M19.16 | 离线降级：LDAP 不可用时 | ☐ |
| M19.17 | 信任关系：双向/单向信任 | ☐ |
| M19.18 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m19_domain/domain.go`

---

## M20 域内渗透检测

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M20.1 | Kerberoasting：SPN 账户统计 | ☐ |
| M20.2 | Kerberoasting：TGS 请求分析 | ☐ |
| M20.3 | AS-REP Roasting：可疑 AS-REP 响应 | ☐ |
| M20.4 | Golden Ticket：TGT 异常检测 | ☐ |
| M20.5 | Silver Ticket：ST 异常检测 | ☐ |
| M20.6 | 账户异常：密码错误次数 | ☐ |
| M20.7 | 账户异常：账户锁定 | ☐ |
| M20.8 | 权限提升：敏感组成员变化 | ☐ |
| M20.9 | 权限提升：Domain Admin 组变化 | ☐ |
| M20.10 | DCSync 攻击检测 | ☐ |
| M20.11 | Pass-the-Hash 检测 | ☐ |
| M20.12 | 远程连接：RDP 横向移动 | ☐ |
| M20.13 | 远程连接：WMI 横向移动 | ☐ |
| M20.14 | 远程连接：PSRemoting | ☐ |
| M20.15 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m20_domainhack/domainhack.go`

---

## M21 WMIC 持久化专项

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M21.1 | WMIC 历史记录读取 | ☐ |
| M21.2 | WMIC 命令：process call create | ☐ |
| M21.3 | WMIC 命令：service name | ☐ |
| M21.4 | WMIC 命令：os get | ☐ |
| M21.5 | WMIC 命令：computersystem | ☐ |
| M21.6 | 可疑命令：删除文件 | ☐ |
| M21.7 | 可疑命令：格式化磁盘 | ☐ |
| M21.8 | 可疑命令：停止服务 | ☐ |
| M21.9 | 批量检测：常见 WMIC 攻击命令 | ☐ |
| M21.10 | 时间线重建 | ☐ |
| M21.11 | Search(keyword) 搜索 | ☐ |

### 实现文件
- `internal/modules/m21_wmic/wmic.go`

---

## M22 报告导出

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M22.1 | HTML 报告：模板引擎 | ☐ |
| M22.2 | HTML 报告：样式美化 | ☐ |
| M22.3 | HTML 报告：图表嵌入 | ☐ |
| M22.4 | HTML 报告：Logo 嵌入 | ☐ |
| M22.5 | PDF 报告：生成功能 | ☐ |
| M22.6 | PDF 报告：分页 | ☐ |
| M22.7 | PDF 报告：目录 | ☐ |
| M22.8 | JSON 导出：完整数据 | ☐ |
| M22.9 | JSON 导出：压缩格式 | ☐ |
| M22.10 | 会话对比报告：Session1 vs Session2 | ☐ |
| M22.11 | 会话对比报告：新增进程 | ☐ |
| M22.12 | 会话对比报告：新增网络连接 | ☐ |
| M22.13 | 会话对比报告：新增注册表 | ☐ |
| M22.14 | 会话对比报告：新增服务 | ☐ |
| M22.15 | 报告加密：AES256 | ☐ |
| M22.16 | 报告签名：数字签名 | ☐ |

### 实现文件
- `internal/modules/m22_report/report.go`

---

## M23 安全基线检查

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M23.1 | 密码策略：最小长度 | ☐ |
| M23.2 | 密码策略：复杂度要求 | ☐ |
| M23.3 | 密码策略：过期时间 | ☐ |
| M23.4 | 密码历史 | ☐ |
| M23.5 | 账户锁定：锁定阈值 | ☐ |
| M23.6 | 账户锁定：锁定时间 | ☐ |
| M23.7 | 审核策略：登录成功审核 | ☐ |
| M23.8 | 审核策略：登录失败审核 | ☐ |
| M23.9 | 审核策略：进程创建审核 | ☐ |
| M23.10 | 审核策略：对象访问审核 | ☐ |
| M23.11 | 网络安全：SMBv1 状态 | ☐ |
| M23.12 | 网络安全：SMBv2 状态 | ☐ |
| M23.13 | 网络安全：防火墙状态 | ☐ |
| M23.14 | 网络安全：远程桌面状态 | ☐ |
| M23.15 | 服务配置：Remote Registry | ☐ |
| M23.16 | 服务配置：Telnet | ☐ |
| M23.17 | 服务配置：Windows Update | ☐ |
| M23.18 | 服务配置：不必要的服务 | ☐ |
| M23.19 | 注册表安全：UAC 设置 | ☐ |
| M23.20 | 注册表安全：LM Hash 存储 | ☐ |
| M23.21 | 基线评分：百分比 | ☐ |
| M23.22 | 基线评分：分类评分 | ☐ |
| M23.23 | 导出报告 | ☐ |

### 实现文件
- `internal/modules/m23_baseline/baseline.go`

---

## M24 IIS/中间件日志

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M24.1 | IIS 日志：W3C 格式解析 | ☐ |
| M24.2 | IIS 日志：时间戳 | ☐ |
| M24.3 | IIS 日志：客户端 IP | ☐ |
| M24.4 | IIS 日志：请求方法 | ☐ |
| M24.5 | IIS 日志：请求 URL | ☐ |
| M24.6 | IIS 日志：状态码 | ☐ |
| M24.7 | IIS 日志：子状态码 | ☐ |
| M24.8 | IIS 日志：字节数 | ☐ |
| M24.9 | IIS 日志：User-Agent | ☐ |
| M24.10 | Apache/Nginx：访问日志解析 | ☐ |
| M24.11 | SQL Server：MSSQL 日志解析 | ☐ |
| M24.12 | 日志统计：访问量 | ☐ |
| M24.13 | 日志统计：状态码分布 | ☐ |
| M24.14 | 日志统计：IP 访问排行 | ☐ |
| M24.15 | 日志统计：URL 访问排行 | ☐ |
| M24.16 | 异常检测：大量 404 | ☐ |
| M24.17 | 异常检测：大量 500 | ☐ |
| M24.18 | 异常检测：SQL 注入尝试 | ☐ |
| M24.19 | 异常检测：XSS 尝试 | ☐ |
| M24.20 | 异常检测：目录遍历尝试 | ☐ |
| M24.21 | 地理统计：IP 归属地 | ☐ |
| M24.22 | 导出报告 | ☐ |

### 实现文件
- `internal/modules/m24_iis/iis.go`

---

## M25 编解码工具

### 功能需求
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| M25.1 | Base64 标准编码 | ☐ |
| M25.2 | Base64 标准解码 | ☐ |
| M25.3 | Base64 URL 安全编码 | ☐ |
| M25.4 | Base64 URL 安全解码 | ☐ |
| M25.5 | Hex 编码 | ☐ |
| M25.6 | Hex 解码 | ☐ |
| M25.7 | Unicode 编码（\uXXXX） | ☐ |
| M25.8 | Unicode 解码 | ☐ |
| M25.9 | Unicode 编码（&#xXXXX;） | ☐ |
| M25.10 | URL 编码 | ☐ |
| M25.11 | URL 解码 | ☐ |
| M25.12 | HTML 实体编码（&lt; &gt; &amp;） | ☐ |
| M25.13 | HTML 实体解码 | ☐ |
| M25.14 | HTML 数字编码（&#60;） | ☐ |
| M25.15 | Binary 二进制字符串 | ☐ |
| M25.16 | Octal 八进制 | ☐ |
| M25.17 | 自动检测编码类型 | ☐ |
| M25.18 | 历史记录：转换历史 | ☐ |
| M25.19 | 批量编解码 | ☐ |
| M25.20 | 编码串联（Base64 -> Hex -> URL） | ☐ |

### 实现文件
- `internal/modules/m25_codec/codec.go`

---

## 核心引擎检查

### 存储引擎
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| C1 | SQLite WAL 模式 | ☐ |
| C2 | Ristretto 缓存 | ☐ |
| C3 | Checkpoint 崩溃恢复 | ☐ |
| C4 | 高并发读写支持 | ☐ |

### 任务调度
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| C5 | 分级线程池：P0 优先级 | ☐ |
| C6 | 分级线程池：P1 优先级 | ☐ |
| C7 | 分级线程池：P2 优先级 | ☐ |
| C8 | 任务老化机制 | ☐ |
| C9 | 任务超时看门狗 | ☐ |
| C10 | 熔断机制 | ☐ |

### 会话管理
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| C11 | 会话创建 | ☐ |
| C12 | 会话状态跟踪 | ☐ |
| C13 | 会话进度追踪 | ☐ |
| C14 | 会话对比分析 | ☐ |
| C15 | 会话导出 | ☐ |

### 攻击时间线
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| C16 | TimelineEvent 事件收集 | ☐ |
| C17 | 时间线重建 | ☐ |
| C18 | 时间线可视化 | ☐ |
| C19 | 事件关联分析 | ☐ |

### 风险引擎
| 功能点 | 描述 | 检查状态 |
|--------|------|----------|
| C20 | RiskLevel 评估 | ☐ |
| C21 | 风险评分计算 | ☐ |
| C22 | 风险告警 | ☐ |

---

## 使用说明

1. 打印此文档
2. 对每个检查项进行验证
3. 在检查状态列标记：✅ 已实现 / ❌ 未实现 / ⚠️ 部分实现
4. 统计完成率

### 完成率计算

- M1-M9: __/90 项 = __%
- M10-M18: __/119 项 = __%
- M19-M25: __/107 项 = __%
- 核心引擎: __/22 项 = __%

**总计: __/338 项 = __%**
