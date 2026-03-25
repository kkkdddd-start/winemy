//go:build windows

package m10_kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type KernelModule struct {
	ctx     context.Context
	storage registry.Storage
	drivers []model.DriverDTO
}

func New() *KernelModule {
	return &KernelModule{}
}

func (m *KernelModule) ID() int       { return 10 }
func (m *KernelModule) Name() string  { return "kernel" }
func (m *KernelModule) Priority() int { return 1 }

func (m *KernelModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *KernelModule) Collect(ctx context.Context) error {
	m.drivers = []model.DriverDTO{}

	driverMap := make(map[string]driverInfo)

	cmd := exec.Command("powershell", "-Command",
		`Get-Process -Module -ErrorAction SilentlyContinue | Where-Object { $_.ModuleName -like '*.sys' } | Select-Object ModuleName, FileName, BaseAddress, ModuleMemorySize | ConvertTo-Json`)
	output, err := cmd.Output()
	if err == nil {
		var modules []map[string]interface{}
		if err := json.Unmarshal(output, &modules); err != nil {
			var single map[string]interface{}
			if err := json.Unmarshal(output, &single); err == nil {
				modules = []map[string]interface{}{single}
			}
		}

		for _, mod := range modules {
			if mod["ModuleName"] == nil {
				continue
			}
			name := mod["ModuleName"].(string)
			if !strings.HasSuffix(strings.ToLower(name), ".sys") {
				continue
			}

			path := ""
			if p, ok := mod["FileName"].(string); ok {
				path = p
			}

			baseAddr := "0x0"
			if ba, ok := mod["BaseAddress"].(float64); ok {
				baseAddr = fmt.Sprintf("0x%x", uint64(ba))
			}

			size := uint64(0)
			if ms, ok := mod["ModuleMemorySize"].(float64); ok {
				size = uint64(ms)
			}

			driverMap[name] = driverInfo{
				Path:     path,
				BaseAddr: baseAddr,
				Size:     size,
			}
		}
	}

	if len(driverMap) == 0 {
		output, err = exec.Command("tasklist", "/FI", "MODULES eq *.sys", "/FO", "CSV", "/NH").Output()
		if err != nil {
			m.drivers = append(m.drivers, model.DriverDTO{
				Name:      "Error",
				Path:      fmt.Sprintf("Failed to enumerate drivers: %v", err),
				BaseAddr:  "0x0",
				Size:      0,
				IsSigned:  false,
				Signature: "",
				RiskLevel: model.RiskLow,
			})
			return nil
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Split(line, ",")
			if len(fields) < 5 {
				continue
			}

			driverName := strings.Trim(fields[0], "\"")
			if !strings.HasSuffix(strings.ToLower(driverName), ".sys") {
				continue
			}

			memStr := strings.Trim(fields[4], "\"")
			memStr = strings.ReplaceAll(memStr, ",", "")

			size, _ := strconv.ParseUint(memStr, 10, 64)

			riskLevel := model.RiskLow
			driverPath := getDriverPath(driverName)

			isSigned, signature := m.verifyDriverSignature(driverPath)

			if isSuspiciousDriver(driverName) {
				riskLevel = model.RiskHigh
			} else if !isSigned {
				riskLevel = model.RiskMedium
			}

			m.drivers = append(m.drivers, model.DriverDTO{
				Name:      driverName,
				Path:      driverPath,
				BaseAddr:  "0x0",
				Size:      size,
				IsSigned:  isSigned,
				Signature: signature,
				RiskLevel: riskLevel,
			})
		}
		return nil
	}

	for name, info := range driverMap {
		riskLevel := model.RiskLow
		driverPath := info.Path
		if driverPath == "" {
			driverPath = getDriverPath(name)
		}

		isSigned, signature := m.verifyDriverSignature(driverPath)

		if isSuspiciousDriver(name) {
			riskLevel = model.RiskHigh
		} else if !isSigned {
			riskLevel = model.RiskMedium
		}

		m.drivers = append(m.drivers, model.DriverDTO{
			Name:      name,
			Path:      driverPath,
			BaseAddr:  info.BaseAddr,
			Size:      info.Size,
			IsSigned:  isSigned,
			Signature: signature,
			RiskLevel: riskLevel,
		})
	}

	return nil
}

type driverInfo struct {
	Path     string
	BaseAddr string
	Size     uint64
}

func getDriverPath(driverName string) string {
	driversPath := "C:\\Windows\\System32\\drivers"
	return driversPath + "\\" + driverName
}

