package service

import (
	"errors"
	"testing"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/repository"
)

func newBudgetCloseFixture(t *testing.T) (*ProjectBudgetService, *CostItemService, *AuditLogService) {
	t.Helper()
	db := newServiceDB(t)
	audit := newAuditService(t, db)
	budgetRepo := repository.NewProjectBudgetRepository(db)
	itemRepo := repository.NewCostItemRepository(db)
	budgetSvc := NewProjectBudgetService(budgetRepo, itemRepo, audit, discardLogger())
	itemSvc := NewCostItemService(itemRepo, budgetRepo, audit, discardLogger())
	return budgetSvc, itemSvc, audit
}

func approveBudget(t *testing.T, svc *ProjectBudgetService) uint {
	t.Helper()
	b, err := svc.Create(dtoCreateBudget(), 1, "会计")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	if _, err := svc.Submit(b.ID, 1, "会计"); err != nil {
		t.Fatalf("submit budget: %v", err)
	}
	if _, err := svc.Approve(b.ID, 2, "财务经理"); err != nil {
		t.Fatalf("approve budget: %v", err)
	}
	return b.ID
}

func TestBudgetCloseRequiresApproved(t *testing.T) {
	budgetSvc, _, _ := newBudgetCloseFixture(t)

	b, err := budgetSvc.Create(dtoCreateBudget(), 1, "会计")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	if _, err := budgetSvc.Close(b.ID, 2, "财务经理"); err == nil {
		t.Fatal("closing a draft budget should fail")
	}
}

func TestBudgetCloseBlockedByAbnormalCosts(t *testing.T) {
	budgetSvc, itemSvc, audit := newBudgetCloseFixture(t)
	budgetID := approveBudget(t, budgetSvc)

	// Accountant records an over-budget cost; it becomes abnormal automatically.
	item, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budgetID,
		Category:     constants.CostCategoryMaterial,
		Name:         "钢材采购",
		BudgetAmount: 100,
		ActualAmount: 130,
		VoucherNo:    "PZ-2026-001",
	}, 3, "成本会计")
	if err != nil {
		t.Fatalf("create cost item: %v", err)
	}
	if !item.IsAbnormal {
		t.Fatal("over-budget item should be abnormal")
	}

	// Close must be refused and must list the voucher number and overrun.
	_, err = budgetSvc.Close(budgetID, 2, "财务经理")
	if err == nil {
		t.Fatal("closing with abnormal costs should fail")
	}
	var appErr *constants.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *AppError", err)
	}
	if appErr.Code != constants.CodeConflict {
		t.Fatalf("error code = %d, want %d", appErr.Code, constants.CodeConflict)
	}
	blocked, ok := appErr.Data.(dto.CloseBudgetBlockedResponse)
	if !ok {
		t.Fatalf("error data type = %T, want CloseBudgetBlockedResponse", appErr.Data)
	}
	if len(blocked.AbnormalItems) != 1 {
		t.Fatalf("abnormal items = %d, want 1", len(blocked.AbnormalItems))
	}
	ab := blocked.AbnormalItems[0]
	if ab.ID != item.ID || ab.VoucherNo != "PZ-2026-001" {
		t.Fatalf("abnormal item = %+v, want id %d voucher PZ-2026-001", ab, item.ID)
	}
	if ab.VarianceAmount != 30 {
		t.Fatalf("variance = %v, want 30", ab.VarianceAmount)
	}

	// Accountant brings the amount back within budget; the flag clears itself.
	updated, err := itemSvc.Update(item.ID, dto.UpdateCostItemRequest{
		Category:     constants.CostCategoryMaterial,
		Name:         "钢材采购",
		BudgetAmount: 100,
		ActualAmount: 95,
		VoucherNo:    "PZ-2026-001",
	}, 3, "成本会计")
	if err != nil {
		t.Fatalf("update cost item: %v", err)
	}
	if updated.IsAbnormal {
		t.Fatal("abnormal flag should clear once back within budget")
	}

	// The finance manager can now close.
	closed, err := budgetSvc.Close(budgetID, 2, "财务经理")
	if err != nil {
		t.Fatalf("close after fix: %v", err)
	}
	if closed.ApprovalStatus != constants.BudgetStatusClosed {
		t.Fatalf("status = %s, want Closed", closed.ApprovalStatus)
	}
	if closed.CloserID != 2 || closed.ClosedAt == nil {
		t.Fatalf("closer = %d, closedAt = %v", closed.CloserID, closed.ClosedAt)
	}

	// Original details and audit records are preserved.
	kept, err := itemSvc.Get(item.ID)
	if err != nil {
		t.Fatalf("cost item should be preserved: %v", err)
	}
	if kept.VoucherNo != "PZ-2026-001" || kept.ActualAmount != 95 {
		t.Fatalf("preserved item = %+v", kept)
	}
	logs, err := audit.List(100)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	var sawCreate, sawUpdate, sawClose bool
	for _, l := range logs {
		switch l.Action {
		case "cost_item.create":
			sawCreate = true
		case "cost_item.update":
			sawUpdate = true
		case "budget.close":
			sawClose = true
		}
	}
	if !sawCreate || !sawUpdate || !sawClose {
		t.Fatalf("audit trail incomplete: create=%v update=%v close=%v", sawCreate, sawUpdate, sawClose)
	}
}

