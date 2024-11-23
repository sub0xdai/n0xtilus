package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sub0xdai/n0xtilus/internal/validation"
)

// OrderCommand represents a trading command to be executed
type OrderCommand struct {
	Type           CommandType
	Symbol         string
	Side           string
	Quantity       string
	Price          string
	OrderID        string
	Timestamp      time.Time
	Leverage       float64
	RiskPercentage float64
}

type CommandType int

const (
	CommandPlaceOrder CommandType = iota
	CommandCancelOrder
	CommandModifyOrder
)

// CommandQueue manages the order execution queue
type CommandQueue struct {
	commands     chan OrderCommand
	wg          sync.WaitGroup
	stateManager *OrderStateManager
	validator   *validation.OrderValidator
}

// NewCommandQueue creates a new command queue with specified buffer size
func NewCommandQueue(bufferSize int) *CommandQueue {
	return &CommandQueue{
		commands:     make(chan OrderCommand, bufferSize),
		stateManager: NewOrderStateManager(),
		validator:    validation.NewOrderValidator(0.00001, 1000000, 0.00001, 1000000, 100, 5), // Example limits
	}
}

// Start begins processing commands from the queue
func (q *CommandQueue) Start(ctx context.Context, executor OrderExecutor) {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case cmd := <-q.commands:
				q.processCommand(cmd, executor)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop gracefully stops the command queue
func (q *CommandQueue) Stop() {
	q.wg.Wait()
}

// Enqueue adds a command to the queue
func (q *CommandQueue) Enqueue(cmd OrderCommand) error {
	// Create atomic order and add to state manager
	atomicOrder := NewAtomicOrder(cmd, q.validator)
	q.stateManager.AddOrder(atomicOrder)

	select {
	case q.commands <- cmd:
		return nil
	default:
		q.stateManager.RemoveOrder(cmd.OrderID)
		return errors.New("command queue is full")
	}
}

// GetStatus returns the status of an order
func (q *CommandQueue) GetStatus(orderID string) (OrderCommand, error) {
	order, exists := q.stateManager.GetOrder(orderID)
	if !exists {
		return OrderCommand{}, errors.New("order not found")
	}

	return OrderCommand{
		OrderID:   order.ID,
		Symbol:    order.Symbol,
		Side:      order.Side,
		Quantity:  order.Quantity,
		Price:     order.Price,
		Timestamp: order.timestamp,
	}, order.GetError()
}

func (q *CommandQueue) processCommand(cmd OrderCommand, executor OrderExecutor) {
	order, exists := q.stateManager.GetOrder(cmd.OrderID)
	if !exists {
		return // Order was removed or doesn't exist
	}

	// Update order to active state
	err := q.stateManager.UpdateOrderState(cmd.OrderID, OrderStateActive)
	if err != nil {
		order.SetError(err)
		return
	}

	var execErr error
	switch cmd.Type {
	case CommandPlaceOrder:
		_, execErr = executor.PlaceOrder(cmd.Symbol, cmd.Side, cmd.Quantity, cmd.Price)
	case CommandCancelOrder:
		execErr = executor.CancelOrder(cmd.OrderID)
	case CommandModifyOrder:
		execErr = executor.ModifyOrder(cmd.OrderID, cmd.Quantity, cmd.Price)
	}

	if execErr != nil {
		order.SetError(execErr)
		return
	}

	// Update to filled state for successful execution
	_ = q.stateManager.UpdateOrderState(cmd.OrderID, OrderStateFilled)
}

// GetPendingOrders returns all pending orders
func (q *CommandQueue) GetPendingOrders() []*AtomicOrder {
	return q.stateManager.GetOrdersByState(OrderStatePending)
}

// GetActiveOrders returns all active orders
func (q *CommandQueue) GetActiveOrders() []*AtomicOrder {
	return q.stateManager.GetOrdersByState(OrderStateActive)
}

// GetFilledOrders returns all filled orders
func (q *CommandQueue) GetFilledOrders() []*AtomicOrder {
	return q.stateManager.GetOrdersByState(OrderStateFilled)
}

// GetFailedOrders returns all failed orders
func (q *CommandQueue) GetFailedOrders() []*AtomicOrder {
	return q.stateManager.GetOrdersByState(OrderStateFailed)
}

// HasFailed checks if an order has failed
func (q *CommandQueue) HasFailed(orderID string) bool {
	if order, exists := q.stateManager.GetOrder(orderID); exists {
		return order.GetState() == OrderStateFailed
	}
	return false
}

// ClearTerminalOrders removes all orders in terminal states
func (q *CommandQueue) ClearTerminalOrders() {
	orders := q.stateManager.GetAllOrders()
	for _, order := range orders {
		if order.IsTerminal() {
			q.stateManager.RemoveOrder(order.ID)
		}
	}
}

// OrderExecutor interface defines methods for executing orders
type OrderExecutor interface {
	PlaceOrder(symbol, side, quantity, price string) (string, error)
	CancelOrder(orderID string) error
	ModifyOrder(orderID, quantity, price string) error
}
