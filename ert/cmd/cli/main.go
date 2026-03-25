//go:build windows

package main

import (
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

	fmt.Println("CLI mode initialized")
	fmt.Println("For GUI mode, run: ert.exe")
}
