package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/yourname/ert/internal/config"
	"github.com/yourname/ert/internal/core/logger"
	"github.com/yourname/ert/internal/core/storage"
)

var (
	configPath = flag.String("config", "config/config.yaml", "Path to config file")
	mode       = flag.String("mode", "gui", "Run mode: gui or cli")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Log.Level, cfg.Log.File, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge, cfg.Log.Compress); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	dbPath := cfg.Database.Main.Path
	if !filepath.IsAbs(dbPath) {
		if dbPath, err = filepath.Abs(dbPath); err != nil {
			logger.Errorf("Failed to get absolute path: %s", err)
			os.Exit(1)
		}
	}

	stor, err := storage.New(dbPath)
	if err != nil {
		logger.Errorf("Failed to create storage: %s", err.Error())
		os.Exit(1)
	}
	defer stor.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Received shutdown signal")
		cancel()
	}()

	switch *mode {
	case "gui":
		runGUI(ctx, cfg, stor)
	case "cli":
		runCLI(ctx, cfg, stor)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

func runGUI(ctx context.Context, cfg *config.Config, stor *storage.Storage) {
	fmt.Println("GUI mode - Wails integration pending")
}

func runCLI(ctx context.Context, cfg *config.Config, stor *storage.Storage) {
	fmt.Println("ERT CLI v13.0")
	fmt.Println("CLI mode not yet implemented")
}
