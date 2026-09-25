package model

import "time"

// ProjectBudget is a budget plan for a construction project.
type ProjectBudget struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ProjectID      uint       `gorm:"index;not null" json:"projectId"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	TotalAmount    float64    `gorm:"type:decimal(16,2);not null" json:"totalAmount"`
	UsedAmount     float64    `gorm:"type:decimal(16,2);not null;default:0" json:"usedAmount"`
	ReservedAmount float64    `gorm:"type:decimal(16,2);not null;default:0" json:"reservedAmount"`
	Currency       string     `gorm:"size:8;not null;default:CNY" json:"currency"`
	ApprovalStatus string     `gorm:"size:16;not null;default:Draft" json:"approvalStatus"`
	ApproverID     uint       `gorm:"index" json:"approverId"`
	ApprovedAt     *time.Time `json:"approvedAt"`
	CloserID       uint       `gorm:"index" json:"closerId"`
	ClosedAt       *time.Time `json:"closedAt"`
	Remarks        string     `gorm:"size:512" json:"remarks"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"-"`
}
