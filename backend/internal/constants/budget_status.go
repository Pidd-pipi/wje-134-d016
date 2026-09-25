package constants

// Budget approval statuses.
const (
	BudgetStatusDraft     = "Draft"
	BudgetStatusSubmitted = "Submitted"
	BudgetStatusApproved  = "Approved"
	BudgetStatusRejected  = "Rejected"
	BudgetStatusClosed    = "Closed"
)

// ValidBudgetStatus reports whether a status is valid.
func ValidBudgetStatus(s string) bool {
	switch s {
	case BudgetStatusDraft, BudgetStatusSubmitted, BudgetStatusApproved, BudgetStatusRejected, BudgetStatusClosed:
		return true
	}
	return false
}
