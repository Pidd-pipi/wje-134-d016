package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
)

// ProjectBudgetRepository persists budgets.
type ProjectBudgetRepository struct {
	db *gorm.DB
}

// NewProjectBudgetRepository builds a ProjectBudgetRepository.
func NewProjectBudgetRepository(db *gorm.DB) *ProjectBudgetRepository {
	return &ProjectBudgetRepository{db: db}
}

// Create inserts a budget.
func (r *ProjectBudgetRepository) Create(b *model.ProjectBudget) error {
	if err := r.db.Create(b).Error; err != nil {
		return fmt.Errorf("create budget: %w", err)
	}
	return nil
}

// List returns budgets, optionally filtered by project.
func (r *ProjectBudgetRepository) List(projectID uint) ([]model.ProjectBudget, error) {
	q := r.db.Order("created_at DESC")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	var budgets []model.ProjectBudget
	if err := q.Find(&budgets).Error; err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	return budgets, nil
}

// FindByID loads a budget by primary key.
func (r *ProjectBudgetRepository) FindByID(id uint) (*model.ProjectBudget, error) {
	var b model.ProjectBudget
	if err := r.db.First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find budget by id: %w", err)
	}
	return &b, nil
}

// Update persists budget changes.
func (r *ProjectBudgetRepository) Update(b *model.ProjectBudget) error {
	if err := r.db.Save(b).Error; err != nil {
		return fmt.Errorf("update budget: %w", err)
	}
	return nil
}
