// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package engine_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/yourorg/riskengine/internal/audit"
	"github.com/yourorg/riskengine/internal/engine"
)

// stubEvaluator implements engine.Evaluator for testing.
type stubEvaluator struct {
	result *engine.DecisionResult
	err    error
}

func (s *stubEvaluator) Execute(_ context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	r := *s.result
	r.RequestID = req.RequestID
	return &r, nil
}

// stubWriter is a no-op audit.Writer.
type stubWriter struct{}

func (s *stubWriter) Write(_ context.Context, _ *audit.Record) error { return nil }
func (s *stubWriter) Flush(_ context.Context) error                  { return nil }
func (s *stubWriter) Close() error                                   { return nil }

func TestDecisionEngine_PassDecision(t *testing.T) {
	eval := &stubEvaluator{result: &engine.DecisionResult{Decision: engine.DecisionPass, RiskScore: 10}}
	eng := engine.NewDecisionEngine(eval, &stubWriter{}, zaptest.NewLogger(t))

	res, err := eng.Evaluate(context.Background(), &engine.DecisionRequest{
		SceneCode: "payment",
		UserID:    "u1",
	})
	require.NoError(t, err)
	assert.Equal(t, engine.DecisionPass, res.Decision)
	assert.NotEmpty(t, res.RequestID, "RequestID should be auto-generated")
}

func TestDecisionEngine_RejectDecision(t *testing.T) {
	eval := &stubEvaluator{result: &engine.DecisionResult{Decision: engine.DecisionReject, RiskScore: 950}}
	eng := engine.NewDecisionEngine(eval, &stubWriter{}, zaptest.NewLogger(t))

	res, err := eng.Evaluate(context.Background(), &engine.DecisionRequest{
		RequestID: "explicit-id",
		SceneCode: "payment",
	})
	require.NoError(t, err)
	assert.Equal(t, engine.DecisionReject, res.Decision)
	assert.Equal(t, "explicit-id", res.RequestID)
}

func TestDecisionEngine_MissingSceneCode(t *testing.T) {
	eng := engine.NewDecisionEngine(&stubEvaluator{result: &engine.DecisionResult{}}, &stubWriter{}, zaptest.NewLogger(t))
	_, err := eng.Evaluate(context.Background(), &engine.DecisionRequest{})
	require.Error(t, err)
}

func TestDecisionEngine_Health(t *testing.T) {
	eng := engine.NewDecisionEngine(&stubEvaluator{result: &engine.DecisionResult{}}, &stubWriter{}, zaptest.NewLogger(t))
	h := eng.Health()
	assert.True(t, h.Healthy)
}
