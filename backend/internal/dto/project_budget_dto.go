// Package dto defines request/response payloads with validator rules.
package dto

// CreateBudgetRequest is the payload for creating a budget.
type CreateBudgetRequest struct {
	ProjectID      uint    `json:"projectId" validate:"required"`
	Name           string  `json:"name" validate:"required,min=1,max=128"`
	TotalAmount    float64 `json:"totalAmount" validate:"required,min=0"`
	ReservedAmount float64 `json:"reservedAmount" validate:"min=0"`
	Currency       string  `json:"currency" validate:"oneof=CNY USD"`
	Remarks        string  `json:"remarks" validate:"max=512"`
}

// UpdateBudgetRequest is the payload for editing a draft budget.
type UpdateBudgetRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=128"`
	TotalAmount    float64 `json:"totalAmount" validate:"required,min=0"`
	ReservedAmount float64 `json:"reservedAmount" validate:"min=0"`
	Remarks        string  `json:"remarks" validate:"max=512"`
}

// AbnormalCostItemInfo describes an abnormal cost item blocking a budget close.
type AbnormalCostItemInfo struct {
	ID             uint    `json:"id"`
	VoucherNo      string  `json:"voucherNo"`
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	BudgetAmount   float64 `json:"budgetAmount"`
	ActualAmount   float64 `json:"actualAmount"`
	VarianceAmount float64 `json:"varianceAmount"`
}
