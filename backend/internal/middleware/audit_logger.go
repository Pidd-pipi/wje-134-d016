package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/service"
)

// AuditLogger records non-GET mutations at the HTTP layer; business-level
// audit entries are additionally written by services with precise context.
func AuditLogger(audit *service.AuditLogService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Writer.Status() >= 400 {
			return
		}
		user := GetCurrentUser(c)
		audit.Record(user.ID, user.Name, c.Request.Method+" "+c.Request.URL.Path, "http", 0, c.Request.URL.Path)
		_ = logger
	}
}
