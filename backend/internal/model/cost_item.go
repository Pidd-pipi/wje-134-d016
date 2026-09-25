package model

import "time"

// CostItem records an actual cost line under a budget.
type CostItem struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	BudgetID        uint      `gorm:"index;not null" json:"budgetId"`
	Category        string    `gorm:"size:32;not null" json:"category"`
	Name            string    `gorm:"size:128;not null" json:"name"`
	BudgetAmount    float64   `gorm:"type:decimal(16,2);not null" json:"budgetAmount"`
	ActualAmount    float64   `gorm:"type:decimal(16,2);not null;default:0" json:"actualAmount"`
	VarianceAmount  float64   `gorm:"type:decimal(16,2);not null;default:0" json:"varianceAmount"`
	IsAbnormal      bool      `gorm:"not null;default:false" json:"isAbnormal"`
	OccurrenceDate  time.Time `json:"occurrenceDate"`
	VoucherNo       string    `gorm:"size:64" json:"voucherNo"`
	MaterialUsageID uint      `gorm:"index" json:"materialUsageId"`
	TimesheetID     uint      `gorm:"index" json:"timesheetId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"-"`
}
