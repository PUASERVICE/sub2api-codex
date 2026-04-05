package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ModelStatusHandler struct {
	modelStatusService *service.ModelStatusService
}

func NewModelStatusHandler(modelStatusService *service.ModelStatusService) *ModelStatusHandler {
	return &ModelStatusHandler{modelStatusService: modelStatusService}
}

type createModelStatusTargetRequest struct {
	Name                 string `json:"name"`
	AccountID            int64  `json:"account_id" binding:"required"`
	ModelID              string `json:"model_id" binding:"required"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	TimeoutSeconds       int    `json:"timeout_seconds"`
	Enabled              *bool  `json:"enabled"`
}

type updateModelStatusTargetRequest struct {
	Name                 string `json:"name"`
	AccountID            int64  `json:"account_id"`
	ModelID              string `json:"model_id"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	TimeoutSeconds       int    `json:"timeout_seconds"`
	Enabled              *bool  `json:"enabled"`
}

func (h *ModelStatusHandler) ListTargets(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"
	targets, err := h.modelStatusService.ListTargets(c.Request.Context(), includeDisabled)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, targets)
}

func (h *ModelStatusHandler) GetOverview(c *gin.Context) {
	includeDisabled := c.Query("include_disabled") == "true"
	overview, err := h.modelStatusService.GetOverview(c.Request.Context(), includeDisabled)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *ModelStatusHandler) CreateTarget(c *gin.Context) {
	var req createModelStatusTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	target := &service.ModelStatusTarget{
		Name:                 strings.TrimSpace(req.Name),
		AccountID:            req.AccountID,
		ModelID:              strings.TrimSpace(req.ModelID),
		CheckIntervalSeconds: req.CheckIntervalSeconds,
		TimeoutSeconds:       req.TimeoutSeconds,
		Enabled:              true,
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	}

	created, err := h.modelStatusService.CreateTarget(c.Request.Context(), target)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *ModelStatusHandler) UpdateTarget(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid target id")
		return
	}

	var req updateModelStatusTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	target := &service.ModelStatusTarget{
		ID:                   targetID,
		Name:                 strings.TrimSpace(req.Name),
		AccountID:            req.AccountID,
		ModelID:              strings.TrimSpace(req.ModelID),
		CheckIntervalSeconds: req.CheckIntervalSeconds,
		TimeoutSeconds:       req.TimeoutSeconds,
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	} else {
		existing, err := h.modelStatusService.GetTarget(c.Request.Context(), targetID)
		if err != nil {
			response.NotFound(c, "target not found")
			return
		}
		target.Enabled = existing.Enabled
	}

	updated, err := h.modelStatusService.UpdateTarget(c.Request.Context(), target)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *ModelStatusHandler) DeleteTarget(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid target id")
		return
	}
	if err := h.modelStatusService.DeleteTarget(c.Request.Context(), targetID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ModelStatusHandler) RunTarget(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid target id")
		return
	}
	target, err := h.modelStatusService.RunTargetCheck(c.Request.Context(), targetID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, target)
}

func (h *ModelStatusHandler) ListChecks(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid target id")
		return
	}
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	checks, err := h.modelStatusService.ListChecks(c.Request.Context(), targetID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, checks)
}
