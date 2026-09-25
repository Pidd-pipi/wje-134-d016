package service

import (
	"strings"
	"testing"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/repository"
)

func newCloseTestServices(t *testing.T) (*ProjectBudgetService, *CostItemService, *AuditLogService) {
	t.Helper()
	db := newServiceDB(t)
	audit := newAuditService(t, db)
	budgetRepo := repository.NewProjectBudgetRepository(db)
	itemRepo := repository.NewCostItemRepository(db)
	budgetSvc := NewProjectBudgetService(budgetRepo, itemRepo, audit, discardLogger())
	itemSvc := NewCostItemService(itemRepo, budgetRepo, audit, discardLogger())
	return budgetSvc, itemSvc, audit
}

func approveBudget(t *testing.T, svc *ProjectBudgetService, id uint) {
	t.Helper()
	if _, err := svc.Submit(id, 1, "成本会计"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := svc.Approve(id, 2, "财务经理"); err != nil {
		t.Fatalf("approve: %v", err)
	}
}

func TestBudgetCloseBlockedByAbnormalCostThenSucceeds(t *testing.T) {
	budgetSvc, itemSvc, _ := newCloseTestServices(t)

	budget, err := budgetSvc.Create(dtoCreateBudget(), 1, "成本会计")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	approveBudget(t, budgetSvc, budget.ID)

	// 录入一笔超支成本，自动标记异常。
	item, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budget.ID,
		Category:     constants.CostCategoryMaterial,
		Name:         "钢材超支",
		BudgetAmount: 100,
		ActualAmount: 150,
		VoucherNo:    "PZ-001",
	}, 1, "成本会计")
	if err != nil {
		t.Fatalf("create cost item: %v", err)
	}
	if !item.IsAbnormal {
		t.Fatal("over-budget item should be abnormal")
	}

	// 有异常成本时不允许封账，且返回单号与超支情况。
	_, abnormalities, err := budgetSvc.Close(budget.ID, 2, "财务经理")
	if err == nil {
		t.Fatal("close with abnormal cost should fail")
	}
	if len(abnormalities) != 1 {
		t.Fatalf("abnormalities = %d, want 1", len(abnormalities))
	}
	ab := abnormalities[0]
	if ab.ID != item.ID || ab.VoucherNo != "PZ-001" {
		t.Fatalf("abnormal info = %+v, want id %d voucher PZ-001", ab, item.ID)
	}
	if ab.VarianceAmount != 50 || ab.BudgetAmount != 100 || ab.ActualAmount != 150 {
		t.Fatalf("abnormal amounts = %+v, want variance 50 of 100/150", ab)
	}

	// 成本会计把金额核回预算内，异常标记自行解除。
	updated, err := itemSvc.Update(item.ID, dto.UpdateCostItemRequest{
		Category:     constants.CostCategoryMaterial,
		Name:         "钢材超支",
		BudgetAmount: 100,
		ActualAmount: 90,
		VoucherNo:    "PZ-001",
	}, 1, "成本会计")
	if err != nil {
		t.Fatalf("update cost item: %v", err)
	}
	if updated.IsAbnormal {
		t.Fatal("abnormal flag should clear once amount is back within budget")
	}

	// 再次封账通过。
	closed, abnormalities, err := budgetSvc.Close(budget.ID, 2, "财务经理")
	if err != nil {
		t.Fatalf("close after fixing costs: %v", err)
	}
	if len(abnormalities) != 0 {
		t.Fatalf("abnormalities after close = %d, want 0", len(abnormalities))
	}
	if closed.ApprovalStatus != constants.BudgetStatusClosed {
		t.Fatalf("status = %s, want Closed", closed.ApprovalStatus)
	}
	if closed.ClosedAt == nil || closed.CloserID != 2 {
		t.Fatalf("closedAt = %v, closerID = %d", closed.ClosedAt, closed.CloserID)
	}
}

func TestBudgetCloseRequiresApprovedStatus(t *testing.T) {
	budgetSvc, _, _ := newCloseTestServices(t)

	budget, err := budgetSvc.Create(dtoCreateBudget(), 1, "成本会计")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	if _, _, err := budgetSvc.Close(budget.ID, 2, "财务经理"); err == nil {
		t.Fatal("closing a draft budget should fail")
	}

	if _, err := budgetSvc.Submit(budget.ID, 1, "成本会计"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, _, err := budgetSvc.Close(budget.ID, 2, "财务经理"); err == nil {
		t.Fatal("closing a submitted (not approved) budget should fail")
	}
}

func TestBudgetCloseRepeatedAndImmutability(t *testing.T) {
	budgetSvc, itemSvc, audit := newCloseTestServices(t)

	budget, err := budgetSvc.Create(dtoCreateBudget(), 1, "成本会计")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	approveBudget(t, budgetSvc, budget.ID)

	item, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budget.ID,
		Category:     constants.CostCategoryLabor,
		Name:         "人工费",
		BudgetAmount: 100,
		ActualAmount: 80,
	}, 1, "成本会计")
	if err != nil {
		t.Fatalf("create cost item: %v", err)
	}

	if _, _, err := budgetSvc.Close(budget.ID, 2, "财务经理"); err != nil {
		t.Fatalf("close: %v", err)
	}

	// 重复封账要有明确提示。
	if _, _, err := budgetSvc.Close(budget.ID, 2, "财务经理"); err == nil ||
		!strings.Contains(err.Error(), "已封账") {
		t.Fatalf("repeated close error = %v, want 已封账 hint", err)
	}

	// 封账后不再接收新成本。
	if _, err := itemSvc.Create(dto.CreateCostItemRequest{
		BudgetID:     budget.ID,
		Category:     constants.CostCategoryLabor,
		Name:         "新增成本",
		BudgetAmount: 10,
		ActualAmount: 10,
	}, 1, "成本会计"); err == nil {
		t.Fatal("creating cost on closed budget should fail")
	}

	// 封账后已有明细不能变更。
	if _, err := itemSvc.Update(item.ID, dto.UpdateCostItemRequest{
		Category:     constants.CostCategoryLabor,
		Name:         "人工费",
		BudgetAmount: 100,
		ActualAmount: 70,
	}, 1, "成本会计"); err == nil {
		t.Fatal("updating cost on closed budget should fail")
	}
	if _, err := itemSvc.MarkAbnormal(item.ID, 1, "成本会计"); err == nil {
		t.Fatal("marking abnormal on closed budget should fail")
	}

	// 原有明细保留。
	items, err := itemSvc.List(budget.ID)
	if err != nil {
		t.Fatalf("list cost items: %v", err)
	}
	if len(items) != 1 || items[0].ActualAmount != 80 {
		t.Fatalf("items after close = %+v, want original item kept", items)
	}

	// 审计记录保留且包含封账动作。
	logs, err := audit.List(100)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	found := false
	for _, l := range logs {
		if l.Action == "budget.close" && l.EntityID == budget.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("audit log should contain budget.close entry")
	}
}

func TestCanCloseBudgetRoles(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{constants.RoleFinanceManager, true},
		{constants.RoleAdmin, true},
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