func TestBudgetCloseRepeatAndFrozenDetails(t *testing.T) {
	budgetSvc, itemSvc, _ := newBudgetCloseFixture(t)
	budgetID := approveBudget(t, budgetSvc)

	item, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budgetID,
		Category:     constants.CostCategoryLabor,
		Name:         "木工工时",
		BudgetAmount: 100,
		ActualAmount: 90,
	}, 3, "成本会计")
	if err != nil {
		t.Fatalf("create cost item: %v", err)
	}

	if _, err := budgetSvc.Close(budgetID, 2, "财务经理"); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Repeated close gets a clear prompt.
	if _, err := budgetSvc.Close(budgetID, 2, "财务经理"); err == nil {
		t.Fatal("repeated close should fail")
	} else {
		var appErr *constants.AppError
		if !errors.As(err, &appErr) || appErr.Message != "预算已封账，请勿重复封账" {
			t.Fatalf("repeated close error = %v", err)
		}
	}

	// No new costs after closing.
	if _, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budgetID,
		Category:     constants.CostCategoryLabor,
		Name:         "新增工时",
		BudgetAmount: 10,
		ActualAmount: 10,
	}, 3, "成本会计"); err == nil {
		t.Fatal("creating cost item on closed budget should fail")
	}

	// Existing details cannot change after closing.
	if _, err := itemSvc.Update(item.ID, dto.UpdateCostItemRequest{
		Category:     constants.CostCategoryLabor,
		Name:         "木工工时",
		BudgetAmount: 100,
		ActualAmount: 80,
	}, 3, "成本会计"); err == nil {
		t.Fatal("updating cost item on closed budget should fail")
	}
	if _, err := itemSvc.MarkAbnormal(item.ID, 3, "成本会计"); err == nil {
		t.Fatal("marking abnormal on closed budget should fail")
	}

	// The stored detail is untouched.
	kept, err := itemSvc.Get(item.ID)
	if err != nil {
		t.Fatalf("get cost item: %v", err)
	}
	if kept.ActualAmount != 90 {
		t.Fatalf("actual amount = %v, want 90 (unchanged)", kept.ActualAmount)
	}
}

func TestCanCloseBudget(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{constants.RoleFinanceManager, true},
		{constants.RoleAdmin, false},
		{constants.RoleAccountant, false},
		{constants.RoleProjectManager, false},
		{constants.RoleViewer, false},
	}
	for _, tt := range tests {
		if got := constants.CanCloseBudget(tt.role); got != tt.want {
			t.Errorf("CanCloseBudget(%s) = %v, want %v", tt.role, got, tt.want)
		}
	}
}
