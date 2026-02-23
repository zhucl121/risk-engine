// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package featurestore provides a gRPC client for the standalone Feature Store
// service, plus a feature.Fetcher adapter that integrates it into the decision
// engine's parallel feature-fetch pipeline.
package featurestore

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	riskv1 "github.com/yourorg/riskengine/api/grpc/v1"
)

// ClientConfig holds gRPC connection settings for the Feature Store.
type ClientConfig struct {
	// Addr is the Feature Store gRPC endpoint, e.g. "feature-store:9100".
	Addr string
	// DialTimeout is the maximum time to wait for the initial connection.
	DialTimeout time.Duration
	// RequestTimeout is the per-RPC deadline (overridden by Fetcher.Timeout()).
	RequestTimeout time.Duration
	// MaxRetries is how many times a retriable RPC is retried on transient errors.
	MaxRetries int
	// KeepAliveTime is the interval between keepalive pings to the server.
	KeepAliveTime time.Duration
	// KeepAliveTimeout is how long to wait for a keepalive pong before closing.
	KeepAliveTimeout time.Duration
}

// DefaultClientConfig returns production-safe defaults.
func DefaultClientConfig(addr string) ClientConfig {
	return ClientConfig{
		Addr:             addr,
		DialTimeout:      5 * time.Second,
		RequestTimeout:   20 * time.Millisecond,
		MaxRetries:       2,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}
}

// Client is a thin, thread-safe gRPC client for FeatureStoreService.
// It manages the connection lifecycle and wraps RPCs with retry logic.
type Client struct {
	conn   *grpc.ClientConn
	stub   riskv1.FeatureStoreServiceClient
	cfg    ClientConfig
	logger *zap.Logger
}

// NewClient dials the Feature Store and returns a ready Client.
// Call Close() when the engine shuts down.
func NewClient(cfg ClientConfig, logger *zap.Logger) (*Client, error) {
	dialCtx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext( //nolint:staticcheck // DialContext deprecated in grpc v2 but still the standard API in v1
		dialCtx,
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepAliveTime,
			Timeout:             cfg.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{"service": "riskengine.v1.FeatureStoreService"}],
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.01s",
					"maxBackoff": "0.1s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
				}
			}]
		}`),
	)
	if err != nil {
		return nil, fmt.Errorf("featurestore: dial %s: %w", cfg.Addr, err)
	}

	logger.Info("feature store client connected", zap.String("addr", cfg.Addr))
	return &Client{
		conn:   conn,
		stub:   riskv1.NewFeatureStoreServiceClient(conn),
		cfg:    cfg,
		logger: logger,
	}, nil
}

// GetFeatures fetches a named feature group for the given entity context.
// The ctx deadline is respected; callers should set an appropriate timeout.
func (c *Client) GetFeatures(ctx context.Context, req *riskv1.GetFeaturesRequest) (*riskv1.GetFeaturesResponse, error) {
	resp, err := c.stub.GetFeatures(ctx, req)
	if err != nil {
		if isRetriable(err) {
			// gRPC service config already handles server-side retry; this path
			// is reached when the retry budget is exhausted or context cancelled.
			c.logger.Warn("feature store GetFeatures failed",
				zap.String("group", req.FeatureGroup),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("featurestore.GetFeatures %q: %w", req.FeatureGroup, err)
	}
	return resp, nil
}

// BatchGetFeatures sends multiple feature requests in one RPC round-trip.
func (c *Client) BatchGetFeatures(ctx context.Context, req *riskv1.BatchGetFeaturesRequest) (*riskv1.BatchGetFeaturesResponse, error) {
	resp, err := c.stub.BatchGetFeatures(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("featurestore.BatchGetFeatures: %w", err)
	}
	return resp, nil
}

// Health checks the Feature Store's readiness.
func (c *Client) Health(ctx context.Context) (*riskv1.FeatureStoreHealthResponse, error) {
	return c.stub.Health(ctx, &riskv1.FeatureStoreHealthRequest{})
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// isRetriable returns true for transient gRPC errors that may succeed on retry.
func isRetriable(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch s.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
		return true
	}
	return false
}
