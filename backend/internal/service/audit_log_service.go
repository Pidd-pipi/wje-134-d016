// Package service implements business logic on top of repositories.
package service

import (
	"fmt"
	"log/slog"

	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
)

// AuditLogService records and reads audit entries.
type AuditLogService struct {
	logs   *repository.AuditLogRepository
	logger *slog.Logger
}

// NewAuditLogService builds an AuditLogService.
func NewAuditLogService(logs *repository.AuditLogRepository, logger *slog.Logger) *AuditLogService {
	return &AuditLogService{logs: logs, logger: logger}
}

// Record writes an audit log entry.
func (s *AuditLogService) Record(userID uint, userName, action, entity string, entityID uint, detail string) {
	entry := &model.AuditLog{
		UserID:   userID,
		UserName: userName,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Detail:   detail,
	}
	if err := s.logs.Create(entry); err != nil {
		s.logger.Warn("record audit log failed", "error", err)
	}
}

// List returns recent audit logs.
func (s *AuditLogService) List(limit int) ([]model.AuditLog, error) {
	logs, err := s.logs.List(limit)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}
