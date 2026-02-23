// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/riskengine/internal/engine"
	"github.com/yourorg/riskengine/internal/feature"
	"github.com/yourorg/riskengine/internal/list"
	"github.com/yourorg/riskengine/internal/orchestrator"
	"github.com/yourorg/riskengine/internal/rule"
	"github.com/yourorg/riskengine/internal/scene"
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
			// amount is now carried via Extra; use extra.amount source expression.
			"amt": "extra.amount",
		},
	}
	pipe := newExtraPipeline(nil, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		UserID:    "u999",
		Extra:     map[string]string{"amount": "5000"},
	})
	require.NoError(t, err)
	assert.Equal(t, "u999", capture.captured.GetString("uid"))
	// amount is injected as extra.amount string; resolveSource returns string value.
	assert.Equal(t, "5000", capture.captured.GetString("amt"))
}

// ── ExtraParamLoader integration ──────────────────────────────────────────────

// fakeExtraParamRepo is an in-memory ExtraParamRepository for testing.
type fakeExtraParamRepo struct {
	specs []*scene.ExtraParamSpec
}

func (r *fakeExtraParamRepo) ListByScene(_ context.Context, sceneCode string) ([]*scene.ExtraParamSpec, error) {
	var out []*scene.ExtraParamSpec
	for _, s := range r.specs {
		if s.SceneCode == sceneCode && s.Status == 1 {
			out = append(out, s)
		}
	}
	return out, nil
}
func (r *fakeExtraParamRepo) ListUpdatedSince(_ context.Context, _ time.Time) ([]*scene.ExtraParamSpec, error) {
	return nil, nil
}
func (r *fakeExtraParamRepo) GetByKey(_ context.Context, sceneCode, key string) (*scene.ExtraParamSpec, error) {
	for _, s := range r.specs {
		if s.SceneCode == sceneCode && s.ParamKey == key {
			return s, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *fakeExtraParamRepo) Upsert(_ context.Context, spec *scene.ExtraParamSpec) error {
	r.specs = append(r.specs, spec)
	return nil
}
func (r *fakeExtraParamRepo) Delete(_ context.Context, sceneCode, key string) error {
	return nil
}

func newExtraPipelineWithLoader(
	specs []*scene.ExtraParamSpec,
	step orchestrator.Step,
	capture *captureRuleCtx,
) orchestrator.Pipeline {
	repo := &fakeExtraParamRepo{specs: specs}
	loader := scene.NewExtraParamLoader(repo, 0, nil)

	reg := orchestrator.NewRegistry(orchestrator.Deps{
		Features:         &stubFeature{m: feature.Map{}},
		Rules:            capture,
		Models:           &stubModelRegistry{score: 0},
		List:             &stubListService{statuses: map[string]list.Status{}},
		Breakers:         nil,
		ExtraParamLoader: loader,
	})
	ps := orchestrator.PolicySet{
		SceneCode: "test",
		Pipeline:  []orchestrator.Step{step},
		Fallback:  engine.DecisionPass,
	}
	_ = reg.Reload(context.Background(), []orchestrator.PolicySet{ps})
	pipe, _ := reg.Get("test")
	return pipe
}

// TestValidateAndFillExtra_RequiredFieldMissing verifies that a missing required
// Extra field causes the pipeline to return an error.
func TestValidateAndFillExtra_RequiredFieldMissing(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		{SceneCode: "test", ParamKey: "merchant_id", ParamType: "string", Required: true, Status: 1},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{}, // merchant_id absent
	})
	require.Error(t, err, "should return error when required Extra field is missing")

	var missingErr *orchestrator.ErrMissingRequiredExtra
	require.ErrorAs(t, err, &missingErr)
	assert.Equal(t, "merchant_id", missingErr.ParamKey)
	assert.Equal(t, "test", missingErr.SceneCode)
}

// TestValidateAndFillExtra_RequiredFieldPresent verifies that the pipeline runs
// normally when a required Extra field is provided.
func TestValidateAndFillExtra_RequiredFieldPresent(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		{SceneCode: "test", ParamKey: "merchant_id", ParamType: "string", Required: true, Status: 1},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{"merchant_id": "MERCH_99"},
	})
	require.NoError(t, err)
	assert.Equal(t, "MERCH_99", capture.captured.GetString("extra.merchant_id"))
}

// TestValidateAndFillExtra_DefaultValue verifies that an optional field's
// default value is injected into the feature map when the field is absent.
func TestValidateAndFillExtra_DefaultValue(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		{
			SceneCode:  "test",
			ParamKey:   "product_type",
			ParamType:  "string",
			Required:   false,
			DefaultVal: "GOODS",
			Status:     1,
		},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{}, // product_type absent
	})
	require.NoError(t, err)
	// Default value should be injected into feature map.
	assert.Equal(t, "GOODS", capture.captured.GetString("extra.product_type"),
		"default value should be injected when optional field is absent")
}

// TestValidateAndFillExtra_DefaultValueNumeric verifies default-value
// injection and type coercion for int and bool field types.
func TestValidateAndFillExtra_DefaultValueNumeric(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		{SceneCode: "test", ParamKey: "order_count", ParamType: "int",  Required: false, DefaultVal: "7",     Status: 1},
		{SceneCode: "test", ParamKey: "is_recurring", ParamType: "bool", Required: false, DefaultVal: "false", Status: 1},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(7), capture.captured.GetInt("extra.order_count"))
	assert.False(t, capture.captured.GetBool("extra.is_recurring"))
}

// TestValidateAndFillExtra_OptionalWithNoDefault verifies that an optional field
// with no default value is simply ignored when absent (no feature entry injected).
func TestValidateAndFillExtra_OptionalWithNoDefault(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		{SceneCode: "test", ParamKey: "coupon_code", ParamType: "string", Required: false, DefaultVal: "", Status: 1},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{},
	})
	require.NoError(t, err)
	// No default → key should not appear in feature map.
	_, ok := capture.captured["extra.coupon_code"]
	assert.False(t, ok, "absent optional field without default should not be injected")
}

// TestValidateAndFillExtra_DisabledSpecIgnored verifies that specs with
// status=0 are completely ignored (neither validated nor filled).
func TestValidateAndFillExtra_DisabledSpecIgnored(t *testing.T) {
	specs := []*scene.ExtraParamSpec{
		// Status=0 (disabled): should be completely ignored.
		{SceneCode: "test", ParamKey: "hidden_field", ParamType: "string", Required: true, Status: 0},
	}
	step := orchestrator.Step{Name: "rules", Kind: orchestrator.StepKindRule}
	capture := &captureRuleCtx{}
	pipe := newExtraPipelineWithLoader(specs, step, capture)

	_, err := pipe.Execute(context.Background(), &engine.DecisionRequest{
		SceneCode: "test",
		Extra:     map[string]string{},
	})
	// Disabled spec should not cause validation error even though the field is absent.
	require.NoError(t, err)
}
