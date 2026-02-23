// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Command featurestore is the standalone Feature Store gRPC server.
// It aggregates features from Redis (velocity counters, user profiles)
// and exposes them via the FeatureStoreService gRPC API.
//
// Usage:
//
//	featurestore -config configs/config.yaml
//
// The decision engine (cmd/server) connects to this service via
// internal/featurestore.Client when feature_store.enabled = true.
package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
	"github.com/zhucl121/risk-engine/internal/config"
	fserver "github.com/zhucl121/risk-engine/internal/featurestore/server"
	"github.com/zhucl121/risk-engine/internal/featurestore/store"
	"github.com/zhucl121/risk-engine/pkg/sliding"

	"github.com/redis/go-redis/v9"
)

var Version = "dev"

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	logger.Info("starting feature store", zap.String("version", Version))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// ── Redis client ──────────────────────────────────────────────────────────
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Redis.Addrs,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer rdb.Close() //nolint:errcheck

	// ── Feature groups ────────────────────────────────────────────────────────
	registry := store.NewRegistry()

	slidingWin := sliding.New(rdb)
	registry.Register(store.NewPaymentVelocityGroup(slidingWin))
	registry.Register(store.NewUserProfileGroup(rdb))

	// Register Redis as a backend health probe.
	registry.RegisterBackend("redis", func() bool {
		return rdb.Ping(context.Background()).Err() == nil
	})

	// ── gRPC server ───────────────────────────────────────────────────────────
	addr := cfg.FeatureStore.Addr
	if addr == "" {
		addr = ":9100"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("listen failed", zap.String("addr", addr), zap.Error(err))
	}

	grpcSrv := grpc.NewServer()
	riskv1.RegisterFeatureStoreServiceServer(grpcSrv, fserver.NewFeatureStoreServer(registry, logger))

	go func() {
		logger.Info("feature store listening", zap.String("addr", addr))
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Error("grpc server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logger.Info("shutting down feature store")
	grpcSrv.GracefulStop()
}
