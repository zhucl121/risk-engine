// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/list"
	"github.com/yourorg/riskengine/internal/model"
	"github.com/yourorg/riskengine/internal/orchestrator"
	"github.com/yourorg/riskengine/internal/rule"
)

// --- stubs ------------------------------------------------------------------

type stubFeature struct{ m feature.Map }

func (s *stubFeature) Fetch(_ context.Context, _ *engine.DecisionRequest) (feature.Map, error) {
	return s.m, nil
}
func (s *stubFeature) Register(_ feature.Fetcher) {}

type stubRuleEvaluator struct{ results []*rule.Result }

func (s *stubRuleEvaluator) Evaluate(_ context.Context, _ *rule.Context) ([]*rule.Result, error) {
	return s.results, nil
}
func (s *stubRuleEvaluator) Reload(_ []rule.Rule) error { return nil }

type stubModelRegistry struct{ score float64 }

func (s *stubModelRegistry) Score(_ context.Context, _ string, _ feature.Map) (float64, error) {
	return s.score, nil
}
func (s *stubModelRegistry) Register(_ string, _ model.Scorer)                    {}
func (s *stubModelRegistry) SetChallenger(_ string, _ model.Scorer, _ float64)    {}
func (s *stubModelRegistry) ClearChallenger(_ string)                             {}

type stubListService struct{ statuses map[string]list.Status }

func (s *stubListService) Check(_ context.Context, q *list.Query) (list.Status, error) {
	return s.statuses[q.Kind+":"+q.Value], nil
}
func (s *stubListService) Add(_ context.Context, _ *list.Query, _ list.Status, _ time.Duration) error {
	return nil
}
func (s *stubListService) Remove(_ context.Context, _ *list.Query) error { return nil }

// --- helpers ----------------------------------------------------------------

func newDeps(
	fm feature.Map,
	ruleResults []*rule.Result,
	modelScore float64,
	listStatuses map[string]list.Status,
) orchestrator.Deps {
	if listStatuses == nil {
		listStatuses = make(map[string]list.Status)
	}
	return orchestrator.Deps{
		Features: &stubFeature{m: fm},
		Rules:    &stubRuleEvaluator{results: ruleResults},
		Models:   &stubModelRegistry{score: modelScore},
		List:     &stubListService{statuses: listStatuses},
		Breakers: nil, // no breakers in tests
	}
}

func newRegistry(deps orchestrator.Deps, policies []orchestrator.PolicySet) orchestrator.Registry {
	reg := orchestrator.NewRegistry(deps)
	_ = reg.Reload(context.Background(), policies)
	return reg
}

// --- tests ------------------------------------------------------------------

func TestPipeline_PassAll(t *testing.T) {
	deps := newDeps(feature.Map{}, nil, 0.1, nil)
	policy := orchestrator.PolicySet{
		SceneCode: "test",
		Pipeline: []orchestrator.Step{
			{Name: "rules", Kind: orchestrator.StepKindRule, OnFailure: orchestrator.FailurePolicySkip},
		},
		Fallback: engine.DecisionPass,
	}
	reg := newRegistry(deps, []orchestrator.PolicySet{policy})
	pipe, err := reg.Get("test")
	require.NoError(t, err)

	res, err := pipe.Execute(context.Background(), &engine.DecisionRequest{RequestID: "req1"})
	require.NoError(t, err)
	assert.Equal(t, engine.DecisionPass, res.Decision)
}

func TestPipeline_RuleReject(t *testing.T) {
	ruleResults := []*rule.Result{
		{RuleID: "FRAUD_001", Hit: true, Decision: engine.DecisionReject, Score: 900, RiskCode: "FRAUD"},
	}
	deps := newDeps(feature.Map{}, ruleResults, 0, nil)
	policy := orchestrator.PolicySet{
		SceneCode: "payment",
		Pipeline: []orchestrator.Step{
			{Name: "rules", Kind: orchestrator.StepKindRule, OnFailure: orchestrator.FailurePolicySkip},
		},
		Fallback: engine.DecisionPass,
	}
	reg := newRegistry(deps, []orchestrator.PolicySet{policy})
	pipe, err := reg.Get("payment")
	require.NoError(t, err)

	res, err := pipe.Execute(context.Background(), &engine.DecisionRequest{RequestID: "req2"})
	require.NoError(t, err)
	assert.Equal(t, engine.DecisionReject, res.Decision)
	assert.Equal(t, 900, res.RiskScore)
	assert.Contains(t, res.HitRules, "FRAUD_001")
}

func TestPipeline_BlacklistReject(t *testing.T) {
	statuses := map[string]list.Status{"user:bad-user": list.StatusBlacklist}
	deps := newDeps(feature.Map{}, nil, 0, statuses)
	policy := orchestrator.PolicySet{
		SceneCode: "scene",
		Pipeline: []orchestrator.Step{
			{Name: "list", Kind: orchestrator.StepKindList, OnFailure: orchestrator.FailurePolicySkip},
		},
		Fallback: engine.DecisionPass,
	}
	reg := newRegistry(deps, []orchestrator.PolicySet{policy})
	pipe, err := reg.Get("scene")
	require.NoError(t, err)

	res, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		RequestID: "req3",
		UserID:    "bad-user",
	})
	require.NoError(t, err)
	assert.Equal(t, engine.DecisionReject, res.Decision)
}

func TestPipeline_SceneNotFound(t *testing.T) {
	reg := orchestrator.NewRegistry(orchestrator.Deps{
		Features: &stubFeature{},
		Rules:    &stubRuleEvaluator{},
		Models:   &stubModelRegistry{},
		List:     &stubListService{statuses: map[string]list.Status{}},
		Breakers: nil,
	})
	_, err := reg.Get("missing_scene")
	require.Error(t, err)
}
