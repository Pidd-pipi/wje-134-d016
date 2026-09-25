package router

import "github.com/gin-gonic/gin"

func registerReports(api *gin.RouterGroup, h *Handlers) {
	reports := api.Group("/reports")
	{
		reports.GET("", h.CostReport.List)
		reports.GET("/:id", h.CostReport.Get)
		reports.POST("/generate", h.CostReport.Generate)
	}
}
