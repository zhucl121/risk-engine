// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/list"
	"github.com/yourorg/riskengine/internal/orchestrator"
	"github.com/yourorg/riskengine/internal/rule"
)

// ── Extra injection ───────────────────────────────────────────────────────────

// captureRuleCtx is a rule.Evaluator that records the feature.Map it received.
type captureRuleCtx struct {
	captured feature.Map
}

func (c *captureRuleCtx) Evaluate(_ context.Context, rctx *rule.Context) ([]*rule.Result, error) {
	c.captured = rctx.Features
	return []*rule.Result{{RuleID: "r1", Hit: false, Decision: engine.DecisionPass}}, nil
}

func (c *captureRuleCtx) Reload(_ []rule.Rule) error { return nil }

func newExtraPipeline(extraSchema orchestrator.ExtraSchema, step orchestrator.Step, capture *captureRuleCtx) orchestrator.Pipeline {
	reg := orchestrator.NewRegistry(orchestrator.Deps{
		Features: &stubFeature{m: feature.Map{}},
		Rules:    capture,
		Models:   &stubModelRegistry{score: 0},
		List:     &stubListService{statuses: map[string]list.Status{}},
		Breakers: nil,
	})
	ps := orchestrator.PolicySet{
		SceneCode:   "test",
		Pipeline:    []orchestrator.Step{step},
		Fallback:    engine.DecisionPass,
		ExtraSchema: extraSchema,
	}
	_ = reg.Reload(context.Background(), []orchestrator.PolicySet{ps})
	pipe, _ := reg.Get("test")
	return pipe
}

func TestExtra_InjectAsString(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{
		Name:      "rules",
		Kind:      orchestrator.StepKindRule,
		RuleGroup: "test",
	}
	pipe := newExtraPipeline(nil, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"merchant_id": "M123"},
	})
	require.NoError(t, err)
	assert.Equal(t, "M123", capture.captured.GetString("extra.merchant_id"),
		"Extra string value should be injected as feature")
}

func TestExtra_InjectWithIntSchema(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	pipe := newExtraPipeline(
		orchestrator.ExtraSchema{"order_count": "int"},
		step,
		capture,
	)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"order_count": "42"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(42), capture.captured.GetInt("extra.order_count"))
}

func TestExtra_InjectWithFloatSchema(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	pipe := newExtraPipeline(
		orchestrator.ExtraSchema{"amount_usd": "float"},
		step,
		capture,
	)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"amount_usd": "9.99"},
	})
	require.NoError(t, err)
	assert.InDelta(t, 9.99, capture.captured.GetFloat("extra.amount_usd"), 0.001)
}

func TestExtra_InjectWithBoolSchema(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	pipe := newExtraPipeline(
		orchestrator.ExtraSchema{"is_vip": "bool"},
		step,
		capture,
	)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"is_vip": "true"},
	})
	require.NoError(t, err)
	assert.True(t, capture.captured.GetBool("extra.is_vip"))
}

func TestExtra_BadIntFallsBackToString(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	pipe := newExtraPipeline(
		orchestrator.ExtraSchema{"count": "int"},
		step,
		capture,
	)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"count": "notanumber"},
	})
	require.NoError(t, err)
	// Falls back to string, int accessor returns 0.
	assert.Equal(t, "notanumber", capture.captured.GetString("extra.count"))
}

// ── ParamMapping ──────────────────────────────────────────────────────────────

func TestParamMapping_OverridesFeatureForRule(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{
		Name: "rules",
		Kind: orchestrator.StepKindRule,
		ParamMapping: orchestrator.ParamMapping{
			"merchant_id": "extra.merchant_id", // maps Extra to downstream key
			"channel":     "WEB",               // literal constant
		},
	}
	pipe := newExtraPipeline(nil, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"merchant_id": "MERCH_42"},
	})
	require.NoError(t, err)
	// Both Extra-injected key AND mapped key should be present.
	assert.Equal(t, "MERCH_42", capture.captured.GetString("extra.merchant_id"))
	assert.Equal(t, "MERCH_42", capture.captured.GetString("merchant_id"))
	assert.Equal(t, "WEB", capture.captured.GetString("channel"))
}

func TestParamMapping_RequestFields(t *testing.T) {
	capture := &captureRuleCtx{}
	step := orchestrator.Step{
		Name: "rules",
		Kind: orchestrator.StepKindRule,
		ParamMapping: orchestrator.ParamMapping{
			"uid": "request.user_id",
			"amt": "request.amount",
		},
	}
	pipe := newExtraPipeline(nil, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		UserID:    "u999",
		Amount:    5000,
	})
	require.NoError(t, err)
	assert.Equal(t, "u999", capture.captured.GetString("uid"))
	assert.Equal(t, int64(5000), capture.captured.GetInt("amt"))
}
