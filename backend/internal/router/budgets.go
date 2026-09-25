package router

import "github.com/gin-gonic/gin"

func registerBudgets(api *gin.RouterGroup, h *Handlers) {
	budgets := api.Group("/budgets")
	{
		budgets.GET("", h.Budget.List)
		budgets.POST("", h.Budget.Create)
		budgets.GET("/:id", h.Budget.Get)
		budgets.PUT("/:id", h.Budget.Update)
		budgets.POST("/:id/submit", h.Budget.Submit)
		budgets.POST("/:id/approve", h.Budget.Approve)
		budgets.POST("/:id/reject", h.Budget.Reject)
	}
}
