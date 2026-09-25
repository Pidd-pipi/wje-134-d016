// Package middleware provides HTTP middlewares shared by the API.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/config"
	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/util"
)

const ctxKeyUser = "auth_user"

// Auth validates the JWT Bearer token and injects the current user.
func Auth(cfg *config.Config, users *repository.UserRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		claims, err := util.ParseToken(cfg.JWTSecret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			logger.Warn("invalid jwt", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		user, err := users.FindByID(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		c.Set(ctxKeyUser, user)
		c.Next()
	}
}

// GetCurrentUser returns the authenticated user from the context.
func GetCurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get(ctxKeyUser); ok {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return &model.User{Name: "anonymous"}
}
