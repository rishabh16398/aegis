package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aegis-av/aegis/internal/alert"
	"github.com/aegis-av/aegis/internal/api"
	"github.com/aegis-av/aegis/internal/config"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/network"
	"github.com/aegis-av/aegis/internal/process"
	"github.com/aegis-av/aegis/internal/quarantine"
	"github.com/aegis-av/aegis/internal/scanner"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	database, err := db.New(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	alertEngine := alert.New()
	scanEngine := scanner.New(cfg.Scanner.Workers, cfg.Scanner.QuarantinePath)
	netMonitor := network.New(cfg.Network.Interface)
	procMonitor := process.New()
	_ = quarantine.New(cfg.Scanner.QuarantinePath)

	if cfg.Network.Enabled {
		if err := netMonitor.Start(); err != nil {
			slog.Warn("network monitor failed to start", "err", err)
		}
		defer netMonitor.Stop()
	}

	procMonitor.Start()
	defer procMonitor.Stop()

	_ = alertEngine
	_ = scanEngine

	router := api.NewRouter()
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("aegisd started", "addr", cfg.Server.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("stopped")
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
