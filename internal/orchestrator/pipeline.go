// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zhucl121/risk-engine/internal/audit"
	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/internal/list"
	"github.com/zhucl121/risk-engine/internal/model"
	"github.com/zhucl121/risk-engine/internal/resilience"
	"github.com/zhucl121/risk-engine/internal/rule"
	"github.com/zhucl121/risk-engine/internal/scene"
	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// Deps bundles all service dependencies the pipeline dispatches to.
type Deps struct {
	Features feature.Service
	Rules    rule.Evaluator
	Models   model.Registry
	List     list.Service
	// Breakers maps step-name → Breaker. Nil entries are treated as no breaker.
	// Keys: "list", "model" (matches StepKindList / StepKindModel lower-case).
	Breakers map[string]*resilience.Breaker

	// ExtraParamLoader loads the parameter specification for a scene's Extra
	// fields from the database.  When nil, only the static PolicySet.ExtraSchema
	// is used.
	ExtraParamLoader *scene.ExtraParamLoader

	// DSLRegistry is used to compile and evaluate Step.Condition expressions.
	// When nil, all Condition checks are skipped (step always runs).
	DSLRegistry *dsl.FunctionRegistry

	// ShadowAuditWriter receives shadow (dry-run) execution records.
	// When nil, shadow results are silently discarded.
	ShadowAuditWriter *audit.ChannelWriter

	// Registry is used by shadow mode to look up shadow policy pipelines.
	// This back-reference is set by atomicRegistry after construction.
	Registry Registry
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
// A/B test routing:
//   - When ABTest.Enabled is true, a random draw determines whether this
//     request is in the experiment group (probability = ABTest.SplitPct).
//   - Experiment-group requests use ABTest.ExperimentPipeline (if set).
//   - The experiment group label "abtest:{experimentID}" is appended to
//     RiskReasons so callers can identify the bucket.
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

	// ── Validate + fill Extra from DB specs ───────────────────────────────────
	// Load per-scene parameter specs from DB (with hot-cache). On success,
	// required fields are validated and optional fields are filled with
	// defaults. The DB schema is then merged with the static YAML schema,
	// with DB entries taking precedence, before type-coercing injection.
	effectiveSchema := p.policy.ExtraSchema
	if p.deps.ExtraParamLoader != nil {
		specs, loadErr := p.deps.ExtraParamLoader.Load(ctx, p.policy.SceneCode)
		if loadErr == nil && len(specs) > 0 {
			if valErr := validateAndFillExtra(req, specs, p.policy.SceneCode); valErr != nil {
				return nil, valErr
			}
			dbSchema := specsToSchema(specs)
			effectiveSchema = mergeSchemas(p.policy.ExtraSchema, dbSchema)
		}
	}

	// ── Inject Extra fields into the feature map ───────────────────────────────
	// After injection, DSL rules can reference "extra.<key>" directly.
	// Type coercion is governed by the merged effectiveSchema.
	injectExtra(req, effectiveSchema, features)

	// ── Traffic-steering priority (highest specificity wins) ──────────────────
	//
	// Priority order:
	//   1. Canary routing (deterministic hash, stable per user) — for gradual rollout
	//   2. A/B test routing (random per request) — for symmetric experiments
	//   3. Main pipeline
	//
	// At most ONE routing decision is made per request.

	steps := p.policy.Pipeline
	routeLabel := ""

	// 1. Canary: hash-stable, per-user deterministic routing.
	if c := p.policy.Canary; c != nil && c.Enabled && c.TrafficPct > 0 {
		if inCanary(c, req.UserID, req.DeviceID, req.SessionID, req.IP, req.Extra) {
			routeLabel = "canary:" + c.CanaryVersion
			if len(c.CanaryPipeline) > 0 {
				steps = c.CanaryPipeline
			}
		}
	}

	// 2. A/B test: random per request (only if canary did not already route).
	if routeLabel == "" {
		if ab := p.policy.ABTest; ab != nil && ab.Enabled && rand.Float64() < ab.SplitPct { //nolint:gosec
			routeLabel = "abtest:" + ab.ExperimentID
			if len(ab.ExperimentPipeline) > 0 {
				steps = ab.ExperimentPipeline
			}
		}
	}

	if routeLabel != "" {
		res.RiskReasons = append(res.RiskReasons, routeLabel)
	}

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

	// ── Shadow (dry-run) policies ─────────────────────────────────────────────
	// Run shadow pipelines asynchronously after the real decision is finalized.
	// They NEVER block or modify the returned result.
	if len(p.policy.ShadowPolicies) > 0 &&
		p.deps.Registry != nil &&
		p.deps.ShadowAuditWriter != nil {
		go p.runShadowPolicies(req, features, res)
	}

	// ── Champion-Challenger ────────────────────────────────────────────────────
	// Run challenger pipelines concurrently in the background.
	// The champion's decision (res) is always what gets returned.
	// Both champion and challenger results are written to cc_audit.
	if cc := p.policy.ChampionChallenger; cc != nil && cc.Enabled &&
		p.deps.ShadowAuditWriter != nil {
		go p.runChallengers(req, features, res, cc)
	}

	return res, nil
}

