package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/middleware"
	"github.com/costguard/costguard/internal/service"
	"github.com/costguard/costguard/internal/util"
)

// CostReportHandler exposes report endpoints.
type CostReportHandler struct {
	svc    *service.AnalyticsService
	logger *slog.Logger
}

// NewCostReportHandler builds a CostReportHandler.
func NewCostReportHandler(svc *service.AnalyticsService, logger *slog.Logger) *CostReportHandler {
	return &CostReportHandler{svc: svc, logger: logger}
}

// List handles GET /reports?projectId=.
func (h *CostReportHandler) List(c *gin.Context) {
	projectID, _ := parseUint(c.Query("projectId"))
	list, err := h.svc.List(uint(projectID))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, list)
}

// Get handles GET /reports/:id (Redis-cached).
func (h *CostReportHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	report, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, report)
}

// Generate handles POST /reports/generate.
func (h *CostReportHandler) Generate(c *gin.Context) {
	var req dto.GenerateReportRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	report, err := h.svc.Generate(req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, report)
}
