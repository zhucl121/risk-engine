// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/list"
	"github.com/yourorg/riskengine/internal/model"
	"github.com/yourorg/riskengine/internal/rule"
)

// Deps bundles all service dependencies the pipeline dispatches to.
type Deps struct {
	Features feature.Service
	Rules    rule.Evaluator
	Models   model.Registry
	List     list.Service
}

// pipeline is the concrete Pipeline implementation.
type pipeline struct {
	policy PolicySet
	deps   Deps
}

// newPipeline constructs a pipeline for the given policy and dependencies.
func newPipeline(policy PolicySet, deps Deps) Pipeline {
	return &pipeline{policy: policy, deps: deps}
}

// Execute runs the PolicySet's DAG and returns a merged DecisionResult.
//
// Execution model:
//   - Consecutive non-parallel steps run sequentially.
//   - A contiguous run of steps marked Parallel=true is executed with errgroup.
//   - Each step has its own per-step timeout derived from Step.Timeout.
func (p *pipeline) Execute(ctx context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error) {
	start := time.Now()
	res := &engine.DecisionResult{
		RequestID:   req.RequestID,
		Decision:    engine.DecisionPass,
		ModelScores: make(map[string]float64),
	}

	// Fetch features once for all steps.
	features, err := p.deps.Features.Fetch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: feature fetch: %w", err)
	}

	steps := p.policy.Pipeline
	i := 0
	for i < len(steps) {
		step := steps[i]

		if !step.Parallel {
			// Sequential step.
			sr, err := p.runStep(ctx, step, req, features)
			if err != nil || sr == nil {
				i++
				continue
			}
			if done := p.apply(res, sr, step); done {
				break
			}
			i++
			continue
		}

		// Collect the contiguous parallel group.
		j := i
		for j < len(steps) && steps[j].Parallel {
			j++
		}
		parallelSteps := steps[i:j]
		results, done := p.runParallel(ctx, parallelSteps, req, features, res)
		_ = results
		if done {
			break
		}
		i = j
	}

	res.RiskLevel = scoreToLevel(res.RiskScore)
	res.CostMs = time.Since(start).Milliseconds()
	return res, nil
}

// runStep executes a single step with its configured timeout.
// Returns nil on SKIP (error swallowed) or when the step produces no result.
func (p *pipeline) runStep(
	ctx context.Context,
	step Step,
	req *engine.DecisionRequest,
	features feature.Map,
) (*StepResult, error) {
	sCtx := ctx
	var cancel context.CancelFunc
	if step.Timeout > 0 {
		sCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	sr, err := p.dispatch(sCtx, step, req, features)
	if err != nil {
		return p.handleStepError(err, step), nil
	}
	return sr, nil
}

// dispatch routes a step to the appropriate service.
func (p *pipeline) dispatch(
	ctx context.Context,
	step Step,
	req *engine.DecisionRequest,
	features feature.Map,
) (*StepResult, error) {
	sr := &StepResult{Step: step}

	switch step.Kind {
	case StepKindRule:
		return p.dispatchRule(ctx, step, req, features, sr)

	case StepKindModel:
		return p.dispatchModel(ctx, step, features, sr)

	case StepKindList:
		return p.dispatchList(ctx, req, sr)

	case StepKindAggregate:
		// Aggregation is done by apply(); no additional dispatch needed.
		return sr, nil

	default:
		return nil, fmt.Errorf("orchestrator: unknown step kind %q", step.Kind)
	}
}

func (p *pipeline) dispatchRule(
	ctx context.Context,
	step Step,
	req *engine.DecisionRequest,
	features feature.Map,
	sr *StepResult,
) (*StepResult, error) {
	rctx := &rule.Context{Request: req, Features: features}
	results, err := p.deps.Rules.Evaluate(ctx, rctx)
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		if !r.Hit {
			continue
		}
		sr.HitRules = append(sr.HitRules, r.RuleID)
		sr.Reasons = append(sr.Reasons, r.RiskCode)
		if r.Score > sr.Score {
			sr.Score = r.Score
		}
		if decisionPriority(r.Decision) > decisionPriority(sr.Decision) {
			sr.Decision = r.Decision
		}
	}
	if sr.Decision == "" {
		sr.Decision = engine.DecisionPass
	}
	return sr, nil
}

func (p *pipeline) dispatchModel(
	ctx context.Context,
	step Step,
	features feature.Map,
	sr *StepResult,
) (*StepResult, error) {
	sr.Models = make(map[string]float64, len(step.Models))
	maxScore := 0.0
	for _, name := range step.Models {
		score, err := p.deps.Models.Score(ctx, name, features)
		if err != nil {
			// Model inference failure is treated as a step error.
			return nil, err
		}
		sr.Models[name] = score
		if score > maxScore {
			maxScore = score
		}
	}
	sr.Score = int(maxScore * 1000)
	sr.Decision = engine.DecisionPass
	return sr, nil
}

