// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package engine provides the top-level decision orchestration interface.
// It wires together the rule evaluator, model scorer, list service, and
// feature service into a single Evaluate call that returns a DecisionResult.
package engine

import (
	"context"
	"time"
)

// Decision represents the outcome of a risk evaluation.
type Decision string

const (
	DecisionPass         Decision = "PASS"
	DecisionReject       Decision = "REJECT"
	DecisionManualReview Decision = "MANUAL_REVIEW"
)

// RiskLevel represents the severity tier of a decision.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// HealthStatus reports engine operational state.
type HealthStatus struct {
	Healthy    bool
	Components map[string]bool
}

// DecisionRequest carries all input data for a single risk evaluation.
//
// Design: only universal identity/routing fields are promoted to top-level.
// All scene-specific business data (e.g. amount, merchant_id, order_type)
// must be passed via Extra so the engine remains domain-agnostic and
// applicable to non-payment scenarios (login, account, content risk, etc.).
type DecisionRequest struct {
	// RequestID is a UUID v4 set by the caller; generated if empty.
	RequestID string
	// SceneCode identifies the business scenario, e.g. "PAYMENT_CHECKOUT".
	SceneCode  string
	UserID     string
	DeviceID   string
	SessionID  string
	IP         string
	// Extra holds all scene-specific key-value pairs, including business
	// fields such as "amount", "merchant_id", "order_type", etc.
	Extra      map[string]string
	ReceivedAt time.Time
}

// StepTrace records the execution details of one pipeline step.
type StepTrace struct {
	Name    string
	CostMs  int64
	Skipped bool
	Error   string
}

// DecisionResult is the output of Engine.Evaluate.
type DecisionResult struct {
	RequestID   string
	Decision    Decision
	RiskScore   int // 0–1000; higher means more risky
	RiskLevel   RiskLevel
	HitRules    []string
	ModelScores map[string]float64
	// RiskReasons contains machine-readable reason codes.
	RiskReasons []string
	// Actions contains recommended enforcement actions.
	Actions  []string
	Path     []StepTrace
	CostMs   int64
}

// Engine is the primary interface for risk decision making.
// All implementations must be safe for concurrent use.
type Engine interface {
	// Evaluate performs a synchronous risk decision.
	// It returns an error only for infrastructure failures; business risk
	// decisions are always encoded in DecisionResult.
	Evaluate(ctx context.Context, req *DecisionRequest) (*DecisionResult, error)

	// Reload triggers a hot-reload of rules, models, and policy configs.
	Reload(ctx context.Context) error

	// Health returns the current operational status of the engine and its
	// sub-components.
	Health() HealthStatus
}
