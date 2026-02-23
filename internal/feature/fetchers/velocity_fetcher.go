// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package fetchers contains concrete feature.Fetcher implementations.
package fetchers

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/pkg/sliding"
)

// velocityDim describes one velocity dimension (window + feature key).
type velocityDim struct {
	window time.Duration
	key    string // feature.Map key, e.g. feature.KeyVelocityPayCount1m
}

// VelocityFetcher computes sliding-window velocity features for the requesting
// entity (identified by UserID, DeviceID, or IP) using a Redis sorted-set
// counter maintained by pkg/sliding.
//
// Each call to Fetch does NOT increment the counter — it only reads the current
// count. The increment must be called by the transaction processing path so that
// the count reflects committed events rather than decision-time speculative ones.
//
// If any dimension lookup fails or times out, that feature is set to 0
// (fail-open / conservative default avoids false positives from infra errors).
type VelocityFetcher struct {
	window  *sliding.Window
	timeout time.Duration
	dims    []velocityDim
	// keyPrefix is prepended to UserID to form the Redis sorted-set key.
	// Format: "riskengine:velocity:{prefix}:{entityID}"
	keyPrefix string
}

// VelocityFetcherOption configures a VelocityFetcher.
type VelocityFetcherOption func(*VelocityFetcher)

// WithTimeout overrides the default per-request timeout (default: 10ms).
func WithTimeout(d time.Duration) VelocityFetcherOption {
	return func(f *VelocityFetcher) { f.timeout = d }
}

// NewPaymentVelocityFetcher returns a VelocityFetcher that tracks payment-event
// velocity for a user across 1-minute, 1-hour, and 24-hour windows.
func NewPaymentVelocityFetcher(w *sliding.Window, opts ...VelocityFetcherOption) *VelocityFetcher {
	f := &VelocityFetcher{
		window:    w,
		timeout:   10 * time.Millisecond,
		keyPrefix: "pay",
		dims: []velocityDim{
			{window: time.Minute, key: feature.KeyVelocityPayCount1m},
			{window: time.Hour, key: feature.KeyVelocityPayCount1h},
			{window: 24 * time.Hour, key: feature.KeyVelocityPayCount24h},
		},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// NewPromoVelocityFetcher returns a VelocityFetcher that tracks promotion-claim
// velocity for a user across a 24-hour window.
func NewPromoVelocityFetcher(w *sliding.Window, opts ...VelocityFetcherOption) *VelocityFetcher {
	f := &VelocityFetcher{
		window:    w,
		timeout:   10 * time.Millisecond,
		keyPrefix: "promo",
		dims: []velocityDim{
			{window: 24 * time.Hour, key: feature.KeyVelocityPromoCount1d},
		},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Name implements feature.Fetcher.
func (f *VelocityFetcher) Name() string {
	return "velocity:" + f.keyPrefix
}

// Timeout implements feature.Fetcher.
func (f *VelocityFetcher) Timeout() time.Duration { return f.timeout }

// Fetch implements feature.Fetcher.
// It reads all configured velocity dimensions for the request's UserID.
// Missing or failed dimensions default to 0.
func (f *VelocityFetcher) Fetch(ctx context.Context, req *engine.DecisionRequest) (feature.Map, error) {
	if req.UserID == "" {
		return feature.Map{}, nil
	}

	result := make(feature.Map, len(f.dims))
	for _, d := range f.dims {
		rkey := f.redisKey(req.UserID, d.window)
		count, err := f.window.Get(ctx, rkey, d.window)
		if err != nil {
			// Fail-open: return 0 for this dimension; caller logs the error via
			// feature.Service's degradation path.
			result[d.key] = feature.Value{Kind: feature.KindInt, IntVal: 0}
			continue
		}
		result[d.key] = feature.Value{Kind: feature.KindInt, IntVal: count}
	}
	return result, nil
}

// Increment adds one event for userID to all configured windows.
// Call this from the transaction commit path, NOT from the decision path.
func (f *VelocityFetcher) Increment(ctx context.Context, userID string) error {
	for _, d := range f.dims {
		rkey := f.redisKey(userID, d.window)
		if _, err := f.window.Count(ctx, rkey, d.window); err != nil {
			return fmt.Errorf("velocity increment %s/%s: %w", f.keyPrefix, userID, err)
		}
	}
	return nil
}

func (f *VelocityFetcher) redisKey(userID string, window time.Duration) string {
	return fmt.Sprintf("riskengine:velocity:%s:%s:%ds",
		f.keyPrefix, userID, int(window.Seconds()))
}
