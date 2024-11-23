package services

import (
	"sync/atomic"
	"time"
	"errors"
	"fmt"
	"sync"
	"strconv"

	"github.com/sub0xdai/n0xtilus/internal/validation"
)

// OrderState represents the current state of an order
type OrderState int32

const (
	OrderStateUnknown OrderState = iota
	OrderStateValidating
	OrderStatePending
	OrderStateActive
	OrderStateFilled
	OrderStateCanceled
	OrderStateFailed
)

// String returns the string representation of OrderState
func (s OrderState) String() string {
	switch s {
	case OrderStateUnknown:
		return "Unknown"
	case OrderStateValidating:
		return "Validating"
	case OrderStatePending:
		return "Pending"
	case OrderStateActive:
		return "Active"
	case OrderStateFilled:
		return "Filled"
	case OrderStateCanceled:
		return "Canceled"
	case OrderStateFailed:
		return "Failed"
	default:
		return fmt.Sprintf("OrderState(%d)", int(s))
	}
}

// OrderStateTransition represents a valid state transition
type OrderStateTransition struct {
	From OrderState
	To   OrderState
}

// ValidStateTransitions defines the valid state transitions for orders
var ValidStateTransitions = []OrderStateTransition{
	{OrderStateValidating, OrderStatePending},
	{OrderStateValidating, OrderStateFailed},
	{OrderStatePending, OrderStateActive},
	{OrderStatePending, OrderStateFailed},
	{OrderStateActive, OrderStateFilled},
	{OrderStateActive, OrderStateCanceled},
	{OrderStateActive, OrderStateFailed},
}

// IsValidTransition checks if a state transition is valid
func IsValidTransition(from, to OrderState) bool {
	for _, transition := range ValidStateTransitions {
		if transition.From == from && transition.To == to {
			return true
		}
	}
	return false
}

// AtomicOrder represents an order with atomic state management
type AtomicOrder struct {
	ID            string
	Symbol        string
	Side          string
	Quantity      string
	Price         string
	Leverage      float64
	RiskPercentage float64
	state         int32
	timestamp     time.Time
	error         atomic.Value // stores error
	fills         atomic.Value // stores []Fill
	validator     *validation.OrderValidator
	mu            sync.RWMutex // for non-atomic fields
}

// Fill represents a partial fill of an order
type Fill struct {
	Quantity    string
	Price       string
	Timestamp   time.Time
}

// NewAtomicOrder creates a new atomic order with validation
func NewAtomicOrder(cmd OrderCommand, validator *validation.OrderValidator) *AtomicOrder {
	order := &AtomicOrder{
		ID:            cmd.OrderID,
		Symbol:        cmd.Symbol,
		Side:          cmd.Side,
		Quantity:      cmd.Quantity,
		Price:         cmd.Price,
		Leverage:      cmd.Leverage,
		RiskPercentage: cmd.RiskPercentage,
		state:         int32(OrderStateValidating),
		timestamp:     time.Now(),
		validator:     validator,
	}
	order.fills.Store(make([]Fill, 0))
	return order
}

// GetState returns the current state of the order
func (o *AtomicOrder) GetState() OrderState {
	return OrderState(atomic.LoadInt32(&o.state))
}

// SetState atomically updates the order state with validation
func (o *AtomicOrder) SetState(newState OrderState) bool {
	currentState := o.GetState()
	
	// Validate state transition
	if !IsValidTransition(currentState, newState) {
		o.SetError(fmt.Errorf("invalid state transition from %s to %s", currentState, newState))
		return false
	}

	return atomic.CompareAndSwapInt32(&o.state, int32(currentState), int32(newState))
}

