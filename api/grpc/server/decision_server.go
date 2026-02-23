// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package server implements the gRPC DecisionService server.
package server

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	riskv1 "github.com/yourorg/riskengine/api/grpc/v1"
	"github.com/yourorg/riskengine/internal/engine"
)

// DecisionServer implements riskv1.DecisionServiceServer.
type DecisionServer struct {
	riskv1.UnimplementedDecisionServiceServer
	eng    engine.Engine
	logger *zap.Logger
}

// NewDecisionServer returns a DecisionServer backed by eng.
func NewDecisionServer(eng engine.Engine, logger *zap.Logger) *DecisionServer {
	return &DecisionServer{eng: eng, logger: logger}
}

// Evaluate implements DecisionService.Evaluate.
func (s *DecisionServer) Evaluate(ctx context.Context, req *riskv1.DecisionRequest) (*riskv1.DecisionResponse, error) {
	eReq := protoToEngineReq(req)

	result, err := s.eng.Evaluate(ctx, eReq)
	if err != nil {
		s.logger.Error("grpc evaluate failed",
			zap.String("request_id", req.RequestId),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Internal, "evaluate: %v", err)
	}

	return engineResToProto(result), nil
}

// BatchEvaluate implements DecisionService.BatchEvaluate.
func (s *DecisionServer) BatchEvaluate(ctx context.Context, req *riskv1.BatchDecisionRequest) (*riskv1.BatchDecisionResponse, error) {
	responses := make([]*riskv1.DecisionResponse, 0, len(req.Requests))
	for _, r := range req.Requests {
		resp, err := s.Evaluate(ctx, r)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return &riskv1.BatchDecisionResponse{Responses: responses}, nil
}

// Health implements DecisionService.Health.
func (s *DecisionServer) Health(_ context.Context, _ *riskv1.HealthRequest) (*riskv1.HealthResponse, error) {
	h := s.eng.Health()
	return &riskv1.HealthResponse{
		Healthy:    h.Healthy,
		Components: boolMapToProto(h.Components),
	}, nil
}

// ── conversion helpers ───────────────────────────────────────────────────────

func protoToEngineReq(req *riskv1.DecisionRequest) *engine.DecisionRequest {
	return &engine.DecisionRequest{
		RequestID:  req.RequestId,
		SceneCode:  req.SceneCode,
		UserID:     req.UserId,
		DeviceID:   req.DeviceId,
		SessionID:  req.SessionId,
		IP:         req.Ip,
		Extra:      req.Extra,
		ReceivedAt: time.Now(),
	}
}

func engineResToProto(r *engine.DecisionResult) *riskv1.DecisionResponse {
	scores := make(map[string]float64, len(r.ModelScores))
	for k, v := range r.ModelScores {
		scores[k] = v
	}
	return &riskv1.DecisionResponse{
		RequestId:   r.RequestID,
		Decision:    decisionToProto(r.Decision),
		RiskScore:   int32(r.RiskScore),
		RiskLevel:   riskLevelToProto(r.RiskLevel),
		HitRules:    r.HitRules,
		ModelScores: scores,
		RiskReasons: r.RiskReasons,
		Actions:     r.Actions,
		CostMs:      r.CostMs,
	}
}

func decisionToProto(d engine.Decision) riskv1.Decision {
	switch d {
	case engine.DecisionPass:
		return riskv1.Decision_DECISION_PASS
	case engine.DecisionReject:
		return riskv1.Decision_DECISION_REJECT
	case engine.DecisionManualReview:
		return riskv1.Decision_DECISION_MANUAL_REVIEW
	}
	return riskv1.Decision_DECISION_UNSPECIFIED
}

func riskLevelToProto(l engine.RiskLevel) riskv1.RiskLevel {
	switch l {
	case engine.RiskLevelLow:
		return riskv1.RiskLevel_RISK_LEVEL_LOW
	case engine.RiskLevelMedium:
		return riskv1.RiskLevel_RISK_LEVEL_MEDIUM
	case engine.RiskLevelHigh:
		return riskv1.RiskLevel_RISK_LEVEL_HIGH
	case engine.RiskLevelCritical:
		return riskv1.RiskLevel_RISK_LEVEL_CRITICAL
	}
	return riskv1.RiskLevel_RISK_LEVEL_UNSPECIFIED
}

func boolMapToProto(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
