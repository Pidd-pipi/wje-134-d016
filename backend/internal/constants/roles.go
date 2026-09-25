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

// CanCloseBudget reports whether a role may close (finalize) a budget.
// Only the finance manager may close a budget at project settlement.
func CanCloseBudget(role string) bool {
	return role == RoleFinanceManager
}
