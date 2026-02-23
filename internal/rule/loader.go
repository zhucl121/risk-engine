// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// DBLoader loads rules from a RuleRepository, compiles their DSL conditions,
// and returns a slice of Rule implementations ready for the Evaluator.
type DBLoader struct {
	repo   RuleRepository
	dslReg *dsl.FunctionRegistry
	logger *zap.Logger
}

// NewDBLoader creates a DBLoader.
// reg must have all required DSL functions registered (via builtins.RegisterAll).
func NewDBLoader(repo RuleRepository, reg *dsl.FunctionRegistry, logger *zap.Logger) *DBLoader {
	return &DBLoader{repo: repo, dslReg: reg, logger: logger}
}

// LoadAll compiles all active rules for the given scene and returns them sorted
// by priority (descending). An empty sceneCode loads all scenes.
func (l *DBLoader) LoadAll(ctx context.Context, sceneCode string) ([]Rule, error) {
	records, err := l.repo.ListActive(ctx, sceneCode)
	if err != nil {
		return nil, fmt.Errorf("DBLoader.LoadAll: %w", err)
	}
	return l.compile(records)
}

// LoadUpdatedSince compiles rules that changed after since.
// Used by the hot-reload watcher for incremental updates.
func (l *DBLoader) LoadUpdatedSince(ctx context.Context, since time.Time) ([]Rule, error) {
	records, err := l.repo.ListUpdatedSince(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("DBLoader.LoadUpdatedSince: %w", err)
	}
	return l.compile(records)
}

func (l *DBLoader) compile(records []*RuleRecord) ([]Rule, error) {
	rules := make([]Rule, 0, len(records))
	for _, rec := range records {
		if rec.Status != 1 {
			continue
		}
		prog, err := dsl.Compile(rec.ConditionDSL, l.dslReg)
		if err != nil {
			// Log and skip the bad rule — don't fail the entire load.
			l.logger.Error("DSL compile failed, rule skipped",
				zap.String("rule_key", rec.RuleKey),
				zap.Error(err),
			)
			continue
		}
		rules = append(rules, &dslRule{
			id:       rec.RuleKey,
			name:     rec.Name,
			priority: rec.Priority,
			program:  prog,
			action: ruleAction{
				decision: engine.Decision(rec.ActionDecision),
				riskCode: rec.ActionRiskCode,
				score:    rec.ActionScore,
			},
		})
	}
	return rules, nil
}

// ─── dslRule: Rule implementation backed by a dsl.Program ────────────────────

type ruleAction struct {
	decision engine.Decision
	riskCode string
	score    int
}

// dslRule is a stateless Rule whose condition is a compiled dsl.Program.
type dslRule struct {
	id       string
	name     string
	priority int
	program  *dsl.Program
	action   ruleAction
}

func (r *dslRule) ID() string       { return r.id }
func (r *dslRule) Name() string     { return r.name }
func (r *dslRule) Priority() int    { return r.priority }

// Evaluate runs the compiled DSL program and, if the condition is satisfied,
// returns a Result with the configured action.
func (r *dslRule) Evaluate(ctx context.Context, rctx *Context) (*Result, error) {
	rt := dsl.AcquireRuntime()
	defer dsl.ReleaseRuntime(rt)
	rt.Features = rctx.Features
	rt.Request = rctx.Request

	hit, err := r.program.Run(ctx, rt)
	if err != nil {
		// Treat evaluation errors as non-fatal; the rule does not fire.
		return &Result{
			RuleID:   r.id,
			RuleName: r.name,
			Hit:      false,
		}, nil
	}

	res := &Result{
		RuleID:   r.id,
		RuleName: r.name,
		Hit:      hit,
	}
	if hit {
		res.Decision = r.action.decision
		res.RiskCode = r.action.riskCode
		res.Score = r.action.score
	}
	return res, nil
}
