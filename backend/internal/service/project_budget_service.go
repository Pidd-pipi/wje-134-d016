package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
)

// ProjectBudgetService manages project budgets and their approvals.
type ProjectBudgetService struct {
	budgets *repository.ProjectBudgetRepository
	items   *repository.CostItemRepository
	audit   *AuditLogService
	logger  *slog.Logger
}

// NewProjectBudgetService builds a ProjectBudgetService.
func NewProjectBudgetService(budgets *repository.ProjectBudgetRepository, items *repository.CostItemRepository, audit *AuditLogService, logger *slog.Logger) *ProjectBudgetService {
	return &ProjectBudgetService{budgets: budgets, items: items, audit: audit, logger: logger}
}

// List returns budgets.
func (s *ProjectBudgetService) List(projectID uint) ([]model.ProjectBudget, error) {
	list, err := s.budgets.List(projectID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	return list, nil
}

// Get loads a budget.
func (s *ProjectBudgetService) Get(id uint) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Create creates a draft budget.
func (s *ProjectBudgetService) Create(req dto.CreateBudgetRequest, userID uint, userName string) (*model.ProjectBudget, error) {
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}
	budget := &model.ProjectBudget{
		ProjectID:      req.ProjectID,
		Name:           req.Name,
		TotalAmount:    req.TotalAmount,
		ReservedAmount: req.ReservedAmount,
		Currency:       currency,
		ApprovalStatus: constants.BudgetStatusDraft,
		Remarks:        req.Remarks,
	}
	if err := s.budgets.Create(budget); err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}
	s.audit.Record(userID, userName, "budget.create", "budget", budget.ID, fmt.Sprintf("编制预算 %s %.2f", budget.Name, budget.TotalAmount))
	return budget, nil
}

// Update edits a draft budget.
func (s *ProjectBudgetService) Update(id uint, req dto.UpdateBudgetRequest, userID uint, userName string) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus != constants.BudgetStatusDraft {
		return nil, constants.NewAppError(constants.CodeConflict, "仅草稿状态的预算可编辑")
	}
	b.Name = req.Name
	b.TotalAmount = req.TotalAmount
	b.ReservedAmount = req.ReservedAmount
	b.Remarks = req.Remarks
	if err := s.budgets.Update(b); err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}
	s.audit.Record(userID, userName, "budget.update", "budget", b.ID, "编辑预算")
	return b, nil
}

// Submit moves a draft budget to Submitted.
func (s *ProjectBudgetService) Submit(id uint, userID uint, userName string) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus != constants.BudgetStatusDraft {
		return nil, constants.NewAppError(constants.CodeConflict, "仅草稿状态的预算可提交审批")
	}
	b.ApprovalStatus = constants.BudgetStatusSubmitted
	if err := s.budgets.Update(b); err != nil {
		return nil, fmt.Errorf("submit budget: %w", err)
	}
	s.audit.Record(userID, userName, "budget.submit", "budget", b.ID, "提交预算审批")
	return b, nil
}

// Approve approves a submitted budget (FinanceManager/Admin only).
func (s *ProjectBudgetService) Approve(id uint, approverID uint, approverName string) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus != constants.BudgetStatusSubmitted {
		return nil, constants.NewAppError(constants.CodeConflict, "仅已提交的预算可审批")
	}
	now := time.Now()
	b.ApprovalStatus = constants.BudgetStatusApproved
	b.ApproverID = approverID
	b.ApprovedAt = &now
	if err := s.budgets.Update(b); err != nil {
		return nil, fmt.Errorf("approve budget: %w", err)
	}
	s.audit.Record(approverID, approverName, "budget.approve", "budget", b.ID, "审批通过预算")
	return b, nil
}

// Reject rejects a submitted budget.
func (s *ProjectBudgetService) Reject(id uint, approverID uint, approverName string) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus != constants.BudgetStatusSubmitted {
		return nil, constants.NewAppError(constants.CodeConflict, "仅已提交的预算可驳回")
	}
	b.ApprovalStatus = constants.BudgetStatusRejected
	b.ApproverID = approverID
	if err := s.budgets.Update(b); err != nil {
		return nil, fmt.Errorf("reject budget: %w", err)
	}
	s.audit.Record(approverID, approverName, "budget.reject", "budget", b.ID, "驳回预算")
	return b, nil
}

// Close finalizes an approved budget at project settlement (FinanceManager only).
// It refuses to close while abnormal costs remain under the budget.
func (s *ProjectBudgetService) Close(id uint, userID uint, userName string) (*model.ProjectBudget, error) {
	b, err := s.budgets.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus == constants.BudgetStatusClosed {
		return nil, constants.NewAppError(constants.CodeConflict, "预算已封账，请勿重复封账")
	}
	if b.ApprovalStatus != constants.BudgetStatusApproved {
		return nil, constants.NewAppError(constants.CodeConflict, "仅已审批通过的预算可封账")
	}
	abnormal, err := s.items.ListAbnormalByBudget(id)
	if err != nil {
		return nil, fmt.Errorf("close budget: %w", err)
	}
	if len(abnormal) > 0 {
		blocked := dto.CloseBudgetBlockedResponse{AbnormalItems: make([]dto.AbnormalCostItem, 0, len(abnormal))}
		for _, it := range abnormal {
			blocked.AbnormalItems = append(blocked.AbnormalItems, dto.AbnormalCostItem{
				ID:             it.ID,
				VoucherNo:      it.VoucherNo,
				Name:           it.Name,
				Category:       it.Category,
				BudgetAmount:   it.BudgetAmount,
				ActualAmount:   it.ActualAmount,
				VarianceAmount: it.VarianceAmount,
			})
		}
		return nil, constants.NewAppErrorWithData(constants.CodeConflict,
			fmt.Sprintf("存在 %d 笔异常成本，不允许封账", len(abnormal)), blocked)
	}
	now := time.Now()
	b.ApprovalStatus = constants.BudgetStatusClosed
	b.CloserID = userID
	b.ClosedAt = &now
	if err := s.budgets.Update(b); err != nil {
		return nil, fmt.Errorf("close budget: %w", err)
	}
	s.audit.Record(userID, userName, "budget.close", "budget", b.ID, "预算封账")
	return b, nil
}
