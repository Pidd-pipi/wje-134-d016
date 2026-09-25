package service

import (
	"io"
	"log/slog"
	"testing"

	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.ProjectBudget{},
		&model.CostItem{},
		&model.ChangeOrder{},
		&model.CostReport{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newAuditService(t *testing.T, db *gorm.DB) *AuditLogService {
	t.Helper()
	return NewAuditLogService(repository.NewAuditLogRepository(db), discardLogger())
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
