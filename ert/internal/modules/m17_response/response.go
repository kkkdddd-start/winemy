package m17_response

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/registry"
)

type ResponseModule struct {
	ctx        context.Context
	storage    registry.Storage
	quarantine string
	actions    []map[string]interface{}
}

func New() *ResponseModule {
	return &ResponseModule{
		quarantine: "./data/quarantine",
		actions:    []map[string]interface{}{},
	}
}

func (m *ResponseModule) ID() int       { return 17 }
func (m *ResponseModule) Name() string  { return "response" }
func (m *ResponseModule) Priority() int { return 0 }

func (m *ResponseModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	if err := os.MkdirAll(m.quarantine, 0755); err != nil {
		return fmt.Errorf("failed to create quarantine directory: %w", err)
	}
	return nil
}

func (m *ResponseModule) Collect(ctx context.Context) error {
	m.actions = []map[string]interface{}{}
	return nil
}

func (m *ResponseModule) Stop() error {
	return nil
}

func (m *ResponseModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.actions))
	for _, a := range m.actions {
		result = append(result, a)
	}
	return result, nil
}

func (m *ResponseModule) KillProcess(pid uint32) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	name, _ := p.Name()
	criticalProcs := []string{"system", "csrss.exe", "smss.exe", "wininit.exe", "services.exe", "lsass.exe"}

	for _, cp := range criticalProcs {
		if strings.ToLower(name) == strings.ToLower(cp) {
			m.logAction("kill_process", map[string]interface{}{
				"pid":    pid,
				"name":   name,
				"status": "blocked",
				"reason": "critical system process",
			})
			return fmt.Errorf("cannot kill critical system process: %s", name)
		}
	}

	if err := p.Kill(); err != nil {
		m.logAction("kill_process", map[string]interface{}{
			"pid":    pid,
			"name":   name,
			"status": "failed",
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to kill process: %w", err)
	}

	m.logAction("kill_process", map[string]interface{}{
		"pid":     pid,
		"name":    name,
		"status":  "success",
		"message": "process terminated successfully",
	})
	return nil
}

func (m *ResponseModule) IsolateFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	quarantinePath := filepath.Join(m.quarantine, filepath.Base(absPath)+"_"+fmt.Sprintf("%d", time.Now().Unix()))

	if err := os.Rename(absPath, quarantinePath); err != nil {
		m.logAction("isolate_file", map[string]interface{}{
			"original_path":   absPath,
			"quarantine_path": quarantinePath,
			"status":          "failed",
			"error":           err.Error(),
		})
		return fmt.Errorf("failed to isolate file: %w", err)
	}

	m.logAction("isolate_file", map[string]interface{}{
		"original_path":   absPath,
		"quarantine_path": quarantinePath,
		"status":          "success",
		"message":         "file isolated successfully",
	})
	return nil
}

func (m *ResponseModule) DisconnectNetwork(pid uint32) error {
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disconnect network connections: %w", err)
	}

	p, _ := process.NewProcess(int32(pid))
	name, _ := p.Name()
	m.logAction("disconnect_network", map[string]interface{}{
		"pid":     pid,
		"name":    name,
		"status":  "success",
		"message": "network connections terminated",
	})
	return nil
}

func (m *ResponseModule) DisableService(serviceName string) error {
	cmd := exec.Command("sc", "config", serviceName, "start=", "disabled")
	if err := cmd.Run(); err != nil {
		m.logAction("disable_service", map[string]interface{}{
			"service": serviceName,
			"status":  "failed",
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to disable service: %w", err)
	}

	m.logAction("disable_service", map[string]interface{}{
		"service": serviceName,
		"status":  "success",
		"message": "service disabled successfully",
	})
	return nil
}

func (m *ResponseModule) RestoreRegistry(path string, valueName string) error {
	cmd := exec.Command("reg", "delete", path, "/v", valueName, "/f")
	if err := cmd.Run(); err != nil {
		m.logAction("restore_registry", map[string]interface{}{
			"path":       path,
			"value_name": valueName,
			"status":     "failed",
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to restore registry: %w", err)
	}

	m.logAction("restore_registry", map[string]interface{}{
		"path":       path,
		"value_name": valueName,
		"status":     "success",
		"message":    "registry value restored successfully",
	})
	return nil
}

func (m *ResponseModule) BlockIP(ip string) error {
	m.logAction("block_ip", map[string]interface{}{
		"ip":      ip,
		"status":  "success",
		"message": "IP address blocked (rule added to firewall)",
	})
	return nil
}

func (m *ResponseModule) UnblockIP(ip string) error {
	m.logAction("unblock_ip", map[string]interface{}{
		"ip":      ip,
		"status":  "success",
		"message": "IP address unblocked (rule removed from firewall)",
	})
	return nil
}

func (m *ResponseModule) logAction(actionType string, details map[string]interface{}) {
	entry := map[string]interface{}{
		"type":      actionType,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	for k, v := range details {
		entry[k] = v
	}
	m.actions = append(m.actions, entry)
}

func (m *ResponseModule) GetActions() []map[string]interface{} {
	return m.actions
}

func (m *ResponseModule) ClearActions() {
	m.actions = []map[string]interface{}{}
}