// runShadowPolicies executes each shadow policy concurrently in the background.
// It uses a detached context (background) so main request cancellation does not
// abort shadow execution; shadow runs are best-effort.
func (p *pipeline) runShadowPolicies(
	req *engine.DecisionRequest,
	features feature.Map,
	prodResult *engine.DecisionResult,
) {
	ctx := context.Background()
	for _, ref := range p.policy.ShadowPolicies {
		ref := ref // capture
		go func() {
			shadowPipeline, err := p.deps.Registry.Get(ref.SceneCode)
			if err != nil {
				return
			}
			shadowStart := time.Now()
			shadowRes, err := shadowPipeline.Execute(ctx, req)
			costMs := time.Since(shadowStart).Milliseconds()
			if err != nil || shadowRes == nil {
				return
			}
			rec := audit.NewShadowRecord(
				req.RequestID,
				p.policy.SceneCode,
				ref.SceneCode,
				ref.Version,
				req.UserID,
				req.DeviceID,
				string(shadowRes.Decision),
				shadowRes.RiskScore,
				string(prodResult.Decision),
				prodResult.RiskScore,
				shadowRes.HitRules,
				shadowRes.ModelScores,
				shadowRes.RiskReasons,
				costMs,
			)
			p.deps.ShadowAuditWriter.WriteShadow(rec)
		}()
	}
}

// runChallengers executes all challenger variants for the champion-challenger
// experiment.  Each challenger whose traffic bucket matches runs concurrently
// in a separate goroutine.  The champion's result is passed in so it can be
// recorded alongside the challenger result in the cc_audit log.
func (p *pipeline) runChallengers(
	req *engine.DecisionRequest,
	features feature.Map,
	champResult *engine.DecisionResult,
	cc *ChampionChallengerConfig,
) {
	ctx := context.Background() // detached – best effort, must not block caller
	for _, variant := range cc.Challengers {
		variant := variant // capture
		if !inCanary(
			&CanaryConfig{
				Enabled:    true,
				TrafficPct: variant.TrafficPct,
				HashKey:    variant.HashKey,
				Salt:       variant.Salt,
			},
			req.UserID, req.DeviceID, req.SessionID, req.IP, req.Extra,
		) {
			continue
		}

		go func() {
			challPipeline := &pipeline{
				policy: PolicySet{
					SceneCode: p.policy.SceneCode,
					Pipeline:  variant.Pipeline,
					Fallback:  p.policy.Fallback,
					Strategy:  p.policy.Strategy,
				},
				deps: p.deps,
			}
			challStart := time.Now()
			challRes, err := challPipeline.runPipelineSteps(ctx, req, features)
			challCostMs := time.Since(challStart).Milliseconds()
			if err != nil || challRes == nil {
				return
			}

			rec := audit.NewChampionChallengerRecord(
				req.RequestID,
				p.policy.SceneCode,
				cc.ExperimentID,
				variant.ChallengerID,
				req.UserID,
				req.DeviceID,
				string(champResult.Decision),
				champResult.RiskScore,
				champResult.CostMs,
				champResult.HitRules,
				champResult.ModelScores,
				string(challRes.Decision),
				challRes.RiskScore,
				challCostMs,
				challRes.HitRules,
				challRes.ModelScores,
				challRes.RiskReasons,
			)
			p.deps.ShadowAuditWriter.WriteCC(rec)
		}()
	}
}

