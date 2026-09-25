package service

import "github.com/costguard/costguard/internal/dto"

func dtoCreateBudget() dto.CreateBudgetRequest {
	return dto.CreateBudgetRequest{
		ProjectID:   1001,
		Name:        "测试预算",
		TotalAmount: 1000000,
		Currency:    "CNY",
	}
}
