package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/service"
	"github.com/costguard/costguard/internal/util"
)

// AuthHandler exposes JWT login endpoints.
type AuthHandler struct {
	svc    *service.AuthService
	logger *slog.Logger
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(svc *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	user, token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, dto.LoginResponse{Token: token, User: user})
}