// runPipelineSteps executes only the pipeline steps (no routing, no shadow/cc)
// and returns the partial DecisionResult.  Used internally by challenger execution.
func (p *pipeline) runPipelineSteps(
	ctx context.Context,
	req *engine.DecisionRequest,
	features feature.Map,
) (*engine.DecisionResult, error) {
	res := &engine.DecisionResult{
		RequestID:   req.RequestID,
		Decision:    engine.DecisionPass,
		ModelScores: make(map[string]float64),
	}
	i := 0
	steps := p.policy.Pipeline
	for i < len(steps) {
		step := steps[i]
		if !step.Parallel {
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
		j := i
		for j < len(steps) && steps[j].Parallel {
			j++
		}
		_, done := p.runParallel(ctx, steps[i:j], req, features, res)
		if done {
			break
		}
		i = j
	}
	res.RiskLevel = scoreToLevel(res.RiskScore)
	return res, nil
}

// runStep executes a single step with its configured timeout, optional
// Condition check, and automatic retry.
// Returns a skipped StepResult (not an error) when:
//   - the Condition evaluates to false
//   - dispatch fails and OnFailure == SKIP
func (p *pipeline) runStep(
	ctx context.Context,
	step Step,
	req *engine.DecisionRequest,
	features feature.Map,
) (*StepResult, error) {
	// ── Step Condition ───────────────────────────────────────────────────────
	if step.Condition != "" && p.deps.DSLRegistry != nil {
		prog, compileErr := dsl.Compile(step.Condition, p.deps.DSLRegistry)
		if compileErr == nil {
			rt := dsl.AcquireRuntime()
			rt.Request = req
			rt.Features = features
			condResult, runErr := prog.Run(ctx, rt)
			dsl.ReleaseRuntime(rt)
			if runErr == nil && !condResult {
				return &StepResult{
					Step:     step,
					Skipped:  true,
					Decision: engine.DecisionPass,
				}, nil
			}
		}
		// On compile/run error, continue executing the step (fail-open).
	}

	sCtx := ctx
	var cancel context.CancelFunc
	if step.Timeout > 0 {
		sCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// ── Retry ────────────────────────────────────────────────────────────────
	maxAttempts := step.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	delay := time.Duration(step.Retry.DelayMs) * time.Millisecond

	var sr *StepResult
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && delay > 0 {
			select {
			case <-ctx.Done():
				return p.handleStepError(ctx.Err(), step), nil
			case <-time.After(delay):
			}
		}
		sr, lastErr = p.dispatch(sCtx, step, req, features)
		if lastErr == nil {
			return sr, nil
		}
	}
	return p.handleStepError(lastErr, step), nil
}


// dispatch routes a step to the appropriate service.
// If the step has a ParamMapping, the resolved params are merged on top of
// the feature map before being forwarded to the downstream service.
func (p *pipeline) dispatch(
	ctx context.Context,
	step Step,
	req *engine.DecisionRequest,
	features feature.Map,
) (*StepResult, error) {
	sr := &StepResult{Step: step}

	// Resolve step-level parameter mapping and produce an effective feature map.
	effectiveFeatures := features
	if len(step.ParamMapping) > 0 {
		params := step.ParamMapping.resolve(req, features)
		effectiveFeatures = mergeParams(features, params)
	}

	switch step.Kind {
	case StepKindRule:
		return p.dispatchRule(ctx, step, req, effectiveFeatures, sr)

	case StepKindModel:
		return p.dispatchModel(ctx, step, effectiveFeatures, sr)

	case StepKindList:
		return p.dispatchList(ctx, req, step, effectiveFeatures, sr)

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

	var callErr error
	innerFn := func(ctx context.Context) error {
		for _, name := range step.Models {
			score, err := p.deps.Models.Score(ctx, name, features)
			if err != nil {
				return err
			}
			sr.Models[name] = score
			if score > maxScore {
				maxScore = score
			}
		}
		return nil
	}

	if b := p.deps.Breakers["model"]; b != nil {
		callErr = b.Execute(ctx, innerFn)
	} else {
		callErr = innerFn(ctx)
	}

	if callErr != nil {
		return nil, callErr
	}
	sr.Score = int(maxScore * 1000)
	sr.Decision = engine.DecisionPass
	return sr, nil
}

func (p *pipeline) dispatchList(
	ctx context.Context,
	req *engine.DecisionRequest,
	step Step,
	features feature.Map,
	sr *StepResult,
) (*StepResult, error) {
	// Build query set: use ListQueryFields when configured,
	// otherwise fall back to the default {user, device, ip} triple.
	var queries []*list.Query
	if len(step.ListQueryFields) > 0 {
		for _, src := range step.ListQueryFields {
			v := resolveSource(src, req, features)
			val := v.StrVal
			if v.Kind == feature.KindInt {
				val = fmt.Sprintf("%d", v.IntVal)
			}
			if val == "" {
				continue
			}
			// Use the source expression as the Kind label for traceability.
			queries = append(queries, &list.Query{Kind: src, Value: val})
		}
	} else {
		queries = []*list.Query{
			{Kind: "user", Value: req.UserID},
			{Kind: "device", Value: req.DeviceID},
			{Kind: "ip", Value: req.IP},
		}
	}

	innerFn := func(ctx context.Context) error {
		for _, q := range queries {
			if q.Value == "" {
				continue
			}
			s, err := p.deps.List.Check(ctx, q)
			if err != nil {
				return err
			}
			switch s {
			case list.StatusBlacklist:
				sr.Decision = engine.DecisionReject
				sr.Score = 1000
				sr.Reasons = append(sr.Reasons, fmt.Sprintf("blacklist:%s:%s", q.Kind, q.Value))
				return nil
			case list.StatusGraylist:
				if decisionPriority(engine.DecisionManualReview) > decisionPriority(sr.Decision) {
					sr.Decision = engine.DecisionManualReview
					sr.Score = 700
				}
			}
		}
		return nil
	}

	var callErr error
	if b := p.deps.Breakers["list"]; b != nil {
		callErr = b.Execute(ctx, innerFn)
	} else {
		callErr = innerFn(ctx)
	}
	if callErr != nil {
		return nil, callErr
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

// apply merges a StepResult into the accumulating DecisionResult using the
// pipeline-level aggregation strategy.
// Returns true if the pipeline should terminate early (short-circuit).
func (p *pipeline) apply(res *engine.DecisionResult, sr *StepResult, step Step) bool {
	return p.applyWithStrategy(res, sr, step, p.policy.Strategy)
}

func (p *pipeline) applyWithStrategy(
	res *engine.DecisionResult,
	sr *StepResult,
	step Step,
	strategy AggregationStrategy,
) bool {
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

	switch strategy {
	case AggregationRuleFirst:
		// Rule steps take priority; model scores are only used when no rule hit.
		if step.Kind == StepKindRule && sr.Decision != engine.DecisionPass {
			res.RiskScore = sr.Score
			res.Decision = sr.Decision
			return sr.Decision == engine.DecisionReject
		}
		// Model step: only update if no rule has produced a decision yet.
		if step.Kind == StepKindModel && res.Decision == engine.DecisionPass {
			if sr.Score > res.RiskScore {
				res.RiskScore = sr.Score
			}
		}

	case AggregationWeighted:
		// Weighted: accumulate a weighted sum of scores; decision is derived from the sum.
		weight := step.Weight
		if weight <= 0 {
			weight = 1.0
		}
		// Store raw weighted score; final normalisation happens in Execute after all steps.
		weightedScore := int(float64(sr.Score) * weight)
		res.RiskScore += weightedScore
		// Clamp to 1000.
		if res.RiskScore > 1000 {
			res.RiskScore = 1000
		}
		// Decision: derive from accumulated score.
		var derived engine.Decision
		switch {
		case res.RiskScore >= 800:
			derived = engine.DecisionReject
		case res.RiskScore >= 500:
			derived = engine.DecisionManualReview
		default:
			derived = engine.DecisionPass
		}
		if decisionPriority(derived) > decisionPriority(res.Decision) {
			res.Decision = derived
		}

	default: // HIGHEST_RISK (default)
		if sr.Score > res.RiskScore {
			res.RiskScore = sr.Score
		}
		if decisionPriority(sr.Decision) > decisionPriority(res.Decision) {
			res.Decision = sr.Decision
		}
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
