package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
)

// ChangeOrderRepository persists change orders.
type ChangeOrderRepository struct {
	db *gorm.DB
}

// NewChangeOrderRepository builds a ChangeOrderRepository.
func NewChangeOrderRepository(db *gorm.DB) *ChangeOrderRepository {
	return &ChangeOrderRepository{db: db}
}

// Create inserts a change order.
func (r *ChangeOrderRepository) Create(c *model.ChangeOrder) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create change order: %w", err)
	}
	return nil
}

// List returns change orders, optionally filtered by project.
func (r *ChangeOrderRepository) List(projectID uint) ([]model.ChangeOrder, error) {
	q := r.db.Order("created_at DESC")
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	var orders []model.ChangeOrder
	if err := q.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list change orders: %w", err)
	}
	return orders, nil
}

// FindByID loads a change order by primary key.
func (r *ChangeOrderRepository) FindByID(id uint) (*model.ChangeOrder, error) {
	var c model.ChangeOrder
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find change order by id: %w", err)
	}
	return &c, nil
}

// Update persists change order changes.
func (r *ChangeOrderRepository) Update(c *model.ChangeOrder) error {
	if err := r.db.Save(c).Error; err != nil {
		return fmt.Errorf("update change order: %w", err)
	}
	return nil
}
