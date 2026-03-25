# Windows 应急响应工具 (ERT)

基于 Go + Wails 的 Windows 系统应急响应工具，支持 25 个独立功能模块。

## 项目状态

**开发中** - 基础框架搭建阶段

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21+ |
| GUI 框架 | Wails v2 |
| 前端 | Vue.js 3.4+ |
| UI 组件 | Element Plus |
| 数据库 | SQLite (modernc.org/sqlite) |

## 项目结构

```
ert/
├── cmd/gui/              # GUI 入口
├── internal/
│   ├── app/             # Wails App
│   ├── config/          # 配置管理
│   ├── core/            # 核心引擎
│   │   ├── storage/     # SQLite 存储
│   │   ├── cache/      # Ristretto 缓存
│   │   ├── concurrency/ # 线程池
│   │   ├── circuit/    # 熔断器
│   │   ├── watchdog/   # 看门狗
│   │   ├── logger/     # Zap 日志
│   │   ├── risk/       # 风险引擎
│   │   └── aging/      # 任务老化
│   ├── registry/        # 模块注册中心
│   └── modules/         # 25 个功能模块
├── config/              # 配置文件
└── data/                # 数据目录
```

## 快速开始

```bash
# 1. 安装依赖
go mod tidy

# 2. 运行开发模式
wails dev

# 3. 构建生产版本
wails build
```

## 核心模块

- M1 系统概览
- M2 进程管理
- M3 网络分析
- M4 注册表分析
- M13 日志分析 (含一键分析)
- M25 编解码工具
- ... (共 25 个模块)

## 开发说明

### 代码规范

- 使用 gofmt 格式化代码
- 所有错误必须处理或记录日志
- 使用 context 控制超时
- 并发操作使用 channel 或 sync 包

### 模块开发

每个模块需实现以下接口:

```go
type Module interface {
    ID() int
    Name() string
    Priority() int
    Init(ctx context.Context, s interface{}) error
    Collect(ctx context.Context) error
    Stop() error
}
```

## 设计文档

详细设计文档请参考: `.monkeycode/specs/2026-03-25-windows-ert/design.md`

## 许可证

MIT
