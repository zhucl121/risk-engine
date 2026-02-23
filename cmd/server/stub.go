// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/yourorg/riskengine/internal/rule"
)

// noopRuleEvaluator is a rule.Evaluator that always returns an empty result set.
// It is used as a placeholder when no rules are loaded at startup.
type noopRuleEvaluator struct{}

func newNoopRuleEvaluator() rule.Evaluator { return &noopRuleEvaluator{} }

func (n *noopRuleEvaluator) Evaluate(_ context.Context, _ *rule.Context) ([]*rule.Result, error) {
	return nil, nil
}

func (n *noopRuleEvaluator) Reload(_ []rule.Rule) error { return nil }
