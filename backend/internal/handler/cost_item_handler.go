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

// CostItemHandler exposes cost item endpoints.
type CostItemHandler struct {
	svc    *service.CostItemService
	logger *slog.Logger
}

// NewCostItemHandler builds a CostItemHandler.
func NewCostItemHandler(svc *service.CostItemService, logger *slog.Logger) *CostItemHandler {
	return &CostItemHandler{svc: svc, logger: logger}
}

// List handles GET /cost-items?budgetId=.
func (h *CostItemHandler) List(c *gin.Context) {
	budgetID, _ := parseUint(c.Query("budgetId"))
	list, err := h.svc.List(uint(budgetID))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, list)
}

// Get handles GET /cost-items/:id.
func (h *CostItemHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

// Create handles POST /cost-items (Accountant/Admin/FinanceManager).
func (h *CostItemHandler) Create(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	if !constants.CanRecordCost(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	var req dto.CreateCostItemRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.Create(req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

// Update handles PUT /cost-items/:id (Accountant/Admin/FinanceManager).
func (h *CostItemHandler) Update(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	if !constants.CanRecordCost(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCostItemRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	item, err := h.svc.Update(id, req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}

// MarkAbnormal handles POST /cost-items/:id/mark-abnormal.
func (h *CostItemHandler) MarkAbnormal(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	if !constants.CanRecordCost(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.MarkAbnormal(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, item)
}
