// Package risk_calculator provides risk management functionality for trading
package risk_calculator

import (
    "math"
    "errors"
)

// RiskCalculatorService defines the interface for risk calculation
type RiskCalculatorService interface {
    // CalculateRisk calculates the risk ratio for a given position
    CalculateRisk(accountBalance, riskPercentage, quantity, price float64) (float64, error)
    
    // CalculatePositionSize calculates the position size based on risk parameters
    CalculatePositionSize(accountBalance, riskPercentage, entryPrice, stopLossPrice float64) (float64, error)
}

// RiskCalculator implements RiskCalculatorService
type RiskCalculator struct{}

// NewRiskCalculator creates a new RiskCalculator
func NewRiskCalculator() RiskCalculatorService {
    return &RiskCalculator{}
}

// CalculateRisk calculates the risk ratio
func (rc *RiskCalculator) CalculateRisk(accountBalance, riskPercentage, quantity, price float64) (float64, error) {
    if accountBalance <= 0 || riskPercentage <= 0 || quantity <= 0 || price <= 0 {
        return 0, errors.New("all input values must be positive")
    }

    riskAmount := accountBalance * (riskPercentage / 100)
    positionSize := quantity * price
    actualRisk := (positionSize / accountBalance) * 100
    return actualRisk / riskAmount, nil // Return risk as a ratio of actual risk to intended risk
}

// CalculatePositionSize calculates the position size based on risk parameters
func (rc *RiskCalculator) CalculatePositionSize(accountBalance, riskPercentage, entryPrice, stopLossPrice float64) (float64, error) {
    if accountBalance <= 0 || riskPercentage <= 0 || entryPrice <= 0 || stopLossPrice <= 0 {
        return 0, errors.New("all input values must be positive")
    }

    if entryPrice == stopLossPrice {
        return 0, errors.New("entry price cannot be equal to stop loss price")
    }

    riskAmount := accountBalance * (riskPercentage / 100)
    riskPerShare := math.Abs(entryPrice - stopLossPrice)
    return riskAmount / riskPerShare, nil
}
