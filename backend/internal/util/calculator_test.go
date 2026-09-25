package util

import "testing"

func TestVariance(t *testing.T) {
	tests := []struct {
		name   string
		budget float64
		actual float64
		want   float64
	}{
		{"over budget", 1000, 1200, 200},
		{"under budget", 1000, 800, -200},
		{"equal", 1000, 1000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Variance(tt.budget, tt.actual); got != tt.want {
				t.Fatalf("Variance(%v, %v) = %v, want %v", tt.budget, tt.actual, got, tt.want)
			}
		})
	}
}

func TestAfterAmount(t *testing.T) {
	if got := AfterAmount(56000000, 3800000); got != 59800000 {
		t.Fatalf("AfterAmount = %v, want 59800000", got)
	}
	if got := AfterAmount(42000000, -1200000); got != 40800000 {
		t.Fatalf("AfterAmount = %v, want 40800000", got)
	}
}
