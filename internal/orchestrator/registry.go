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

type shadowPolicyYAML struct {
	SceneCode string `yaml:"sceneCode"`
	Version   string `yaml:"version"`
}

// policySetYAML mirrors PolicySet for YAML unmarshalling.
type policySetYAML struct {
	SceneCode string `yaml:"sceneCode"`
	Version   string `yaml:"version"`
	Fallback  string `yaml:"fallback"`
	Strategy  string `yaml:"strategy"` // HIGHEST_RISK | WEIGHTED | RULE_FIRST
	Pipeline  []stepYAML `yaml:"pipeline"`
	ABTest    *abTestYAML `yaml:"abTest"`
	// Canary enables deterministic hash-based traffic splitting.
	Canary *canaryYAML `yaml:"canary"`
	// ChampionChallenger enables simultaneous execution of challenger pipelines.
	ChampionChallenger *championChallengerYAML `yaml:"championChallenger"`
	// ShadowPolicies lists policies to run in shadow/dry-run mode.
	ShadowPolicies []shadowPolicyYAML `yaml:"shadowPolicies"`
	// ExtraSchema declares the type for each Extra key.
	// Supported types: string (default), int, float, bool.
	ExtraSchema map[string]string `yaml:"extraSchema"`
}

type abTestYAML struct {
	Enabled            bool       `yaml:"enabled"`
	ExperimentID       string     `yaml:"experimentId"`
	SplitPct           float64    `yaml:"splitPct"`
	ExperimentPipeline []stepYAML `yaml:"experimentPipeline"`
}

type canaryYAML struct {
	Enabled        bool       `yaml:"enabled"`
	CanaryVersion  string     `yaml:"canaryVersion"`
	TrafficPct     int        `yaml:"trafficPct"`
	HashKey        string     `yaml:"hashKey"`
	Salt           string     `yaml:"salt"`
	CanaryPipeline []stepYAML `yaml:"canaryPipeline"`
}

type challengerVariantYAML struct {
	ChallengerID string     `yaml:"challengerID"`
	TrafficPct   int        `yaml:"trafficPct"`
	HashKey      string     `yaml:"hashKey"`
	Salt         string     `yaml:"salt"`
	Pipeline     []stepYAML `yaml:"pipeline"`
}

type championChallengerYAML struct {
	Enabled      bool                    `yaml:"enabled"`
	ExperimentID string                  `yaml:"experimentID"`
	Challengers  []challengerVariantYAML `yaml:"challengers"`
}

type stepRetryYAML struct {
	MaxAttempts int `yaml:"maxAttempts"`
	DelayMs     int `yaml:"delayMs"`
}

