package constants

// Cost categories.
const (
	CostCategoryMaterial    = "Material"
	CostCategoryLabor       = "Labor"
	CostCategoryEquipment   = "Equipment"
	CostCategorySubcontract = "Subcontract"
	CostCategoryOverhead    = "Overhead"
	CostCategoryOther       = "Other"
)

// ValidCostCategory reports whether a category is valid.
func ValidCostCategory(s string) bool {
	switch s {
	case CostCategoryMaterial, CostCategoryLabor, CostCategoryEquipment, CostCategorySubcontract, CostCategoryOverhead, CostCategoryOther:
		return true
	}
	return false
}
