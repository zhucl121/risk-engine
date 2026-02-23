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

// Step is one node in the execution DAG.
type Step struct {
	Name      string
	Kind      StepKind
	// RuleGroup is the rule group name for StepKindRule steps.
	RuleGroup string
	// Models lists model names for StepKindModel steps.
	Models    []string
	Timeout   time.Duration
	// Parallel marks this step to run concurrently with its siblings.
	Parallel  bool
	// OnFailure defines behaviour when this step errors or times out.
	OnFailure FailurePolicy
	Strategy  AggregationStrategy // used only for StepKindAggregate
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
}

// ABTestConfig routes a fraction of traffic to an experimental pipeline.
type ABTestConfig struct {
	Enabled      bool
	ExperimentID string
	// SplitPct is the fraction of traffic routed to the experiment (0.0–1.0).
	SplitPct float64
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
