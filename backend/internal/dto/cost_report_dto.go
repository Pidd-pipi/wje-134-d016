package dto

// GenerateReportRequest asks the analytics service to generate a report.
type GenerateReportRequest struct {
	ProjectID    uint   `json:"projectId" validate:"required"`
	ReportType   string `json:"reportType" validate:"required,oneof=Monthly Quarterly Annual Custom"`
	ReportPeriod string `json:"reportPeriod" validate:"max=32"`
}
