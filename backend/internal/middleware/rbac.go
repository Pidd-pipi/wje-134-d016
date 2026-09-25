package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/util"
)

// RequireRole rejects callers whose role is not in the allowed set.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if !allowed[user.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, util.Response{
				Code:    constants.CodeForbidden,
				Message: constants.MsgForbidden,
			})
			return
		}
		c.Next()
	}
}
