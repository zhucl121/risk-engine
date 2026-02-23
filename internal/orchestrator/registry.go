// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/yourorg/riskengine/internal/engine"
)

// policySetYAML mirrors PolicySet for YAML unmarshalling.
type policySetYAML struct {
	SceneCode string       `yaml:"sceneCode"`
	Version   string       `yaml:"version"`
	Fallback  string       `yaml:"fallback"`
	Pipeline  []stepYAML   `yaml:"pipeline"`
	ABTest    *abTestYAML  `yaml:"abTest"`
}

type abTestYAML struct {
	Enabled            bool       `yaml:"enabled"`
	ExperimentID       string     `yaml:"experimentId"`
	SplitPct           float64    `yaml:"splitPct"`
	ExperimentPipeline []stepYAML `yaml:"experimentPipeline"`
}

type stepYAML struct {
	Name      string  `yaml:"name"`
	Kind      string  `yaml:"kind"`
	RuleGroup string  `yaml:"ruleGroup"`
	Models    []string `yaml:"models"`
	TimeoutMs int     `yaml:"timeoutMs"`
	Parallel  bool    `yaml:"parallel"`
	OnFailure string  `yaml:"onFailure"`
	Strategy  string  `yaml:"strategy"`
}

// atomicRegistry is the concrete Registry implementation.
// It uses an atomic pointer to an immutable map for lock-free reads.
type atomicRegistry struct {
	// pipelines is a pointer to an immutable map[sceneCode]Pipeline.
	pipelines atomic.Pointer[map[string]Pipeline]
	deps      Deps
}

// NewRegistry returns an empty Registry. Call Reload to populate it.
func NewRegistry(deps Deps) Registry {
	r := &atomicRegistry{deps: deps}
	empty := make(map[string]Pipeline)
	r.pipelines.Store(&empty)
	return r
}

// Get returns the Pipeline for the given sceneCode.
func (r *atomicRegistry) Get(sceneCode string) (Pipeline, error) {
	m := *r.pipelines.Load()
	p, ok := m[sceneCode]
	if !ok {
		return nil, fmt.Errorf("orchestrator: no policy found for scene %q", sceneCode)
	}
	return p, nil
}

// Reload atomically replaces all registered pipelines.
func (r *atomicRegistry) Reload(_ context.Context, policies []PolicySet) error {
	m := make(map[string]Pipeline, len(policies))
	for _, ps := range policies {
		m[ps.SceneCode] = newPipeline(ps, r.deps)
	}
	r.pipelines.Store(&m)
	return nil
}

// LoadFromYAML reads policy definitions from a YAML file and calls Reload.
func (r *atomicRegistry) LoadFromYAML(ctx context.Context, path string) error {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return fmt.Errorf("orchestrator: open %s: %w", path, err)
	}
	defer f.Close() //nolint:errcheck
	return r.LoadFromReader(ctx, f)
}

// LoadFromReader reads YAML-encoded policy definitions from r and reloads.
func (r *atomicRegistry) LoadFromReader(ctx context.Context, rd io.Reader) error {
	var raw []policySetYAML
	if err := yaml.NewDecoder(rd).Decode(&raw); err != nil {
		return fmt.Errorf("orchestrator: decode yaml: %w", err)
	}

	policies := make([]PolicySet, 0, len(raw))
	for _, y := range raw {
		ps, err := convertPolicy(y)
		if err != nil {
			return err
		}
		policies = append(policies, ps)
	}
	return r.Reload(ctx, policies)
}

func convertPolicy(y policySetYAML) (PolicySet, error) {
	steps := make([]Step, 0, len(y.Pipeline))
	for _, sy := range y.Pipeline {
		s, err := convertStep(sy)
		if err != nil {
			return PolicySet{}, fmt.Errorf("orchestrator: scene %s: %w", y.SceneCode, err)
		}
		steps = append(steps, s)
	}
	fallback := parseFallback(y.Fallback)

	var abTest *ABTestConfig
	if y.ABTest != nil && y.ABTest.Enabled {
		expSteps := make([]Step, 0, len(y.ABTest.ExperimentPipeline))
		for _, sy := range y.ABTest.ExperimentPipeline {
			s, err := convertStep(sy)
			if err != nil {
				return PolicySet{}, fmt.Errorf("orchestrator: scene %s abtest: %w", y.SceneCode, err)
			}
			expSteps = append(expSteps, s)
		}
		abTest = &ABTestConfig{
			Enabled:            y.ABTest.Enabled,
			ExperimentID:       y.ABTest.ExperimentID,
			SplitPct:           y.ABTest.SplitPct,
			ExperimentPipeline: expSteps,
		}
	}

	return PolicySet{
		SceneCode: y.SceneCode,
		Version:   y.Version,
		Pipeline:  steps,
		Fallback:  fallback,
		ABTest:    abTest,
	}, nil
}

func convertStep(sy stepYAML) (Step, error) {
	var kind StepKind
	switch sy.Kind {
	case "LIST":
		kind = StepKindList
	case "RULE":
		kind = StepKindRule
	case "MODEL":
		kind = StepKindModel
	case "AGGREGATE":
		kind = StepKindAggregate
	case "CUSTOM":
		kind = StepKindCustom
	default:
		return Step{}, fmt.Errorf("unknown step kind %q", sy.Kind)
	}

	var fp FailurePolicy
	switch sy.OnFailure {
	case "REJECT":
		fp = FailurePolicyReject
	case "FALLBACK":
		fp = FailurePolicyFallback
	default:
		fp = FailurePolicySkip
	}

	return Step{
		Name:      sy.Name,
		Kind:      kind,
		RuleGroup: sy.RuleGroup,
		Models:    sy.Models,
		Timeout:   millisToDuration(sy.TimeoutMs),
		Parallel:  sy.Parallel,
		OnFailure: fp,
		Strategy:  AggregationStrategy(sy.Strategy),
	}, nil
}

func millisToDuration(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func parseFallback(s string) engine.Decision {
	switch s {
	case "REJECT":
		return engine.DecisionReject
	case "MANUAL_REVIEW":
		return engine.DecisionManualReview
	default:
		return engine.DecisionPass
	}
}
