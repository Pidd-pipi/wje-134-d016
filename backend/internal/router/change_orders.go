package router

import "github.com/gin-gonic/gin"

func registerChangeOrders(api *gin.RouterGroup, h *Handlers) {
	orders := api.Group("/change-orders")
	{
		orders.GET("", h.ChangeOrder.List)
		orders.POST("", h.ChangeOrder.Create)
		orders.GET("/:id", h.ChangeOrder.Get)
		orders.PUT("/:id", h.ChangeOrder.Update)
		orders.POST("/:id/submit", h.ChangeOrder.Submit)
		orders.POST("/:id/approve", h.ChangeOrder.Approve)
		orders.POST("/:id/reject", h.ChangeOrder.Reject)
		orders.POST("/:id/cancel", h.ChangeOrder.Cancel)
	}
}
