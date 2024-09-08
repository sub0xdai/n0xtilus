package services

import "math"

type RiskCalculator struct{}

func NewRiskCalculator() *RiskCalculator {
	return &RiskCalculator{}
}

func (rc *RiskCalculator) CalculateRisk(accountBalance, riskPercentage, quantity, price float64) float64 {
	riskAmount := accountBalance * (riskPercentage / 100)
	positionSize := quantity * price
	actualRisk := (positionSize / accountBalance) * 100
	return actualRisk / riskAmount // Return risk as a ratio of actual risk to intended risk
}

func (rc *RiskCalculator) CalculatePositionSize(accountBalance, riskPercentage, entryPrice, stopLossPrice float64) float64 {
	riskAmount := accountBalance * (riskPercentage / 100)
	riskPerShare := math.Abs(entryPrice - stopLossPrice)
	return riskAmount / riskPerShare
}
