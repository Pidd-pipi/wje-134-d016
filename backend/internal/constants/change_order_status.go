package constants

// Change order approval statuses.
const (
	ChangeOrderStatusDraft     = "Draft"
	ChangeOrderStatusSubmitted = "Submitted"
	ChangeOrderStatusApproved  = "Approved"
	ChangeOrderStatusRejected  = "Rejected"
	ChangeOrderStatusCancelled = "Cancelled"
)

// ValidChangeOrderStatus reports whether a status is valid.
func ValidChangeOrderStatus(s string) bool {
	switch s {
	case ChangeOrderStatusDraft, ChangeOrderStatusSubmitted, ChangeOrderStatusApproved, ChangeOrderStatusRejected, ChangeOrderStatusCancelled:
		return true
	}
	return false
}
