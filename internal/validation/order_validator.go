package validation

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrInvalidSymbol     = errors.New("invalid trading symbol")
	ErrInvalidSide       = errors.New("invalid order side")
	ErrInvalidQuantity   = errors.New("invalid order quantity")
	ErrInvalidPrice      = errors.New("invalid order price")
	ErrInvalidLeverage   = errors.New("invalid leverage")
	ErrInvalidRisk       = errors.New("invalid risk percentage")
	ErrInsufficientFunds = errors.New("insufficient funds for order")
)

// OrderValidator provides validation for order parameters
type OrderValidator struct {
	minQuantity float64
	maxQuantity float64
	minPrice    float64
	maxPrice    float64
	maxLeverage float64
	maxRisk     float64
}

// NewOrderValidator creates a new order validator with specified limits
func NewOrderValidator(minQty, maxQty, minPrice, maxPrice, maxLev, maxRisk float64) *OrderValidator {
	return &OrderValidator{
		minQuantity: minQty,
		maxQuantity: maxQty,
		minPrice:    minPrice,
		maxPrice:    maxPrice,
		maxLeverage: maxLev,
		maxRisk:     maxRisk,
	}
}

// ValidateSymbol checks if the trading symbol is valid
func (v *OrderValidator) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return ErrInvalidSymbol
	}
	
	// Check symbol format (e.g., BTC/USDT)
	parts := strings.Split(symbol, "/")
	if len(parts) != 2 {
		return fmt.Errorf("%w: invalid format", ErrInvalidSymbol)
	}
	
	// Validate base and quote currencies
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return fmt.Errorf("%w: empty base or quote currency", ErrInvalidSymbol)
	}
	
	return nil
}

// ValidateSide checks if the order side is valid
func (v *OrderValidator) ValidateSide(side string) error {
	side = strings.ToUpper(side)
	if side != "BUY" && side != "SELL" {
		return fmt.Errorf("%w: must be BUY or SELL", ErrInvalidSide)
	}
	return nil
}

// ValidateQuantity checks if the order quantity is valid
func (v *OrderValidator) ValidateQuantity(quantity string) error {
	qty, err := strconv.ParseFloat(quantity, 64)
	if err != nil {
		return fmt.Errorf("%w: parse error", ErrInvalidQuantity)
	}

	if qty < v.minQuantity || qty > v.maxQuantity {
		return fmt.Errorf("%w: quantity must be between %v and %v", 
			ErrInvalidQuantity, v.minQuantity, v.maxQuantity)
	}

	// Check decimal places
	decimalPlaces := len(strings.Split(quantity, ".")[1])
	if decimalPlaces > 8 {
		return fmt.Errorf("%w: maximum 8 decimal places allowed", ErrInvalidQuantity)
	}

	return nil
}

// ValidatePrice checks if the order price is valid
func (v *OrderValidator) ValidatePrice(price string) error {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return fmt.Errorf("%w: parse error", ErrInvalidPrice)
	}

	if p < v.minPrice || p > v.maxPrice {
		return fmt.Errorf("%w: price must be between %v and %v", 
			ErrInvalidPrice, v.minPrice, v.maxPrice)
	}

	// Check for reasonable price precision
	decimalPlaces := len(strings.Split(price, ".")[1])
	if decimalPlaces > 8 {
		return fmt.Errorf("%w: maximum 8 decimal places allowed", ErrInvalidPrice)
	}

	return nil
}

// ValidateRisk checks if the risk percentage is valid
func (v *OrderValidator) ValidateRisk(riskPercentage float64) error {
	if riskPercentage <= 0 || riskPercentage > v.maxRisk {
		return fmt.Errorf("%w: must be between 0 and %v%%", ErrInvalidRisk, v.maxRisk)
	}
	return nil
}

// ValidateLeverage checks if the leverage is valid
func (v *OrderValidator) ValidateLeverage(leverage float64) error {
	if leverage < 1 || leverage > v.maxLeverage {
		return fmt.Errorf("%w: must be between 1 and %vx", ErrInvalidLeverage, v.maxLeverage)
	}
	return nil
}

// ValidateOrder performs comprehensive order validation
func (v *OrderValidator) ValidateOrder(order *Order) error {
	if err := v.ValidateSymbol(order.Symbol); err != nil {
		return err
	}
	if err := v.ValidateSide(order.Side); err != nil {
		return err
	}
	if err := v.ValidateQuantity(order.Quantity); err != nil {
		return err
	}
	if err := v.ValidatePrice(order.Price); err != nil {
		return err
	}
	if err := v.ValidateRisk(order.RiskPercentage); err != nil {
		return err
	}
	if err := v.ValidateLeverage(order.Leverage); err != nil {
		return err
	}

	// Validate position size against account balance
	qty, _ := strconv.ParseFloat(order.Quantity, 64)
	price, _ := strconv.ParseFloat(order.Price, 64)
	positionSize := qty * price

	if positionSize > order.AccountBalance*order.Leverage {
		return fmt.Errorf("%w: position size exceeds available margin", ErrInsufficientFunds)
	}

	return nil
}

// Order represents the order parameters for validation
type Order struct {
	Symbol         string
	Side           string
	Quantity       string
	Price          string
	RiskPercentage float64
	Leverage       float64
	AccountBalance float64
}

// ValidateStopLoss ensures the stop loss is valid for the position
func (v *OrderValidator) ValidateStopLoss(entryPrice, stopLoss float64, side string) error {
	if stopLoss <= 0 {
		return fmt.Errorf("%w: stop loss must be greater than 0", ErrInvalidPrice)
	}

	// For long positions, stop loss must be below entry price
	if side == "BUY" && stopLoss >= entryPrice {
		return errors.New("stop loss must be below entry price for long positions")
	}

	// For short positions, stop loss must be above entry price
	if side == "SELL" && stopLoss <= entryPrice {
		return errors.New("stop loss must be above entry price for short positions")
	}

	// Calculate stop loss percentage
	slPercentage := math.Abs((entryPrice - stopLoss) / entryPrice * 100)
	if slPercentage > 50 { // Example: max 50% stop loss
		return errors.New("stop loss percentage too large")
	}

	return nil
}
