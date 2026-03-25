//go:build windows

package main

import (
	"context"

	"github.com/yourname/ert/internal/config"
	"github.com/yourname/ert/internal/core/storage"
	"github.com/yourname/ert/internal/modules/m10_kernel"
	"github.com/yourname/ert/internal/modules/m11_filesystem"
	"github.com/yourname/ert/internal/modules/m12_activity"
	"github.com/yourname/ert/internal/modules/m13_logging"
	"github.com/yourname/ert/internal/modules/m14_account"
	"github.com/yourname/ert/internal/modules/m15_memory"
	"github.com/yourname/ert/internal/modules/m16_threat"
	"github.com/yourname/ert/internal/modules/m17_response"
	"github.com/yourname/ert/internal/modules/m18_autostart"
	"github.com/yourname/ert/internal/modules/m19_domain"
	"github.com/yourname/ert/internal/modules/m1_system"
	"github.com/yourname/ert/internal/modules/m20_domainhack"
	"github.com/yourname/ert/internal/modules/m21_wmic"
	"github.com/yourname/ert/internal/modules/m22_report"
	"github.com/yourname/ert/internal/modules/m23_baseline"
	"github.com/yourname/ert/internal/modules/m24_iis"
	"github.com/yourname/ert/internal/modules/m25_codec"
	"github.com/yourname/ert/internal/modules/m2_process"
	"github.com/yourname/ert/internal/modules/m3_network"
	"github.com/yourname/ert/internal/modules/m4_registry"
	"github.com/yourname/ert/internal/modules/m5_service"
	"github.com/yourname/ert/internal/modules/m6_schedule"
	"github.com/yourname/ert/internal/modules/m7_monitor"
	"github.com/yourname/ert/internal/modules/m8_patch"
	"github.com/yourname/ert/internal/modules/m9_software"
	"github.com/yourname/ert/internal/registry"
)

type App struct {
	ctx     context.Context
	config  *config.Config
	storage *storage.Storage
	reg     *registry.Registry
}

func NewApp(ctx context.Context, cfg *config.Config, stor *storage.Storage) *App {
	app := &App{
		ctx:     ctx,
		config:  cfg,
		storage: stor,
		reg:     registry.New(stor),
	}
	app.registerModules()
	return app
}

func (a *App) registerModules() {
	a.reg.Register(m1_system.New())
	a.reg.Register(m2_process.New())
	a.reg.Register(m3_network.New())
	a.reg.Register(m4_registry.New())
	a.reg.Register(m5_service.New())
	a.reg.Register(m6_schedule.New())
	a.reg.Register(m7_monitor.New())
	a.reg.Register(m8_patch.New())
	a.reg.Register(m9_software.New())
	a.reg.Register(m10_kernel.New())
	a.reg.Register(m11_filesystem.New())
	a.reg.Register(m12_activity.New())
	a.reg.Register(m13_logging.New())
	a.reg.Register(m14_account.New())
	a.reg.Register(m15_memory.New())
	a.reg.Register(m16_threat.New())
	a.reg.Register(m17_response.New())
	a.reg.Register(m18_autostart.New())
	a.reg.Register(m19_domain.New())
	a.reg.Register(m20_domainhack.New())
	a.reg.Register(m21_wmic.New())
	a.reg.Register(m22_report.New())
	a.reg.Register(m23_baseline.New())
	a.reg.Register(m24_iis.New())
	a.reg.Register(m25_codec.New())
}

func (a *App) Run() {
}

func (a *App) GetSystemInfo(ctx context.Context) ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(ctx, 1, "")
}

func (a *App) CollectModule(ctx context.Context, moduleID int) error {
	return a.reg.CollectModule(ctx, moduleID)
}

func (a *App) GetModuleData(ctx context.Context, moduleID int, query string) ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(ctx, moduleID, query)
}
