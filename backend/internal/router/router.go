// Package router wires up the Gin engine and all routes.
package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/config"
	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/docs"
	"github.com/costguard/costguard/internal/handler"
	"github.com/costguard/costguard/internal/middleware"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/service"
	"github.com/costguard/costguard/internal/util"
)

// Handlers aggregates every HTTP handler for assembly.
type Handlers struct {
	Auth        *handler.AuthHandler
	Budget      *handler.ProjectBudgetHandler
	CostItem    *handler.CostItemHandler
	ChangeOrder *handler.ChangeOrderHandler
	CostReport  *handler.CostReportHandler
	Audit       *handler.AuditHandler
}

// New assembles the Gin engine and registers all routes.
func New(cfg *config.Config, logger *slog.Logger, h *Handlers, users *repository.UserRepository, rdb *redis.Client, audit *service.AuditLogService, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.ErrorHandler(logger), middleware.RequestLogger(logger))
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
	}))

	healthz := func(c *gin.Context) {
		c.JSON(http.StatusOK, util.Response{Code: constants.CodeSuccess, Message: constants.MsgOK, Data: gin.H{"status": "ok", "service": "costguard"}})
	}
	engine.GET("/healthz", healthz)
	engine.GET("/api/healthz", healthz)
	engine.GET("/readyz", readiness(db, rdb, logger))
	engine.GET("/api/readyz", readiness(db, rdb, logger))

	engine.GET("/docs", docs.IndexHandler())
	engine.GET("/docs/openapi.json", docs.OpenAPIHandler())

	api := engine.Group("/api/v1")
	auth := api.Group("/auth", middleware.RateLimit(rdb, cfg.AuthRateLimit, logger))
	auth.POST("/login", h.Auth.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(cfg, users, logger))
	protected.Use(middleware.RateLimit(rdb, cfg.APIRateLimit, logger))
	protected.Use(middleware.AuditLogger(audit, logger))
	{
		registerBudgets(protected, h)
		registerCostItems(protected, h)
		registerChangeOrders(protected, h)
		registerReports(protected, h)
		protected.GET("/audit-logs", h.Audit.List)
	}

	return engine
}

func readiness(db *gorm.DB, rdb *redis.Client, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("readiness: get sql db", "error", err)
			c.JSON(http.StatusServiceUnavailable, util.Response{Code: constants.CodeInternal, Message: "database unavailable", Data: nil})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			logger.Error("readiness: ping database", "error", err)
			c.JSON(http.StatusServiceUnavailable, util.Response{Code: constants.CodeInternal, Message: "database unavailable", Data: nil})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Error("readiness: ping redis", "error", err)
			c.JSON(http.StatusServiceUnavailable, util.Response{Code: constants.CodeInternal, Message: "redis unavailable", Data: nil})
			return
		}
		c.JSON(http.StatusOK, util.Response{Code: constants.CodeSuccess, Message: constants.MsgOK, Data: gin.H{"status": "ok", "service": "costguard"}})
	}
}
