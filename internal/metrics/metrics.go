// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package metrics defines all Prometheus collectors for the riskengine.
// Call Register() once at startup to register them with the default registry.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collectors holds all application-level Prometheus metrics.
// Use the package-level functions (RecordDecision, IncRuleHit, …) rather
// than accessing these fields directly.
var (
	// DecisionDuration tracks end-to-end decision latency per scene + outcome.
	DecisionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "riskengine",
			Name:      "decision_duration_seconds",
			Help:      "End-to-end decision latency in seconds.",
			Buckets:   []float64{.005, .01, .025, .05, .075, .1, .25, .5, 1},
		},
		[]string{"scene_code", "decision"},
	)

	// DecisionTotal counts completed decisions.
	DecisionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "riskengine",
			Name:      "decisions_total",
			Help:      "Total number of decisions made.",
		},
		[]string{"scene_code", "decision"},
	)

	// RuleHitsTotal counts individual rule trigger events.
	RuleHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "riskengine",
			Name:      "rule_hits_total",
			Help:      "Total number of rule hit events.",
		},
		[]string{"rule_id", "scene_code"},
	)

	// FeatureFetchErrors counts per-fetcher degradation events.
	FeatureFetchErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "riskengine",
			Name:      "feature_fetch_errors_total",
			Help:      "Total number of feature fetch errors (degraded to default).",
		},
		[]string{"fetcher"},
	)

	// FeatureFetchDuration tracks per-fetcher latency.
	FeatureFetchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "riskengine",
			Name:      "feature_fetch_duration_seconds",
			Help:      "Feature fetch latency per fetcher.",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1},
		},
		[]string{"fetcher"},
	)

	// ActiveRequests is the current number of in-flight decision requests.
	ActiveRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "riskengine",
		Name:      "active_requests",
		Help:      "Number of decision requests currently being processed.",
	})

	// HTTPRequestDuration tracks per-endpoint HTTP latency (used by metrics middleware).
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "riskengine",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestTotal counts HTTP requests.
	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "riskengine",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// RateLimitedTotal counts requests rejected by the rate limiter.
	RateLimitedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "riskengine",
			Name:      "rate_limited_total",
			Help:      "Total number of requests rejected by the rate limiter.",
		},
		[]string{"path"},
	)

	// CircuitBreakerState tracks breaker state changes (0=closed, 1=open, 2=half-open).
	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "riskengine",
		Name:      "circuit_breaker_state",
		Help:      "Current circuit breaker state (0=closed, 1=open, 2=half-open).",
	}, []string{"name"})
)
