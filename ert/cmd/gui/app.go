//go:build windows

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourname/ert/internal/config"
	"github.com/yourname/ert/internal/core/storage"
	"github.com/yourname/ert/internal/model"
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

// GetModuleData retrieves data from a specific module
// Wails binding for frontend API call
func (a *App) GetModuleData(moduleID int, query string) ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), moduleID, query)
}

// CollectModule triggers data collection for a specific module
// Wails binding for frontend API call
func (a *App) CollectModule(moduleID int) error {
	return a.reg.CollectModule(context.Background(), moduleID)
}

// GetSystemInfo retrieves system information
// Wails binding for frontend API call
func (a *App) GetSystemInfo() (map[string]interface{}, error) {
	data, err := a.reg.GetModuleData(context.Background(), 1, "")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// GetProcessList retrieves process list
// Wails binding for frontend API call
func (a *App) GetProcessList() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 2, "")
}

// KillProcess kills a process by PID
// Wails binding for frontend API call
func (a *App) KillProcess(pid uint32) error {
	module, err := a.reg.Get(2)
	if err != nil {
		return err
	}
	if procModule, ok := module.(*m2_process.ProcessModule); ok {
		return procModule.KillProcess(pid)
	}
	return nil
}

// GetProcessTree retrieves process tree structure
// Wails binding for frontend API call
func (a *App) GetProcessTree() ([]map[string]interface{}, error) {
	module, err := a.reg.Get(2)
	if err != nil {
		return nil, err
	}
	if procModule, ok := module.(*m2_process.ProcessModule); ok {
		tree := procModule.GetProcessTree()
		result := make([]map[string]interface{}, len(tree))
		for i, node := range tree {
			result[i] = map[string]interface{}{
				"pid":      node.PID,
				"name":     node.Name,
				"children": flattenProcessTree(node.Children),
			}
		}
		return result, nil
	}
	return nil, nil
}

func flattenProcessTree(nodes []*model.ProcessTreeNode) []map[string]interface{} {
	result := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		result[i] = map[string]interface{}{
			"pid":      node.PID,
			"name":     node.Name,
			"children": flattenProcessTree(node.Children),
		}
	}
	return result
}

// StartService starts a Windows service
// Wails binding for frontend API call
func (a *App) StartService(serviceName string) error {
	module, err := a.reg.Get(5)
	if err != nil {
		return err
	}
	if svcModule, ok := module.(*m5_service.ServiceModule); ok {
		return svcModule.StartService(serviceName)
	}
	return nil
}

// StopService stops a Windows service
// Wails binding for frontend API call
func (a *App) StopService(serviceName string) error {
	module, err := a.reg.Get(5)
	if err != nil {
		return err
	}
	if svcModule, ok := module.(*m5_service.ServiceModule); ok {
		return svcModule.StopService(serviceName)
	}
	return nil
}

// RestartService restarts a Windows service
// Wails binding for frontend API call
func (a *App) RestartService(serviceName string) error {
	module, err := a.reg.Get(5)
	if err != nil {
		return err
	}
	if svcModule, ok := module.(*m5_service.ServiceModule); ok {
		return svcModule.RestartService(serviceName)
	}
	return nil
}

// GetNetworkList retrieves network connections
// Wails binding for frontend API call
func (a *App) GetNetworkList() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 3, "")
}

// GetRegistryKeys retrieves registry keys
// Wails binding for frontend API call
func (a *App) GetRegistryKeys() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 4, "")
}

// GetServices retrieves services list
// Wails binding for frontend API call
func (a *App) GetServices() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 5, "")
}

// GetScheduledTasks retrieves scheduled tasks
// Wails binding for frontend API call
func (a *App) GetScheduledTasks() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 6, "")
}

// ExportScheduledTaskToXML exports a scheduled task to XML
// Wails binding for frontend API call
func (a *App) ExportScheduledTaskToXML(taskName string, outputPath string) error {
	module, err := a.reg.Get(6)
	if err != nil {
		return err
	}
	if schedModule, ok := module.(*m6_schedule.ScheduleModule); ok {
		return schedModule.ExportTaskToXML(taskName, outputPath)
	}
	return nil
}

// GetMonitorData retrieves real-time monitoring data
// Wails binding for frontend API call
func (a *App) GetMonitorData() (map[string]interface{}, error) {
	data, err := a.reg.GetModuleData(context.Background(), 7, "")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// GetPatches retrieves installed patches
// Wails binding for frontend API call
func (a *App) GetPatches() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 8, "")
}

// GetSoftwareList retrieves installed software
// Wails binding for frontend API call
func (a *App) GetSoftwareList() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 9, "")
}

// GetDrivers retrieves kernel drivers
// Wails binding for frontend API call
func (a *App) GetDrivers() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 10, "")
}

// GetFiles retrieves file system entries
// Wails binding for frontend API call
func (a *App) GetFiles(path string) ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 11, path)
}

// GetActivity retrieves activity history
// Wails binding for frontend API call
func (a *App) GetActivity() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 12, "")
}

// GetEventLogs retrieves event logs
// Wails binding for frontend API call
func (a *App) GetEventLogs(channel string, level string, eventID int) ([]map[string]interface{}, error) {
	query := channel + ":" + level + ":" + string(rune(eventID))
	return a.reg.GetModuleData(context.Background(), 13, query)
}

// GetAccounts retrieves user accounts
// Wails binding for frontend API call
func (a *App) GetAccounts() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 14, "")
}

// GetMemoryDumps retrieves memory dump list
// Wails binding for frontend API call
func (a *App) GetMemoryDumps() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 15, "")
}