func (p *pipeline) dispatchList(
	ctx context.Context,
	req *engine.DecisionRequest,
	sr *StepResult,
) (*StepResult, error) {
	queries := []*list.Query{
		{Kind: "user", Value: req.UserID},
		{Kind: "device", Value: req.DeviceID},
		{Kind: "ip", Value: req.IP},
	}
	for _, q := range queries {
		if q.Value == "" {
			continue
		}
		status, err := p.deps.List.Check(ctx, q)
		if err != nil {
			return nil, err
		}
		switch status {
		case list.StatusBlacklist:
			sr.Decision = engine.DecisionReject
			sr.Score = 1000
			sr.Reasons = append(sr.Reasons, fmt.Sprintf("blacklist:%s:%s", q.Kind, q.Value))
			return sr, nil
		case list.StatusGraylist:
			if decisionPriority(engine.DecisionManualReview) > decisionPriority(sr.Decision) {
				sr.Decision = engine.DecisionManualReview
				sr.Score = 700
			}
		}
	}
	if sr.Decision == "" {
		sr.Decision = engine.DecisionPass
	}
	return sr, nil
}

// handleStepError applies the FailurePolicy and returns a synthesised StepResult.
func (p *pipeline) handleStepError(err error, step Step) *StepResult {
	sr := &StepResult{Step: step, Err: err, Skipped: true}
	switch step.OnFailure {
	case FailurePolicyReject:
		sr.Decision = engine.DecisionReject
		sr.Score = 1000
		sr.Skipped = false
	case FailurePolicyFallback:
		sr.Decision = p.policy.Fallback
		sr.Skipped = false
	default: // SKIP
		sr.Decision = engine.DecisionPass
	}
	return sr
}

// apply merges a StepResult into the accumulating DecisionResult.
// Returns true if the pipeline should terminate early.
func (p *pipeline) apply(res *engine.DecisionResult, sr *StepResult, step Step) bool {
	trace := engine.StepTrace{
		Name:    step.Name,
		CostMs:  sr.CostMs,
		Skipped: sr.Skipped,
	}
	if sr.Err != nil {
		trace.Error = sr.Err.Error()
	}
	res.Path = append(res.Path, trace)

	if sr.Skipped {
		return false
	}

	for _, h := range sr.HitRules {
		res.HitRules = append(res.HitRules, h)
	}
	for _, r := range sr.Reasons {
		res.RiskReasons = append(res.RiskReasons, r)
	}
	for k, v := range sr.Models {
		res.ModelScores[k] = v
	}
	if sr.Score > res.RiskScore {
		res.RiskScore = sr.Score
	}
	if decisionPriority(sr.Decision) > decisionPriority(res.Decision) {
		res.Decision = sr.Decision
	}

	// Short-circuit on REJECT from a non-SKIP policy.
	return sr.Decision == engine.DecisionReject && step.OnFailure != FailurePolicySkip
}

// runParallel executes a group of steps concurrently using errgroup.
func (p *pipeline) runParallel(
	ctx context.Context,
	steps []Step,
	req *engine.DecisionRequest,
	features feature.Map,
	res *engine.DecisionResult,
) ([]*StepResult, bool) {
	type indexed struct {
		idx int
		sr  *StepResult
	}
	ch := make(chan indexed, len(steps))

	g, gCtx := errgroup.WithContext(ctx)
	for i, step := range steps {
		i, step := i, step
		g.Go(func() error {
			sr, _ := p.runStep(gCtx, step, req, features)
			if sr == nil {
				sr = &StepResult{Step: step, Skipped: true, Decision: engine.DecisionPass}
			}
			ch <- indexed{i, sr}
			return nil
		})
	}
	_ = g.Wait()
	close(ch)

	ordered := make([]*StepResult, len(steps))
	for item := range ch {
		ordered[item.idx] = item.sr
	}

	done := false
	for _, sr := range ordered {
		if sr == nil {
			continue
		}
		if p.apply(res, sr, sr.Step) {
			done = true
			break
		}
	}
	return ordered, done
}

// decisionPriority maps a Decision to a numeric ordering for aggregation.
func decisionPriority(d engine.Decision) int {
	switch d {
	case engine.DecisionReject:
		return 3
	case engine.DecisionManualReview:
		return 2
	case engine.DecisionPass:
		return 1
	}
	return 0
}

// scoreToLevel maps a risk score (0–1000) to a RiskLevel.
func scoreToLevel(score int) engine.RiskLevel {
	switch {
	case score >= 800:
		return engine.RiskLevelCritical
	case score >= 600:
		return engine.RiskLevelHigh
	case score >= 400:
		return engine.RiskLevelMedium
	default:
		return engine.RiskLevelLow
	}
}
