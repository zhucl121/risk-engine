// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package orchestrator executes a DAG of decision steps and aggregates
// their results into a single DecisionResult. Step definitions are loaded
// from YAML PolicySet configurations and can be hot-reloaded.
package orchestrator

import (
	"context"
	"time"

	"github.com/yourorg/riskengine/internal/engine"
)

// StepKind identifies the type of work performed in a pipeline step.
type StepKind string

const (
	StepKindList      StepKind = "LIST"
	StepKindRule      StepKind = "RULE"
	StepKindModel     StepKind = "MODEL"
	StepKindAggregate StepKind = "AGGREGATE"
	StepKindCustom    StepKind = "CUSTOM"
)

// FailurePolicy controls what happens when a step errors or times out.
type FailurePolicy string

const (
	FailurePolicySkip     FailurePolicy = "SKIP"     // log and continue with partial result
	FailurePolicyReject   FailurePolicy = "REJECT"   // treat as high-risk, return REJECT
	FailurePolicyFallback FailurePolicy = "FALLBACK" // use scene-level fallback decision
)

// AggregationStrategy controls how sub-results are merged.
type AggregationStrategy string

const (
	AggregationHighestRisk AggregationStrategy = "HIGHEST_RISK"
	AggregationWeighted    AggregationStrategy = "WEIGHTED"
	AggregationRuleFirst   AggregationStrategy = "RULE_FIRST"
)

// RetryConfig controls automatic retry behaviour for a step.
type RetryConfig struct {
	// MaxAttempts is the maximum number of total attempts (1 = no retry).
	MaxAttempts int
	// DelayMs is the fixed delay between attempts in milliseconds.
	DelayMs int
}

// Step is one node in the execution DAG.
type Step struct {
	Name string
	Kind StepKind
	// RuleGroup is the rule group name for StepKindRule steps.
	RuleGroup string
	// Models lists model names for StepKindModel steps.
	Models  []string
	Timeout time.Duration
	// Parallel marks this step to run concurrently with its siblings.
	Parallel bool
	// OnFailure defines behaviour when this step errors or times out.
	OnFailure FailurePolicy
	Strategy  AggregationStrategy // used only for StepKindAggregate

	// Weight is used by the WEIGHTED aggregation strategy.
	// Steps with higher Weight contribute proportionally more to the final score.
	// Defaults to 1.0 when zero.
	Weight float64

	// Condition is an optional DSL expression evaluated before the step runs.
	// If the expression evaluates to false the step is skipped (treated as SKIP).
	// Example: "extra.amount > 10000"  — only run this step for large amounts.
	Condition string

	// Retry configures automatic retry on transient errors.
	// Zero value means no retry (single attempt).
	Retry RetryConfig

	// ParamMapping controls how downstream input parameters are populated
	// for this step.  Keys are the downstream parameter names; values are
	// source expressions (see mapping.go for the resolution rules).
	//
	// Example: {"merchant_id": "extra.merchant_id", "channel": "WEB"}
	ParamMapping ParamMapping

	// ListQueryFields overrides the default list-check fields for StepKindList.
	// Each entry is a source expression (same syntax as ParamMapping values).
	// If empty, the default {user, device, ip} triple is used.
	//
	// Example: ["extra.merchant_id", "request.ip"]
	ListQueryFields []string
}

// ShadowPolicyRef references another PolicySet to run in shadow (dry-run) mode.
// The shadow pipeline executes concurrently with the real pipeline but its
// decision is NEVER returned to the caller; results are written to the audit
// log for offline analysis.
type ShadowPolicyRef struct {
	// SceneCode is the scene code of the shadow policy to execute.
	SceneCode string
	// Version tags the shadow policy for traceability in audit records.
	Version string
}

// PolicySet is the complete configuration for one business scene.
type PolicySet struct {
	SceneCode string
	Version   string
	Pipeline  []Step
	// Fallback is the decision returned when the pipeline cannot complete.
	Fallback engine.Decision
	// ABTest holds optional A/B experiment configuration.
	ABTest *ABTestConfig

	// Strategy is the pipeline-level aggregation strategy.
	// Defaults to HIGHEST_RISK when empty.
	Strategy AggregationStrategy

	// ShadowPolicies lists policies to execute in parallel in shadow/dry-run mode.
	// Their results do NOT affect the real decision; they are recorded in the audit
	// log under the "shadow_audit" key for offline comparison and analysis.
	ShadowPolicies []ShadowPolicyRef

	// ExtraSchema declares the intended type for each key in
	// DecisionRequest.Extra.  Declared keys are type-coerced before being
	// injected into the feature.Map as "extra.<key>".
	// Undeclared keys are always injected as string.
	ExtraSchema ExtraSchema
}

// ABTestConfig routes a fraction of traffic to an experimental pipeline.
type ABTestConfig struct {
	Enabled      bool
	ExperimentID string
	// SplitPct is the fraction of traffic routed to the experiment (0.0–1.0).
	SplitPct float64
	// ExperimentPipeline is the alternative step sequence for the experiment group.
	// If empty, the control Pipeline is reused (useful for testing config changes).
	ExperimentPipeline []Step
}

// StepResult captures the outcome of one pipeline step.
type StepResult struct {
	Step     Step
	Decision engine.Decision
	Score    int
	HitRules []string
	Models   map[string]float64
	Reasons  []string
	CostMs   int64
	Skipped  bool
	Err      error
}

// Pipeline executes a PolicySet's DAG and returns a merged DecisionResult.
// All implementations must be safe for concurrent use.
type Pipeline interface {
	Execute(ctx context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error)
}

// Registry maps scene codes to their compiled Pipeline instances.
// It supports atomic hot-reload when PolicySet YAML files change.
type Registry interface {
	// Get returns the Pipeline for sceneCode, or an error if not found.
	Get(sceneCode string) (Pipeline, error)

	// Reload replaces all registered pipelines atomically.
	Reload(ctx context.Context, policies []PolicySet) error
}
