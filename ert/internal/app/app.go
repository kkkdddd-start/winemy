package app

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/yourname/ert/internal/config"
	"github.com/yourname/ert/internal/core/logger"
	"github.com/yourname/ert/internal/core/storage"
	"github.com/yourname/ert/internal/registry"
)

type App struct {
	wailsApp       *wails.App
	cfg            *config.Config
	storage        *storage.Storage
	moduleRegistry *registry.Registry
	ctx            context.Context
}

func New(cfg *config.Config, stor *storage.Storage, reg *registry.Registry) *App {
	return &App{
		cfg:            cfg,
		storage:        stor,
		moduleRegistry: reg,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	logger.Info("ERT application starting",
		logger.GetLogger().With(
			logger.GetLogger().Field("app", a.cfg.App.Name),
			logger.GetLogger().Field("version", a.cfg.App.Version),
		)...)

	if err := a.moduleRegistry.Init(ctx); err != nil {
		logger.Error("Failed to initialize module registry",
			logger.GetLogger().Field("error", err.Error()),
		)
	}
}

func (a *App) Shutdown(ctx context.Context) {
	logger.Info("ERT application shutting down")

	if err := a.moduleRegistry.Stop(ctx); err != nil {
		logger.Error("Failed to stop module registry",
			logger.GetLogger().Field("error", err.Error()),
		)
	}
}

func (a *App) Run() error {
	app, err := wails.NewApp(a)
	if err != nil {
		return fmt.Errorf("failed to create wails app: %w", err)
	}

	a.wailsApp = app

	_, err = app.NewWebviewWindowWithOptions(&windows.Options{
		Title:  a.cfg.App.Name,
		Width:  1400,
		Height: 900,
	})
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	return app.Run()
}

func (a *App) CollectModule(ctx context.Context, moduleID int) error {
	return a.moduleRegistry.Collect(ctx, moduleID)
}

func (a *App) GetModuleData(ctx context.Context, moduleID int, query string) ([]map[string]interface{}, error) {
	return a.moduleRegistry.GetData(ctx, moduleID, query)
}

func (a *App) GetSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	module, err := a.moduleRegistry.Get(1)
	if err != nil {
		return nil, err
	}

	collector, ok := module.(SystemCollector)
	if !ok {
		return nil, fmt.Errorf("module %d does not implement SystemCollector", moduleID)
	}

	return collector.GetSystemInfo()
}

type SystemCollector interface {
	GetSystemInfo() (*model.SystemInfo, error)
}
