// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package resilience provides circuit-breaker primitives for protecting
// external service calls (Redis, model inference, etc.) from cascading failures.
//
// Design: wraps github.com/sony/gobreaker with a Prometheus state gauge and
// a simple Execute(fn) helper so callers don't need to import gobreaker directly.
package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"

	"github.com/yourorg/riskengine/internal/metrics"
)

// Breaker is a named circuit breaker that tracks its state in Prometheus.
type Breaker struct {
	name string
	cb   *gobreaker.CircuitBreaker
}

// BreakerConfig controls the circuit breaker thresholds.
type BreakerConfig struct {
	// MaxRequests is the max number of requests allowed in half-open state.
	MaxRequests uint32
	// Interval is the cyclic period of the closed state to clear counts (0 = never clears).
	Interval time.Duration
	// Timeout is how long the breaker stays open before switching to half-open.
	Timeout time.Duration
	// FailureThreshold is the minimum number of failures in a closed window to trip the breaker.
	FailureThreshold uint32
}

// DefaultBreakerConfig returns conservative production defaults:
// open after 5 consecutive failures, retry after 30s.
func DefaultBreakerConfig() BreakerConfig {
	return BreakerConfig{
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
	}
}

// New returns a Breaker with the given name and configuration.
// State changes are automatically recorded in the riskengine_circuit_breaker_state gauge.
func New(name string, cfg BreakerConfig) *Breaker {
	b := &Breaker{name: name}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.FailureThreshold
		},
		OnStateChange: func(_ string, from, to gobreaker.State) {
			metrics.CircuitBreakerState.WithLabelValues(name).Set(float64(to))
			_ = from
		},
	}
	b.cb = gobreaker.NewCircuitBreaker(settings)

	// Initialise to closed (0).
	metrics.CircuitBreakerState.WithLabelValues(name).Set(0)
	return b
}

// Execute runs fn through the circuit breaker.
// If the breaker is open, fn is not called and ErrOpen is returned.
// The context is passed to fn for timeout/cancellation propagation.
func (b *Breaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	_, err := b.cb.Execute(func() (interface{}, error) {
		return nil, fn(ctx)
	})
	if err == gobreaker.ErrOpenState {
		return fmt.Errorf("circuit breaker %q is open: %w", b.name, ErrOpen)
	}
	if err == gobreaker.ErrTooManyRequests {
		return fmt.Errorf("circuit breaker %q too many requests in half-open: %w", b.name, ErrOpen)
	}
	return err
}

// State returns the current breaker state as a human-readable string.
func (b *Breaker) State() string {
	return b.cb.State().String()
}

// ErrOpen is returned when a call is rejected because the breaker is open.
var ErrOpen = fmt.Errorf("circuit breaker open")
