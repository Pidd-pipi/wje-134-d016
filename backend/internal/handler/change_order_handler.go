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

// ChangeOrderHandler exposes change order endpoints.
type ChangeOrderHandler struct {
	svc    *service.ChangeOrderService
	logger *slog.Logger
}

// NewChangeOrderHandler builds a ChangeOrderHandler.
func NewChangeOrderHandler(svc *service.ChangeOrderService, logger *slog.Logger) *ChangeOrderHandler {
	return &ChangeOrderHandler{svc: svc, logger: logger}
}

// List handles GET /change-orders?projectId=.
func (h *ChangeOrderHandler) List(c *gin.Context) {
	projectID, _ := parseUint(c.Query("projectId"))
	list, err := h.svc.List(uint(projectID))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, list)
}

// Get handles GET /change-orders/:id.
func (h *ChangeOrderHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	order, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Create handles POST /change-orders.
func (h *ChangeOrderHandler) Create(c *gin.Context) {
	var req dto.CreateChangeOrderRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	order, err := h.svc.Create(req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Update handles PUT /change-orders/:id.
func (h *ChangeOrderHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateChangeOrderRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	order, err := h.svc.Update(id, req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Submit handles POST /change-orders/:id/submit.
func (h *ChangeOrderHandler) Submit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	order, err := h.svc.Submit(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Approve handles POST /change-orders/:id/approve.
func (h *ChangeOrderHandler) Approve(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	if !constants.CanManageChangeOrder(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	order, err := h.svc.Approve(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Reject handles POST /change-orders/:id/reject.
func (h *ChangeOrderHandler) Reject(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	if !constants.CanManageChangeOrder(u.Role) {
		util.Fail(c, constants.ErrForbidden)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	order, err := h.svc.Reject(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}

// Cancel handles POST /change-orders/:id/cancel.
func (h *ChangeOrderHandler) Cancel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	order, err := h.svc.Cancel(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, order)
}
