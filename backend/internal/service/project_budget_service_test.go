package service

import (
	"log/slog"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
)

func newBudgetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ProjectBudget{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestBudgetApprovalFlow(t *testing.T) {
	db := newBudgetTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	budgetRepo := repository.NewProjectBudgetRepository(db)
	itemRepo := repository.NewCostItemRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)
	auditSvc := NewAuditLogService(auditRepo, logger)
	svc := NewProjectBudgetService(budgetRepo, itemRepo, auditSvc, logger)

	created, err := svc.Create(dtoCreateBudget(), 1, "测试员")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ApprovalStatus != "Draft" {
		t.Fatalf("initial status = %s, want Draft", created.ApprovalStatus)
	}

	submitted, err := svc.Submit(created.ID, 1, "测试员")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if submitted.ApprovalStatus != "Submitted" {
		t.Fatalf("status after submit = %s, want Submitted", submitted.ApprovalStatus)
	}

	approved, err := svc.Approve(created.ID, 2, "财务经理")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.ApprovalStatus != "Approved" || approved.ApproverID != 2 {
		t.Fatalf("status after approve = %s, approver %d", approved.ApprovalStatus, approved.ApproverID)
	}

	// Re-approve should fail (not submitted).
	if _, err := svc.Approve(created.ID, 2, "财务经理"); err == nil {
		t.Fatal("approving an approved budget should fail")
	}
}
