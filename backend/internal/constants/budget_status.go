package constants

// Budget approval statuses.
const (
	BudgetStatusDraft     = "Draft"
	BudgetStatusSubmitted = "Submitted"
	BudgetStatusApproved  = "Approved"
	BudgetStatusRejected  = "Rejected"
)

// ValidBudgetStatus reports whether a status is valid.
func ValidBudgetStatus(s string) bool {
	switch s {
	case BudgetStatusDraft, BudgetStatusSubmitted, BudgetStatusApproved, BudgetStatusRejected:
		return true
	}
	return false
}
