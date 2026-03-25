# Windows 应急响应工具 (ERT) 用户手册

## 目录

1. [简介](#简介)
2. [功能特性](#功能特性)
3. [系统要求](#系统要求)
4. [安装指南](#安装指南)
5. [快速开始](#快速开始)
6. [模块功能详解](#模块功能详解)
7. [命令行使用](#命令行使用)
8. [配置文件](#配置文件)
9. [常见问题](#常见问题)
10. [技术架构](#技术架构)

---

## 简介

Windows 应急响应工具 (ERT) 是一款基于 Go 语言开发的 Windows 系统应急响应工具，采用 Wails v2 + Vue3 + Element Plus 前后端分离架构，支持 **25 个独立功能模块**。

### 版本信息

- 版本号：v13.0
- 构建日期：2026-03-25

---

## 功能特性

### 25 个功能模块

| ID | 模块名称 | 功能描述 | 优先级 |
|----|---------|---------|--------|
| M1 | 系统概览 | 主机信息、资源监控、实时图表 | P0 |
| M2 | 进程管理 | 进程列表/树、查杀、Dump | P0 |
| M3 | 网络分析 | 连接列表、端口监听、IP 地理 | P0 |
| M4 | 注册表分析 | 关键项检测、持久化、自启动 | P0 |
| M5 | 服务管理 | 服务列表、启停操作 | P0 |
| M6 | 计划任务 | 任务列表、异常检测 | P0 |
| M7 | 系统监控 | CPU/内存/磁盘/网络实时监控 | P0 |
| M8 | 系统补丁 | 已安装补丁、缺失补丁 | P1 |
| M9 | 软件列表 | 已安装软件、异常检测 | P0 |
| M10 | 内核分析 | 驱动列表、签名状态 | P1 |
| M11 | 文件系统 | 文件枚举、哈希、大文件处理 | P0 |
| M12 | 活动痕迹 | 最近打开、USB使用、浏览器历史 | P0 |
| M13 | 日志分析 | 事件日志、EVTX解析、全文搜索 | P0 |
| M14 | 账户分析 | 本地/域账户、组、权限 | P0 |
| M15 | 内存取证 | 进程/系统内存 Dump | P1 |
| M16 | 威胁检测 | 恶意进程、可疑网络、敏感行为 | P0 |
| M17 | 应急处置 | 进程查杀、文件隔离、审计日志 | P1 |
| M18 | 自启动项目 | 注册表/启动文件夹/服务/WMI | P0 |
| M19 | 域控检测专项 | 域用户/组/OU/GPO、离线降级 | P0 |
| M20 | 域内渗透检测 | Kerberoasting、Golden Ticket | P0 |
| M21 | WMIC 持久化专项 | WMIC 命令历史检测 | P0 |
| M22 | 报告导出 | HTML/PDF/JSON 导出 | P0 |
| M23 | 安全基线检查 | 密码/账户/审核/网络安全 | P0 |
| M24 | IIS/中间件日志 | IIS/Apache/SQL Server 日志 | P0 |
| M25 | 编解码工具 | Base64/Hex/Unicode/URL/HTML | P0 |

### 核心引擎

- **SQLite 存储** - WAL 模式，支持高并发读写
- **Ristretto 缓存** - 高性能内存缓存
- **熔断机制** - 防止故障扩散
- **看门狗监控** - 任务超时控制
- **分级线程池** - P0/P1/P2 优先级调度
- **任务老化机制** - 防止低优先级任务饿死
- **风险引擎** - 自动化风险评估
- **进度管理** - 实时进度追踪
- **攻击时间线** - 安全事件时间线重建
- **会话对比** - 多会话对比分析

---

## 系统要求

### 硬件要求

| 组件 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 1 核心 | 4 核心以上 |
| 内存 | 2 GB | 8 GB 以上 |
| 磁盘 | 500 MB | 10 GB 以上 |
| 显示器 | 1024x768 | 1920x1080 |

### 软件要求

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| 操作系统 | Windows 10/11, Windows Server 2016+ | 仅支持 64 位 |
| WebView2 | 最新版本 | 预安装在 Windows 10/11 |
| 管理员权限 | 必须 | 部分功能需要 |

---

## 安装指南

### 方法一：从头构建

#### 1. 克隆代码

```bash
git clone https://github.com/kkkdddd-start/winemy.git
cd winemy
```

#### 2. 安装依赖

```bash
# 安装 Go 1.21+
# https://go.dev/dl/

# 安装 Node.js 18+
# https://nodejs.org/

# 安装 pnpm (可选)
npm install -g pnpm

# 安装前端依赖
cd app && pnpm install && cd ..
```

#### 3. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 4. 构建应用

```bash
# 开发模式运行
wails dev

# 生产构建
wails build
```

### 方法二：直接编译

```bash
# 设置 Windows 目标平台
set GOOS=windows
set GOARCH=amd64

# 编译
go build -ldflags="-s -w -H=windowsgui -trimpath" -o ert.exe ./cmd/gui/
```

### 一键打包脚本

项目已包含自动化打包脚本，见 [打包脚本](#打包脚本)

---

## 快速开始

### 首次启动

1. 双击 `ert.exe` 启动程序
2. 如果提示 WebView2 未安装，请先安装 WebView2 运行时
3. 程序启动后显示主界面

### 主界面布局

```
┌─────────────────────────────────────────────────────────┐
│  ERT v13.0                                    [≡]    │
├─────────┬─────────────────────────────────────────────┤
│         │                                             │
│ M1 系统 │                                             │
│ M2 进程 │                                             │
│ M3 网络 │           主内容区域                        │
│ M4 注册 │                                             │
│ M5 服务 │                                             │
│ ...     │                                             │
│         │                                             │
├─────────┴─────────────────────────────────────────────┤
│ 状态栏: 就绪 | 会话: abc123 | 时间: 2026-03-25       │
└─────────────────────────────────────────────────────┘
```

### 快捷键

| 快捷键 | 功能 | 适用范围 |
|--------|------|----------|
| `Ctrl+Shift+T` | 一键 Triage 采集 | 全局 |
| `Ctrl+E` | 全局搜索 | 全局 |
| `Ctrl+S` | 数据导出 | 全局 |
| `Ctrl+R` | 刷新当前模块 | 全局 |
| `Ctrl+F` | 页面内搜索 | 当前视图 |
| `F5` | 刷新 | 全局 |
| `F11` | 全屏切换 | 全局 |
| `Esc` | 取消/关闭 | 全局 |
| `Ctrl+1~9` | 切换到模块 M1~M9 | 全局 |

---

## 模块功能详解

### M1 系统概览

显示主机基本信息和关键指标的总览面板。

**核心功能**：
- 主机基本信息（计算机名、操作系统、当前用户、启动时间）
- 系统资源（CPU 使用率、内存使用率、磁盘使用率）
- 网络状态（网卡信息、IP 地址、网络连接数）
- 实时监控图表

### M2 进程管理

进程查看、分析、处置。

**核心功能**：
- 进程列表：PID、名称、路径、用户、CPU、内存、启动时间
- 进程树：父子进程关系树形展示
- 进程搜索：按名称/PID/路径搜索
- 风险标记：无签名标黄，可疑进程标红
- 进程查杀：终止进程（需管理员+二次确认）
- 进程 Dump：导出进程内存用于分析

### M3 网络分析

网络连接查看、异常检测。

**核心功能**：
- 连接列表：协议、本地地址/端口、远程地址/端口、状态、PID
- 监听端口：所有监听中的端口及对应进程
- IP 地理位置：显示远程 IP 的地理位置
- 风险检测：可疑端口、可疑外部连接

### M4 注册表分析

注册表关键项检测、持久化分析。

**重点检测路径**：
```
HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run
HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce
HKLM\SYSTEM\CurrentControlSet\Services
HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run
```

### M5 服务管理

Windows 服务查看、分析。

**核心功能**：
- 服务列表：名称、显示名、状态、启动类型、路径
- 服务详情：依赖关系、描述、触发条件
- 风险检测：异常服务、禁用安全服务
- 服务操作：启动/停止/重启（需管理员+二次确认）

### M6 计划任务

计划任务查看、持久化检测。

**核心功能**：
- 任务列表：任务名称、状态、上次运行、下次运行
- 异常检测：隐藏任务、异常路径任务
- XML 导出：导出任务配置用于分析

### M7 系统监控

实时系统状态监控、告警。

**核心功能**：
- CPU 监控：使用率实时曲线、历史峰值
- 内存监控：使用量/总量、换页率
- 磁盘监控：读写速度、空间使用
- 网络监控：流量实时曲线
- 告警规则：自定义阈值告警

### M8 系统补丁

系统补丁查看、漏洞检测。

**核心功能**：
- 已安装补丁：KB 编号、描述、安装日期
- 缺失补丁：已知的重要安全补丁
- 漏洞关联：补丁对应的 CVE 漏洞

### M9 软件列表

已安装软件查看、异常检测。

**核心功能**：
- 软件列表：名称、版本、发布者、安装日期
- 安装位置：软件安装路径
- 异常检测：可疑软件、无版本软件

### M10 内核分析

内核对象查看（用户态降级）。

**核心功能**：
- SSDT 查看：系统服务描述符表
- 驱动列表：已加载内核驱动
- 驱动签名：驱动签名状态
- 异常驱动：无签名驱动、签名无效驱动

### M11 文件系统

文件分析、取证。

**核心功能**：
- 文件枚举：目录浏览、文件列表
- 文件详情：大小、创建/修改/访问时间、属性
- 文件搜索：按名称/大小/日期范围搜索
- 文件哈希：MD5、SHA1、SHA256 计算
- 大文件处理：GB 级文件流式读取

### M12 活动痕迹

用户操作痕迹检测。

**核心功能**：
- 最近打开：最近打开的文档、程序
- USB 使用：USB 设备使用记录
- 网络浏览：浏览器历史记录
- 文件操作：文件创建、修改、删除记录

### M13 日志分析

Windows 事件日志分析。

**核心功能**：
- 事件日志：安全、系统、应用日志
- 日志筛选：按级别/来源/时间/事件ID
- 关键词搜索：日志内容全文搜索
- 一键分析：自动分析可疑日志并导出报告

### M14 账户分析

用户账户检测、权限分析。

**核心功能**：
- 本地账户：账户列表、最后登录
- 账户组：用户组及成员
- 特殊账户：Guest、Administrator、隐藏账户
- SID 分析：账户 SID 解析

### M15 内存取证

内存 dump 和分析。

**核心功能**：
- 进程内存 Dump：指定进程内存导出
- 系统内存 Dump：完整系统内存（需管理员）
- Dump 列表：历次 Dump 记录
- 完整性校验：SHA256 哈希校验

### M16 威胁检测

基于威胁情报的检测。

**核心功能**：
- 恶意进程检测：基于哈希的恶意进程检测
- 可疑网络检测：恶意 IP/域名连接
- 敏感行为：关键注册表修改、关键目录访问
- 威胁情报：离线威胁情报库匹配

### M17 应急处置

安全的处置操作。

**核心功能**：
- 进程查杀：终止恶意进程
- 文件隔离：移动恶意文件到隔离区
- 网络断开：断开网络连接
- 服务禁用：禁用恶意服务
- 注册表修复：恢复被篡改的注册表

**安全机制**：
- 二次确认：所有操作需要前端弹窗确认
- 审计日志：记录操作人、时间、对象、结果
- 关键保护：禁止查杀系统关键进程（lsass.exe, winlogon.exe 等）

### M18 自启动项目

持久化驻留检测。

**核心功能**：
- 注册表自启动：Run、RunOnce 等键值
- 启动文件夹：启动目录快捷方式
- 计划任务：自启动计划任务
- 服务自启动：自启动服务
- WMI 自启动：WMI Event Subscriber

### M19 域控检测专项

Active Directory 检测。

**核心功能**：
- 域信息：域名称、域控制器
- 域用户：域用户列表
- 域组：域组及成员
- OU 结构：组织单位结构
- GPO：组策略对象
- 离线降级：LDAP 不可用时降级为本地分析

### M20 域内渗透检测

Kerberos 攻击检测。

**核心功能**：
- Kerberoasting：SPN 账户请求统计
- AS-REP Roasting：可疑 AS-REP 响应
- Golden Ticket：TGT 异常检测
- Silver Ticket：ST 异常检测
- 账户异常：大量密码错误、账户锁定
- 权限提升：敏感组成员变化

### M21 WMIC 持久化专项

WMIC 命令历史检测。

**核心功能**：
- WMIC 历史：WMIC 命令执行历史
- 可疑命令：创建进程、删除文件等
- 批量检测：常见 WMIC 攻击命令

### M22 报告导出

生成分析报告。

**核心功能**：
- HTML 报告：生成 HTML 格式报告
- PDF 报告：生成 PDF 格式报告
- JSON 导出：结构化数据导出
- 会话对比报告：两个会话对比分析报告

### M23 安全基线检查

安全配置检测。

**核心功能**：
- 密码策略：密码复杂度、长度、过期时间
- 账户策略：锁定阈值、登录尝试
- 审核策略：安全日志审核配置
- 网络安全：SMB 版本、防火墙状态
- 服务配置：不必要的服务

### M24 IIS/中间件日志

Web 服务器日志分析。

**核心功能**：
- IIS 日志：IIS 日志解析（W3C 格式）
- Apache/Nginx：访问日志解析
- SQL Server 日志：MSSQL 日志分析
- 日志统计：访问量、状态码、IP 统计
- 异常检测：大量 404、500 错误
- 攻击检测：SQL 注入、XSS 尝试

### M25 编解码工具

常用编码转换工具。

**核心功能**：
- Base64：标准/URL 安全 Base64
- Hex：十六进制转换
- Unicode：Unicode 编码/解码
- URL：URL 编码/解码
- HTML：HTML 实体编码/解码
- Binary：二进制字符串转换
- 自动检测：尝试所有解码器
- 历史记录：转换历史（默认开启）

---

## 命令行使用

### GUI 模式

```powershell
# 启动 GUI
.\ert.exe

# 指定配置文件
.\ert.exe -config "C:\ERT\config.yaml"
```

### CLI 模式

```powershell
# 查看帮助
.\ert-cli.exe --help

# 采集所有模块
.\ert-cli.exe collect --all --output ./output

# 采集指定模块
.\ert-cli.exe collect --module process --module network

# Triage 模式（快速采集关键指标）
.\ert-cli.exe triage --output ./triage

# 导出报告
.\ert-cli.exe export --session <session-id> --format html

# 执行单项检查
.\ert-cli.exe check --type suspicious_processes
```

### CLI 返回码

| 返回码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 参数错误 |
| 3 | 超时 |
| 4 | 权限不足 |
| 5 | 采集失败 |
| 6 | 输出错误 |

---

## 配置文件

配置文件位于 `config/config.yaml`。

### 配置结构

```yaml
app:
  name: "Windows 应急响应工具"
  version: "13.0"
  debug: false

server:
  host: "localhost"
  port: 9277

database:
  main:
    path: "./data/ert.db"
    wal_mode: true
    busy_timeout: 5000
    max_open_conns: 10
  cache:
    max_cost: 100 * 1024 * 1024
    buffer_size: 32 * 1024
    ttl: 10m

storage:
  data_dir: "./data"
  dump_dir: "./data/memory"
  report_dir: "./data/reports"
  max_storage: 50

concurrency:
  high_priority_workers: 10
  medium_priority_workers: 5
  low_priority_workers: 2
  aging:
    max_wait_time: 5m
    boost_threshold: 10
    priority_boost: 2
    check_interval: 30s

timeout:
  global: 5m
  module:
    process: 30s
    network: 30s
    registry: 60s
    filesystem: 300s
    logging: 600s
    memory_dump: 600s

security:
  readonly: true
  response:
    require_confirmation: true
    allow_kill_critical: false
    backup_before_action: true
    critical_processes:
      - "lsass.exe"
      - "winlogon.exe"
      - "csrss.exe"
      - "smss.exe"

ui:
  theme: "dark"
  language: "zh-CN"
  refresh_rate: 1s
  table:
    page_size: 50
    virtual_scroll: true
  shortcuts:
    enabled: true

log:
  level: "info"
  file: "./logs/ert.log"
  max_size: 500
  max_backups: 5
  max_age: 30
  compress: true
```

---

## 常见问题

### Q1: 程序启动提示 WebView2 未安装

**解决方法**：
1. 下载并安装 [WebView2 运行时](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
2. 如果是企业环境，可以使用离线安装包

### Q2: 部分功能提示需要管理员权限

**解决方法**：
1. 右键点击 `ert.exe`
2. 选择"以管理员身份运行"

### Q3: SQLite 数据库损坏

**解决方法**：
```powershell
# 运行修复命令（如果支持）
.\ert.exe --repair

# 或者删除数据库重新启动
Remove-Item "./data/ert.db"
.\ert.exe
```

### Q4: 内存占用过高

**解决方法**：
调整 `config.yaml` 中的缓存配置：
```yaml
database:
  cache:
    max_cost: 50 * 1024 * 1024  # 减少到 50MB
```

### Q5: 采集卡死

**解决方法**：
1. 检查 `config.yaml` 中的 timeout 配置
2. 使用 Triage 模式进行快速采集

---

## 技术架构

### 技术栈

| 层级 | 技术选型 | 版本 |
|------|----------|------|
| 后端框架 | Wails v2 | v2.7+ |
| 后端语言 | Go | 1.21+ |
| 前端框架 | Vue.js | 3.4+ |
| UI 组件库 | Element Plus | 2.4+ |
| 状态管理 | Pinia | 2.1+ |
| 图表库 | ECharts | 5.4+ |
| 数据库 | SQLite | v1.30+ |
| 缓存 | Ristretto | v0.2.0 |
| 日志 | Zap | v1.27+ |

### 项目结构

```
ert/
├── app/                      # 前端源码 (Vue3)
│   ├── src/
│   │   ├── components/       # 公共 UI 组件
│   │   ├── views/            # 25 个模块独立视图
│   │   ├── stores/           # 25 个独立 Pinia Store
│   │   └── router/          # 路由配置
│   └── package.json
├── internal/                  # 后端内部包
│   ├── modules/              # 25 个独立模块
│   │   ├── m1_system/       # 系统概览
│   │   ├── m2_process/      # 进程管理
│   │   └── ...              # 其他模块
│   ├── core/                 # 核心引擎
│   │   ├── storage/         # SQLite 存储
│   │   ├── cache/           # Ristretto 缓存
│   │   ├── logger/          # Zap 日志
│   │   ├── codec/           # 编解码引擎
│   │   └── ...
│   └── registry/             # 模块注册中心
├── config/
│   └── config.yaml          # 配置文件
├── data/                     # 数据目录
│   ├── ipdb/                # IP 地理位置库
│   ├── memory/              # 内存 dump 文件
│   └── reports/             # 报告导出
├── cmd/
│   └── gui/
│       └── main.go          # 应用入口
├── wails.json               # Wails 配置
├── go.mod
└── go.sum
```

### 数据存储

采用 MainDB + Cache + Checkpoint 三库架构：

- **MainDB (WAL 模式)**：主要数据存储，支持高并发读写
- **Ristretto 缓存**：高性能内存缓存，减少磁盘 IO
- **Checkpoint**：崩溃恢复点，保证数据完整性

---

## 联系方式

- 项目主页：https://github.com/kkkdddd-start/winemy
- 问题反馈：https://github.com/kkkdddd-start/winemy/issues

---

## 许可证

本项目仅供学习和研究使用。