// DumpProcess creates a memory dump for a process
// Wails binding for frontend API call
func (a *App) DumpProcess(pid uint32) (string, error) {
	module, err := a.reg.Get(15)
	if err != nil {
		return "", err
	}
	if memModule, ok := module.(*m15_memory.MemoryModule); ok {
		return memModule.DumpProcess(pid)
	}
	return "", nil
}

// GetThreats retrieves threat detection results
// Wails binding for frontend API call
func (a *App) GetThreats() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 16, "")
}

// ResponseAction performs a response action
// Wails binding for frontend API call
func (a *App) ResponseAction(action string, target string) error {
	module, err := a.reg.Get(17)
	if err != nil {
		return err
	}
	if respModule, ok := module.(*m17_response.ResponseModule); ok {
		switch action {
		case "kill_process":
			pid := uint32(0)
			fmt.Sscanf(target, "%d", &pid)
			if pid != 0 {
				return respModule.KillProcess(pid)
			}
		case "isolate_file":
			return respModule.IsolateFile(target)
		case "disconnect_network":
			pid := uint32(0)
			fmt.Sscanf(target, "%d", &pid)
			if pid != 0 {
				return respModule.DisconnectNetwork(pid)
			}
		case "disable_service":
			return respModule.DisableService(target)
		case "block_ip":
			return respModule.BlockIP(target)
		case "unblock_ip":
			return respModule.UnblockIP(target)
		case "restore_registry":
			parts := strings.Split(target, "|")
			if len(parts) >= 2 {
				return respModule.RestoreRegistry(parts[0], parts[1])
			}
		case "backup_file":
			_, err := respModule.BackupFile(target)
			return err
		case "restore_file":
			_, err := respModule.RestoreFile(target)
			return err
		}
	}
	return fmt.Errorf("unsupported action: %s", action)
}

// GetAutostartItems retrieves autostart items
// Wails binding for frontend API call
func (a *App) GetAutostartItems() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 18, "")
}

// GetDomainInfo retrieves domain information
// Wails binding for frontend API call
func (a *App) GetDomainInfo() (map[string]interface{}, error) {
	data, err := a.reg.GetModuleData(context.Background(), 19, "")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// GetDomainHackDetections retrieves domain hack detections
// Wails binding for frontend API call
func (a *App) GetDomainHackDetections() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 20, "")
}

// GetWMICHistory retrieves WMIC command history
// Wails binding for frontend API call
func (a *App) GetWMICHistory() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 21, "")
}

// ExportReport exports a report
// Wails binding for frontend API call
func (a *App) ExportReport(format string, sessionID string) (string, error) {
	module, err := a.reg.Get(22)
	if err != nil {
		return "", err
	}
	if reportModule, ok := module.(*m22_report.ReportModule); ok {
		return reportModule.ExportReport(format, sessionID)
	}
	return "", nil
}

// GetReportHistory retrieves report generation history
// Wails binding for frontend API call
func (a *App) GetReportHistory() ([]map[string]interface{}, error) {
	module, err := a.reg.Get(22)
	if err != nil {
		return nil, err
	}
	if reportModule, ok := module.(*m22_report.ReportModule); ok {
		return reportModule.ListReports(), nil
	}
	return nil, nil
}

// GetBaselineResults retrieves security baseline check results
// Wails binding for frontend API call
func (a *App) GetBaselineResults() ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 23, "")
}

// GetIISLogs retrieves IIS logs
// Wails binding for frontend API call
func (a *App) GetIISLogs(logPath string) ([]map[string]interface{}, error) {
	return a.reg.GetModuleData(context.Background(), 24, logPath)
}

// CodecEncode encodes a string using specified codec
// Wails binding for frontend API call
func (a *App) CodecEncode(input string, codecType string) (string, error) {
	module, err := a.reg.Get(25)
	if err != nil {
		return "", err
	}
	if codecModule, ok := module.(*m25_codec.CodecModule); ok {
		return codecModule.Encode(input, codecType)
	}
	return "", nil
}

// CodecDecode decodes a string using specified codec
// Wails binding for frontend API call
func (a *App) CodecDecode(input string, codecType string) (string, error) {
	module, err := a.reg.Get(25)
	if err != nil {
		return "", err
	}
	if codecModule, ok := module.(*m25_codec.CodecModule); ok {
		return codecModule.Decode(input, codecType)
	}
	return "", nil
}

// CodecAutoDetect auto-detects encoding type
// Wails binding for frontend API call
func (a *App) CodecAutoDetect(input string) ([]map[string]interface{}, error) {
	module, err := a.reg.Get(25)
	if err != nil {
		return nil, err
	}
	if codecModule, ok := module.(*m25_codec.CodecModule); ok {
		return codecModule.AutoDetect(input)
	}
	return nil, nil
}

// GetCodecHistory retrieves codec history
// Wails binding for frontend API call
func (a *App) GetCodecHistory() ([]map[string]interface{}, error) {
	module, err := a.reg.Get(25)
	if err != nil {
		return nil, err
	}
	if codecModule, ok := module.(*m25_codec.CodecModule); ok {
		return codecModule.GetHistory(), nil
	}
	return nil, nil
}

// ClearCodecHistory clears codec history
// Wails binding for frontend API call
func (a *App) ClearCodecHistory() error {
	module, err := a.reg.Get(25)
	if err != nil {
		return err
	}
	if codecModule, ok := module.(*m25_codec.CodecModule); ok {
		codecModule.ClearHistory()
	}
	return nil
}
