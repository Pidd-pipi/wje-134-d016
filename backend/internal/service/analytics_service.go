package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/dto"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
)

const reportCacheTTL = 10 * time.Minute

// AnalyticsService generates cost reports and caches them in Redis.
type AnalyticsService struct {
	items   *repository.CostItemRepository
	budgets *repository.ProjectBudgetRepository
	reports *repository.CostReportRepository
	rdb     *redis.Client
	audit   *AuditLogService
	logger  *slog.Logger
}

// NewAnalyticsService builds an AnalyticsService.
func NewAnalyticsService(items *repository.CostItemRepository, budgets *repository.ProjectBudgetRepository, reports *repository.CostReportRepository, rdb *redis.Client, audit *AuditLogService, logger *slog.Logger) *AnalyticsService {
	return &AnalyticsService{items: items, budgets: budgets, reports: reports, rdb: rdb, audit: audit, logger: logger}
}

// List returns generated reports.
func (s *AnalyticsService) List(projectID uint) ([]model.CostReport, error) {
	list, err := s.reports.List(projectID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	return list, nil
}

// Get returns a report, preferring the Redis cache.
func (s *AnalyticsService) Get(id uint) (*model.CostReport, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("costguard:report:%d", id)
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var report model.CostReport
		if json.Unmarshal([]byte(cached), &report) == nil {
			return &report, nil
		}
	}
	report, err := s.reports.FindByID(id)
	if err != nil {
		return nil, err
	}
	if raw, err := json.Marshal(report); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, raw, reportCacheTTL).Err(); err != nil {
			s.logger.Warn("cache report failed", "error", err)
		}
	}
	return report, nil
}

// Generate builds a cost report from actual cost items and caches it.
func (s *AnalyticsService) Generate(req dto.GenerateReportRequest, userID uint, userName string) (*model.CostReport, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("costguard:report:gen:%d:%s:%s", req.ProjectID, req.ReportType, req.ReportPeriod)
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var report model.CostReport
		if json.Unmarshal([]byte(cached), &report) == nil {
			return &report, nil
		}
	}

	budgets, err := s.budgets.List(req.ProjectID)
	if err != nil {
		return nil, err
	}
	var labor, material, equipment, other float64
	for _, b := range budgets {
		byCat, err := s.items.SumActualByCategory(b.ID)
		if err != nil {
			return nil, err
		}
		labor += byCat[constants.CostCategoryLabor]
		material += byCat[constants.CostCategoryMaterial]
		equipment += byCat[constants.CostCategoryEquipment]
		other += byCat[constants.CostCategorySubcontract] + byCat[constants.CostCategoryOverhead] + byCat[constants.CostCategoryOther]
	}
	total := labor + material + equipment + other
	profit := "成本合计 " + fmt.Sprintf("%.2f", total) +
		"；其中人工 " + fmt.Sprintf("%.2f", labor) +
		"、材料 " + fmt.Sprintf("%.2f", material) +
		"、设备 " + fmt.Sprintf("%.2f", equipment) +
		"、其他 " + fmt.Sprintf("%.2f", other) +
		"。请结合合同收入进行盈亏分析。"

	report := &model.CostReport{
		ProjectID:      req.ProjectID,
		ReportType:     req.ReportType,
		ReportPeriod:   req.ReportPeriod,
		LaborCost:      labor,
		MaterialCost:   material,
		EquipmentCost:  equipment,
		OtherCost:      other,
		TotalCost:      total,
		ProfitAnalysis: profit,
		GeneratedAt:    time.Now(),
	}
	if err := s.reports.Create(report); err != nil {
		return nil, fmt.Errorf("create cost report: %w", err)
	}
	if raw, err := json.Marshal(report); err == nil {
		if err := s.rdb.Set(ctx, cacheKey, raw, reportCacheTTL).Err(); err != nil {
			s.logger.Warn("cache generated report failed", "error", err)
		}
		if err := s.rdb.Set(ctx, fmt.Sprintf("costguard:report:%d", report.ID), raw, reportCacheTTL).Err(); err != nil {
			s.logger.Warn("cache report by id failed", "error", err)
		}
	}
	s.audit.Record(userID, userName, "report.generate", "cost_report", report.ID, fmt.Sprintf("生成成本报表 %s %s", req.ReportType, req.ReportPeriod))
	s.logger.Info("cost report generated", "report_id", report.ID, "project_id", req.ProjectID)
	return report, nil
}
