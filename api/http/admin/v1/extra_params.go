// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package adminv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yourorg/riskengine/internal/scene"
)

// ExtraParamsHandler manages /admin/v1/scenes/:scene/extra-params endpoints.
//
//	GET    /admin/v1/scenes/:scene/extra-params        — list specs for a scene
//	POST   /admin/v1/scenes/:scene/extra-params        — create / upsert a spec
//	PUT    /admin/v1/scenes/:scene/extra-params/:key   — update a spec
//	DELETE /admin/v1/scenes/:scene/extra-params/:key   — soft-delete a spec
type ExtraParamsHandler struct {
	repo   scene.ExtraParamRepository
	loader *scene.ExtraParamLoader // invalidated after write
	logger *zap.Logger
}

// NewExtraParamsHandler constructs the handler.
// loader may be nil; when non-nil the cache is invalidated on writes.
func NewExtraParamsHandler(
	repo scene.ExtraParamRepository,
	loader *scene.ExtraParamLoader,
	logger *zap.Logger,
) *ExtraParamsHandler {
	return &ExtraParamsHandler{repo: repo, loader: loader, logger: logger.Named("admin.extra_params")}
}

// Register mounts the handler routes on the given router group.
func (h *ExtraParamsHandler) Register(g *gin.RouterGroup) {
	g.GET("/scenes/:scene/extra-params", h.List)
	g.POST("/scenes/:scene/extra-params", h.Upsert)
	g.PUT("/scenes/:scene/extra-params/:key", h.UpsertByKey)
	g.DELETE("/scenes/:scene/extra-params/:key", h.Delete)
}

// extraParamRequest is the request body for create / update.
type extraParamRequest struct {
	ParamKey    string `json:"param_key"   binding:"required"`
	ParamType   string `json:"param_type"`  // string|int|float|bool; defaults to "string"
	Required    bool   `json:"required"`
	DefaultVal  string `json:"default_val"`
	Description string `json:"description"`
}

func (r *extraParamRequest) toSpec(sceneCode string) *scene.ExtraParamSpec {
	pt := r.ParamType
	if pt == "" {
		pt = "string"
	}
	return &scene.ExtraParamSpec{
		SceneCode:   sceneCode,
		ParamKey:    r.ParamKey,
		ParamType:   pt,
		Required:    r.Required,
		DefaultVal:  r.DefaultVal,
		Description: r.Description,
		Status:      1,
	}
}

// List godoc
// @Summary  List Extra parameter specs for a scene
// @Tags     extra-params
// @Produce  json
// @Param    scene  path  string  true  "Scene code"
// @Success  200  {array}  scene.ExtraParamSpec
// @Router   /admin/v1/scenes/{scene}/extra-params [get]
func (h *ExtraParamsHandler) List(c *gin.Context) {
	sceneCode := c.Param("scene")
	specs, err := h.repo.ListByScene(c.Request.Context(), sceneCode)
	if err != nil {
		h.logger.Error("ListByScene failed", zap.String("scene", sceneCode), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, specs)
}

// Upsert godoc
// @Summary  Create or update an Extra parameter spec
// @Tags     extra-params
// @Accept   json
// @Produce  json
// @Param    scene  path  string            true  "Scene code"
// @Param    body   body  extraParamRequest  true  "Spec payload"
// @Success  200  {object}  map[string]string
// @Router   /admin/v1/scenes/{scene}/extra-params [post]
func (h *ExtraParamsHandler) Upsert(c *gin.Context) {
	sceneCode := c.Param("scene")
	var req extraParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	spec := req.toSpec(sceneCode)
	if err := h.repo.Upsert(c.Request.Context(), spec); err != nil {
		h.logger.Error("Upsert failed", zap.String("scene", sceneCode), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate(sceneCode)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// UpsertByKey godoc
// @Summary  Update an Extra parameter spec by key
// @Tags     extra-params
// @Accept   json
// @Produce  json
// @Param    scene  path  string            true  "Scene code"
// @Param    key    path  string            true  "Param key"
// @Param    body   body  extraParamRequest  true  "Spec payload"
// @Success  200  {object}  map[string]string
// @Router   /admin/v1/scenes/{scene}/extra-params/{key} [put]
func (h *ExtraParamsHandler) UpsertByKey(c *gin.Context) {
	sceneCode := c.Param("scene")
	key := c.Param("key")
	var req extraParamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Path param wins over body for the key.
	req.ParamKey = key
	spec := req.toSpec(sceneCode)
	if err := h.repo.Upsert(c.Request.Context(), spec); err != nil {
		h.logger.Error("UpsertByKey failed", zap.String("scene", sceneCode), zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate(sceneCode)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Delete godoc
// @Summary  Soft-delete an Extra parameter spec
// @Tags     extra-params
// @Produce  json
// @Param    scene  path  string  true  "Scene code"
// @Param    key    path  string  true  "Param key"
// @Success  200  {object}  map[string]string
// @Router   /admin/v1/scenes/{scene}/extra-params/{key} [delete]
func (h *ExtraParamsHandler) Delete(c *gin.Context) {
	sceneCode := c.Param("scene")
	key := c.Param("key")
	if err := h.repo.Delete(c.Request.Context(), sceneCode, key); err != nil {
		h.logger.Error("Delete failed", zap.String("scene", sceneCode), zap.String("key", key), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.invalidate(sceneCode)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// invalidate removes the scene's cached specs so the next request re-fetches.
func (h *ExtraParamsHandler) invalidate(sceneCode string) {
	if h.loader != nil {
		h.loader.Invalidate(sceneCode)
	}
}
