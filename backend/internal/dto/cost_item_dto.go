package dto

// CreateCostItemRequest is the payload for recording a cost item.
type CreateCostItemRequest struct {
	BudgetID        uint    `json:"budgetId" validate:"required"`
	Category        string  `json:"category" validate:"required"`
	Name            string  `json:"name" validate:"required,min=1,max=128"`
	BudgetAmount    float64 `json:"budgetAmount" validate:"min=0"`
	ActualAmount    float64 `json:"actualAmount" validate:"min=0"`
	OccurrenceDate  string  `json:"occurrenceDate"`
	VoucherNo       string  `json:"voucherNo" validate:"max=64"`
	MaterialUsageID uint    `json:"materialUsageId"`
	TimesheetID     uint    `json:"timesheetId"`
}

// UpdateCostItemRequest is the payload for editing a cost item.
type UpdateCostItemRequest struct {
	Category       string  `json:"category" validate:"required"`
	Name           string  `json:"name" validate:"required,min=1,max=128"`
	BudgetAmount   float64 `json:"budgetAmount" validate:"min=0"`
	ActualAmount   float64 `json:"actualAmount" validate:"min=0"`
	VoucherNo      string  `json:"voucherNo" validate:"max=64"`
	OccurrenceDate string  `json:"occurrenceDate"`
}
