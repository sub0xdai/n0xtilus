package services

import (
	"fmt"

	"github.com/tehuticode/b0xfi/internal/api"
)

type OrderService struct {
	client         *api.APIClient
	riskCalculator *RiskCalculator
}

func NewOrderService(client *api.APIClient, riskCalculator *RiskCalculator) *OrderService {
	return &OrderService{
		client:         client,
		riskCalculator: riskCalculator,
	}
}

func (s *OrderService) CalculatePositionSize(riskPercentage, entryPrice, stopLossPrice float64) (float64, error) {
	balance, err := s.client.GetBalance()
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	return s.riskCalculator.CalculatePositionSize(balance, riskPercentage, entryPrice, stopLossPrice), nil
}

func (s *OrderService) PlaceOrder(symbol, side, quantity, price string) (string, error) {
	// Place the order
	return s.client.PlaceOrder(symbol, side, quantity, price)
}

