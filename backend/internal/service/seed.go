package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/model"
)

// SeedService idempotently inserts demo users, budgets, cost items and change
// orders on first boot.
type SeedService struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSeedService builds a SeedService.
func NewSeedService(db *gorm.DB, logger *slog.Logger) *SeedService {
	return &SeedService{db: db, logger: logger}
}

// EnsureSeedData creates the demo dataset if the users table is empty.
func (s *SeedService) EnsureSeedData(ctx context.Context) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("seed: count users: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash := func(pwd string) string {
		b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("seed: hash password failed", "error", err)
		}
		return string(b)
	}

	users := []*model.User{
		{Username: "admin", PasswordHash: hash("admin123"), Name: "系统管理员", Role: constants.RoleAdmin},
		{Username: "finance", PasswordHash: hash("finance123"), Name: "财务经理", Role: constants.RoleFinanceManager},
		{Username: "pm", PasswordHash: hash("pm123456"), Name: "项目经理", Role: constants.RoleProjectManager},
		{Username: "accountant", PasswordHash: hash("accountant123"), Name: "成本会计", Role: constants.RoleAccountant},
		{Username: "viewer", PasswordHash: hash("viewer123"), Name: "只读用户", Role: constants.RoleViewer},
	}
	for _, u := range users {
		if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
			return fmt.Errorf("seed: create user: %w", err)
		}
	}

	date := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	ts := func(s string) *time.Time {
		t := date(s)
		return &t
	}

	// Two projects: 1001 (示范市文化中心), 1002 (地铁车辆段).
	budgets := []*model.ProjectBudget{
		{ProjectID: 1001, Name: "主体结构预算", TotalAmount: 56000000, ReservedAmount: 2000000, Currency: "CNY", ApprovalStatus: constants.BudgetStatusApproved, ApproverID: 2, ApprovedAt: ts("2026-03-10"), Remarks: "含钢筋、混凝土、模板"},
		{ProjectID: 1001, Name: "机电安装预算", TotalAmount: 18000000, ReservedAmount: 500000, Currency: "CNY", ApprovalStatus: constants.BudgetStatusSubmitted, Remarks: "待审批"},
		{ProjectID: 1002, Name: "站场轨道预算", TotalAmount: 42000000, ReservedAmount: 1500000, Currency: "CNY", ApprovalStatus: constants.BudgetStatusApproved, ApproverID: 2, ApprovedAt: ts("2026-02-20"), Remarks: "轨道与道床"},
		{ProjectID: 1002, Name: "设备采购预算", TotalAmount: 25000000, ReservedAmount: 800000, Currency: "CNY", ApprovalStatus: constants.BudgetStatusDraft, Remarks: "编制中"},
	}
	for i := range budgets {
		if err := s.db.WithContext(ctx).Create(budgets[i]).Error; err != nil {
			return fmt.Errorf("seed: create budget: %w", err)
		}
	}

	costItems := []*model.CostItem{
		{BudgetID: budgets[0].ID, Category: constants.CostCategoryMaterial, Name: "主体钢筋采购", BudgetAmount: 18500000, ActualAmount: 19200000, VarianceAmount: 700000, IsAbnormal: true, OccurrenceDate: date("2026-08-05"), VoucherNo: "V2026080501"},
		{BudgetID: budgets[0].ID, Category: constants.CostCategoryLabor, Name: "主体结构劳务", BudgetAmount: 9800000, ActualAmount: 9500000, VarianceAmount: -300000, IsAbnormal: false, OccurrenceDate: date("2026-08-10"), VoucherNo: "V2026081002"},
		{BudgetID: budgets[0].ID, Category: constants.CostCategoryEquipment, Name: "塔吊租赁", BudgetAmount: 3600000, ActualAmount: 3700000, VarianceAmount: 100000, IsAbnormal: true, OccurrenceDate: date("2026-08-08"), VoucherNo: "V2026080803"},
		{BudgetID: budgets[2].ID, Category: constants.CostCategoryMaterial, Name: "钢轨采购", BudgetAmount: 16800000, ActualAmount: 17100000, VarianceAmount: 300000, IsAbnormal: true, OccurrenceDate: date("2026-07-28"), VoucherNo: "V2026072804"},
		{BudgetID: budgets[2].ID, Category: constants.CostCategorySubcontract, Name: "道床施工分包", BudgetAmount: 5200000, ActualAmount: 5000000, VarianceAmount: -200000, IsAbnormal: false, OccurrenceDate: date("2026-08-01"), VoucherNo: "V2026080105"},
	}
	for i := range costItems {
		if err := s.db.WithContext(ctx).Create(costItems[i]).Error; err != nil {
			return fmt.Errorf("seed: create cost item: %w", err)
		}
	}

	changeOrders := []*model.ChangeOrder{
		{ProjectID: 1001, ChangeType: constants.ChangeTypeScopeChange, Description: "增加地下二层人防面积", OriginalAmount: 56000000, ChangeAmount: 3800000, AfterAmount: 59800000, Reason: "政府人防要求", ApprovalStatus: constants.ChangeOrderStatusApproved, ApplicantID: 3, ApproverID: 2, AppliedAt: date("2026-06-15"), ApprovedAt: ts("2026-06-20")},
		{ProjectID: 1002, ChangeType: constants.ChangeTypePriceAdjustment, Description: "钢材价格调差", OriginalAmount: 42000000, ChangeAmount: -1200000, AfterAmount: 40800000, Reason: "材料价格回落", ApprovalStatus: constants.ChangeOrderStatusSubmitted, ApplicantID: 3, AppliedAt: date("2026-08-01")},
		{ProjectID: 1001, ChangeType: constants.ChangeTypeDesignChange, Description: "幕墙深化设计变更", OriginalAmount: 18000000, ChangeAmount: 900000, AfterAmount: 18900000, Reason: "设计优化", ApprovalStatus: constants.ChangeOrderStatusDraft, ApplicantID: 3, AppliedAt: date("2026-08-12")},
	}
	for i := range changeOrders {
		if err := s.db.WithContext(ctx).Create(changeOrders[i]).Error; err != nil {
			return fmt.Errorf("seed: create change order: %w", err)
		}
	}

	s.logger.Info("seed data inserted",
		"users", len(users), "budgets", len(budgets),
		"cost_items", len(costItems), "change_orders", len(changeOrders))
	return nil
}
