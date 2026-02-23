// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package rule provides the rule evaluation subsystem. Rules are stateless
// functions that inspect a feature map and return a risk verdict. They are
// loaded from YAML configuration and can be hot-reloaded atomically.
package rule

import (
	"context"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
)

// Context bundles all inputs available to a rule during evaluation.
type Context struct {
	Request  *engine.DecisionRequest
	Features feature.Map
}

// Result is the output of a single rule evaluation.
type Result struct {
	RuleID   string
	RuleName string
	Hit      bool
	// Score is added to the aggregate risk score only when Hit is true.
	Score    int
	Decision engine.Decision
	// RiskCode is a machine-readable reason code, e.g. "DEVICE_MULTI_ACCOUNT".
	RiskCode string
	CostMs   int64
}

// Rule is a single, stateless risk check.
// Implementations must be safe for concurrent use.
type Rule interface {
	ID()       string
	Name()     string
	Priority() int
	// Evaluate inspects rctx and returns a Result.
	// It must not mutate rctx.
	Evaluate(ctx context.Context, rctx *Context) (*Result, error)
}

// Evaluator executes a collection of rules and aggregates their results.
type Evaluator interface {
	// Evaluate runs all configured rules and returns the full result set.
	// Short-circuit behaviour (stop on first hard reject) is controlled by
	// the Evaluator implementation's configuration.
	Evaluate(ctx context.Context, rctx *Context) ([]*Result, error)

	// Reload atomically replaces the active rule set with the provided rules.
	// In-flight evaluations continue with the old set until completion.
	Reload(rules []Rule) error
}
