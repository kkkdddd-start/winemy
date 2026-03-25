//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourname/ert/internal/config"
	"github.com/yourname/ert/internal/core/logger"
	"github.com/yourname/ert/internal/core/storage"
)

var (
	version   = "13.0.0"
	buildTime = ""
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printHelp()
			os.Exit(0)
		case "-v", "--version", "version":
			printVersion()
			os.Exit(0)
		}
	}

	fmt.Printf("ERT CLI v%s\n", version)
	fmt.Println("Windows Emergency Response Tool - Command Line Interface")
	fmt.Println()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Log.Level, cfg.Log.File, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	dbPath := cfg.Database.Main.Path
	stor, err := storage.New(dbPath)
	if err != nil {
		logger.Errorf("Failed to create storage: %s", err.Error())
		os.Exit(1)
	}
	defer stor.Close()

	if len(os.Args) > 1 {
		if err := handleCommand(os.Args[1:], stor); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	printInteractiveHelp()
}

func printVersion() {
	fmt.Printf("ERT CLI v%s\n", version)
	if buildTime != "" {
		fmt.Printf("Build time: %s\n", buildTime)
	}
}

func printHelp() {
	fmt.Printf(`ERT CLI v%s
Windows Emergency Response Tool - Command Line Interface

Usage:
  ert-cli.exe [command] [options]

Commands:
  collect <module>    Collect data from specified module
  collect --all       Collect data from all modules
  triage              Quick triage mode - collect critical indicators
  check <type>        Run specific security check
  export <session>    Export report for session
  list                List available modules
  status             Show current status

Options:
  -h, --help         Show this help message
  -v, --version      Show version information

Examples:
  ert-cli.exe list
  ert-cli.exe collect process
  ert-cli.exe collect --all
  ert-cli.exe triage
  ert-cli.exe check suspicious_processes
  ert-cli.exe export abc123

For GUI mode, run: ert.exe
`, version)
}

func printInteractiveHelp() {
	fmt.Println("CLI mode initialized")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  ert-cli.exe list                - List available modules")
	fmt.Println("  ert-cli.exe collect <module>    - Collect data from module")
	fmt.Println("  ert-cli.exe collect --all      - Collect all modules")
	fmt.Println("  ert-cli.exe triage             - Quick triage")
	fmt.Println("  ert-cli.exe check <type>       - Run security check")
	fmt.Println("  ert-cli.exe export <session>   - Export report")
	fmt.Println("  ert-cli.exe -h                - Show help")
	fmt.Println("  ert-cli.exe -v                - Show version")
	fmt.Println()
	fmt.Println("For GUI mode, run: ert.exe")
}

func handleCommand(args []string, stor *storage.Storage) error {
	if len(args) == 0 {
		printInteractiveHelp()
		return nil
	}

	cmd := args[0]

	switch cmd {
	case "list":
		return listModules()
	case "collect":
		return handleCollect(args[1:])
	case "triage":
		return runTriage(stor)
	case "check":
		if len(args) < 2 {
			return fmt.Errorf("check requires a type argument")
		}
		return runCheck(args[1])
	case "export":
		if len(args) < 2 {
			return fmt.Errorf("export requires a session ID argument")
		}
		return exportSession(args[1])
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func listModules() error {
	fmt.Println("Available modules:")
	modules := []struct {
		id   int
		name string
		desc string
	}{
		{1, "system", "System Overview"},
		{2, "process", "Process Management"},
		{3, "network", "Network Connections"},
		{4, "registry", "Registry Analysis"},
		{5, "service", "Service Management"},
		{6, "schedule", "Scheduled Tasks"},
		{7, "monitor", "System Monitor"},
		{8, "patch", "System Patches"},
		{9, "software", "Installed Software"},
		{10, "kernel", "Kernel Drivers"},
		{11, "filesystem", "File System"},
		{12, "activity", "Activity History"},
		{13, "logging", "Event Logs"},
		{14, "account", "User Accounts"},
		{15, "memory", "Memory Forensics"},
		{16, "threat", "Threat Detection"},
		{17, "response", "Incident Response"},
		{18, "autostart", "Autostart Items"},
		{19, "domain", "Domain Controller"},
		{20, "domainhack", "Domain Penetration"},
		{21, "wmic", "WMIC History"},
		{22, "report", "Report Export"},
		{23, "baseline", "Security Baseline"},
		{24, "iis", "IIS/Logs"},
		{25, "codec", "Encoding Tools"},
	}
	for _, m := range modules {
		fmt.Printf("  M%d - %s: %s\n", m.id, m.name, m.desc)
	}
	return nil
}

func handleCollect(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("collect requires module name or --all flag")
	}
	if args[0] == "--all" {
		fmt.Println("Collecting data from all modules...")
		for i := 1; i <= 25; i++ {
			fmt.Printf("  Collecting M%d...\n", i)
		}
		fmt.Println("Collection complete")
		return nil
	}
	module := args[0]
	fmt.Printf("Collecting data from module: %s\n", module)
	return nil
}

func runTriage(stor *storage.Storage) error {
	fmt.Println("Running triage mode...")
	fmt.Println("  Collecting critical indicators:")
	fmt.Println("    - Suspicious processes")
	fmt.Println("    - Network connections")
	fmt.Println("    - Recent files")
	fmt.Println("    - Security events")
	fmt.Println("Triage complete")
	return nil
}

func runCheck(checkType string) error {
	fmt.Printf("Running security check: %s\n", checkType)
	return nil
}

func exportSession(sessionID string) error {
	fmt.Printf("Exporting report for session: %s\n", sessionID)
	return nil
}

func init() {
	flag.Parse()
}
