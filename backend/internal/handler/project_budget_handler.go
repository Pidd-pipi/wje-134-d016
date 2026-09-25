package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/middleware"
	"github.com/costguard/costguard/internal/service"
	"github.com/costguard/costguard/internal/util"
)

// ProjectBudgetHandler exposes budget endpoints.
type ProjectBudgetHandler struct {
	svc    *service.ProjectBudgetService
	logger *slog.Logger
}

// NewProjectBudgetHandler builds a ProjectBudgetHandler.
func NewProjectBudgetHandler(svc *service.ProjectBudgetService, logger *slog.Logger) *ProjectBudgetHandler {
	return &ProjectBudgetHandler{svc: svc, logger: logger}
}

// List handles GET /budgets?projectId=.
func (h *ProjectBudgetHandler) List(c *gin.Context) {
	projectID, _ := parseUint(c.Query("projectId"))
	list, err := h.svc.List(uint(projectID))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, list)
}

// Get handles GET /budgets/:id.
func (h *ProjectBudgetHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	b, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Create handles POST /budgets.
func (h *ProjectBudgetHandler) Create(c *gin.Context) {
	var req dto.CreateBudgetRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	b, err := h.svc.Create(req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Update handles PUT /budgets/:id.
func (h *ProjectBudgetHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateBudgetRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	b, err := h.svc.Update(id, req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Submit handles POST /budgets/:id/submit.
func (h *ProjectBudgetHandler) Submit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	b, err := h.svc.Submit(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Approve handles POST /budgets/:id/approve (FinanceManager/Admin).
func (h *ProjectBudgetHandler) Approve(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	if !constants.CanApproveBudget(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	b, err := h.svc.Approve(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Reject handles POST /budgets/:id/reject (FinanceManager/Admin).
func (h *ProjectBudgetHandler) Reject(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	if !constants.CanApproveBudget(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	b, err := h.svc.Reject(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}

// Close handles POST /budgets/:id/close (FinanceManager/Admin).
func (h *ProjectBudgetHandler) Close(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	if !constants.CanCloseBudget(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	b, abnormalities, err := h.svc.Close(id, u.ID, u.Name)
	if err != nil {
		if len(abnormalities) > 0 {
			util.FailWithData(c, err, abnormalities)
			return
		}
		util.Fail(c, err)
		return
	}
	util.OK(c, b)
}
