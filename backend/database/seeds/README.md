# 种子数据说明

用户 5 人（admin/finance/pm/accountant/viewer）、预算 4 条、成本项 5 条、变更单 3 条，
由后端启动时通过 `service/seed.go`（GORM）幂等写入。

登录账号：
- admin / admin123（Admin）
- finance / finance123（FinanceManager，可审批预算）
- accountant / accountant123（Accountant，可录入成本）
- pm / pm123456（ProjectManager）
- viewer / viewer123（Viewer，只读）
