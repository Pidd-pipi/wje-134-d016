package service

import (
	"testing"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/repository"
)

func TestChangeOrderServiceApprovalFlow(t *testing.T) {
	db := newServiceDB(t)
	audit := newAuditService(t, db)
	svc := NewChangeOrderService(repository.NewChangeOrderRepository(db), audit, discardLogger())

	created, err := svc.Create(dto.CreateChangeOrderRequest{
		ProjectID:      1001,
		ChangeType:     constants.ChangeTypeDesignChange,
		Description:    "外立面设计变更",
		OriginalAmount: 1000000,
		ChangeAmount:   200000,
	}, 1, "项目经理")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.AfterAmount != 1200000 {
		t.Fatalf("AfterAmount = %v, want 1200000", created.AfterAmount)
	}

	tests := []struct {
		name       string
		action     func() (string, error)
		wantStatus string
	}{
		{
			name: "submit draft",
			action: func() (string, error) {
				order, err := svc.Submit(created.ID, 1, "项目经理")
				if err != nil {
					return "", err
				}
				return order.ApprovalStatus, nil
			},
			wantStatus: constants.ChangeOrderStatusSubmitted,
		},
		{
			name: "approve submitted",
			action: func() (string, error) {
				order, err := svc.Approve(created.ID, 2, "财务经理")
				if err != nil {
					return "", err
				}
				return order.ApprovalStatus, nil
			},
			wantStatus: constants.ChangeOrderStatusApproved,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := tt.action()
			if err != nil {
				t.Fatalf("action error = %v", err)
			}
			if status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}

	if _, err := svc.Cancel(created.ID, 1, "项目经理"); err == nil {
		t.Fatal("approved change order should not be cancellable")
	}
}
