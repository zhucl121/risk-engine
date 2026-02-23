// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package v1 provides the HTTP API handlers for the risk decision engine.
// All endpoints are versioned under /api/v1/.
package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yourorg/riskengine/internal/engine"
)

// Handler holds dependencies for the HTTP API handlers.
type Handler struct {
	engine engine.Engine
	logger *zap.Logger
}

// NewHandler returns an initialised Handler.
func NewHandler(e engine.Engine, logger *zap.Logger) *Handler {
	return &Handler{engine: e, logger: logger}
}

// RegisterRoutes mounts all v1 routes onto the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/decision", h.Evaluate)
	rg.GET("/health", h.Health)
}

// decisionRequest is the JSON body for POST /api/v1/decision.
type decisionRequest struct {
	RequestID string            `json:"request_id"`
	SceneCode string            `json:"scene_code" binding:"required"`
	UserID    string            `json:"user_id"`
	DeviceID  string            `json:"device_id"`
	SessionID string            `json:"session_id"`
	IP        string            `json:"ip"`
	Amount    int64             `json:"amount"`
	Extra     map[string]string `json:"extra"`
}

// decisionResponse is the JSON response for POST /api/v1/decision.
type decisionResponse struct {
	RequestID   string             `json:"request_id"`
	Decision    engine.Decision    `json:"decision"`
	RiskScore   int                `json:"risk_score"`
	RiskLevel   engine.RiskLevel   `json:"risk_level"`
	HitRules    []string           `json:"hit_rules,omitempty"`
	ModelScores map[string]float64 `json:"model_scores,omitempty"`
	RiskReasons []string           `json:"risk_reasons,omitempty"`
	Actions     []string           `json:"actions,omitempty"`
	CostMs      int64              `json:"cost_ms"`
}

// Evaluate handles POST /api/v1/decision.
func (h *Handler) Evaluate(c *gin.Context) {
	var req decisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	engineReq := &engine.DecisionRequest{
		RequestID:  req.RequestID,
		SceneCode:  req.SceneCode,
		UserID:     req.UserID,
		DeviceID:   req.DeviceID,
		SessionID:  req.SessionID,
		IP:         req.IP,
		Amount:     req.Amount,
		Extra:      req.Extra,
		ReceivedAt: time.Now(),
	}

	result, err := h.engine.Evaluate(c.Request.Context(), engineReq)
	if err != nil {
		h.logger.Error("engine evaluate failed",
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, decisionResponse{
		RequestID:   result.RequestID,
		Decision:    result.Decision,
		RiskScore:   result.RiskScore,
		RiskLevel:   result.RiskLevel,
		HitRules:    result.HitRules,
		ModelScores: result.ModelScores,
		RiskReasons: result.RiskReasons,
		Actions:     result.Actions,
		CostMs:      result.CostMs,
	})
}

// healthResponse is the JSON response for GET /api/v1/health.
type healthResponse struct {
	Healthy    bool            `json:"healthy"`
	Components map[string]bool `json:"components,omitempty"`
}

// Health handles GET /api/v1/health.
func (h *Handler) Health(c *gin.Context) {
	status := h.engine.Health()
	code := http.StatusOK
	if !status.Healthy {
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, healthResponse{
		Healthy:    status.Healthy,
		Components: status.Components,
	})
}
