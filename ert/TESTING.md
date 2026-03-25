# ERT 测试指南

## 编译测试

### 在 Windows 上编译和测试

```powershell
# 编译项目
go build -o ert.exe ./cmd/gui/

# 运行所有测试
go test -v ./...

# 运行特定模块测试
go test -v ./internal/core/codec/
go test -v ./internal/core/memory/
go test -v ./internal/core/progress/
```

### 在 Windows 上运行测试的先决条件

1. 安装 Go 1.21+
2. 安装 Windows 操作系统

## 测试覆盖范围

### 单元测试

| 模块 | 测试文件 | 状态 |
|------|----------|------|
| codec | internal/core/codec/*_test.go | ✅ |
| memory | internal/core/memory/*_test.go | ✅ |
| progress | internal/core/progress/*_test.go | ✅ |
| model | internal/model/*_test.go | ✅ |
| timeline | internal/core/timeline/*_test.go | ✅ |
| compare | internal/core/compare/*_test.go | ✅ |

### 模块测试

由于 ERT 是 Windows 专用工具，所有模块测试都需要在 Windows 环境下运行。

## CI/CD

GitHub Actions 工作流配置示例：

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Build
        run: go build -v ./...
      - name: Test
        run: go test -v ./...
```

## 测试数据

测试使用的示例数据位于 `testdata/` 目录。

## Mock 数据

对于无法在测试环境获取的系统 API，模块使用 mock 数据进行测试。
