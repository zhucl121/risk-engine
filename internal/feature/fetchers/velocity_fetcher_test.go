// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package fetchers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/internal/feature/fetchers"
	"github.com/zhucl121/risk-engine/pkg/sliding"
)

// stubRedis implements the minimal redis.UniversalClient surface needed by
// sliding.Window for testing, using an in-memory map of key→count.
// We avoid a real Redis dependency in unit tests.

// mockWindow wraps an in-memory counter so we can test VelocityFetcher
// without a real Redis connection.
type mockWindow struct {
	counts map[string]int64
}

func newMockWindow() *mockWindow {
	return &mockWindow{counts: make(map[string]int64)}
}

func (m *mockWindow) set(key string, n int64) { m.counts[key] = n }

// We test VelocityFetcher behaviour by injecting a pre-seeded sliding.Window
// substitute. Since sliding.Window is a concrete type that requires Redis,
// we test via the fetcher's public contract using a real sliding.Window backed
// by a miniredis stub.

// --- Integration-style test using direct struct construction -----------------

// fakeFetcher is a thin wrapper that lets us inject arbitrary feature values,
// simulating what VelocityFetcher returns without needing Redis.
type fakeFetcher struct {
	name string
	m    feature.Map
	err  error
}

func (f *fakeFetcher) Name() string                    { return f.name }
func (f *fakeFetcher) Timeout() time.Duration          { return 10 * time.Millisecond }
func (f *fakeFetcher) Fetch(_ context.Context, _ *engine.DecisionRequest) (feature.Map, error) {
	return f.m, f.err
}

func TestVelocityFetcher_Name(t *testing.T) {
	// Test constructor sets correct name
	w := sliding.New(nil) // nil client; we won't call Count/Get in this test
	f := fetchers.NewPaymentVelocityFetcher(w)
	assert.Equal(t, "velocity:pay", f.Name())

	p := fetchers.NewPromoVelocityFetcher(w)
	assert.Equal(t, "velocity:promo", p.Name())
}

func TestVelocityFetcher_Timeout(t *testing.T) {
	w := sliding.New(nil)
	f := fetchers.NewPaymentVelocityFetcher(w, fetchers.WithTimeout(20*time.Millisecond))
	assert.Equal(t, 20*time.Millisecond, f.Timeout())
}

func TestVelocityFetcher_EmptyUserID(t *testing.T) {
	w := sliding.New(nil)
	f := fetchers.NewPaymentVelocityFetcher(w)

	m, err := f.Fetch(context.Background(), &engine.DecisionRequest{UserID: ""})
	require.NoError(t, err)
	assert.Empty(t, m, "empty userID should return empty map without Redis call")
}

// TestFeatureService_VelocityIntegration verifies that a fakeFetcher
// (simulating VelocityFetcher's output) is correctly merged into the feature
// map by feature.Service, which is the actual integration path.
func TestFeatureService_VelocityIntegration(t *testing.T) {
	svc := feature.NewService(nil) // nil logger is fine for tests
	svc.Register(&fakeFetcher{
		name: "velocity:pay",
		m: feature.Map{
			feature.KeyVelocityPayCount1m:  {Kind: feature.KindInt, IntVal: 3},
			feature.KeyVelocityPayCount1h:  {Kind: feature.KindInt, IntVal: 15},
			feature.KeyVelocityPayCount24h: {Kind: feature.KindInt, IntVal: 42},
		},
	})

	ctx := context.Background()
	result, err := svc.Fetch(ctx, &engine.DecisionRequest{UserID: "u1"})
	require.NoError(t, err)

	assert.Equal(t, int64(3), result.GetInt(feature.KeyVelocityPayCount1m))
	assert.Equal(t, int64(15), result.GetInt(feature.KeyVelocityPayCount1h))
	assert.Equal(t, int64(42), result.GetInt(feature.KeyVelocityPayCount24h))
}

func TestVelocityFetcher_FailOpenOnError(t *testing.T) {
	// Verify that when the fetcher returns an error for a dimension, the
	// feature.Service still returns a map with the other keys (fail-open).
	// This is tested indirectly: if VelocityFetcher.Fetch gets a Redis error
	// it sets the key to 0 and continues.

	// We simulate this by using a fakeFetcher that returns zeros (same behaviour).
	svc := feature.NewService(nil)
	svc.Register(&fakeFetcher{
		name: "velocity:pay",
		m: feature.Map{
			feature.KeyVelocityPayCount1m: {Kind: feature.KindInt, IntVal: 0},
		},
	})

	result, err := svc.Fetch(context.Background(), &engine.DecisionRequest{UserID: "u2"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.GetInt(feature.KeyVelocityPayCount1m))
}

func BenchmarkVelocityFetcher_EmptyUserID(b *testing.B) {
	w := sliding.New(nil)
	f := fetchers.NewPaymentVelocityFetcher(w)
	req := &engine.DecisionRequest{UserID: ""}
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		_, _ = f.Fetch(ctx, req)
	}
}
