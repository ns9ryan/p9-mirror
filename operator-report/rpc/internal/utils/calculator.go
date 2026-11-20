package utils

import (
	"github.com/shopspring/decimal"
)

type Calculator struct {
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// GetKillAmount 计算杀数
func (Calculator) GetKillAmount(winLossAmount, validAmount int64) string {
	if validAmount == 0 {
		return "-"
	}
	if winLossAmount == 0 {
		return "0%"
	}
	result := decimal.NewFromInt(winLossAmount).
		Div(decimal.NewFromInt(validAmount)).
		Mul(decimal.NewFromInt(100)).
		StringFixed(2) + "%"
	return result
}

// GetReDepositRate 计算复充率
func (Calculator) GetReDepositRate(reDepositUser, firstDepositUser int64) string {
	if firstDepositUser == 0 {
		return "-"
	}
	if reDepositUser == 0 {
		return "0%"
	}
	result := decimal.NewFromInt(reDepositUser).
		Div(decimal.NewFromInt(firstDepositUser)).
		Mul(decimal.NewFromInt(100)).
		StringFixed(2) + "%"
	return result
}

// GetRegCount 计算注册总数
func (Calculator) GetRegCount(userCount, agentCount int64) int64 {
	return userCount + agentCount
}

// GetDepWithDiff 计算充提差
func (Calculator) GetDepWithDiff(depositAmount, withdrawAmount int64) int64 {
	return depositAmount - withdrawAmount
}

// GetGrowthRate 增长率计算方法
func (Calculator) GetGrowthRate(current, previous int64) float64 {
	if previous == 0 {
		return 0
	}
	if current == 0 {
		return 0
	}
	dCurrent := decimal.NewFromInt(current)
	dPrevious := decimal.NewFromInt(previous)
	hundred := decimal.NewFromInt(100)
	result, _ := dCurrent.Sub(dPrevious).Div(dPrevious).Mul(hundred).Round(4).Float64()
	return result
}

// GetAvgAmount 计算用户平均金额
func (Calculator) GetAvgAmount(amount int64, users int64) (avgAmount int64) {
	if users == 0 {
		return 0
	}
	if amount == 0 {
		return 0
	}
	return amount / users
}
