package service

import (
	"testing"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/repository"
)

func TestCostItemServiceCreateValidationAndVariance(t *testing.T) {
	db := newServiceDB(t)
	audit := newAuditService(t, db)
	budgetRepo := repository.NewProjectBudgetRepository(db)
	itemRepo := repository.NewCostItemRepository(db)
	itemSvc := NewCostItemService(itemRepo, budgetRepo, audit, discardLogger())

	budget, err := NewProjectBudgetService(budgetRepo, itemRepo, audit, discardLogger()).Create(dtoCreateBudget(), 1, "tester")
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}

	tests := []struct {
		name         string
		category     string
		budgetAmt    float64
		actualAmt    float64
		wantAbnormal bool
		wantErr      bool
	}{
		{name: "valid material category", category: constants.CostCategoryMaterial, budgetAmt: 100, actualAmt: 90},
		{name: "valid labor over budget", category: constants.CostCategoryLabor, budgetAmt: 100, actualAmt: 130, wantAbnormal: true},
		{name: "invalid category", category: "Bad", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := itemSvc.Create(dto.CreateCostItemRequest{
				BudgetID:     budget.ID,
				Category:     tt.category,
				Name:         "测试成本项",
				BudgetAmount: tt.budgetAmt,
				ActualAmount: tt.actualAmt,
			}, 1, "tester")
			if tt.wantErr {
				if err == nil {
					t.Fatal("Create() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if item.IsAbnormal != tt.wantAbnormal {
				t.Fatalf("IsAbnormal = %v, want %v", item.IsAbnormal, tt.wantAbnormal)
			}
			if item.VarianceAmount != tt.actualAmt-tt.budgetAmt {
				t.Fatalf("VarianceAmount = %v, want %v", item.VarianceAmount, tt.actualAmt-tt.budgetAmt)
			}
		})
	}
}