type stepYAML struct {
	Name      string  `yaml:"name"`
	Kind      string  `yaml:"kind"`
	RuleGroup string  `yaml:"ruleGroup"`
	Models    []string `yaml:"models"`
	TimeoutMs int     `yaml:"timeoutMs"`
	Parallel  bool    `yaml:"parallel"`
	OnFailure string  `yaml:"onFailure"`
	Strategy  string  `yaml:"strategy"` // per-step aggregation (for AGGREGATE kind)
	Weight    float64 `yaml:"weight"`   // used by WEIGHTED pipeline strategy
	// Condition is a DSL expression; step is skipped when it evaluates to false.
	Condition string `yaml:"condition"`
	// Retry configures automatic retry.
	Retry stepRetryYAML `yaml:"retry"`
	// ParamMapping maps downstream parameter names to source expressions.
	// e.g. {"merchant_id": "extra.merchant_id", "channel": "WEB"}
	ParamMapping map[string]string `yaml:"params"`
	// ListQueryFields overrides the default user/device/ip list queries.
	// e.g. ["extra.merchant_id", "request.ip"]
	ListQueryFields []string `yaml:"listQueryFields"`
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
	// Inject back-reference so pipelines can look up shadow policies.
	deps := r.deps
	deps.Registry = r

	m := make(map[string]Pipeline, len(policies))
	for _, ps := range policies {
		m[ps.SceneCode] = newPipeline(ps, deps)
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

// LoadFromReader reads YAML-encoded policy definitions from rd and reloads.
//
// Supports two formats:
//  1. A YAML sequence of PolicySet objects (single document, list root).
//  2. Multiple YAML documents separated by "---", each being a single PolicySet
//     or a sequence of PolicySets.
func (r *atomicRegistry) LoadFromReader(ctx context.Context, rd io.Reader) error {
	dec := yaml.NewDecoder(rd)
	var all []PolicySet

	for {
		// Try decoding as a sequence first.
		var raw []policySetYAML
		err := dec.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Maybe it's a single object, not a list.
			var single policySetYAML
			if err2 := dec.Decode(&single); err2 != nil {
				return fmt.Errorf("orchestrator: decode yaml: %w", err)
			}
			raw = []policySetYAML{single}
		}
		for _, y := range raw {
			if y.SceneCode == "" {
				continue // skip empty documents
			}
			ps, err := convertPolicy(y)
			if err != nil {
				return err
			}
			all = append(all, ps)
		}
	}
	return r.Reload(ctx, all)
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

	// Parse ExtraSchema.
	schema := make(ExtraSchema, len(y.ExtraSchema))
	for k, v := range y.ExtraSchema {
		schema[k] = ExtraFieldType(v)
	}

	var pipelineStrategy AggregationStrategy
	switch y.Strategy {
	case "WEIGHTED":
		pipelineStrategy = AggregationWeighted
	case "RULE_FIRST":
		pipelineStrategy = AggregationRuleFirst
	default:
		pipelineStrategy = AggregationHighestRisk
	}

	// Parse ShadowPolicies.
	shadowPolicies := make([]ShadowPolicyRef, 0, len(y.ShadowPolicies))
	for _, sp := range y.ShadowPolicies {
		if sp.SceneCode != "" {
			shadowPolicies = append(shadowPolicies, ShadowPolicyRef{
				SceneCode: sp.SceneCode,
				Version:   sp.Version,
			})
		}
	}

	// Parse Canary config.
	var canary *CanaryConfig
	if y.Canary != nil && y.Canary.Enabled {
		canarySteps := make([]Step, 0, len(y.Canary.CanaryPipeline))
		for _, sy := range y.Canary.CanaryPipeline {
			s, err := convertStep(sy)
			if err != nil {
				return PolicySet{}, fmt.Errorf("orchestrator: scene %s canary: %w", y.SceneCode, err)
			}
			canarySteps = append(canarySteps, s)
		}
		canary = &CanaryConfig{
			Enabled:        y.Canary.Enabled,
			CanaryVersion:  y.Canary.CanaryVersion,
			TrafficPct:     y.Canary.TrafficPct,
			HashKey:        y.Canary.HashKey,
			Salt:           y.Canary.Salt,
			CanaryPipeline: canarySteps,
		}
	}

	// Parse ChampionChallenger config.
	var cc *ChampionChallengerConfig
	if y.ChampionChallenger != nil && y.ChampionChallenger.Enabled {
		variants := make([]ChallengerVariant, 0, len(y.ChampionChallenger.Challengers))
		for _, vy := range y.ChampionChallenger.Challengers {
			challSteps := make([]Step, 0, len(vy.Pipeline))
			for _, sy := range vy.Pipeline {
				s, err := convertStep(sy)
				if err != nil {
					return PolicySet{}, fmt.Errorf("orchestrator: scene %s challenger %s: %w",
						y.SceneCode, vy.ChallengerID, err)
				}
				challSteps = append(challSteps, s)
			}
			variants = append(variants, ChallengerVariant{
				ChallengerID: vy.ChallengerID,
				TrafficPct:   vy.TrafficPct,
				HashKey:      vy.HashKey,
				Salt:         vy.Salt,
				Pipeline:     challSteps,
			})
		}
		cc = &ChampionChallengerConfig{
			Enabled:      y.ChampionChallenger.Enabled,
			ExperimentID: y.ChampionChallenger.ExperimentID,
			Challengers:  variants,
		}
	}

	return PolicySet{
		SceneCode:          y.SceneCode,
		Version:            y.Version,
		Pipeline:           steps,
		Fallback:           fallback,
		Strategy:           pipelineStrategy,
		ABTest:             abTest,
		Canary:             canary,
		ChampionChallenger: cc,
		ShadowPolicies:     shadowPolicies,
		ExtraSchema:        schema,
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
		return Step{}, fmt.Errorf("orchestrator: unknown step kind %q", sy.Kind)
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

	// Parse ParamMapping.
	pm := make(ParamMapping, len(sy.ParamMapping))
	for k, v := range sy.ParamMapping {
		pm[k] = v
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
		Weight:    sy.Weight,
		Condition: sy.Condition,
		Retry: RetryConfig{
			MaxAttempts: sy.Retry.MaxAttempts,
			DelayMs:     sy.Retry.DelayMs,
		},
		ParamMapping:    pm,
		ListQueryFields: sy.ListQueryFields,
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
