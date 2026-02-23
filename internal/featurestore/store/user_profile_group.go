// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
	"github.com/zhucl121/risk-engine/internal/feature"
)

// userProfileRecord is the JSON schema stored in Redis for each user.
// Keys align with feature.Key* constants so rule conditions can reference them directly.
type userProfileRecord struct {
	RegisterDays        int64   `json:"register_days"`
	HistoryFraudCount   int64   `json:"history_fraud_count"`
	CreditScore         float64 `json:"credit_score"`
	ActiveOrderCount30d int64   `json:"active_order_count_30d"`
}

// UserProfileGroup implements FeatureGroup for user profile features backed by Redis.
// Records are stored as JSON hashes under key "riskengine:profile:user:{userID}".
// TTL is set by the upstream user-profile update job (typically 24h).
type UserProfileGroup struct {
	rdb     redis.UniversalClient
	timeout time.Duration
}

// NewUserProfileGroup returns a UserProfileGroup using the given Redis client.
func NewUserProfileGroup(rdb redis.UniversalClient) *UserProfileGroup {
	return &UserProfileGroup{rdb: rdb, timeout: 10 * time.Millisecond}
}

// Name implements FeatureGroup.
func (g *UserProfileGroup) Name() string { return "user_profile" }

// Fetch implements FeatureGroup.
// Returns user profile features; missing entries default to zero values.
func (g *UserProfileGroup) Fetch(ctx context.Context, entity *riskv1.EntityContext) (map[string]*riskv1.FeatureValue, []string, error) {
	if entity.UserId == "" {
		return map[string]*riskv1.FeatureValue{}, nil, nil
	}

	key := fmt.Sprintf("riskengine:profile:user:%s", entity.UserId)
	raw, err := g.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// No profile data yet — return zero defaults for all keys.
		return defaultUserProfile(), allUserProfileKeys(), nil
	}
	if err != nil {
		return defaultUserProfile(), allUserProfileKeys(), fmt.Errorf("user_profile: redis get %s: %w", key, err)
	}

	var rec userProfileRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return defaultUserProfile(), allUserProfileKeys(), fmt.Errorf("user_profile: unmarshal %s: %w", key, err)
	}

	return map[string]*riskv1.FeatureValue{
		feature.KeyUserRegisterDays:        intFV(rec.RegisterDays),
		feature.KeyUserHistoryFraud:        intFV(rec.HistoryFraudCount),
		feature.KeyUserCreditScore:         floatFV(rec.CreditScore),
		feature.KeyUserActiveOrderCount30d: intFV(rec.ActiveOrderCount30d),
	}, nil, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func intFV(v int64) *riskv1.FeatureValue {
	return &riskv1.FeatureValue{Value: &riskv1.FeatureValue_IntVal{IntVal: v}}
}

func floatFV(v float64) *riskv1.FeatureValue {
	return &riskv1.FeatureValue{Value: &riskv1.FeatureValue_FloatVal{FloatVal: v}}
}

func allUserProfileKeys() []string {
	return []string{
		feature.KeyUserRegisterDays,
		feature.KeyUserHistoryFraud,
		feature.KeyUserCreditScore,
		feature.KeyUserActiveOrderCount30d,
	}
}

func defaultUserProfile() map[string]*riskv1.FeatureValue {
	return map[string]*riskv1.FeatureValue{
		feature.KeyUserRegisterDays:        intFV(0),
		feature.KeyUserHistoryFraud:        intFV(0),
		feature.KeyUserCreditScore:         floatFV(0),
		feature.KeyUserActiveOrderCount30d: intFV(0),
	}
}
