package util

// Variance computes the cost variance between actual and budgeted amounts.
// Positive means over budget (超支).
func Variance(budget, actual float64) float64 {
	return actual - budget
}

// AfterAmount computes the resulting amount after a change order.
func AfterAmount(original, change float64) float64 {
	return original + change
}
