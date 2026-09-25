package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/costguard/costguard/internal/model"
)

// AuditLogRepository persists audit logs.
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository builds an AuditLogRepository.
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create inserts an audit log.
func (r *AuditLogRepository) Create(l *model.AuditLog) error {
	if err := r.db.Create(l).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// List returns recent audit logs.
func (r *AuditLogRepository) List(limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var logs []model.AuditLog
	if err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}
