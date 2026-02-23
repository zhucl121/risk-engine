// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package orchestrator

import (
	"context"
	"fmt"

	"github.com/yourorg/riskengine/internal/engine"
)

// RegistryEvaluator wraps a Registry so it satisfies engine.Evaluator.
// It looks up the pipeline for req.SceneCode on every call, which supports
// hot-reload without requiring a restart.
type RegistryEvaluator struct {
	reg Registry
}

// NewRegistryEvaluator returns an engine.Evaluator backed by reg.
func NewRegistryEvaluator(reg Registry) *RegistryEvaluator {
	return &RegistryEvaluator{reg: reg}
}

// Execute implements engine.Evaluator.
func (e *RegistryEvaluator) Execute(ctx context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error) {
	pipe, err := e.reg.Get(req.SceneCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", engine.ErrSceneNotFound, req.SceneCode)
	}
	return pipe.Execute(ctx, req)
}
