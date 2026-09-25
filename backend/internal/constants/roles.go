package constants

// User roles.
const (
	RoleAdmin          = "Admin"
	RoleFinanceManager = "FinanceManager"
	RoleProjectManager = "ProjectManager"
	RoleAccountant     = "Accountant"
	RoleViewer         = "Viewer"
)

// CanApproveBudget reports whether a role may approve budgets.
func CanApproveBudget(role string) bool {
	return role == RoleAdmin || role == RoleFinanceManager
}

// CanRecordCost reports whether a role may record cost items.
func CanRecordCost(role string) bool {
	return role == RoleAdmin || role == RoleAccountant || role == RoleFinanceManager
}

// CanManageChangeOrder reports whether a role may approve change orders.
func CanManageChangeOrder(role string) bool {
	return role == RoleAdmin || role == RoleFinanceManager || role == RoleProjectManager
}
