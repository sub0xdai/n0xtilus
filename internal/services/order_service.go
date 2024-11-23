package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sub0xdai/n0xtilus/internal/api"
	"github.com/sub0xdai/n0xtilus/internal/services/risk_calculator"
)

type OrderService struct {
	client         *api.APIClient
	riskCalculator risk_calculator.RiskCalculatorService
}

func NewOrderService(client *api.APIClient, riskCalculator risk_calculator.RiskCalculatorService) *OrderService {
	return &OrderService{
		client:         client,
		riskCalculator: riskCalculator,
	}
}

type OrderServicer interface {
	CalculatePositionSize(riskPercentage, entryPrice, stopLossPrice float64) (float64, error)
	PlaceOrder(symbol, side, quantity, price string) (string, error)
	CancelOrder(orderID string) error
	ModifyOrder(orderID, quantity, price string) error
}

type TradeExecutor struct {
	client         *api.APIClient
	orderService   OrderServicer
	riskPercentage float64
	symbol         string
	side           string
	entryPrice     float64
	stopLossPrice  float64
	commandQueue   *CommandQueue
}

func NewTradeExecutor(client *api.APIClient, orderService OrderServicer, riskPercentage float64, symbol string, side string, entryPrice float64, stopLossPrice float64) *TradeExecutor {
	return &TradeExecutor{
		client:         client,
		orderService:   orderService,
		riskPercentage: riskPercentage,
		symbol:         symbol,
		side:           side,
		entryPrice:     entryPrice,
		stopLossPrice:  stopLossPrice,
		commandQueue:   NewCommandQueue(100), // Buffer size of 100 commands
	}
}

func (te *TradeExecutor) Execute() error {
	// Validate trade parameters
	if err := te.validateTrade(); err != nil {
		return fmt.Errorf("trade validation failed: %w", err)
	}

	// Calculate position size
	posSize, err := te.orderService.CalculatePositionSize(
		te.riskPercentage,
		te.entryPrice,
		te.stopLossPrice,
	)
	if err != nil {
		return fmt.Errorf("position size calculation failed: %w", err)
	}

	// Start the command queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	te.commandQueue.Start(ctx, te.orderService)
	defer te.commandQueue.Stop()

	// Create main order command
	mainOrderCmd := OrderCommand{
		Type:      CommandPlaceOrder,
		Symbol:    te.symbol,
		Side:      te.side,
		Quantity:  fmt.Sprintf("%.8f", posSize),
		Price:     fmt.Sprintf("%.8f", te.entryPrice),
		OrderID:   generateOrderID(),
		Timestamp: time.Now(),
	}

	// Enqueue main order
	if err := te.commandQueue.Enqueue(mainOrderCmd); err != nil {
		return fmt.Errorf("failed to enqueue main order: %w", err)
	}

	// Wait for main order to complete
	mainOrderStatus, err := te.waitForOrderCompletion(mainOrderCmd.OrderID)
	if err != nil {
		return fmt.Errorf("main order failed: %w", err)
	}

	// Create stop loss order command
	stopLossCmd := OrderCommand{
		Type:      CommandPlaceOrder,
		Symbol:    te.symbol,
		Side:      te.getOpposingSide(),
		Quantity:  fmt.Sprintf("%.8f", posSize),
		Price:     fmt.Sprintf("%.8f", te.stopLossPrice),
		OrderID:   generateOrderID(),
		Timestamp: time.Now(),
	}

	// Enqueue stop loss order
	if err := te.commandQueue.Enqueue(stopLossCmd); err != nil {
		// If stop loss fails, try to cancel the main order
		cancelCmd := OrderCommand{
			Type:      CommandCancelOrder,
			OrderID:   mainOrderStatus.OrderID,
			Timestamp: time.Now(),
		}
		_ = te.commandQueue.Enqueue(cancelCmd)
		return fmt.Errorf("failed to enqueue stop loss order: %w", err)
	}

	return nil
}

func (te *TradeExecutor) waitForOrderCompletion(orderID string) (OrderCommand, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		cmd, err := te.commandQueue.GetStatus(orderID)
		if err == nil {
			return cmd, nil
		}
		if te.commandQueue.HasFailed(orderID) {
			return OrderCommand{}, err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return OrderCommand{}, errors.New("order timed out")
}

func (te *TradeExecutor) validateTrade() error {
	if te.symbol == "" || te.side == "" {
		return errors.New("invalid trade parameters")
	}
	if te.entryPrice <= 0 || te.stopLossPrice <= 0 {
		return errors.New("invalid prices")
	}
	if te.riskPercentage <= 0 || te.riskPercentage > 100 {
		return errors.New("invalid risk percentage")
	}
	return nil
}

func (te *TradeExecutor) getOpposingSide() string {
	if te.side == "BUY" {
		return "SELL"
	}
	return "BUY"
}

func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

func (s *OrderService) PlaceOrder(symbol, side, quantity, price string) (string, error) {
	// Implement the order placement logic here
	// For now, we'll just call the API client's PlaceOrder method
	return s.client.PlaceOrder(symbol, side, quantity, price)
}

func (s *OrderService) CancelOrder(orderID string) error {
	return s.client.CancelOrder(orderID)
}

func (s *OrderService) ModifyOrder(orderID, quantity, price string) error {
	// Implement order modification logic
	return errors.New("modify order not implemented")
}
