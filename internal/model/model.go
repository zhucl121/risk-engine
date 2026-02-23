// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package model provides ML model inference interfaces and the
// champion-challenger routing registry. Models are loaded as ONNX files
// and executed in a CGO-isolated goroutine pool to prevent scheduler starvation.
package model

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/yourorg/riskengine/internal/feature"
)

// Scorer runs ML inference for a single model and returns a risk score in [0, 1].
// Implementations must be safe for concurrent use.
type Scorer interface {
	Name() string
	// Score returns a probability in [0.0, 1.0]; higher means more risky.
	Score(ctx context.Context, features feature.Map) (float64, error)
}

// challenger pairs a Scorer with its traffic allocation percentage.
type challenger struct {
	scorer     Scorer
	trafficPct float64 // 0.0–1.0
}

// entry holds the active champion and optional challenger for one model name.
type entry struct {
	mu         sync.RWMutex
	champion   Scorer
	challenger *challenger
}

func (e *entry) route() Scorer {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.challenger != nil && rand.Float64() < e.challenger.trafficPct { //nolint:gosec
		return e.challenger.scorer
	}
	return e.champion
}

// Registry manages named scorers with champion-challenger routing.
// All methods are safe for concurrent use.
type Registry interface {
	// Score routes the request to the champion (or challenger) for modelName.
	Score(ctx context.Context, modelName string, features feature.Map) (float64, error)

	// Register adds or replaces the champion scorer for modelName.
	Register(name string, s Scorer)

	// SetChallenger activates a challenger for modelName that receives
	// trafficPct fraction of traffic (0.0–1.0).
	SetChallenger(name string, s Scorer, trafficPct float64)

	// ClearChallenger removes the challenger for modelName.
	ClearChallenger(name string)
}

type registry struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

// NewRegistry returns a new empty model Registry.
func NewRegistry() Registry {
	return &registry{entries: make(map[string]*entry)}
}

func (r *registry) Register(name string, s Scorer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.entries[name]; ok {
		e.mu.Lock()
		e.champion = s
		e.mu.Unlock()
		return
	}
	r.entries[name] = &entry{champion: s}
}

func (r *registry) SetChallenger(name string, s Scorer, trafficPct float64) {
	r.mu.RLock()
	e, ok := r.entries[name]
	r.mu.RUnlock()
	if !ok {
		return
	}
	e.mu.Lock()
	e.challenger = &challenger{scorer: s, trafficPct: trafficPct}
	e.mu.Unlock()
}

func (r *registry) ClearChallenger(name string) {
	r.mu.RLock()
	e, ok := r.entries[name]
	r.mu.RUnlock()
	if !ok {
		return
	}
	e.mu.Lock()
	e.challenger = nil
	e.mu.Unlock()
}

func (r *registry) Score(ctx context.Context, modelName string, features feature.Map) (float64, error) {
	r.mu.RLock()
	e, ok := r.entries[modelName]
	r.mu.RUnlock()
	if !ok {
		return 0, &ErrModelNotFound{Name: modelName}
	}
	return e.route().Score(ctx, features)
}

// requestCounter is a simple atomic counter for round-robin or sampling.
// Kept here as a utility; use rand for champion-challenger instead.
type requestCounter struct{ n atomic.Uint64 }

func (c *requestCounter) Inc() uint64 { return c.n.Add(1) }
