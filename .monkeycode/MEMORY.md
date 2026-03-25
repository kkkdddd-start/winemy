# 用户指令记忆

本文件记录了用户的指令、偏好和教导，用于在未来的交互中提供参考。

## 格式

### 用户指令条目
用户指令条目应遵循以下格式：

[用户指令摘要]
- Date: [YYYY-MM-DD]
- Context: [提及的场景或时间]
- Instructions:
  - [用户教导或指示的内容，逐行描述]

### 项目知识条目
Agent 在任务执行过程中发现的条目应遵循以下格式：

[项目知识摘要]
- Date: [YYYY-MM-DD]
- Context: Agent 在执行 [具体任务描述] 时发现
- Category: [代码结构|代码模式|代码生成|构建方法|测试方法|依赖关系|环境配置]
- Instructions:
  - [具体的知识点，逐行描述]

## 去重策略
- 添加新条目前，检查是否存在相似或相同的指令
- 若发现重复，跳过新条目或与已有条目合并
- 合并时，更新上下文或日期信息
- 这有助于避免冗余条目，保持记忆文件整洁

## 条目

[项目知识摘要]
- Date: 2026-03-25
- Context: Agent 在初始化 Windows 应急响应工具 (ERT) 项目时发现
- Category: 代码结构
- Instructions:
  - 项目采用 Wails v2 + Vue3 + Element Plus 前后端分离架构
  - 后端使用纯 Go 实现 (modernc.org/sqlite)，无 CGO 依赖
  - 25 个独立模块位于 internal/modules/ 目录
  - 核心引擎位于 internal/core/ 目录
  - 前端代码位于 app/ 目录 (需创建)
  - 数据库采用 MainDB + Cache + Checkpoint 三库架构
  - 模块接口统一: ID() int, Name() string, Priority() int, Init, Collect, Stop

[项目知识摘要]
- Date: 2026-03-25
- Context: Agent 在阅读项目需求文档时发现
- Category: 构建方法
- Instructions:
  - 使用 `wails dev` 启动开发模式
  - 使用 `wails build` 进行生产构建
  - Go 编译: `go build -ldflags="-s -w -H=windowsgui -trimpath" -o ert.exe ./cmd/gui`
  - 前端依赖安装: `cd app && npm install`
  - 前端构建: `cd app && npm run build`
