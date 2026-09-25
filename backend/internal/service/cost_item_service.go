package service

import (
	"fmt"
	"log/slog"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/util"
)

// CostItemService manages cost items with automatic variance calculation.
type CostItemService struct {
	items   *repository.CostItemRepository
	budgets *repository.ProjectBudgetRepository
	audit   *AuditLogService
	logger  *slog.Logger
}

// NewCostItemService builds a CostItemService.
func NewCostItemService(items *repository.CostItemRepository, budgets *repository.ProjectBudgetRepository, audit *AuditLogService, logger *slog.Logger) *CostItemService {
	return &CostItemService{items: items, budgets: budgets, audit: audit, logger: logger}
}

// List returns cost items.
func (s *CostItemService) List(budgetID uint) ([]model.CostItem, error) {
	list, err := s.items.List(budgetID)
	if err != nil {
		return nil, fmt.Errorf("list cost items: %w", err)
	}
	return list, nil
}

// Get loads a cost item.
func (s *CostItemService) Get(id uint) (*model.CostItem, error) {
	return s.items.FindByID(id)
}

// Create records a cost item and auto-computes variance.
func (s *CostItemService) Create(req dto.CreateCostItemRequest, userID uint, userName string) (*model.CostItem, error) {
	if !constants.ValidCostCategory(req.Category) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的成本类别")
	}
	if _, err := s.budgets.FindByID(req.BudgetID); err != nil {
		return nil, err
	}
	item := &model.CostItem{
		BudgetID:       req.BudgetID,
		Category:       req.Category,
		Name:           req.Name,
		BudgetAmount:   req.BudgetAmount,
		ActualAmount:   req.ActualAmount,
		VarianceAmount: util.Variance(req.BudgetAmount, req.ActualAmount),
		OccurrenceDate: dto.ParseDate(req.OccurrenceDate),
		VoucherNo:      req.VoucherNo,
		MaterialUsageID: req.MaterialUsageID,
		TimesheetID:    req.TimesheetID,
	}
	if item.VarianceAmount > 0 {
		item.IsAbnormal = true
	}
	if err := s.items.Create(item); err != nil {
		return nil, fmt.Errorf("create cost item: %w", err)
	}
	s.audit.Record(userID, userName, "cost_item.create", "cost_item", item.ID, fmt.Sprintf("录入成本 %s %s %.2f", item.Category, item.Name, item.ActualAmount))
	return item, nil
}

// Update edits a cost item and recomputes variance.
func (s *CostItemService) Update(id uint, req dto.UpdateCostItemRequest, userID uint, userName string) (*model.CostItem, error) {
	if !constants.ValidCostCategory(req.Category) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的成本类别")
	}
	item, err := s.items.FindByID(id)
	if err != nil {
		return nil, err
	}
	item.Category = req.Category
	item.Name = req.Name
	item.BudgetAmount = req.BudgetAmount
	item.ActualAmount = req.ActualAmount
	item.VarianceAmount = util.Variance(req.BudgetAmount, req.ActualAmount)
	item.VoucherNo = req.VoucherNo
	if req.OccurrenceDate != "" {
		item.OccurrenceDate = dto.ParseDate(req.OccurrenceDate)
	}
	if item.VarianceAmount > 0 {
		item.IsAbnormal = true
	}
	if err := s.items.Update(item); err != nil {
		return nil, fmt.Errorf("update cost item: %w", err)
	}
	s.audit.Record(userID, userName, "cost_item.update", "cost_item", item.ID, "核对成本")
	return item, nil
}

// MarkAbnormal flags a cost item as abnormal.
func (s *CostItemService) MarkAbnormal(id uint, userID uint, userName string) (*model.CostItem, error) {
	item, err := s.items.FindByID(id)
	if err != nil {
		return nil, err
	}
	item.IsAbnormal = true
	if err := s.items.Update(item); err != nil {
		return nil, fmt.Errorf("mark cost item abnormal: %w", err)
	}
	s.audit.Record(userID, userName, "cost_item.abnormal", "cost_item", item.ID, "标记异常成本")
	return item, nil
}
