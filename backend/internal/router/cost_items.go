package router

import "github.com/gin-gonic/gin"

func registerCostItems(api *gin.RouterGroup, h *Handlers) {
	items := api.Group("/cost-items")
	{
		items.GET("", h.CostItem.List)
		items.POST("", h.CostItem.Create)
		items.GET("/:id", h.CostItem.Get)
		items.PUT("/:id", h.CostItem.Update)
		items.POST("/:id/mark-abnormal", h.CostItem.MarkAbnormal)
	}
}
