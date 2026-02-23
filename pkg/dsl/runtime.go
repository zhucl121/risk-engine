// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package dsl

import (
	"context"
	"sync"

	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
)

// Runtime holds all per-request dependencies that DSL functions can access.
// It is passed to Program.Run on every evaluation call.
//
// Runtime instances should be obtained from RuntimePool to minimise allocations
// in the hot path. Callers must call RuntimePool.Put after use.
type Runtime struct {
	Features feature.Map
	Request  *engine.DecisionRequest

	// Service dependencies injected at startup; safe for concurrent read.
	// Functions that need these (inList, velocity, modelScore) receive *Runtime.
	ListChecker  ListChecker
	VelocityFunc VelocityFunc
	ModelScorer  ModelScorer
}

// ListChecker is the dependency interface that inList() calls.
type ListChecker interface {
	InList(ctx context.Context, listName string, value string) (bool, error)
}

// VelocityFunc is the dependency interface that velocity() calls.
type VelocityFunc interface {
	Count(ctx context.Context, event string, window string, key string) (int64, error)
}

// ModelScorer is the dependency interface that modelScore() calls.
type ModelScorer interface {
	Score(ctx context.Context, modelName string, features feature.Map) (float64, error)
}

// reset clears mutable fields before returning to pool.
func (r *Runtime) reset() {
	r.Features = nil
	r.Request = nil
}

// runtimePool is a process-wide pool of Runtime objects.
var runtimePool = sync.Pool{
	New: func() interface{} { return &Runtime{} },
}

// AcquireRuntime obtains a Runtime from the pool.
// Callers MUST call ReleaseRuntime when done to avoid leaking memory.
func AcquireRuntime() *Runtime {
	return runtimePool.Get().(*Runtime)
}

// ReleaseRuntime returns a Runtime to the pool.
func ReleaseRuntime(r *Runtime) {
	r.reset()
	runtimePool.Put(r)
}
