// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package health provides liveness and readiness probes for Kubernetes.
//
// Liveness (/livez):   always returns 200 while the process is running.
// Readiness (/readyz): runs all registered Checkers; returns 503 if any fail.
package health

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Status is the result of a single health check.
type Status struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// Checker probes a single dependency.
type Checker interface {
	Check(ctx context.Context) Status
}

// CompositeChecker runs multiple Checkers and aggregates results.
type CompositeChecker struct {
	checkers []Checker
}

// NewCompositeChecker returns a CompositeChecker with the given Checkers.
func NewCompositeChecker(checkers ...Checker) *CompositeChecker {
	return &CompositeChecker{checkers: checkers}
}

// Add appends a Checker to the composite.
func (c *CompositeChecker) Add(ch Checker) {
	c.checkers = append(c.checkers, ch)
}

// CheckAll runs all Checkers with the given context and returns their results.
// The boolean return is true only when ALL checkers are healthy.
func (c *CompositeChecker) CheckAll(ctx context.Context) ([]Status, bool) {
	results := make([]Status, 0, len(c.checkers))
	allHealthy := true
	for _, ch := range c.checkers {
		s := ch.Check(ctx)
		results = append(results, s)
		if !s.Healthy {
			allHealthy = false
		}
	}
	return results, allHealthy
}

// ─── Concrete checkers ───────────────────────────────────────────────────────

// RedisChecker pings a Redis client.
type RedisChecker struct {
	name   string
	client redis.UniversalClient
}

// NewRedisChecker returns a Checker that pings the given Redis client.
func NewRedisChecker(name string, client redis.UniversalClient) *RedisChecker {
	return &RedisChecker{name: name, client: client}
}

// Check implements Checker.
func (r *RedisChecker) Check(ctx context.Context) Status {
	pCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := r.client.Ping(pCtx).Err(); err != nil {
		return Status{Name: r.name, Healthy: false, Message: fmt.Sprintf("ping failed: %v", err)}
	}
	return Status{Name: r.name, Healthy: true}
}

// FuncChecker wraps an arbitrary check function.
type FuncChecker struct {
	name string
	fn   func(ctx context.Context) error
}

// NewFuncChecker returns a Checker backed by fn.
func NewFuncChecker(name string, fn func(ctx context.Context) error) *FuncChecker {
	return &FuncChecker{name: name, fn: fn}
}

// Check implements Checker.
func (f *FuncChecker) Check(ctx context.Context) Status {
	if err := f.fn(ctx); err != nil {
		return Status{Name: f.name, Healthy: false, Message: err.Error()}
	}
	return Status{Name: f.name, Healthy: true}
}
