package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/util"
)

// ChangeOrderService manages change orders with auto after-amount.
type ChangeOrderService struct {
	orders *repository.ChangeOrderRepository
	audit  *AuditLogService
	logger *slog.Logger
}

// NewChangeOrderService builds a ChangeOrderService.
func NewChangeOrderService(orders *repository.ChangeOrderRepository, audit *AuditLogService, logger *slog.Logger) *ChangeOrderService {
	return &ChangeOrderService{orders: orders, audit: audit, logger: logger}
}

// List returns change orders.
func (s *ChangeOrderService) List(projectID uint) ([]model.ChangeOrder, error) {
	list, err := s.orders.List(projectID)
	if err != nil {
		return nil, fmt.Errorf("list change orders: %w", err)
	}
	return list, nil
}

// Get loads a change order.
func (s *ChangeOrderService) Get(id uint) (*model.ChangeOrder, error) {
	return s.orders.FindByID(id)
}

// Create creates a draft change order.
func (s *ChangeOrderService) Create(req dto.CreateChangeOrderRequest, userID uint, userName string) (*model.ChangeOrder, error) {
	if !constants.ValidChangeType(req.ChangeType) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的变更类型")
	}
	order := &model.ChangeOrder{
		ProjectID:      req.ProjectID,
		ChangeType:     req.ChangeType,
		Description:    req.Description,
		OriginalAmount: req.OriginalAmount,
		ChangeAmount:   req.ChangeAmount,
		AfterAmount:    util.AfterAmount(req.OriginalAmount, req.ChangeAmount),
		Reason:         req.Reason,
		ApprovalStatus: constants.ChangeOrderStatusDraft,
		ApplicantID:    userID,
		AppliedAt:      time.Now(),
	}
	if err := s.orders.Create(order); err != nil {
		return nil, fmt.Errorf("create change order: %w", err)
	}
	s.audit.Record(userID, userName, "change_order.create", "change_order", order.ID, fmt.Sprintf("发起变更 %s %s", order.ChangeType, order.Description))
	return order, nil
}

// Update edits a draft change order.
func (s *ChangeOrderService) Update(id uint, req dto.UpdateChangeOrderRequest, userID uint, userName string) (*model.ChangeOrder, error) {
	if !constants.ValidChangeType(req.ChangeType) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的变更类型")
	}
	order, err := s.orders.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.ApprovalStatus != constants.ChangeOrderStatusDraft {
		return nil, constants.NewAppError(constants.CodeConflict, "仅草稿状态的变更单可编辑")
	}
	order.ChangeType = req.ChangeType
	order.Description = req.Description
	order.OriginalAmount = req.OriginalAmount
	order.ChangeAmount = req.ChangeAmount
	order.AfterAmount = util.AfterAmount(req.OriginalAmount, req.ChangeAmount)
	order.Reason = req.Reason
	if err := s.orders.Update(order); err != nil {
		return nil, fmt.Errorf("update change order: %w", err)
	}
	s.audit.Record(userID, userName, "change_order.update", "change_order", order.ID, "编辑变更单")
	return order, nil
}

// Submit submits a draft change order for approval.
func (s *ChangeOrderService) Submit(id uint, userID uint, userName string) (*model.ChangeOrder, error) {
	order, err := s.orders.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.ApprovalStatus != constants.ChangeOrderStatusDraft {
		return nil, constants.NewAppError(constants.CodeConflict, "仅草稿状态的变更单可提交")
	}
	order.ApprovalStatus = constants.ChangeOrderStatusSubmitted
	if err := s.orders.Update(order); err != nil {
		return nil, fmt.Errorf("submit change order: %w", err)
	}
	s.audit.Record(userID, userName, "change_order.submit", "change_order", order.ID, "提交变更审批")
	return order, nil
}

// Approve approves a submitted change order.
func (s *ChangeOrderService) Approve(id uint, approverID uint, approverName string) (*model.ChangeOrder, error) {
	order, err := s.orders.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.ApprovalStatus != constants.ChangeOrderStatusSubmitted {
		return nil, constants.NewAppError(constants.CodeConflict, "仅已提交的变更单可审批")
	}
	now := time.Now()
	order.ApprovalStatus = constants.ChangeOrderStatusApproved
	order.ApproverID = approverID
	order.ApprovedAt = &now
	if err := s.orders.Update(order); err != nil {
		return nil, fmt.Errorf("approve change order: %w", err)
	}
	s.audit.Record(approverID, approverName, "change_order.approve", "change_order", order.ID, "审批通过变更单")
	return order, nil
}

// Reject rejects a submitted change order.
func (s *ChangeOrderService) Reject(id uint, approverID uint, approverName string) (*model.ChangeOrder, error) {
	order, err := s.orders.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.ApprovalStatus != constants.ChangeOrderStatusSubmitted {
		return nil, constants.NewAppError(constants.CodeConflict, "仅已提交的变更单可驳回")
	}
	order.ApprovalStatus = constants.ChangeOrderStatusRejected
	order.ApproverID = approverID
	if err := s.orders.Update(order); err != nil {
		return nil, fmt.Errorf("reject change order: %w", err)
	}
	s.audit.Record(approverID, approverName, "change_order.reject", "change_order", order.ID, "驳回变更单")
	return order, nil
}

// Cancel cancels a change order.
func (s *ChangeOrderService) Cancel(id uint, userID uint, userName string) (*model.ChangeOrder, error) {
	order, err := s.orders.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.ApprovalStatus == constants.ChangeOrderStatusApproved || order.ApprovalStatus == constants.ChangeOrderStatusCancelled {
		return nil, constants.NewAppError(constants.CodeConflict, "该变更单不可作废")
	}
	order.ApprovalStatus = constants.ChangeOrderStatusCancelled
	if err := s.orders.Update(order); err != nil {
		return nil, fmt.Errorf("cancel change order: %w", err)
	}
	s.audit.Record(userID, userName, "change_order.cancel", "change_order", order.ID, "作废变更单")
	return order, nil
}
