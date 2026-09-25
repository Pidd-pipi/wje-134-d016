-- 001_init.sql
-- Reference DDL for the cost control schema (PostgreSQL 15).
-- In production the Go service runs GORM AutoMigrate on startup.
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(64),
  role VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS project_budgets (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  project_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  total_amount NUMERIC(16,2) NOT NULL,
  used_amount NUMERIC(16,2) NOT NULL DEFAULT 0,
  reserved_amount NUMERIC(16,2) NOT NULL DEFAULT 0,
  currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
  approval_status VARCHAR(16) NOT NULL DEFAULT 'Draft',
  approver_id BIGINT,
  approved_at TIMESTAMPTZ,
  closer_id BIGINT,
  closed_at TIMESTAMPTZ,
  remarks VARCHAR(512)
);

CREATE TABLE IF NOT EXISTS cost_items (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  budget_id BIGINT NOT NULL,
  category VARCHAR(32) NOT NULL,
  name VARCHAR(128) NOT NULL,
  budget_amount NUMERIC(16,2) NOT NULL,
  actual_amount NUMERIC(16,2) NOT NULL DEFAULT 0,
  variance_amount NUMERIC(16,2) NOT NULL DEFAULT 0,
  is_abnormal BOOLEAN NOT NULL DEFAULT FALSE,
  occurrence_date TIMESTAMPTZ,
  voucher_no VARCHAR(64),
  material_usage_id BIGINT,
  timesheet_id BIGINT
);

CREATE TABLE IF NOT EXISTS change_orders (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  project_id BIGINT NOT NULL,
  change_type VARCHAR(32) NOT NULL,
  description VARCHAR(512),
  original_amount NUMERIC(16,2) NOT NULL,
  change_amount NUMERIC(16,2) NOT NULL,
  after_amount NUMERIC(16,2) NOT NULL,
  reason VARCHAR(512),
  approval_status VARCHAR(16) NOT NULL DEFAULT 'Draft',
  applicant_id BIGINT NOT NULL,
  approver_id BIGINT,
  applied_at TIMESTAMPTZ,
  approved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cost_reports (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  project_id BIGINT NOT NULL,
  report_period VARCHAR(32),
  report_type VARCHAR(16) NOT NULL,
  labor_cost NUMERIC(16,2) NOT NULL DEFAULT 0,
  material_cost NUMERIC(16,2) NOT NULL DEFAULT 0,
  equipment_cost NUMERIC(16,2) NOT NULL DEFAULT 0,
  other_cost NUMERIC(16,2) NOT NULL DEFAULT 0,
  total_cost NUMERIC(16,2) NOT NULL DEFAULT 0,
  profit_analysis TEXT,
  generated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  user_id BIGINT,
  user_name VARCHAR(64),
  action VARCHAR(64),
  entity VARCHAR(64),
  entity_id BIGINT,
  detail VARCHAR(512)
);
