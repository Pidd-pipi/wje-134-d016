// Command server is the entrypoint of the cost control API service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/costguard/costguard/internal/config"
	"github.com/costguard/costguard/internal/handler"
	"github.com/costguard/costguard/internal/logger"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/router"
	"github.com/costguard/costguard/internal/service"
)

func main() {
	logger := logger.New(os.Getenv("APP_ENV"))
	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := openDB(cfg, logger)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ProjectBudget{},
		&model.CostItem{},
		&model.ChangeOrder{},
		&model.CostReport{},
		&model.AuditLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer func() { _ = rdb.Close() }()
	if err := pingRedis(cfg, rdb, logger); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	seedSvc := service.NewSeedService(db, logger)
	if err := seedSvc.EnsureSeedData(context.Background()); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}

	h, users, auditSvc := buildHandlers(cfg, db, rdb, logger)

	engine := router.New(cfg, logger, h, users, rdb, auditSvc, db)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "port", cfg.ServerPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		logger.Info("shutting down", "signal", sig.String())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

func openDB(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	retries := cfg.DBConnectRetries
	if retries < 0 {
		retries = 0
	}
	for attempt := 0; ; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				if pingErr = sqlDB.Ping(); pingErr == nil {
					break
				}
			}
			err = pingErr
		}
		if attempt >= retries {
			return nil, fmt.Errorf("connect postgres after %d retries: %w", retries, err)
		}
		logger.Warn("database not ready, retrying", "attempt", attempt+1, "error", err)
		time.Sleep(time.Duration(cfg.DBConnectRetryIntervalSec) * time.Second)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute)
	return db, nil
}

func pingRedis(cfg *config.Config, rdb *redis.Client, logger *slog.Logger) error {
	var err error
	retries := cfg.RedisConnectRetries
	if retries < 0 {
		retries = 0
	}
	for attempt := 0; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = rdb.Ping(ctx).Err()
		cancel()
		if err == nil {
			return nil
		}
		if attempt >= retries {
			return err
		}
		logger.Warn("redis not ready, retrying", "attempt", attempt+1, "error", err)
		time.Sleep(time.Duration(cfg.RedisConnectRetryInterval) * time.Second)
	}
}

func buildHandlers(cfg *config.Config, db *gorm.DB, rdb *redis.Client, logger *slog.Logger) (*router.Handlers, *repository.UserRepository, *service.AuditLogService) {
	userRepo := repository.NewUserRepository(db)
	budgetRepo := repository.NewProjectBudgetRepository(db)
	itemRepo := repository.NewCostItemRepository(db)
	orderRepo := repository.NewChangeOrderRepository(db)
	reportRepo := repository.NewCostReportRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)

	auditSvc := service.NewAuditLogService(auditRepo, logger)
	authSvc := service.NewAuthService(cfg, userRepo, logger)
	budgetSvc := service.NewProjectBudgetService(budgetRepo, itemRepo, auditSvc, logger)
	itemSvc := service.NewCostItemService(itemRepo, budgetRepo, auditSvc, logger)
	orderSvc := service.NewChangeOrderService(orderRepo, auditSvc, logger)
	analyticsSvc := service.NewAnalyticsService(itemRepo, budgetRepo, reportRepo, rdb, auditSvc, logger)

	return &router.Handlers{
		Auth:        handler.NewAuthHandler(authSvc, logger),
		Budget:      handler.NewProjectBudgetHandler(budgetSvc, logger),
		CostItem:    handler.NewCostItemHandler(itemSvc, logger),
		ChangeOrder: handler.NewChangeOrderHandler(orderSvc, logger),
		CostReport:  handler.NewCostReportHandler(analyticsSvc, logger),
		Audit:       handler.NewAuditHandler(auditSvc, logger),
	}, userRepo, auditSvc
}
