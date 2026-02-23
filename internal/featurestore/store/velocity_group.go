// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"fmt"
	"time"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/pkg/sliding"
)

// velocityWindow is one sliding-window dimension exposed to the Feature Store.
type velocityWindow struct {
	featureKey string        // the key clients receive in the response map
	redisKey   string        // Redis sorted-set key template (%s = entityID)
	window     time.Duration // window size
}

// VelocityGroup implements FeatureGroup for sliding-window velocity features.
// It re-uses pkg/sliding.Window (the same Redis Lua script used by
// the in-process VelocityFetcher) so Redis sorted-sets are shared.
type VelocityGroup struct {
	name    string
	win     *sliding.Window
	windows []velocityWindow
}

// NewPaymentVelocityGroup returns a FeatureGroup for payment velocity features.
func NewPaymentVelocityGroup(w *sliding.Window) *VelocityGroup {
	return &VelocityGroup{
		name: "velocity",
		win:  w,
		windows: []velocityWindow{
			{
				featureKey: feature.KeyVelocityPayCount1m,
				redisKey:   "riskengine:velocity:pay:%s:60s",
				window:     time.Minute,
			},
			{
				featureKey: feature.KeyVelocityPayCount1h,
				redisKey:   "riskengine:velocity:pay:%s:3600s",
				window:     time.Hour,
			},
			{
				featureKey: feature.KeyVelocityPayCount24h,
				redisKey:   "riskengine:velocity:pay:%s:86400s",
				window:     24 * time.Hour,
			},
			{
				featureKey: feature.KeyVelocityPromoCount1d,
				redisKey:   "riskengine:velocity:promo:%s:86400s",
				window:     24 * time.Hour,
			},
		},
	}
}

// Name implements FeatureGroup.
func (g *VelocityGroup) Name() string { return g.name }

// Fetch implements FeatureGroup.
// Reads each configured window from Redis; missing/failed entries default to 0.
func (g *VelocityGroup) Fetch(ctx context.Context, entity *riskv1.EntityContext) (map[string]*riskv1.FeatureValue, []string, error) {
	if entity.UserId == "" {
		return map[string]*riskv1.FeatureValue{}, nil, nil
	}
	out := make(map[string]*riskv1.FeatureValue, len(g.windows))
	var missing []string

	for _, w := range g.windows {
		key := fmt.Sprintf(w.redisKey, entity.UserId)
		count, err := g.win.Get(ctx, key, w.window)
		if err != nil {
			missing = append(missing, w.featureKey)
			out[w.featureKey] = &riskv1.FeatureValue{Value: &riskv1.FeatureValue_IntVal{IntVal: 0}}
			continue
		}
		out[w.featureKey] = &riskv1.FeatureValue{Value: &riskv1.FeatureValue_IntVal{IntVal: count}}
	}
	return out, missing, nil
}
