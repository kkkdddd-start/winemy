//go:build windows

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 是 Wails 应用的主结构
type App struct {
	ctx context.Context
}

// NewApp 创建一个新的 App 实例
func NewApp() *App {
	return &App{}
}

// Startup 在应用启动时调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	runtime.Log.Info(ctx, "ERT Application started")
}

// Shutdown 在应用关闭时调用
func (a *App) Shutdown(ctx context.Context) {
	runtime.Log.Info(ctx, "ERT Application shutdown")
}

// Greet 返回问候信息
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to ERT!", name)
}

// GetVersion 返回应用版本
func (a *App) GetVersion() string {
	return "13.0.0"
}

func main() {
	// 创建应用实例
	app := NewApp()

	// 创建并运行应用
	err := wails.Run(app, &wails.AppConfig{
		Title:     "ERT - Windows 应急响应工具",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 600,
		HomeDir:   "",
		AssetDir:  "",
	})
	if err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
