// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/zhucl121/risk-engine/internal/engine"
)

// atomicEvaluator is a thread-safe Evaluator that supports hot-reload via
// an atomic pointer swap (double-buffer pattern).
type atomicEvaluator struct {
	// rules holds a pointer to the current immutable rule slice.
	rules atomic.Pointer[[]Rule]
	// shortCircuit stops evaluation on the first hard REJECT when true.
	shortCircuit bool
}

// NewEvaluator returns a new Evaluator. If shortCircuit is true, evaluation
// stops immediately upon the first rule returning DecisionReject.
func NewEvaluator(rules []Rule, shortCircuit bool) Evaluator {
	e := &atomicEvaluator{shortCircuit: shortCircuit}
	sorted := sortByPriority(rules)
	e.rules.Store(&sorted)
	return e
}

func (e *atomicEvaluator) Evaluate(ctx context.Context, rctx *Context) ([]*Result, error) {
	rules := *e.rules.Load()
	results := make([]*Result, 0, len(rules))

	for _, r := range rules {
		if err := ctx.Err(); err != nil {
			return results, fmt.Errorf("rule evaluator: context cancelled: %w", err)
		}

		start := time.Now()
		res, err := r.Evaluate(ctx, rctx)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", r.ID(), err)
		}
		res.CostMs = time.Since(start).Milliseconds()
		results = append(results, res)

		if res.Hit && res.Decision == engine.DecisionReject && e.shortCircuit {
			return results, nil
		}
	}
	return results, nil
}

func (e *atomicEvaluator) Reload(rules []Rule) error {
	sorted := sortByPriority(rules)
	e.rules.Store(&sorted)
	return nil
}

// sortByPriority returns a copy of rules sorted descending by Priority().
func sortByPriority(rules []Rule) []Rule {
	cp := make([]Rule, len(rules))
	copy(cp, rules)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Priority() > cp[j].Priority()
	})
	return cp
}
