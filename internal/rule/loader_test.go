// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package rule_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/zhucl121/risk-engine/internal/engine"
	"github.com/zhucl121/risk-engine/internal/feature"
	"github.com/zhucl121/risk-engine/internal/rule"
	"github.com/zhucl121/risk-engine/pkg/dsl"
	"github.com/zhucl121/risk-engine/pkg/dsl/builtins"
)

func newTestReg(t *testing.T) *dsl.FunctionRegistry {
	t.Helper()
	reg := dsl.NewFunctionRegistry()
	require.NoError(t, builtins.RegisterAll(reg, builtins.Deps{}))
	return reg
}

func TestDBLoader_LoadAll(t *testing.T) {
	repo := rule.NewFakeRepository()
	logger := zap.NewNop()
	reg := newTestReg(t)

	// Seed two rules.
	ctx := context.Background()
	_, err := repo.Create(ctx, &rule.RuleRecord{
		RuleKey:        "TEST_RULE_001",
		Name:           "测试规则1",
		GroupName:      "test",
		SceneCode:      "TEST_SCENE",
		Priority:       100,
		ConditionDSL:   "amount > 1000",
		ActionDecision: "REJECT",
		ActionRiskCode: "TEST_REJECT",
		ActionScore:    900,
		Status:         1,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, &rule.RuleRecord{
		RuleKey:        "TEST_RULE_002",
		Name:           "测试规则2",
		GroupName:      "test",
		SceneCode:      "TEST_SCENE",
		Priority:       50,
		ConditionDSL:   "amount > 500",
		ActionDecision: "MANUAL_REVIEW",
		ActionRiskCode: "TEST_REVIEW",
		ActionScore:    500,
		Status:         1,
	})
	require.NoError(t, err)

	loader := rule.NewDBLoader(repo, reg, logger)
	rules, err := loader.LoadAll(ctx, "TEST_SCENE")
	require.NoError(t, err)
	assert.Len(t, rules, 2)
}

func TestDBLoader_InvalidDSL_Skipped(t *testing.T) {
	repo := rule.NewFakeRepository()
	logger, _ := zap.NewDevelopment()
	reg := newTestReg(t)

	ctx := context.Background()
	_, err := repo.Create(ctx, &rule.RuleRecord{
		RuleKey:        "VALID_RULE",
		Name:           "Valid",
		GroupName:      "test",
		SceneCode:      "S",
		Priority:       100,
		ConditionDSL:   "amount > 0",
		ActionDecision: "PASS",
		Status:         1,
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, &rule.RuleRecord{
		RuleKey:        "INVALID_RULE",
		Name:           "Bad DSL",
		GroupName:      "test",
		SceneCode:      "S",
		Priority:       200,
		ConditionDSL:   "amount > > 0", // syntax error
		ActionDecision: "REJECT",
		Status:         1,
	})
	require.NoError(t, err)

	loader := rule.NewDBLoader(repo, reg, logger)
	rules, err := loader.LoadAll(ctx, "S")
	require.NoError(t, err)
	// Only the valid rule should be loaded; invalid DSL is logged and skipped.
	assert.Len(t, rules, 1)
	assert.Equal(t, "VALID_RULE", rules[0].ID())
}

func TestDBLoader_HotReload(t *testing.T) {
	repo := rule.NewFakeRepository()
	logger := zap.NewNop()
	reg := newTestReg(t)
	ctx := context.Background()

	id, err := repo.Create(ctx, &rule.RuleRecord{
		RuleKey: "RELOAD_RULE", Name: "R", GroupName: "g",
		SceneCode: "S", Priority: 100,
		ConditionDSL:   "amount > 100",
		ActionDecision: "REJECT", Status: 1,
	})
	require.NoError(t, err)

	loader := rule.NewDBLoader(repo, reg, logger)
	before := time.Now()

	// Update the rule and verify incremental load sees it.
	time.Sleep(2 * time.Millisecond)
	rec, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	rec.ConditionDSL = "amount > 999"
	require.NoError(t, repo.Update(ctx, rec))

	updated, err := loader.LoadUpdatedSince(ctx, before)
	require.NoError(t, err)
	assert.Len(t, updated, 1)
	assert.Equal(t, "RELOAD_RULE", updated[0].ID())
}

func TestDSLRule_Evaluate(t *testing.T) {
	repo := rule.NewFakeRepository()
	logger := zap.NewNop()
	reg := newTestReg(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, &rule.RuleRecord{
		RuleKey: "EVAL_RULE", Name: "E", GroupName: "g",
		SceneCode: "S", Priority: 100,
		ConditionDSL:   "amount > 100",
		ActionDecision: "REJECT",
		ActionRiskCode: "HIGH_AMOUNT",
		ActionScore:    800,
		Status:         1,
	})
	require.NoError(t, err)

	loader := rule.NewDBLoader(repo, reg, logger)
	rules, err := loader.LoadAll(ctx, "S")
	require.NoError(t, err)
	require.Len(t, rules, 1)

	rctx := &rule.Context{
		Request: &engine.DecisionRequest{
			Extra: map[string]string{"amount": "200"},
		},
		Features: feature.Map{
			"extra.amount": feature.Value{Kind: feature.KindInt, IntVal: 200},
		},
	}

	// amount 200 > 100: rule should hit.
	res, err := rules[0].Evaluate(ctx, rctx)
	require.NoError(t, err)
	assert.True(t, res.Hit)
	assert.Equal(t, engine.DecisionReject, res.Decision)
	assert.Equal(t, "HIGH_AMOUNT", res.RiskCode)

	// amount 50 <= 100: rule should not hit.
	rctx.Request.Extra["amount"] = "50"
	rctx.Features["extra.amount"] = feature.Value{Kind: feature.KindInt, IntVal: 50}
	res, err = rules[0].Evaluate(ctx, rctx)
	require.NoError(t, err)
	assert.False(t, res.Hit)
}
