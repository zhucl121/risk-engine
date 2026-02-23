// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package featurestore

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/internal/health"
	"github.com/zhucl121/risk-engine/internal/metrics"
)

// Fetcher adapts Client to the feature.Fetcher interface so it can be
// registered directly into feature.Service alongside local fetchers.
//
// Each Fetcher instance targets one feature group (e.g. "user_profile",
// "device_risk").  Create one Fetcher per group and register all of them:
//
//	featureSvc.Register(featurestore.NewFetcher(client, "user_profile", 20*time.Millisecond, logger))
//	featureSvc.Register(featurestore.NewFetcher(client, "device_risk",  15*time.Millisecond, logger))
type Fetcher struct {
	client  *Client
	group   string
	timeout time.Duration
	logger  *zap.Logger
}

// NewFetcher returns a Fetcher that calls the Feature Store for featureGroup.
// timeout is the per-request deadline (should be well below the overall
// feature.Service total timeout so other fetchers are not starved).
func NewFetcher(client *Client, featureGroup string, timeout time.Duration, logger *zap.Logger) *Fetcher {
	return &Fetcher{
		client:  client,
		group:   featureGroup,
		timeout: timeout,
		logger:  logger,
	}
}

// Name implements feature.Fetcher.
func (f *Fetcher) Name() string {
	return "featurestore:" + f.group
}

// Timeout implements feature.Fetcher.
func (f *Fetcher) Timeout() time.Duration { return f.timeout }

// Fetch implements feature.Fetcher.
// It calls FeatureStoreService.GetFeatures and converts the protobuf response
// into a feature.Map.  On any error it returns an empty map (fail-open) so
// the decision engine continues with partial features rather than failing.
func (f *Fetcher) Fetch(ctx context.Context, req *engine.DecisionRequest) (feature.Map, error) {
	start := time.Now()

	grpcReq := &riskv1.GetFeaturesRequest{
		RequestId:    req.RequestID,
		FeatureGroup: f.group,
		Entity:       entityContext(req),
	}

	resp, err := f.client.GetFeatures(ctx, grpcReq)
	elapsed := time.Since(start).Seconds()
	metrics.FeatureFetchDuration.WithLabelValues(f.Name()).Observe(elapsed)

	if err != nil {
		metrics.FeatureFetchErrors.WithLabelValues(f.Name()).Inc()
		f.logger.Warn("feature store fetch failed, using empty features",
			zap.String("group", f.group),
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)
		// Fail-open: return empty map, non-fatal.
		return feature.Map{}, nil
	}

	if len(resp.MissingKeys) > 0 {
		f.logger.Debug("feature store partial response",
			zap.String("group", f.group),
			zap.Strings("missing_keys", resp.MissingKeys),
		)
	}

	return protoToFeatureMap(resp.Features), nil
}

// ── conversion helpers ───────────────────────────────────────────────────────

// entityContext builds the gRPC EntityContext from a DecisionRequest.
func entityContext(req *engine.DecisionRequest) *riskv1.EntityContext {
	extra := make(map[string]string, len(req.Extra))
	for k, v := range req.Extra {
		extra[k] = v
	}
	return &riskv1.EntityContext{
		UserId:    req.UserID,
		DeviceId:  req.DeviceID,
		SessionId: req.SessionID,
		Ip:        req.IP,
		Extra:     extra,
	}
}

// ProtoMapToFeatureMap converts a proto map[string]*FeatureValue to feature.Map.
// Exported for use in tests and external adapters.
func ProtoMapToFeatureMap(protoMap map[string]*riskv1.FeatureValue) feature.Map {
	return protoToFeatureMap(protoMap)
}

// protoToFeatureMap converts a proto map[string]*FeatureValue to feature.Map.
func protoToFeatureMap(protoMap map[string]*riskv1.FeatureValue) feature.Map {
	if len(protoMap) == 0 {
		return feature.Map{}
	}
	m := make(feature.Map, len(protoMap))
	for key, pv := range protoMap {
		if pv == nil {
			continue
		}
		switch v := pv.Value.(type) {
		case *riskv1.FeatureValue_IntVal:
			m[key] = feature.Value{Kind: feature.KindInt, IntVal: v.IntVal}
		case *riskv1.FeatureValue_FloatVal:
			m[key] = feature.Value{Kind: feature.KindFloat, FltVal: v.FloatVal}
		case *riskv1.FeatureValue_StringVal:
			m[key] = feature.Value{Kind: feature.KindString, StrVal: v.StringVal}
		case *riskv1.FeatureValue_BoolVal:
			m[key] = feature.Value{Kind: feature.KindBool, BoolVal: v.BoolVal}
		default:
			// Unknown type — skip silently.
		}
	}
	return m
}

// HealthChecker wraps Client.Health to satisfy health.Checker.
// Register it with health.CompositeChecker in main.go:
//
//	healthChecker.Add(featurestore.NewHealthChecker(fsClient))
type HealthChecker struct {
	client *Client
	name   string
}

// NewHealthChecker returns a health.Checker for the Feature Store gRPC client.
func NewHealthChecker(client *Client) *HealthChecker {
	return &HealthChecker{client: client, name: "feature-store"}
}

// Check implements health.Checker.
func (h *HealthChecker) Check(ctx context.Context) health.Status {
	resp, err := h.client.Health(ctx)
	if err != nil {
		return health.Status{Name: h.name, Healthy: false, Message: fmt.Sprintf("grpc health: %v", err)}
	}
	if !resp.Healthy {
		return health.Status{Name: h.name, Healthy: false, Message: resp.Message}
	}
	return health.Status{Name: h.name, Healthy: true}
}
