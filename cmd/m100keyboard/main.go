package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/msollanych-tt/TRS80-M100-keyboard-driver/internal/config"
	"github.com/msollanych-tt/TRS80-M100-keyboard-driver/internal/keyboard"
)

const version = "0.0.3"

func main() {
	// Parse configuration
	cfg := config.ParseFlags()
	cfg.SetupLogging()

	slog.Info("TRS-80 Model 100 Keyboard Driver starting", "version", version)

	// Create scanner
	scanner, err := keyboard.New(cfg)
	if err != nil {
		slog.Error("Failed to create keyboard scanner", "error", err)
		os.Exit(1)
	}
	defer scanner.Close()

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Run the scanner
	if err := scanner.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("Scanner error", "error", err)
		os.Exit(1)
	}

	slog.Info("Shutdown complete")
}
