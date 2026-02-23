// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Command server is the riskengine HTTP/gRPC server entry point.
// It wires all dependencies, starts the decision engine, and serves traffic
// until a SIGTERM/SIGINT is received, then drains in-flight requests gracefully.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/yourorg/riskengine/api/http/v1"
	"github.com/yourorg/riskengine/internal/config"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	logger.Info("starting riskengine", zap.String("version", Version))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// TODO: wire engine, feature service, list service, audit writer
	// using cfg. Replace stub below with real wiring.
	eng := newStubEngine()

	router := gin.New()
	router.Use(gin.Recovery())
	// TODO: add middleware: rate-limit, trace injection, access log

	apiV1 := router.Group("/api/v1")
	v1.NewHandler(eng, logger).RegisterRoutes(apiV1)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in background goroutine.
	go func() {
		logger.Info("http server listening", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	// Block until shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
	logger.Info("server stopped")
}
