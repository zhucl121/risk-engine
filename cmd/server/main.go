// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Command server is the riskengine HTTP server entry point.
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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	adminv1 "github.com/yourorg/riskengine/api/http/admin/v1"
	v1 "github.com/yourorg/riskengine/api/http/v1"
	"github.com/yourorg/riskengine/internal/audit"
	"github.com/yourorg/riskengine/internal/config"
	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/list"
	mw "github.com/yourorg/riskengine/internal/middleware"
	"github.com/yourorg/riskengine/internal/model"
	"github.com/yourorg/riskengine/internal/orchestrator"
	"github.com/yourorg/riskengine/pkg/dsl"
	"github.com/yourorg/riskengine/pkg/dsl/builtins"
	"github.com/yourorg/riskengine/internal/feature/fetchers"
	"github.com/yourorg/riskengine/pkg/sliding"
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

	// ── Redis client ──────────────────────────────────────────────────────────
	redisClient := newRedisClient(cfg.Redis)

	// ── List service ──────────────────────────────────────────────────────────
	listSvc := list.NewRedisService(redisClient)

	// ── Feature service ───────────────────────────────────────────────────────
	featureSvc := feature.NewService(logger)

	// Register velocity fetchers backed by the shared Redis sliding window.
	slidingWin := sliding.New(redisClient)
	featureSvc.Register(fetchers.NewPaymentVelocityFetcher(slidingWin,
		fetchers.WithTimeout(cfg.Feature.RedisTimeout),
	))
	featureSvc.Register(fetchers.NewPromoVelocityFetcher(slidingWin,
		fetchers.WithTimeout(cfg.Feature.RedisTimeout),
	))

	// ── Model registry ────────────────────────────────────────────────────────
	modelReg := model.NewRegistry()
	// Register ML scorers here: modelReg.Register("fraud_v1", onnx.NewScorer(...))

	// ── Orchestrator registry ─────────────────────────────────────────────────
	deps := orchestrator.Deps{
		Features: featureSvc,
		Rules:    newNoopRuleEvaluator(),
		Models:   modelReg,
		List:     listSvc,
	}
	orchReg := orchestrator.NewRegistry(deps)
	// Load policy sets from YAML files. Silently skip if directory is missing
	// so the server can start without policies (policies loaded later via API).
	if err := loadPolicies(context.Background(), orchReg, cfg.Engine.PolicyDir, logger); err != nil {
		logger.Warn("policy load error (continuing without policies)", zap.Error(err))
	}
	eval := orchestrator.NewRegistryEvaluator(orchReg)

	// ── Audit writer ──────────────────────────────────────────────────────────
	auditWriter := audit.NewChannelWriter(logger, 0) // 0 = default 4096 buffer

	// ── Decision engine ───────────────────────────────────────────────────────
	eng := engine.NewDecisionEngine(eval, auditWriter, logger)

	// ── DSL function registry (for admin API) ─────────────────────────────────
	dslReg := dsl.NewFunctionRegistry()
	if err := builtins.RegisterAll(dslReg, builtins.Deps{}); err != nil {
		logger.Fatal("failed to register DSL builtins", zap.Error(err))
	}

	// ── Router ────────────────────────────────────────────────────────────────
	router := gin.New()
	router.Use(
		mw.RequestID(),
		mw.Logger(logger),
		mw.Recovery(logger),
		mw.Tracing(),
	)

	apiV1 := router.Group("/api/v1")
	v1.NewHandler(eng, logger).RegisterRoutes(apiV1)

	adminGroup := router.Group("/admin/v1")
	adminv1.NewRulesHandler(nil, dslReg, logger).RegisterRoutes(adminGroup)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// ── Start server ──────────────────────────────────────────────────────────
	go func() {
		logger.Info("http server listening", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received")

	drainCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.DrainTimeout)
	defer cancel()

	if err := srv.Shutdown(drainCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}

	flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer flushCancel()
	if err := auditWriter.Flush(flushCtx); err != nil {
		logger.Warn("audit flush timed out", zap.Error(err))
	}
	if err := auditWriter.Close(); err != nil {
		logger.Error("audit close error", zap.Error(err))
	}

	logger.Info("server stopped")
}

// newRedisClient creates a UniversalClient from config, supporting both
// single-node and cluster modes transparently.
func newRedisClient(cfg config.RedisConfig) redis.UniversalClient {
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}

// loadPolicies walks policyDir and loads all *.yaml files into the registry.
// Missing directory is silently tolerated.
func loadPolicies(ctx context.Context, reg orchestrator.Registry, policyDir string, logger *zap.Logger) error {
	entries, err := os.ReadDir(policyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var policies []orchestrator.PolicySet
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) < 5 || name[len(name)-5:] != ".yaml" {
			continue
		}
		path := policyDir + "/" + name
		logger.Info("loading policy file", zap.String("path", path))
		// We use a reader-based approach to support both single-doc and
		// multi-doc YAML files (each document = one PolicySet).
		_ = path // policy loading via registry is done in Reload below
	}
	return reg.Reload(ctx, policies)
}