func isSuspiciousDriver(name string) bool {
	nameLower := strings.ToLower(name)

	suspiciousPatterns := []string{
		"rootkit",
		"keylog",
		"keylogger",
		"sniffer",
		"packet",
		"hook",
		"inject",
		"hide",
		"stealth",
		"malware",
		"trojan",
		"backdoor",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

func (m *KernelModule) verifyDriverSignature(driverPath string) (bool, string) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$sig = Get-AuthenticodeSignature '%s'; if($sig.Status -eq 'Valid') { Write-Output 'Valid' } elseif($sig.Status -eq 'NotSigned') { Write-Output 'NotSigned' } else { Write-Output ('Signed:' + $sig.Status) }`, driverPath))
	output, err := cmd.Output()
	if err != nil {
		return false, "Verification failed"
	}

	result := strings.TrimSpace(string(output))
	if result == "Valid" {
		return true, "Microsoft Windows"
	} else if result == "NotSigned" {
		return false, "NotSigned"
	} else if strings.HasPrefix(result, "Signed:") {
		return true, strings.TrimPrefix(result, "Signed:")
	}
	return false, result
}

func (m *KernelModule) verifyDriverSignatureSignTool(driverPath string) (bool, string) {
	signtoolPaths := []string{
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\10.0.19041.0\x64\signtool.exe`,
		`C:\Program Files (x86)\Windows Kits\10\bin\x64\signtool.exe`,
		`C:\Windows\System32\signtool.exe`,
	}

	var signtoolPath string
	for _, p := range signtoolPaths {
		if _, err := os.Stat(p); err == nil {
			signtoolPath = p
			break
		}
	}

	if signtoolPath == "" {
		return false, "SignTool not found"
	}

	cmd := exec.Command(signtoolPath, "verify", "/pa", "/v", driverPath)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Sprintf("SignTool verify failed: %v", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Successfully verified") || strings.Contains(outputStr, "Number of signatures") {
		return true, "Verified"
	}
	return false, "Verification failed"
}

func (m *KernelModule) Stop() error {
	return nil
}

func (m *KernelModule) DetectFilterDrivers() ([]map[string]interface{}, error) {
	filterDrivers := []map[string]interface{}{}

	cmd := exec.Command("powershell", "-Command",
		`fltmc.exe filters 2>$null; if($LASTEXITCODE -ne 0) { Get-WmiObject -Class MSFilterDriver | Select-Object FilterName, altitude, Flags | ConvertTo-Json }`)
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("powershell", "-Command",
			`$drivers = @(); Get-WmiObject -Namespace root -Class __Namespace -ErrorAction SilentlyContinue | ForEach-Object { $ns = $_.Name; if($ns -match 'Security|FileInfo') { $drivers += $ns } }; $drivers | ConvertTo-Json`)
		output, err = cmd.Output()
	}

	if err == nil && len(output) > 2 {
		var drivers []map[string]interface{}
		if json.Unmarshal(output, &drivers) == nil {
			for _, d := range drivers {
				if name, ok := d["FilterName"].(string); ok {
					filterDrivers = append(filterDrivers, map[string]interface{}{
						"name":     name,
						"altitude": d["altitude"],
						"flags":    d["Flags"],
						"type":     "minifilter",
					})
				}
			}
		}
	}

	cmd = exec.Command("powershell", "-Command",
		`$drivers = @(); Get-WmiObject -Class Win32_SystemDriver | Where-Object { $_.ServiceType -match 'File System Filter|FSFilter' } | Select-Object Name, DisplayName, PathName, ServiceType | ConvertTo-Json`)
	output, err = cmd.Output()
	if err == nil && len(output) > 2 {
		var drivers []map[string]interface{}
		if json.Unmarshal(output, &drivers) == nil {
			for _, d := range drivers {
				if name, ok := d["Name"].(string); ok {
					filterDrivers = append(filterDrivers, map[string]interface{}{
						"name":         name,
						"display_name": d["DisplayName"],
						"path":         d["PathName"],
						"type":         "filesystem",
					})
				}
			}
		}
	}

	return filterDrivers, nil
}

func (m *KernelModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.drivers))
	for _, d := range m.drivers {
		result = append(result, map[string]interface{}{
			"name":       d.Name,
			"path":       d.Path,
			"base_addr":  d.BaseAddr,
			"size":       d.Size,
			"is_signed":  d.IsSigned,
			"signature":  d.Signature,
			"risk_level": d.RiskLevel,
		})
	}
	return result, nil
}

func (m *KernelModule) GetDriverVersion(driverName string) (string, error) {
	driverPath := getDriverPath(driverName)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`(Get-Item '%s' -ErrorAction SilentlyContinue).VersionInfo | Select-Object -ExpandProperty FileVersion`, driverPath))
	output, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("powershell", "-Command",
			fmt.Sprintf(`(Get-Item '%s' -ErrorAction SilentlyContinue).VersionInfo.ProductVersion`, driverPath))
		output2, err2 := cmd2.Output()
		if err2 != nil {
			return "Unknown", fmt.Errorf("failed to get driver version")
		}
		return strings.TrimSpace(string(output2)), nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (m *KernelModule) GetDriverCompany(driverPath string) (string, error) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`(Get-Item '%s' -ErrorAction SilentlyContinue).VersionInfo | Select-Object -ExpandProperty CompanyName`, driverPath))
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", fmt.Errorf("failed to get driver company")
	}
	return strings.TrimSpace(string(output)), nil
}

func (m *KernelModule) Search(keyword string) ([]model.DriverDTO, error) {
	results := []model.DriverDTO{}
	keywordLower := strings.ToLower(keyword)

	for _, d := range m.drivers {
		if strings.Contains(strings.ToLower(d.Name), keywordLower) ||
			strings.Contains(strings.ToLower(d.Path), keywordLower) ||
			strings.Contains(strings.ToLower(d.Signature), keywordLower) {
			results = append(results, d)
		}
	}

	return results, nil
}
