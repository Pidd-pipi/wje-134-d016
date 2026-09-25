package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
)

// CostItemRepository persists cost items.
type CostItemRepository struct {
	db *gorm.DB
}

// NewCostItemRepository builds a CostItemRepository.
func NewCostItemRepository(db *gorm.DB) *CostItemRepository {
	return &CostItemRepository{db: db}
}

// Create inserts a cost item.
func (r *CostItemRepository) Create(c *model.CostItem) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create cost item: %w", err)
	}
	return nil
}

// List returns cost items, optionally filtered by budget.
func (r *CostItemRepository) List(budgetID uint) ([]model.CostItem, error) {
	q := r.db.Order("occurrence_date DESC")
	if budgetID > 0 {
		q = q.Where("budget_id = ?", budgetID)
	}
	var items []model.CostItem
	if err := q.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list cost items: %w", err)
	}
	return items, nil
}

// FindByID loads a cost item by primary key.
func (r *CostItemRepository) FindByID(id uint) (*model.CostItem, error) {
	var c model.CostItem
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find cost item by id: %w", err)
	}
	return &c, nil
}

// Update persists cost item changes.
func (r *CostItemRepository) Update(c *model.CostItem) error {
	if err := r.db.Save(c).Error; err != nil {
		return fmt.Errorf("update cost item: %w", err)
	}
	return nil
}

// SumActualByCategory aggregates actual amounts by category for a budget.
func (r *CostItemRepository) SumActualByCategory(budgetID uint) (map[string]float64, error) {
	type row struct {
		Category string
		Total    float64
	}
	var rows []row
	if err := r.db.Model(&model.CostItem{}).
		Select("category, SUM(actual_amount) AS total").
		Where("budget_id = ?", budgetID).
		Group("category").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("sum actual by category: %w", err)
	}
	out := map[string]float64{}
	for _, rw := range rows {
		out[rw.Category] = rw.Total
	}
	return out, nil
}
