package constants

// Change order types.
const (
	ChangeTypeScopeChange         = "ScopeChange"
	ChangeTypeDesignChange        = "DesignChange"
	ChangeTypePriceAdjustment     = "PriceAdjustment"
	ChangeTypeUnforeseenCondition = "UnforeseenCondition"
)

// ValidChangeType reports whether a change type is valid.
func ValidChangeType(s string) bool {
	switch s {
	case ChangeTypeScopeChange, ChangeTypeDesignChange, ChangeTypePriceAdjustment, ChangeTypeUnforeseenCondition:
		return true
	}
	return false
}
