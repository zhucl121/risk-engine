// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package adminv1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zhucl121/risk-engine/internal/rule"
	"github.com/zhucl121/risk-engine/pkg/dsl"
)

// RulesHandler handles all /admin/v1/rules endpoints.
type RulesHandler struct {
	repo    rule.RuleRepository
	dslReg  *dsl.FunctionRegistry
	watcher interface {
		ForceReload(ctx interface{ Deadline() (interface{}, bool) }) error
	}
	logger *zap.Logger
}

// Watcher is the subset of watcher.DBWatcher that RulesHandler needs.
// Defined as an interface to allow test mocking.
type Watcher interface {
	ForceReload(ctx interface{}) error
}

// NewRulesHandler creates a RulesHandler.
func NewRulesHandler(
	repo rule.RuleRepository,
	dslReg *dsl.FunctionRegistry,
	logger *zap.Logger,
) *RulesHandler {
	return &RulesHandler{repo: repo, dslReg: dslReg, logger: logger}
}

// RegisterRoutes wires all rule management routes into the given router group.
// group should be the /admin/v1 gin.RouterGroup.
func (h *RulesHandler) RegisterRoutes(group *gin.RouterGroup) {
	g := group.Group("/rules")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.POST("/validate", h.Validate)
	g.POST("/:id/enable", h.Enable)
	g.POST("/:id/disable", h.Disable)
}

// SetWatcher injects the DBWatcher so the reload endpoint can trigger it.
// Call after NewRulesHandler and before serving traffic.
func (h *RulesHandler) SetWatcher(w interface {
	ForceReload(ctx interface{ Deadline() (interface{}, bool) }) error
}) {
	h.watcher = w
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// List godoc
// GET /admin/v1/rules?scene_code=X&status=1&page=1&page_size=20
func (h *RulesHandler) List(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	records, err := h.repo.ListActive(c.Request.Context(), q.SceneCode)
	if err != nil {
		h.logger.Error("List rules failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: "DB_ERROR", Message: err.Error()})
		return
	}

	out := make([]RuleResponse, 0, len(records))
	for _, r := range records {
		out = append(out, toResponse(r))
	}
	c.JSON(http.StatusOK, gin.H{"data": out, "total": len(out)})
}

// Get godoc
// GET /admin/v1/rules/:id
func (h *RulesHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: "invalid id"})
		return
	}
	rec, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toResponse(rec))
}

// Create godoc
// POST /admin/v1/rules
func (h *RulesHandler) Create(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	// Validate DSL before persisting.
	if _, err := dsl.Compile(req.ConditionDSL, h.dslReg); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Code: "INVALID_DSL", Message: err.Error()})
		return
	}

	rec := fromCreateRequest(&req)
	id, err := h.repo.Create(c.Request.Context(), rec)
	if err != nil {
		h.logger.Error("Create rule failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: "DB_ERROR", Message: err.Error()})
		return
	}
	rec.ID = id
	h.logger.Info("Rule created", zap.String("rule_key", rec.RuleKey), zap.Int64("id", id))
	c.JSON(http.StatusCreated, toResponse(rec))
}

// Update godoc
// PUT /admin/v1/rules/:id
func (h *RulesHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: "invalid id"})
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}

	if _, err := dsl.Compile(req.ConditionDSL, h.dslReg); err != nil {
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Code: "INVALID_DSL", Message: err.Error()})
		return
	}

	rec := fromUpdateRequest(id, &req)
	if err := h.repo.Update(c.Request.Context(), rec); err != nil {
		h.logger.Error("Update rule failed", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusConflict, ErrorResponse{Code: "CONFLICT", Message: err.Error()})
		return
	}
	h.logger.Info("Rule updated", zap.Int64("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete godoc
// DELETE /admin/v1/rules/:id
func (h *RulesHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: "invalid id"})
		return
	}
	if err := h.repo.SoftDelete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Code: "DB_ERROR", Message: err.Error()})
		return
	}
	h.logger.Info("Rule soft-deleted", zap.Int64("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Validate godoc
// POST /admin/v1/rules/validate
func (h *RulesHandler) Validate(c *gin.Context) {
	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: err.Error()})
		return
	}
	if _, err := dsl.Compile(req.ConditionDSL, h.dslReg); err != nil {
		c.JSON(http.StatusOK, ValidateResponse{Valid: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ValidateResponse{Valid: true})
}

// Enable godoc
// POST /admin/v1/rules/:id/enable
func (h *RulesHandler) Enable(c *gin.Context) {
	h.setStatus(c, 1)
}

// Disable godoc
// POST /admin/v1/rules/:id/disable
func (h *RulesHandler) Disable(c *gin.Context) {
	h.setStatus(c, 0)
}

func (h *RulesHandler) setStatus(c *gin.Context, status int8) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Code: "BAD_REQUEST", Message: "invalid id"})
		return
	}
	rec, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
		return
	}
	rec.Status = status
	if err := h.repo.Update(c.Request.Context(), rec); err != nil {
		c.JSON(http.StatusConflict, ErrorResponse{Code: "CONFLICT", Message: err.Error()})
		return
	}
	action := "enabled"
	if status == 0 {
		action = "disabled"
	}
	h.logger.Info("Rule status changed", zap.Int64("id", id), zap.String("action", action))
	c.JSON(http.StatusOK, gin.H{"message": action})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func toResponse(r *rule.RuleRecord) RuleResponse {
	resp := RuleResponse{
		ID:             r.ID,
		RuleKey:        r.RuleKey,
		Name:           r.Name,
		GroupName:      r.GroupName,
		SceneCode:      r.SceneCode,
		Priority:       r.Priority,
		ConditionDSL:   r.ConditionDSL,
		ActionDecision: r.ActionDecision,
		ActionRiskCode: r.ActionRiskCode,
		ActionScore:    r.ActionScore,
		Status:         r.Status,
		Version:        r.Version,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
	}
	if len(r.ConditionAST) > 0 {
		var ast any
		if err := json.Unmarshal(r.ConditionAST, &ast); err == nil {
			resp.ConditionAST = ast
		}
	}
	return resp
}

func fromCreateRequest(req *CreateRuleRequest) *rule.RuleRecord {
	rec := &rule.RuleRecord{
		RuleKey:        req.RuleKey,
		Name:           req.Name,
		GroupName:      req.GroupName,
		SceneCode:      req.SceneCode,
		Priority:       req.Priority,
		ConditionDSL:   req.ConditionDSL,
		ActionDecision: req.ActionDecision,
		ActionRiskCode: req.ActionRiskCode,
		ActionScore:    req.ActionScore,
	}
	if req.ConditionAST != nil {
		if b, err := json.Marshal(req.ConditionAST); err == nil {
			rec.ConditionAST = b
		}
	}
	return rec
}

func fromUpdateRequest(id int64, req *UpdateRuleRequest) *rule.RuleRecord {
	rec := &rule.RuleRecord{
		ID:             id,
		Name:           req.Name,
		GroupName:      req.GroupName,
		SceneCode:      req.SceneCode,
		Priority:       req.Priority,
		ConditionDSL:   req.ConditionDSL,
		ActionDecision: req.ActionDecision,
		ActionRiskCode: req.ActionRiskCode,
		ActionScore:    req.ActionScore,
		Version:        req.Version,
		Status:         1,
	}
	if req.ConditionAST != nil {
		if b, err := json.Marshal(req.ConditionAST); err == nil {
			rec.ConditionAST = b
		}
	}
	return rec
}
