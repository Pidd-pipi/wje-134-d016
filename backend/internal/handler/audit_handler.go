package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/service"
	"github.com/costguard/costguard/internal/util"
)

// AuditHandler exposes audit log endpoints.
type AuditHandler struct {
	svc    *service.AuditLogService
	logger *slog.Logger
}

// NewAuditHandler builds an AuditHandler.
func NewAuditHandler(svc *service.AuditLogService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{svc: svc, logger: logger}
}

// List handles GET /audit-logs?limit=.
func (h *AuditHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	logs, err := h.svc.List(limit)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, logs)
}
