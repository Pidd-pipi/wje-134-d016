package model

import "time"

// ChangeOrder records a project scope/price change.
type ChangeOrder struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ProjectID       uint       `gorm:"index;not null" json:"projectId"`
	ChangeType      string     `gorm:"size:32;not null" json:"changeType"`
	Description     string     `gorm:"size:512" json:"description"`
	OriginalAmount  float64    `gorm:"type:decimal(16,2);not null" json:"originalAmount"`
	ChangeAmount    float64    `gorm:"type:decimal(16,2);not null" json:"changeAmount"`
	AfterAmount     float64    `gorm:"type:decimal(16,2);not null" json:"afterAmount"`
	Reason          string     `gorm:"size:512" json:"reason"`
	ApprovalStatus  string     `gorm:"size:16;not null;default:Draft" json:"approvalStatus"`
	ApplicantID     uint       `gorm:"index;not null" json:"applicantId"`
	ApproverID      uint       `gorm:"index" json:"approverId"`
	AppliedAt       time.Time  `json:"appliedAt"`
	ApprovedAt      *time.Time `json:"approvedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"-"`
}
