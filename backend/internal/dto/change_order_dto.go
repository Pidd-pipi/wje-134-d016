package dto

// CreateChangeOrderRequest is the payload for creating a change order.
type CreateChangeOrderRequest struct {
	ProjectID      uint    `json:"projectId" validate:"required"`
	ChangeType     string  `json:"changeType" validate:"required"`
	Description    string  `json:"description" validate:"required,min=1,max=512"`
	OriginalAmount float64 `json:"originalAmount" validate:"min=0"`
	ChangeAmount   float64 `json:"changeAmount" validate:"required"`
	Reason         string  `json:"reason" validate:"max=512"`
}

// UpdateChangeOrderRequest is the payload for editing a draft change order.
type UpdateChangeOrderRequest struct {
	ChangeType     string  `json:"changeType" validate:"required"`
	Description    string  `json:"description" validate:"required,min=1,max=512"`
	OriginalAmount float64 `json:"originalAmount" validate:"min=0"`
	ChangeAmount   float64 `json:"changeAmount" validate:"required"`
	Reason         string  `json:"reason" validate:"max=512"`
}
