//go:build windows

package main

import (
	"context"
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
	version   = "13.0.0"
	buildTime = ""
)

func main() {
	fmt.Printf("ERT v%s\n", version)
	fmt.Println("Windows Emergency Response Tool")
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

	runApp(ctx, cfg, stor)
}

func runApp(ctx context.Context, cfg *config.Config, stor *storage.Storage) {
	app := NewApp(ctx, cfg, stor)

	fmt.Println("Initializing modules...")
	app.registerModules()
	fmt.Println("Application initialized successfully")
	fmt.Println()
	fmt.Println("Note: This is a GUI application, please run with Wails runtime")
	fmt.Println("For development: wails dev")
	fmt.Println("For build: wails build")

	select {
	case <-ctx.Done():
		logger.Info("Application shutdown")
	}
}