// AddFill atomically adds a fill to the order
func (o *AtomicOrder) AddFill(fill Fill) error {
	if o.GetState() != OrderStateActive {
		return errors.New("cannot add fill: order not active")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	fills := o.GetFills()
	totalFilled := 0.0
	for _, f := range fills {
		qty, _ := strconv.ParseFloat(f.Quantity, 64)
		totalFilled += qty
	}

	newQty, _ := strconv.ParseFloat(fill.Quantity, 64)
	orderQty, _ := strconv.ParseFloat(o.Quantity, 64)

	if totalFilled + newQty > orderQty {
		return errors.New("fill would exceed order quantity")
	}

	newFills := append(fills, fill)
	o.fills.Store(newFills)

	// Check if order is completely filled
	if totalFilled + newQty == orderQty {
		o.SetState(OrderStateFilled)
	}

	return nil
}

// GetFills returns all fills for the order
func (o *AtomicOrder) GetFills() []Fill {
	return o.fills.Load().([]Fill)
}

// SetError atomically sets an error and updates state
func (o *AtomicOrder) SetError(err error) {
	o.error.Store(err)
	o.SetState(OrderStateFailed)
}

// GetError returns the error associated with the order
func (o *AtomicOrder) GetError() error {
	if err := o.error.Load(); err != nil {
		return err.(error)
	}
	return nil
}

// Validate performs comprehensive order validation
func (o *AtomicOrder) Validate(accountBalance float64) error {
	orderParams := &validation.Order{
		Symbol:         o.Symbol,
		Side:           o.Side,
		Quantity:       o.Quantity,
		Price:          o.Price,
		RiskPercentage: o.RiskPercentage,
		Leverage:       o.Leverage,
		AccountBalance: accountBalance,
	}

	if err := o.validator.ValidateOrder(orderParams); err != nil {
		o.SetError(err)
		return err
	}

	// Atomically update state after validation
	if !o.SetState(OrderStatePending) {
		return errors.New("failed to transition to pending state")
	}

	return nil
}

// IsTerminal returns true if the order is in a terminal state
func (o *AtomicOrder) IsTerminal() bool {
	state := o.GetState()
	return state == OrderStateFilled || state == OrderStateCanceled || state == OrderStateFailed
}

// GetFilledQuantity returns the total filled quantity
func (o *AtomicOrder) GetFilledQuantity() float64 {
	fills := o.GetFills()
	var total float64
	for _, fill := range fills {
		qty, _ := strconv.ParseFloat(fill.Quantity, 64)
		total += qty
	}
	return total
}

// GetAverageFilledPrice returns the average filled price
func (o *AtomicOrder) GetAverageFilledPrice() float64 {
	fills := o.GetFills()
	var totalQuantity, totalValue float64
	for _, fill := range fills {
		qty, _ := strconv.ParseFloat(fill.Quantity, 64)
		price, _ := strconv.ParseFloat(fill.Price, 64)
		totalQuantity += qty
		totalValue += qty * price
	}
	if totalQuantity == 0 {
		return 0
	}
	return totalValue / totalQuantity
}

// OrderStateManager manages the state of multiple orders
type OrderStateManager struct {
	orders sync.Map // map[string]*AtomicOrder
}

// NewOrderStateManager creates a new order state manager
func NewOrderStateManager() *OrderStateManager {
	return &OrderStateManager{}
}

// AddOrder adds a new order to the manager
func (m *OrderStateManager) AddOrder(order *AtomicOrder) {
	m.orders.Store(order.ID, order)
}

// GetOrder retrieves an order from the manager
func (m *OrderStateManager) GetOrder(orderID string) (*AtomicOrder, bool) {
	value, exists := m.orders.Load(orderID)
	if !exists {
		return nil, false
	}
	return value.(*AtomicOrder), true
}

// UpdateOrderState attempts to update an order's state
func (m *OrderStateManager) UpdateOrderState(orderID string, newState OrderState) error {
	order, exists := m.GetOrder(orderID)
	if !exists {
		return errors.New("order not found")
	}

	currentState := order.GetState()
	if !IsValidTransition(currentState, newState) {
		return fmt.Errorf("invalid state transition from %s to %s", currentState, newState)
	}

	if !order.SetState(newState) {
		return fmt.Errorf("failed to update order state from %s to %s", currentState, newState)
	}

	return nil
}

// RemoveOrder removes an order from the manager
func (m *OrderStateManager) RemoveOrder(orderID string) {
	m.orders.Delete(orderID)
}

// GetAllOrders returns all orders in the manager
func (m *OrderStateManager) GetAllOrders() []*AtomicOrder {
	var orders []*AtomicOrder
	m.orders.Range(func(key, value interface{}) bool {
		orders = append(orders, value.(*AtomicOrder))
		return true
	})
	return orders
}

// GetOrdersByState returns all orders in a specific state
func (m *OrderStateManager) GetOrdersByState(state OrderState) []*AtomicOrder {
	var orders []*AtomicOrder
	m.orders.Range(func(key, value interface{}) bool {
		order := value.(*AtomicOrder)
		if order.GetState() == state {
			orders = append(orders, order)
		}
		return true
	})
	return orders
}
