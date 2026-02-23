// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package featurestore_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	riskv1 "github.com/yourorg/riskengine/api/grpc/v1"
	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/featurestore"
)

// fakeFetcher stubs the feature.Fetcher interface without a real gRPC client.
// This lets us test feature.Service integration and fail-open behaviour
// without a real Feature Store process.
type fakeFetcher struct {
	name    string
	retMap  feature.Map
	retErr  bool
	timeout time.Duration
}

func (f *fakeFetcher) Name() string           { return f.name }
func (f *fakeFetcher) Timeout() time.Duration { return f.timeout }
func (f *fakeFetcher) Fetch(_ context.Context, _ *engine.DecisionRequest) (feature.Map, error) {
	if f.retErr {
		return feature.Map{}, status.Error(codes.Unavailable, "store down")
	}
	return f.retMap, nil
}

// ── Fetcher metadata ─────────────────────────────────────────────────────────

func TestFetcher_Name(t *testing.T) {
	f := featurestore.NewFetcher(nil, "user_profile", 20*time.Millisecond, zaptest.NewLogger(t))
	assert.Equal(t, "featurestore:user_profile", f.Name())
}

func TestFetcher_Timeout(t *testing.T) {
	f := featurestore.NewFetcher(nil, "velocity", 15*time.Millisecond, zaptest.NewLogger(t))
	assert.Equal(t, 15*time.Millisecond, f.Timeout())
}

// ── Feature service integration ───────────────────────────────────────────────

func TestFeatureService_WithStoreFetcher(t *testing.T) {
	svc := feature.NewService(zaptest.NewLogger(t))
	svc.Register(&fakeFetcher{
		name:    "featurestore:user_profile",
		timeout: 20 * time.Millisecond,
		retMap: feature.Map{
			feature.KeyUserRegisterDays: {Kind: feature.KindInt, IntVal: 180},
			feature.KeyUserCreditScore:  {Kind: feature.KindFloat, FltVal: 720.5},
		},
	})

	result, err := svc.Fetch(context.Background(), &engine.DecisionRequest{UserID: "u1"})
	require.NoError(t, err)
	assert.Equal(t, int64(180), result.GetInt(feature.KeyUserRegisterDays))
	assert.InDelta(t, 720.5, result.GetFloat(feature.KeyUserCreditScore), 0.001)
}

func TestFetcher_FailOpen(t *testing.T) {
	svc := feature.NewService(zaptest.NewLogger(t))
	svc.Register(&fakeFetcher{
		name:    "featurestore:user_profile",
		timeout: 20 * time.Millisecond,
		retErr:  true,
	})

	result, err := svc.Fetch(context.Background(), &engine.DecisionRequest{UserID: "u1"})
	// Service must not propagate fetcher errors.
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.GetInt(feature.KeyUserRegisterDays))
}

// ── ProtoMapToFeatureMap type conversion ──────────────────────────────────────

func TestProtoMapToFeatureMap_Int(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(map[string]*riskv1.FeatureValue{
		"k": {Value: &riskv1.FeatureValue_IntVal{IntVal: 42}},
	})
	assert.Equal(t, int64(42), m.GetInt("k"))
}

func TestProtoMapToFeatureMap_Float(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(map[string]*riskv1.FeatureValue{
		"k": {Value: &riskv1.FeatureValue_FloatVal{FloatVal: 3.14}},
	})
	assert.InDelta(t, 3.14, m.GetFloat("k"), 0.001)
}

func TestProtoMapToFeatureMap_String(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(map[string]*riskv1.FeatureValue{
		"k": {Value: &riskv1.FeatureValue_StringVal{StringVal: "CN"}},
	})
	assert.Equal(t, "CN", m.GetString("k"))
}

func TestProtoMapToFeatureMap_Bool(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(map[string]*riskv1.FeatureValue{
		"k": {Value: &riskv1.FeatureValue_BoolVal{BoolVal: true}},
	})
	assert.True(t, m.GetBool("k"))
}

func TestProtoMapToFeatureMap_Nil(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(nil)
	assert.Empty(t, m)
}

func TestProtoMapToFeatureMap_NilValue(t *testing.T) {
	m := featurestore.ProtoMapToFeatureMap(map[string]*riskv1.FeatureValue{"k": nil})
	assert.Empty(t, m) // nil entries are skipped
}
