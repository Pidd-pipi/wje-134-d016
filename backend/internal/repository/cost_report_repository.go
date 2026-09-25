package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
)

// CostReportRepository persists generated reports.
type CostReportRepository struct {
	db *gorm.DB
}

// NewCostReportRepository builds a CostReportRepository.
func NewCostReportRepository(db *gorm.DB) *CostReportRepository {
	return &CostReportRepository{db: db}
}

// Create inserts a report.
func (r *CostReportRepository) Create(c *model.CostReport) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create cost report: %w", err)
	}
	return nil
}

// List returns reports, optionally filtered by project.
func (r *CostReportRepository) List(projectID uint) ([]model.CostReport, error) {
	q := r.db.Order("generated_at DESC")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	var reports []model.CostReport
	if err := q.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("list cost reports: %w", err)
	}
	return reports, nil
}

// FindByID loads a report by primary key.
func (r *CostReportRepository) FindByID(id uint) (*model.CostReport, error) {
	var c model.CostReport
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find cost report by id: %w", err)
	}
	return &c, nil
}
