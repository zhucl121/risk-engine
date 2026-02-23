// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package store defines the extension interfaces for the Feature Store server.
// Each feature group (e.g. "user_profile", "device_risk", "velocity") is backed
// by a FeatureGroup implementation that knows how to fetch from its data source.
package store

import (
	"context"
	"fmt"
	"sync"

	riskv1 "github.com/zhucl121/risk-engine/api/grpc/v1"
)

// FeatureGroup fetches features for a named group given an entity context.
// Implementations should be safe for concurrent use.
type FeatureGroup interface {
	// Name returns the group identifier (must match what clients request).
	Name() string
	// Fetch returns a map of feature key → proto FeatureValue, plus any
	// key names that were requested but had no data.
	Fetch(ctx context.Context, entity *riskv1.EntityContext) (map[string]*riskv1.FeatureValue, []string, error)
}

// GroupRegistry maps group names to their FeatureGroup implementations.
type GroupRegistry interface {
	// Get returns the FeatureGroup for the given name, or an error if unknown.
	Get(name string) (FeatureGroup, error)
	// BackendStatuses returns a map of backend name → healthy for /health.
	BackendStatuses() map[string]bool
}

// DefaultRegistry is the in-memory GroupRegistry implementation.
// Use NewRegistry() to construct; the type is exported so callers can call
// Register() and RegisterBackend() beyond the GroupRegistry interface.
type DefaultRegistry struct {
	mu       sync.RWMutex
	groups   map[string]FeatureGroup
	backends map[string]func() bool // name → liveness probe
}

// NewRegistry returns an empty *DefaultRegistry as a GroupRegistry.
func NewRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		groups:   make(map[string]FeatureGroup),
		backends: make(map[string]func() bool),
	}
}

// Register adds a FeatureGroup. Safe to call before serving traffic.
func (r *DefaultRegistry) Register(g FeatureGroup) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[g.Name()] = g
}

// RegisterBackend adds a named liveness probe for health reporting.
func (r *DefaultRegistry) RegisterBackend(name string, probe func() bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[name] = probe
}

// Get implements GroupRegistry.
func (r *DefaultRegistry) Get(name string) (FeatureGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[name]
	if !ok {
		return nil, fmt.Errorf("store: feature group %q not registered", name)
	}
	return g, nil
}

// BackendStatuses implements GroupRegistry.
func (r *DefaultRegistry) BackendStatuses() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]bool, len(r.backends))
	for name, probe := range r.backends {
		out[name] = probe()
	}
	return out
}
