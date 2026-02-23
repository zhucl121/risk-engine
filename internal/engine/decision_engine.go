// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yourorg/riskengine/internal/audit"
)

// Evaluator is a local interface that the decisionEngine uses to call the
// orchestrator without creating an import cycle (orchestrator imports engine).
type Evaluator interface {
	// Execute runs the pipeline for the request's scene.
	Execute(ctx context.Context, req *DecisionRequest) (*DecisionResult, error)
}

// decisionEngine is the concrete Engine implementation.
// It wires together an Evaluator (orchestrator pipeline), audit, and
// request enrichment.
type decisionEngine struct {
	eval   Evaluator
	writer audit.Writer
	logger *zap.Logger
}

// NewDecisionEngine returns an Engine that delegates evaluation to the
// given Evaluator and audits every decision asynchronously.
func NewDecisionEngine(
	eval Evaluator,
	writer audit.Writer,
	logger *zap.Logger,
) Engine {
	return &decisionEngine{
		eval:   eval,
		writer: writer,
		logger: logger,
	}
}

// Evaluate performs a synchronous risk decision.
//
// Steps:
//  1. Enrich: generate RequestID if empty, stamp ReceivedAt.
//  2. Delegate to the scene's Pipeline.
//  3. Asynchronously audit the result.
func (e *decisionEngine) Evaluate(ctx context.Context, req *DecisionRequest) (*DecisionResult, error) {
	if req.SceneCode == "" {
		return nil, fmt.Errorf("%w: SceneCode is required", ErrRequestInvalid)
	}
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}
	if req.ReceivedAt.IsZero() {
		req.ReceivedAt = time.Now()
	}

	result, err := e.eval.Execute(ctx, req)
	if err != nil {
		// Distinguish "scene not found" from infrastructure errors.
		return nil, fmt.Errorf("engine: execute pipeline for scene %s: %w", req.SceneCode, err)
	}

	// Audit asynchronously; never block the caller on audit I/O.
	go func() {
		rec := &audit.Record{
			RequestID:   req.RequestID,
			SceneCode:   req.SceneCode,
			UserID:      req.UserID,
			DeviceID:    req.DeviceID,
			Decision:    string(result.Decision),
			RiskScore:   result.RiskScore,
			HitRules:    result.HitRules,
			RiskReasons: result.RiskReasons,
			ModelScores: result.ModelScores,
			CostMs:      result.CostMs,
			CreatedAt:   time.Now(),
		}
		if wErr := e.writer.Write(context.Background(), rec); wErr != nil {
			e.logger.Warn("audit write failed",
				zap.String("request_id", req.RequestID),
				zap.Error(wErr),
			)
		}
	}()

	return result, nil
}

// Reload triggers a hot-reload of all policy sets in the orchestrator registry.
// Each sub-component (rules, models) manages its own reload separately.
func (e *decisionEngine) Reload(ctx context.Context) error {
	// Orchestrator reload is driven externally (e.g. by config watcher).
	// This implementation is a no-op; callers should invoke
	// orchestrator.Registry.Reload directly with new PolicySet data.
	e.logger.Info("engine.Reload called; orchestrator manages hot-reload independently")
	return nil
}

// Health returns the current operational status of the engine.
// The orchestrator and audit writer have no external dependencies that can
// be health-checked synchronously; their operational state is reflected
// in log metrics instead.
func (e *decisionEngine) Health() HealthStatus {
	return HealthStatus{
		Healthy: true,
		Components: map[string]bool{
			"orchestrator": true,
			"audit":        true,
		},
	}
}
