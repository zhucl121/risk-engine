// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package server implements the FeatureStoreService gRPC server.
// It acts as the authoritative source of pre-computed and real-time features,
// aggregating data from Redis, relational databases, and external APIs.
//
// Deployment options:
//   - Sidecar: run in the same pod as riskengine, communicate over localhost.
//   - Standalone: run as a separate Kubernetes Deployment behind a ClusterIP service.
package server

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
	"github.com/zhucl121/risk-engine/internal/featurestore/store"
)

// FeatureStoreServer implements riskv1.FeatureStoreServiceServer.
type FeatureStoreServer struct {
	riskv1.UnimplementedFeatureStoreServiceServer
	registry store.GroupRegistry
	logger   *zap.Logger
}

// NewFeatureStoreServer returns a server backed by the given registry.
func NewFeatureStoreServer(registry store.GroupRegistry, logger *zap.Logger) *FeatureStoreServer {
	return &FeatureStoreServer{registry: registry, logger: logger}
}

// GetFeatures implements FeatureStoreService.GetFeatures.
func (s *FeatureStoreServer) GetFeatures(ctx context.Context, req *riskv1.GetFeaturesRequest) (*riskv1.GetFeaturesResponse, error) {
	start := time.Now()

	group, err := s.registry.Get(req.FeatureGroup)
	if err != nil {
		return nil, fmt.Errorf("feature group %q not registered: %w", req.FeatureGroup, err)
	}

	raw, missing, err := group.Fetch(ctx, req.Entity)
	if err != nil {
		s.logger.Error("feature group fetch error",
			zap.String("group", req.FeatureGroup),
			zap.String("request_id", req.RequestId),
			zap.Error(err),
		)
		return nil, fmt.Errorf("fetch %q: %w", req.FeatureGroup, err)
	}

	return &riskv1.GetFeaturesResponse{
		RequestId:    req.RequestId,
		FeatureGroup: req.FeatureGroup,
		Features:     raw,
		MissingKeys:  missing,
		FetchMs:      time.Since(start).Milliseconds(),
	}, nil
}

// BatchGetFeatures implements FeatureStoreService.BatchGetFeatures.
func (s *FeatureStoreServer) BatchGetFeatures(ctx context.Context, req *riskv1.BatchGetFeaturesRequest) (*riskv1.BatchGetFeaturesResponse, error) {
	responses := make([]*riskv1.GetFeaturesResponse, 0, len(req.Requests))
	for _, r := range req.Requests {
		resp, err := s.GetFeatures(ctx, r)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return &riskv1.BatchGetFeaturesResponse{Responses: responses}, nil
}

// Health implements FeatureStoreService.Health.
func (s *FeatureStoreServer) Health(_ context.Context, _ *riskv1.FeatureStoreHealthRequest) (*riskv1.FeatureStoreHealthResponse, error) {
	statuses := s.registry.BackendStatuses()
	allHealthy := true
	for _, ok := range statuses {
		if !ok {
			allHealthy = false
			break
		}
	}
	return &riskv1.FeatureStoreHealthResponse{
		Healthy:         allHealthy,
		BackendStatuses: statuses,
	}, nil
}
