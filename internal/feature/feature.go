// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package feature provides the feature fetching subsystem.
// All feature sources implement the Fetcher interface. The Service orchestrates
// parallel fetching with per-source timeouts and graceful degradation.
package feature

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/yourorg/riskengine/internal/engine"
)

// ValueKind is a tag for the Value union, avoiding interface{} boxing in hot paths.
type ValueKind uint8

const (
	KindInt ValueKind = iota
	KindFloat
	KindString
	KindBool
)

// Value is a typed union storing a single feature value.
type Value struct {
	Kind    ValueKind
	IntVal  int64
	FltVal  float64
	StrVal  string
	BoolVal bool
}

// Map is the unified feature container passed through the decision pipeline.
type Map map[string]Value

// GetInt returns the int64 value for key, or 0 if absent or wrong type.
func (m Map) GetInt(key string) int64 {
	if v, ok := m[key]; ok && v.Kind == KindInt {
		return v.IntVal
	}
	return 0
}

// GetFloat returns the float64 value for key, or 0 if absent or wrong type.
func (m Map) GetFloat(key string) float64 {
	if v, ok := m[key]; ok && v.Kind == KindFloat {
		return v.FltVal
	}
	return 0
}

// GetString returns the string value for key, or "" if absent or wrong type.
func (m Map) GetString(key string) string {
	if v, ok := m[key]; ok && v.Kind == KindString {
		return v.StrVal
	}
	return ""
}

// GetBool returns the bool value for key, or false if absent or wrong type.
func (m Map) GetBool(key string) bool {
	if v, ok := m[key]; ok && v.Kind == KindBool {
		return v.BoolVal
	}
	return false
}

// Fetcher retrieves a subset of features for a given request.
// Implementations must be safe for concurrent use and must honour ctx cancellation.
type Fetcher interface {
	Name()    string
	Timeout() time.Duration
	Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error)
}

// Service orchestrates parallel fetching across all registered Fetchers.
type Service interface {
	Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error)
	Register(f Fetcher)
}

// service is the default Service implementation.
type service struct {
	mu       sync.RWMutex
	fetchers []Fetcher
	logger   *zap.Logger
}

// NewService returns a new feature Service.
func NewService(logger *zap.Logger) Service {
	return &service{logger: logger}
}

// Register adds a Fetcher to the service. Safe to call before serving traffic.
func (s *service) Register(f Fetcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetchers = append(s.fetchers, f)
}

// Fetch queries all registered Fetchers in parallel.
// Fetcher timeouts are enforced independently; a timeout in one source
// does not cancel others and never prevents returning a partial result.
func (s *service) Fetch(ctx context.Context, req *engine.DecisionRequest) (Map, error) {
	s.mu.RLock()
	fetchers := s.fetchers
	s.mu.RUnlock()

	result := make(Map, len(fetchers)*8)
	var mu sync.Mutex

	g, _ := errgroup.WithContext(ctx)
	for _, f := range fetchers {
		f := f
		g.Go(func() error {
			fCtx, cancel := context.WithTimeout(ctx, f.Timeout())
			defer cancel()

			m, err := f.Fetch(fCtx, req)
			if err != nil {
				s.logger.Warn("feature fetch degraded",
					zap.String("fetcher", f.Name()),
					zap.Error(err),
				)
				return nil // non-fatal; continue with partial features
			}
			mu.Lock()
			for k, v := range m {
				result[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()
	return result, nil
}
