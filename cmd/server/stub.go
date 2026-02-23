// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/yourorg/riskengine/internal/engine"
)

// stubEngine is a minimal Engine implementation used during initial bringup.
// Replace with the real wired engine before production use.
type stubEngine struct{}

func newStubEngine() engine.Engine { return &stubEngine{} }

func (s *stubEngine) Evaluate(_ context.Context, req *engine.DecisionRequest) (*engine.DecisionResult, error) {
	return &engine.DecisionResult{
		RequestID: req.RequestID,
		Decision:  engine.DecisionPass,
		RiskScore: 0,
		RiskLevel: engine.RiskLevelLow,
	}, nil
}

func (s *stubEngine) Reload(_ context.Context) error { return nil }

func (s *stubEngine) Health() engine.HealthStatus {
	return engine.HealthStatus{Healthy: true, Components: map[string]bool{"stub": true}}
}
