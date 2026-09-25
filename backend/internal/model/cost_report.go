package model

import "time"

// CostReport is a generated cost analysis report.
type CostReport struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProjectID      uint      `gorm:"index;not null" json:"projectId"`
	ReportPeriod   string    `gorm:"size:32" json:"reportPeriod"`
	ReportType     string    `gorm:"size:16;not null" json:"reportType"`
	LaborCost      float64   `gorm:"type:decimal(16,2);not null;default:0" json:"laborCost"`
	MaterialCost   float64   `gorm:"type:decimal(16,2);not null;default:0" json:"materialCost"`
	EquipmentCost  float64   `gorm:"type:decimal(16,2);not null;default:0" json:"equipmentCost"`
	OtherCost      float64   `gorm:"type:decimal(16,2);not null;default:0" json:"otherCost"`
	TotalCost      float64   `gorm:"type:decimal(16,2);not null;default:0" json:"totalCost"`
	ProfitAnalysis string    `gorm:"type:text" json:"profitAnalysis"`
	GeneratedAt    time.Time `json:"generatedAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"-"`
}
